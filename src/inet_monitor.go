package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/vishvananda/netlink"
	"time"
	"github.com/projectcalico/felix/set"
	"syscall"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type netlinkStub interface {
	Subscribe(
		linkUpdates chan netlink.LinkUpdate,
		addrUpdates chan netlink.AddrUpdate,
	) error
	LinkList() ([]netlink.Link, error)
	AddrList(link netlink.Link, family int) ([]netlink.Addr, error)
}

type State string

const (
	StateUp   = "up"
	StateDown = "down"
)

type InterfaceStateCallback func(ifaceName string, ifaceState State)
type AddrStateCallback func(ifaceName string, addrs set.Set)

type InterfaceMonitor struct {
	netlinkStub  netlinkStub
	resyncC      <-chan time.Time
	upIfaces     set.Set
	Callback     InterfaceStateCallback
	AddrCallback AddrStateCallback
	ifaceName    map[int]string
	ifaceAddrs   map[int]set.Set
}

func New() *InterfaceMonitor {
	// Interface monitor using the real netlink, and resyncing every 10 seconds.
	resyncTicker := time.NewTicker(10 * time.Second)
	return NewWithStubs(&netlinkReal{}, resyncTicker.C)
}

func NewWithStubs(netlinkStub netlinkStub, resyncC <-chan time.Time) *InterfaceMonitor {
	return &InterfaceMonitor{
		netlinkStub: netlinkStub,
		resyncC:     resyncC,
		upIfaces:    set.New(),
		ifaceName:   map[int]string{},
		ifaceAddrs:  map[int]set.Set{},
	}
}

func main() {

}

func (m *InterfaceMonitor) MonitorInterfaces() {
	log.Info("Interface monitoring thread started.")

	updates := make(chan netlink.LinkUpdate)
	addrUpdates := make(chan netlink.AddrUpdate)
	if err := m.netlinkStub.Subscribe(updates, addrUpdates); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to netlink stub")
	}
	log.Info("Subscribed to netlink updates.")

	// Start of day, do a resync to notify all our existing interfaces.  We also do periodic
	// resyncs because it's not clear what the ordering guarantees are for our netlink
	// subscription vs a list operation as used by resync().
	err := m.resync()
	if err != nil {
		log.WithError(err).Fatal("Failed to read link states from netlink.")
	}

readLoop:
	for {
		log.WithFields(log.Fields{
			"updates":     updates,
			"addrUpdates": addrUpdates,
			"resyncC":     m.resyncC,
		}).Debug("About to select on possible triggers")
		select {
		case update, ok := <-updates:
			log.WithField("update", update).Debug("Link update")
			if !ok {
				log.Warn("Failed to read a link update")
				break readLoop
			}
			m.handleNetlinkUpdate(update)
		case addrUpdate, ok := <-addrUpdates:
			log.WithField("addrUpdate", addrUpdate).Debug("Address update")
			if !ok {
				log.Warn("Failed to read an address update")
				break readLoop
			}
			m.handleNetlinkAddrUpdate(addrUpdate)
		case <-m.resyncC:
			log.Debug("Resync trigger")
			err := m.resync()
			if err != nil {
				log.WithError(err).Fatal("Failed to read link states from netlink.")
			}
		}
	}
	log.Fatal("Failed to read events from Netlink.")
}

func (m *InterfaceMonitor) handleNetlinkUpdate(update netlink.LinkUpdate) {
	attrs := update.Link.Attrs()
	if attrs == nil {
		// Defensive, some sort of interface that the netlink lib doesn't understand?
		log.WithField("update", update).Warn("Missing attributes on netlink update.")
		return
	}
	msgType := update.Header.Type
	ifaceExists := msgType == syscall.RTM_NEWLINK // Alternative is an RTM_DELLINK
	m.storeAndNotifyLink(ifaceExists, update.Link)
}

func (m *InterfaceMonitor) storeAndNotifyLink(ifaceExists bool, link netlink.Link) {
	log.WithFields(log.Fields{
		"ifaceExists": ifaceExists,
		"link":        link,
	}).Debug("storeAndNotifyLink called")

	attrs := link.Attrs()
	ifIndex := attrs.Index
	oldName := m.ifaceName[ifIndex]
	newName := attrs.Name
	if oldName != "" && oldName != newName {
		log.WithFields(log.Fields{
			"oldName": oldName,
			"newName": newName,
		}).Info("Interface renamed, simulating deletion of old copy.")
		m.storeAndNotifyLinkInner(false, oldName, link)
	}

	m.storeAndNotifyLinkInner(ifaceExists, newName, link)
}

func (m *InterfaceMonitor) storeAndNotifyLinkInner(ifaceExists bool, ifaceName string, link netlink.Link) {
	log.WithFields(log.Fields{
		"ifaceExists": ifaceExists,
		"ifaceName":   ifaceName,
		"link":        link,
	}).Debug("storeAndNotifyLinkInner called")

	// Store or remove mapping between this interface's index and name.
	attrs := link.Attrs()
	ifIndex := attrs.Index
	if ifaceExists {
		m.ifaceName[ifIndex] = ifaceName
	} else {
		log.Debug("Notify link non-existence to address callback consumers")
		delete(m.ifaceAddrs, ifIndex)
		m.notifyIfaceAddrs(ifIndex)
		delete(m.ifaceName, ifIndex)
	}

	// We need the operstate of the interface; this is carried in the IFF_RUNNING flag.  The
	// IFF_UP flag contains the admin state, which doesn't tell us whether we can program routes
	// etc.
	rawFlags := attrs.RawFlags
	ifaceIsUp := ifaceExists && rawFlags&syscall.IFF_RUNNING != 0
	ifaceWasUp := m.upIfaces.Contains(ifaceName)
	logCxt := log.WithField("ifaceName", ifaceName)
	if ifaceIsUp && !ifaceWasUp {
		logCxt.Debug("Interface now up")
		m.upIfaces.Add(ifaceName)
		m.Callback(ifaceName, StateUp)
	} else if ifaceWasUp && !ifaceIsUp {
		logCxt.Debug("Interface now down")
		m.upIfaces.Discard(ifaceName)
		m.Callback(ifaceName, StateDown)
	} else {
		logCxt.WithField("ifaceIsUp", ifaceIsUp).Debug("Nothing to notify")
	}

	// If the link now exists, get addresses for the link and store and notify those too; then
	// we don't have to worry about a possible race between the link and address update
	// channels.  We deliberately do this regardless of the link state, as in some cases this
	// will allow us to secure a Host Endpoint interface _before_ it comes up, and so eliminate
	// a small window of insecurity.
	if ifaceExists {
		newAddrs := set.New()
		for _, family := range [2]int{netlink.FAMILY_V4, netlink.FAMILY_V6} {
			addrs, err := m.netlinkStub.AddrList(link, family)
			if err != nil {
				log.WithError(err).Warn("Netlink addr list operation failed.")
			}
			for _, addr := range addrs {
				newAddrs.Add(addr.IPNet.IP.String())
			}
		}
		if (m.ifaceAddrs[ifIndex] == nil) || !m.ifaceAddrs[ifIndex].Equals(newAddrs) {
			m.ifaceAddrs[ifIndex] = newAddrs
			m.notifyIfaceAddrs(ifIndex)
		}
	}
}


/*
func notify(ch <-chan netlink.LinkUpdate) {
	for {
		select {
		case update := <-ch:
			fmt.Println(update.Link.Attrs().Name, update.Link.Attrs().HardwareAddr, update.IfInfomsg.Flags, syscall.IFF_UP)
		case <-time.After(1 * time.Minute):
			fmt.Println("timeout")
		}
	}
}

func MonitorLinks() {
	ch := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	defer close(done)
	if err := netlink.LinkSubscribe(ch, done); err != nil {
		log.Fatal(err)
	}
	notify(ch)
}
*/
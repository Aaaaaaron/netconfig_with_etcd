package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/vishvananda/netlink"
	"time"
	"github.com/projectcalico/felix/set"
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

func NewInetMonitor() *InterfaceMonitor {
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
	GetLinkDetails()

}

func (m *InterfaceMonitor) MonitorInterfaces() {

	log.Info("Interface monitoring thread started.")

	updates := make(chan netlink.LinkUpdate)
	addrUpdates := make(chan netlink.AddrUpdate)
	if err := m.netlinkStub.Subscribe(updates, addrUpdates); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to netlink stub")
	}
	log.Info("Subscribed to netlink updates.")

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
	name := attrs.Name
	if attrs == nil {
		// Defensive, some sort of interface that the netlink lib doesn't understand?
		log.WithField("update", update).Warn("Missing attributes on netlink update.")
		return
	}

	ifId := GetHostId() + "_" + GetEthBusInfo(name)
	link := getLinkById(ifId)
	//m.updateEtcd()
	m.updateMap(link)
}
func getLinkById(ifId string) (LinkAttrs) {
	result, ok := LinkMap.Get(ifId);
	if !ok {
		log.Warn("can not retrieve value from key:", ifId)
		return nil
	}
	return result.(LinkAttrs)
}

func (m *InterfaceMonitor) updateMap( link netlink.Link) {
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

package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/vishvananda/netlink"
	"time"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type UpdateChan struct {
	LinkUpdateChan chan LinkUpdate
	//AddrUpdateChan <-chan LinkUpdate
	//RouteUpdateChan <-chan LinkUpdate
}

type LinkUpdate struct {
	Action   string
	LinkId   string
	DevName  string
	Command  string
	Argument string
	link     netlink.Link
}

func main() {
	GetLinkDetails()
	link, _ := LinkMap.Get(GetLinkId("eth0"))
	netlink := link.(LinkWrapper)
	log.WithField("link",netlink.link).Debug("get link")
	linkUpdate := LinkUpdate{"update", "1", "eth0", "set", "down", netlink.link}
	linkUpdateChan := make(chan LinkUpdate)
	updateChan := UpdateChan{linkUpdateChan}
	go UpdateKernel(updateChan, time.NewTicker(10 * time.Second).C)
	updateChan.LinkUpdateChan <- linkUpdate
	time.Sleep(100000 * time.Millisecond)
}

func UpdateKernel(updateChan UpdateChan, resyncC <-chan time.Time) {
	log.Info("Interface monitoring thread started.")

	for {
		select {
		case linkUpdate := <-updateChan.LinkUpdateChan:
			log.WithField("update", linkUpdate).Debug("Link update")
			handldLinkUpdate(linkUpdate)
		case <-resyncC:
			log.Debug("Resync trigger")
		//err := resync()
		//if err != nil {
		//	log.WithError(err).Fatal("Failed to read link states from netlink.")
		//}
		}
	}
	log.Fatal("Failed to read events from Netlink.")
}

func handldLinkUpdate(update LinkUpdate) {
	log.Debug("into handldLinkUpdate")
	link := update.link
	switch update.Action {
	case "update":
		log.Debug("into update")
		if update.Command == "set" {
			log.Debug("into set")
			if update.Command == "up" {
				netlink.LinkSetUp(link)
			}
			if update.Command == "down" {
				log.Debug("into down")
				netlink.LinkSetDown(update.link)
			}
		}
	//case:
	//	return
	//case:
	//	return
	}
}

func getLinkById(ifId string) (LinkWrapper) {
	result, ok := LinkMap.Get(ifId);
	if !ok {
		log.Fatal("can not retrieve value from key:", ifId) //todo how to do is better ?
	}
	return result.(LinkWrapper)
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

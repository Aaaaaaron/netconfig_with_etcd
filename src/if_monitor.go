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

type Update struct {
	LinkUpdateChan chan LinkUpdate
	//AddrUpdateChan <-chan LinkUpdate
	//RouteUpdateChan <-chan LinkUpdate
}

type LinkUpdate struct {
	Action   string
	LinkId   string
	Object   string
	Command  string
	Argument string
	link     netlink.Link
}

func main() {
	link, _ := GetLinkByName("eth0")
	updateChan := Update{make(chan LinkUpdate)}
	go UpdateKernel(updateChan, time.NewTicker(10 * time.Second).C)

	linkUpdate := LinkUpdate{"update", "1", "eth0", "set", "down", link}
	updateChan.LinkUpdateChan <- linkUpdate
	time.Sleep(100000 * time.Millisecond)
}

func UpdateKernel(update Update, resyncC <-chan time.Time) {
	log.Info("Interface monitoring thread started.")

	for {
		select {
		case linkUpdate := <-update.LinkUpdateChan:
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
	link := update.link
	switch update.Action {
	case "update":
		if update.Command == "set" {
			if update.Argument == "up" {
				netlink.LinkSetUp(link)
			}
			if update.Argument == "down" {
				netlink.LinkSetDown(update.link)
			}
		}
	//case:
	//	return
	//case:
	//	return
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

package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/vishvananda/netlink"
	"time"
	"errors"
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
			if err := handldLinkUpdate(linkUpdate); err != nil {
				//handle fail.retry or alert
			}
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

func handldLinkUpdate(update LinkUpdate) error {
	link := update.link
	updateError := errors.New("update fail, " + update.Command + update.Argument + update.link.Attrs().Name)
	switch update.Action {
	case "update":
		if update.Command == "set" {
			if update.Argument == "up" {
				if err := netlink.LinkSetUp(link); err != nil {
					log.Error("update fail", err)
					return updateError
				}
			}
			if update.Argument == "down" {
				if err := netlink.LinkSetDown(link); err != nil {
					log.Error("update fail", err)
					return updateError
				}
			}
		}
	//case:
	//	return
	//case:
	//	return
	}
	return errors.New("no action found")
}

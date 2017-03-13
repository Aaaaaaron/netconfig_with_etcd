package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/vishvananda/netlink"
	"time"
	"errors"
	"fmt"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type Update interface {
	handleUpdate() error
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
	/*	link, _ := GetLinkByName("eth0")
		updateChan := Update{make(chan LinkUpdate)}
		go UpdateKernel(updateChan, time.NewTicker(10 * time.Second).C)

		linkUpdate := LinkUpdate{"update", "1", "eth0", "set", "down", link}
		updateChan.LinkUpdateChan <- linkUpdate
		time.Sleep(100000 * time.Millisecond)*/
	link, _ := GetLinkByName("eth0")
	LinkUpdate{"update", "1", "eth0", "set", "up", link}.handleUpdate()
	fmt.Println(link.Attrs().Flags, "	", link.Attrs().RawFlags)
	LinkUpdate{"update", "1", "eth0", "set", "down", link}.handleUpdate()
	fmt.Println(link.Attrs().Flags, "	", link.Attrs().RawFlags)

}

func UpdateKernel(updateChan chan Update, resyncC <-chan time.Time) {
	log.Info("Interface monitoring thread started.")

	for {
		select {
		case update := <-updateChan:
			//log.WithField("update", linkUpdate).Debug("Link update")
			if err := update.handleUpdate(); err != nil {
				//handle fail.retry or alert
			}
		// update linux success,update map and etcd

		//periodic resyncs
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

func UpdateMap(id string, updatedLink netlink.Link) {
	//todo
}

//func UpdateEtcd(id string,updatedLink netlink.Link){
//
//}

func (update *LinkUpdate) handleUpdate() error {
	link := update.link
	updateError := errors.New("update fail, " + update.Command + update.Argument + link.Attrs().Name)
	switch update.Action {
	case "update":
		if update.Command == "set" {
			if update.Argument == "up" {
				if err := netlink.LinkSetUp(link); err != nil { // link will not be update,you should retrieve the link by your self.should i do this here or return the updated link?
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
		link, _ = GetLinkByName(link.Attrs().Name)
		log.WithFields(log.Fields{
			"link name":   link.Attrs().Name,
			"link status": string(link.Attrs().Flags),
		}).Debug("set ", update.Argument, " linux success")
	case "del":
	//return
	case "add":
	//return
	}
	return errors.New("no action found")
}

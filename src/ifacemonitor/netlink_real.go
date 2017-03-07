package ifacemonitor

import (
	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	//"syscall"
)

type netlinkReal struct {
}

func (nl *netlinkReal) Subscribe(
	linkUpdates chan netlink.LinkUpdate,
	addrUpdates chan netlink.AddrUpdate,
) error {
	cancel := make(chan struct{})

	if err := netlink.LinkSubscribe(linkUpdates, cancel); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to link updates")
		return err
	}
	if err := netlink.AddrSubscribe(addrUpdates, cancel); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to addr updates")
		return err
	}

	return nil
}

func (nl *netlinkReal) LinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}

func (nl *netlinkReal) AddrList(link netlink.Link, family int) ([]netlink.Addr, error) {
	return netlink.AddrList(link, family)
}

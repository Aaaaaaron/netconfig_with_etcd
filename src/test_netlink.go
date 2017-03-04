package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"syscall"
	"time"
)

func notify(ch <-chan netlink.LinkUpdate) {
	for {
		timeout := time.After(time.Minute)
		select {
		case update := <-ch:
			fmt.Println(update.Link.Attrs().Name, update.Link.Attrs().HardwareAddr, update.IfInfomsg.Flags, syscall.IFF_UP)
		case <-timeout:
			fmt.Println("timeout")
		}
	}
}

func MonitorLink() {
	ch := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	defer close(done)
	if err := netlink.LinkSubscribe(ch, done); err != nil {
		log.Fatal(err)
	}
	notify(ch)
}
func main() {
	MonitorLink()
}

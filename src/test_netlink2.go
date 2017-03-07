package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"github.com/safchain/ethtool"
	"net"
	//"encoding/json"
)

type LinkAttrs struct {
	Id           string // equals HostId + BusInfo
	HostId       string
	BusInfo      string
	Name         string
	DisplayName  string
	HardwareAddr net.HardwareAddr
	MTU          int
	TxQLen       int
	AdminStat    netlink.LinkOperState
	OperStat     netlink.LinkOperState
	ParentId     int
	BypassId     string
}

func GetLinkDetails() {
	//var links []*string
	linkList, err := LinkList()
	if err != nil {
		log.Fatal(err)
	}

	for _, link := range linkList {
		fmt.Print(link)
		la := NewLinkAttrs(link)
		fmt.Print(la.Name)
		//data, err := json.Marshal(la)
		//if err != nil {
		//	log.Fatalf("JSON marshaling failed: %s", err)
		//}
		//fmt.Printf("%s\n", data)

		//links = append(links, la)
		//fmt.Println(linkAttrs.Name, linkAttrs.HardwareAddr, linkAttrs.MTU, linkAttrs.TxQLen, linkAttrs.OperState, linkAttrs.ParentIndex)
	}
	//return
}

func LinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}

func NewLinkAttrs(link netlink.Link) (*LinkAttrs) {
	linkAttrs := link.Attrs()
	name := linkAttrs.Name
	la := new(LinkAttrs)
	la.HostId = getHostId()
	la.BusInfo = getEthBusInfo(name)
	la.Id = la.HostId + la.BusInfo
	la.Name = linkAttrs.Name
	la.DisplayName = la.Name //need to retrieve from etcd if etcd has, or equals name
	la.HardwareAddr = linkAttrs.HardwareAddr
	la.MTU = linkAttrs.MTU
	la.TxQLen = linkAttrs.TxQLen
	la.AdminStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals operState
	la.OperStat = linkAttrs.OperState
	la.ParentId = linkAttrs.ParentIndex
	//la.BypassId=
	return la
}

func getEthBusInfo(ethName string) string {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatal(err)
	}

	busInfo, err := ethHandle.BusInfo(ethName)
	if err != nil {
		busInfo=""
	}

	return busInfo
}

func getHostId() string {
	return "1"
}

func main() {
	GetLinkDetails()
}

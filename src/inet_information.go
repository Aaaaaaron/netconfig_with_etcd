package main

import (
	"github.com/vishvananda/netlink"
	"github.com/safchain/ethtool"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	//"math/rand"
	//"strconv"
	"os"
	"github.com/orcaman/concurrent-map"
	"fmt"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

//warpper netlink's Link,add some fields
type LinkWrapper struct {
	link        netlink.Link
	Id          string // equals HostId +"_" +BusInfo
	HostId      string
	BusInfo     string
	DisplayName string // not equals alias
	SysStat     netlink.LinkOperState
	adminStat   netlink.LinkOperState //uint8
	execStat    netlink.LinkOperState
	BypassId    string
}

func main() {
	GetLinkDetails()
	fmt.Print(LinkMap)
}

func GetLinkDetails() cmap.ConcurrentMap {
	linkList := getLinkList()

	for _, link := range linkList {
		la := NewLink(link)
		data, err := json.MarshalIndent(la, "", "\t")
		if err != nil {
			log.Fatalf("JSON marshaling failed: %s", err)
		}

		LinkMap.Set(la.Id, data)

		log.WithFields(log.Fields{
			"kye":        la.Id,
			"link value": la,
		}).Debug("插入etcd的link的value值信息")
	}
	return LinkMap
}

func getLinkList() ([]netlink.Link) {
	linkList, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("get link list from netlink failed: %s", err)
	}
	return linkList
}

func NewLink(link netlink.Link) (*LinkWrapper) {
	name := link.Attrs().Name
	lw := new(LinkWrapper)

	lw.link = link
	lw.Id = GetHostId() + "_" + GetEthBusInfo(name)
	lw.HostId = GetHostId()
	lw.BusInfo = GetEthBusInfo(name)
	lw.DisplayName = name //need to retrieve from etcd if etcd has, or equals name
	//lw.SysStat =
	//lw.AdminStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals SysStat
	//lw.ExecStat=
	lw.BypassId = ""
	return lw
}

func GetHostId() string {
	//return strconv.Itoa(rand.Int())
	return "1"
}

func GetEthBusInfo(ethName string) string {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatal("can not get ethtoll",err)
	}

	busInfo, err := ethHandle.BusInfo(ethName)
	if err != nil {
		busInfo = ""
	}

	return busInfo
}

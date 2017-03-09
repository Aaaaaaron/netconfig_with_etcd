package main

import (
	"github.com/vishvananda/netlink"
	"github.com/safchain/ethtool"
	log "github.com/Sirupsen/logrus"
	//"math/rand"
	//"strconv"
	"os"
	"github.com/orcaman/concurrent-map"
	"fmt"
	"net"
	"encoding/json"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

//warpper netlink's Link,and attrs add some fields
type LinkWrapper struct {
	link  *netlink.Link
	attrs *LinkAttrs
}

type LinkAttrs struct {
	Id           string // equals HostId +"_" +BusInfo
	HostId       string
	BusInfo      string
	Name         string
	DisplayName  string // not equals alias
	HardwareAddr net.HardwareAddr
	MTU          int
	TxQLen       int
	Statistics   *netlink.LinkStatistics
	AdminStat    netlink.LinkOperState //uint8
	ExecStat     netlink.LinkOperState
	SysStat      netlink.LinkOperState
	ParentIndex  int
	MasterIndex  int
	BypassId     string
}

func main() {
	GetLinkDetails()
	//fmt.Print(LinkMap)
}

func GetLinkDetails() cmap.ConcurrentMap {
	linkList := getLinkList()

	for _, link := range linkList {
		linkWrapper := NewLink(link)
		data, err := json.MarshalIndent(linkWrapper.attrs, "", "\t")
		if err != nil {
			log.Fatalf("JSON marshaling failed: %s", err)
		}

		fmt.Println(data)
		//LinkMap.Set(linkWrapper.Id, data)
		LinkMap.Set(linkWrapper.attrs.Id, linkWrapper.attrs)

		//log.WithFields(log.Fields{
		//	"kye":        linkWrapper.attrs.Id,
		//	"link value": linkWrapper,
		//	"josn value": data,
		//}).Debug("插入etcd的link的value值信息")
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

	lw.link = &link
	lw.attrs.Id = GetHostId() + "_" + GetEthBusInfo(name)
	lw.attrs.HostId = GetHostId()
	lw.attrs.BusInfo = GetEthBusInfo(name)
	lw.attrs.Name = name
	lw.attrs.DisplayName = name //need to retrieve from etcd if etcd has, or equals name
	lw.attrs.HardwareAddr = link.Attrs().HardwareAddr
	lw.attrs.MTU = link.Attrs().MTU
	lw.attrs.TxQLen = link.Attrs().TxQLen
	lw.attrs.Statistics = link.Attrs().Statistics
	lw.attrs.ParentIndex = link.Attrs().ParentIndex
	lw.attrs.MasterIndex = link.Attrs().MasterIndex
	//lw.attrs.SysStat =
	//lw.attrs.AdminStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals SysStat
	//lw.attrs.ExecStat=
	lw.attrs.BypassId = ""
	return lw
}

func GetHostId() string {
	//return strconv.Itoa(rand.Int())
	return "1"
}

func GetEthBusInfo(ethName string) string {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatal("can not get ethtoll", err)
	}

	busInfo, err := ethHandle.BusInfo(ethName)
	if err != nil {
		busInfo = ""
	}

	return busInfo
}

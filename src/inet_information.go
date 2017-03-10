package main

import (
	"github.com/vishvananda/netlink"
	"github.com/safchain/ethtool"
	log "github.com/Sirupsen/logrus"
	//"math/rand"
	//"strconv"
	"os"
	"github.com/orcaman/concurrent-map"
	"net"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

//warpper netlink's Link,and Attrs add some fields
type LinkWrapper struct {
	link  netlink.Link
	Attrs LinkAttrs
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

//func main() {
//	GetLinkDetails()
//	//fmt.Print(LinkMap)
//}

func GetLinkDetails() cmap.ConcurrentMap {
	linkList := getLinkList()

	for _, link := range linkList {
		linkWrapper := NewLink(link)
		//data, err := json.MarshalIndent(linkWrapper.Attrs, "", "\t")
		//if err != nil {
		//	log.Fatalf("JSON marshaling failed: %s", err)
		//}
		//fmt.Println(string(data))
		LinkMap.Set(linkWrapper.Attrs.Id, linkWrapper)
		log.WithFields(log.Fields{
			"kye":        linkWrapper.Attrs.Id,
			"link value": linkWrapper.Attrs,
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
	lw.Attrs.Id = GetLinkId(name)
	lw.Attrs.HostId = GetHostId()
	lw.Attrs.BusInfo = GetEthBusInfo(name)
	lw.Attrs.Name = name
	lw.Attrs.DisplayName = name //need to retrieve from etcd if etcd has, or equals name
	lw.Attrs.HardwareAddr = link.Attrs().HardwareAddr
	lw.Attrs.MTU = link.Attrs().MTU
	lw.Attrs.TxQLen = link.Attrs().TxQLen
	lw.Attrs.Statistics = link.Attrs().Statistics
	lw.Attrs.ParentIndex = link.Attrs().ParentIndex
	lw.Attrs.MasterIndex = link.Attrs().MasterIndex
	//lw.Attrs.SysStat =
	//lw.Attrs.AdminStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals SysStat
	//lw.Attrs.ExecStat=
	lw.Attrs.BypassId = ""
	return lw
}

func GetHostId() string {
	//return strconv.Itoa(rand.Int())
	return "1"
}

func GetLinkId(name string) string {
	ifId := GetHostId() + "_" + GetEthBusInfo(name)
	return ifId
}

func GetEthBusInfo(ethName string) string {
	if ethName == "lo" {
		return "lo"
	}
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

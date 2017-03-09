package main

import (
	"github.com/vishvananda/netlink"
	"github.com/safchain/ethtool"
	"net"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	//"math/rand"
	//"strconv"
	"os"
	"github.com/orcaman/concurrent-map"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type LinkAttrs struct {
	Id           string // equals HostId +"_" +BusInfo
	HostId       string
	BusInfo      string
	Name         string
	DisplayName  string
	HardwareAddr net.HardwareAddr
	MTU          int
	TxQLen       int
	Statistics   *netlink.LinkStatistics
	SysStat      netlink.LinkOperState
	OperStat     netlink.LinkOperState //flags+OperState
	ParentId     int
	MasterId     int
	BypassId     string
}

func main() {
	GetLinkDetailsInJSON()
}

func GetLinkDetailsInJSON() []string {
	var links []string
	linkList := getLinkList()

	for _, link := range linkList {
		la := NewLinkAttrs(link)
		data, err := json.MarshalIndent(la, "", "\t")
		if err != nil {
			log.Fatalf("JSON marshaling failed: %s", err)
		}

		log.WithFields(log.Fields{
			"link JSON数据": la,
		}).Debug("插入etcd的link的value值信息")

		links = append(links, string(data))
	}
	return links
}

func getLinkList() ([]netlink.Link) {
	linkList, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("get link list from netlink failed: %s", err)
	}
	return linkList
}

func NewLinkAttrs(link netlink.Link) (*LinkAttrs) {
	linkAttrs := link.Attrs()
	name := linkAttrs.Name
	la := new(LinkAttrs)

	la.Id = la.HostId + "_" + la.BusInfo
	la.HostId = GetHostId()
	la.BusInfo = GetEthBusInfo(name)
	la.Name = name
	la.DisplayName = name //need to retrieve from etcd if etcd has, or equals name
	la.HardwareAddr = linkAttrs.HardwareAddr
	la.MTU = linkAttrs.MTU
	la.TxQLen = linkAttrs.TxQLen
	la.SysStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals operState
	la.OperStat = linkAttrs.OperState
	la.ParentId = linkAttrs.ParentIndex
	la.BypassId = ""
	return la
}

func GetHostId() string {
	//return strconv.Itoa(rand.Int())
	return "1"
}

func GetEthBusInfo(ethName string) string {
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatal(err)
	}

	busInfo, err := ethHandle.BusInfo(ethName)
	if err != nil {
		busInfo = ""
	}

	return busInfo
}

package main

import (
	"errors"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/orcaman/concurrent-map"
	"github.com/safchain/ethtool"
	"github.com/vishvananda/netlink"
	"syscall"
	"encoding/json"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

//warpper netlink's Link,and Attrs add some fields
//type LinkWrapper struct {
//	link  netlink.Link
//	Attrs LinkAttrs
//}

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
	AdminStat    string //uint8
	ExecStat     string
	SysStat      string
	ParentIndex  int
	MasterIndex  int
	BypassId     string
}

func main() {
	GetLinkDetails()
}

func GetLinkDetails() cmap.ConcurrentMap {
	linkList := GetLinkList()

	for _, link := range linkList {
		la := NewLink(link)
		data, err := json.MarshalIndent(la, "", "\t")
		if err != nil {
			log.Fatalf("JSON marshaling failed: %s", err)
		}
		//fmt.Println(string(data))
		log.WithFields(log.Fields{
			"kye":        la.Id,
			"link value": la,
		}).Debug("插入etcd的link的value值信息")
		EtcdPut(la.Id, string(data))
		LinkMap.Set(la.Id, link)
	}
	return LinkMap
}

func GetLinkList() ([]netlink.Link) {
	linkList, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("get link list from netlink failed: %s", err)
	}
	return linkList
}

func NewLink(link netlink.Link) (LinkAttrs) {
	name := link.Attrs().Name
	var la LinkAttrs
	la.Id = GetLinkId(name)
	la.HostId = GetHostId()
	la.BusInfo = GetEthBusInfo(name)
	la.Name = name
	la.DisplayName = getDisplayName(name) //need to retrieve from etcd if etcd has, or equals name
	la.HardwareAddr = link.Attrs().HardwareAddr
	la.MTU = link.Attrs().MTU
	la.TxQLen = link.Attrs().TxQLen
	la.Statistics = link.Attrs().Statistics
	la.ParentIndex = link.Attrs().ParentIndex
	la.MasterIndex = link.Attrs().MasterIndex
	la.SysStat = getSysStat(link)
	la.AdminStat = getSysStat(link) //need to retrieve from etcd if etcd has, or equals SysStat
	//la.AdminStat = getAdminStat(link) //need to retrieve from etcd if etcd has, or equals SysStat
	//la.ExecStat=
	la.BypassId = ""
	return la
}

func getSysStat(link netlink.Link) string {
	//现阶段只取一个interface是否开启
	if (link.Attrs().RawFlags & syscall.IFF_UP) != 0 {
		return "up"
	}
	return "down"
}

func getAdminStat(link netlink.Link) string {
	//kvs := EtcdGet(name)
	return ""
}

func getDisplayName(name string) string {
	kvs := EtcdGet(name)
	for _, ev := range kvs {
		return string(ev.Value)
	}
	return name
}

func GetLinkId(name string) string {
	ifId := GetHostId() + "_" + GetEthBusInfo(name)
	return ifId
}

//限制 ethName只能从GetLinkList()中的link中取,不能自己指定
func GetHostId() string {
	//return strconv.Itoa(rand.Int())
	return "1"
}

func GetEthBusInfo(ethName string) string {
	if ethName == "lo" {
		return "lo"
	}
	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatal("can not get ethtool", err)
	}

	busInfo, err := ethHandle.BusInfo(ethName)
	if err != nil {
		busInfo = ""
	}

	return busInfo
}

func GetLinkByName(name string) (netlink.Link, error) {
	links := GetLinkList()
	for _, link := range links {
		if link.Attrs().Name == name {
			return link, nil
		}
	}
	return nil, errors.New("can not find link named:" + name)
}

func GetLinkById(ifId string) (LinkAttrs) {
	result, ok := LinkMap.Get(ifId);
	if !ok {
		log.Fatal("can not retrieve value from key:", ifId) //todo how to do is better ?
	}
	return result.(LinkAttrs)
}

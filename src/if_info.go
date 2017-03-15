package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/orcaman/concurrent-map"
	"github.com/safchain/ethtool"
	"github.com/vishvananda/netlink"
)

var LinkMap = cmap.New()

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
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
	linkList := GetLinkList()
	m := make(map[string]LinkAttrs)
	for _, link := range linkList {
		id, value := GetLinkDetails(link)
		m[id] = value
	}
	data, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	putEtcd("linklist", string(data))

	fmt.Print(EtcdGet("linklist"))
}

func putEtcd(id, value string) {
	EtcdPut(id, value)
}

func putMap(id string, value LinkAttrs) {
	LinkMap.Set(id, value)
}

func GetLinkDetails(link netlink.Link) (string, LinkAttrs) {
	la := NewLink(link)
	log.WithFields(log.Fields{
		"kye":        la.Id,
		"link value": la,
	}).Debug("插入etcd的link的value值信息")
	return la.Id, la
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
	la.Id = GetLinkId(link)
	la.HostId = GetHostId()
	la.BusInfo = GetEthBusInfo(name)
	la.Name = name
	la.DisplayName = getDisplayName(link) //need to retrieve from etcd if etcd has, or equals name
	la.HardwareAddr = link.Attrs().HardwareAddr
	la.MTU = link.Attrs().MTU
	la.TxQLen = link.Attrs().TxQLen
	la.Statistics = link.Attrs().Statistics
	la.ParentIndex = link.Attrs().ParentIndex
	la.MasterIndex = link.Attrs().MasterIndex
	la.SysStat = getSysStat(link)
	la.AdminStat = getAdminStat(link) //need to retrieve from etcd if etcd has, or equals SysStat
	la.ExecStat = getExecStat(link)
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

func getExecStat(link netlink.Link) string {
	return ""
}

func getDisplayName(link netlink.Link) string {
	id := GetLinkId(link)
	la := EtcdGet(id)
	if la.DisplayName != "" {
		return la.DisplayName
	}

	return link.Attrs().Name
}

func GetLinkId(link netlink.Link) string {
	name := link.Attrs().Name
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

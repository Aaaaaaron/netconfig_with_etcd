package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

type Links struct {
	Stat string
	Link map[string]LinkWrapper
}

type LinkWrapper struct {
	Id      string // equals HostId +"_" +BusInfo
	HostId  string
	BusInfo string
	//DisplayName string //equals alias
	BypassId string
	link     netlink.Link
}

func main() {
	linkList := GetLinkList()
	linkMap := make(map[string]LinkWrapper)
	for _, link := range linkList {
		id, value := GetLinkDetails(link)
		linkMap[id] = value
	}

	data, err := json.MarshalIndent(Links{"done", linkMap}, "", "\t")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	putEtcd("linklist", string(data))

	fmt.Print(EtcdGet("linklist"))
}

func putEtcd(id, value string) {
	EtcdPut(id, value)
}

func putMap(id string, value LinkWrapper) {
	LinkMap.Set(id, value)
}

func GetLinkDetails(link netlink.Link) (string, LinkWrapper) {
	la := NewLinkWrapper(link)
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

func NewLinkWrapper(link netlink.Link) (LinkWrapper) {
	var la LinkWrapper
	la.Id = GetLinkId(link.Attrs().Name)
	la.HostId = GetHostId()
	la.BusInfo = GetEthBusInfo(link.Attrs().Name)
	la.BypassId = ""
	la.link = link
	return la
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
		//busInfo = ""
		log.Fatal("can not retrieve " + ethName + "'s bus info")
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

func GetLinkById(ifId string) (LinkWrapper) {
	result, ok := LinkMap.Get(ifId);
	if !ok {
		log.Fatal("can not retrieve value from key:", ifId) //todo how to do is better ?
	}
	return result.(LinkWrapper)
}

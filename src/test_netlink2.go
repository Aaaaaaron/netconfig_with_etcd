package main

import (
	"github.com/vishvananda/netlink"
	"github.com/safchain/ethtool"
	"net"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"math/rand"
	"strconv"
	"os"
	"fmt"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.TextFormatter{})
	log.SetFormatter(new(MyJSONFormatter))

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

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

type MyJSONFormatter struct {
}

func (f *MyJSONFormatter) Format(entry *log.Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	serialized, err := json.MarshalIndent(entry.Data,"", "    ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func GetLinkDetailsInJSON() []string {
	var links []string
	linkList, err := LinkList()
	if err != nil {
		log.Fatal(err)
	}

	for _, link := range linkList {
		la := NewLinkAttrs(link)
		data, err := json.MarshalIndent(la, "", "    ")

		//data, err := json.Marshal(la)
		if err != nil {
			log.Fatalf("JSON marshaling failed: %s", err)
		}

		log.WithFields(log.Fields{
			"link JSON数据": string(data),
		}).Debug("插入etcd的link的value值")

		links = append(links, string(data))
	}
	return links
}

func LinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}

func NewLinkAttrs(link netlink.Link) (*LinkAttrs) {
	linkAttrs := link.Attrs()
	name := linkAttrs.Name
	la := new(LinkAttrs)

	la.Id = la.HostId + la.BusInfo
	la.HostId = getHostId()
	la.BusInfo = getEthBusInfo(name)
	la.Name = linkAttrs.Name
	la.DisplayName = la.Name //need to retrieve from etcd if etcd has, or equals name
	la.HardwareAddr = linkAttrs.HardwareAddr
	la.MTU = linkAttrs.MTU
	la.TxQLen = linkAttrs.TxQLen
	la.AdminStat = linkAttrs.OperState //need to retrieve from etcd if etcd has, or equals operState
	la.OperStat = linkAttrs.OperState
	la.ParentId = linkAttrs.ParentIndex
	la.BypassId = ""
	return la
}

func getHostId() string {
	return strconv.Itoa(rand.Int())
}

func getEthBusInfo(ethName string) string {
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

func main() {
	GetLinkDetailsInJSON()
}

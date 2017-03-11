package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestGetLinkByName(t *testing.T) {
	links := GetLinkList()
	link1, _ := GetLinkByName(links[1].Attrs().Name)
	assert.Equal(t, links[1], link1, "they should be equal")
	_, err := GetLinkByName("eth")
	assert.NotNil(t, err, "error should not be nil")
}

func TestGetBusInfo(t *testing.T) {
	assert.Equal(t, "lo", GetEthBusInfo("lo"))
	assert.Equal(t, "", GetEthBusInfo("loo"))
	assert.Equal(t, "0000:00:03.0", GetEthBusInfo("eth0"))
}

func TestHandldLinkUpdate(t *testing.T) {
	link, _ := GetLinkByName("eth0")
	linkUpdate := LinkUpdate{"update", "1", "eth0", "set", "up", link}
	handldLinkUpdate(linkUpdate)
	updatedLink, _ := GetLinkByName("eth0")
	//assert.Equal(t,updatedLink.Attrs().)
	fmt.Print(updatedLink.Attrs().Flags, updatedLink.Attrs().RawFlags)
	linkUpdate2 := LinkUpdate{"update", "1", "eth0", "set", "down", link}
	handldLinkUpdate(linkUpdate2)
	fmt.Print(updatedLink.Attrs().Flags, updatedLink.Attrs().RawFlags)
}

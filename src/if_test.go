package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	links := GetLinkList()
	link1, _ := GetLinkByName(links[1].Attrs().Name)
	assert.Equal(t, links[1], link1, "they should be equal")
	_,err := GetLinkByName("eth")
	assert.NotNil(t,err, "error should not be nil")
}

func TestGetBusInfo(t *testing.T) {
	assert.Equal(t,"lo",GetEthBusInfo("lo"))
	assert.Equal(t,"",GetEthBusInfo("loo"))
	assert.Equal(t,"0000:00:03.0",GetEthBusInfo("eth0"))
}


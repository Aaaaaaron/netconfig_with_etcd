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
	/*
	assert.Equal(t, 123, 123, "they should be equal")
	assert.NotEqual(t, 123, 456, "they should not be equal")
	assert.Nil(t, object)
	if assert.NotNil(t, object) {
		assert.Equal(t, "Something", object.Value)
	}
	*/
}

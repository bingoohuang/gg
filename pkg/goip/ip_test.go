package goip_test

import (
	"log"
	"testing"

	"github.com/bingoohuang/gg/pkg/goip"
	"github.com/stretchr/testify/assert"
)

func TestListAllIPv4(t *testing.T) {
	ips, err := goip.ListAllIPv4()

	assert.Nil(t, err)
	log.Printf("ListAllIPv4 %+v", ips)
}

func TestListAllIPv6(t *testing.T) {
	ips, err := goip.ListAllIPv6()

	assert.Nil(t, err)
	log.Printf("ListAllIPv6 %+v", ips)
}

func TestGetOutboundIP(t *testing.T) {
	log.Printf("Outbound:%s", goip.Outbound())
	mainIP, ipList := goip.MainIP()
	log.Printf("MainIP:%s", mainIP)
	log.Printf("ipList:%+v", ipList)
}

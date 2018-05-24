package client

import (
	"fmt"
	"testing"
)

func TestParseConf(t *testing.T) {
	ParseDSN("wby:p:s@w@localhost:2379,localhost:2479/call/s")
	ParseDSN("localhost:2379/call/s")
	ParseDSN("localhost:2379,localhost:2479")
	ParseDSN("wby:p:s@w@localhost:2379/")
	ParseDSN("wby:p:s@w@localhost:2379")
}

func TestClient_Get(t *testing.T) {
	cli, _ := NewClient("localhost:2379")
	val, err := cli.Get("/test")
	t.Log(err)
	t.Log(val)
	fmt.Printf("%", val)

}

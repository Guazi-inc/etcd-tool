package client

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestParseConf(t *testing.T) {
	assert.Equal(t,
		ParseDSN("wby:p:s@w@localhost:2379,localhost:2479/call/s"),
		&Config{
			Username: "wby",
			Password: "p:s@w",
			Addrs:    "localhost:2379,localhost:2479",
			Path:     "/call/s/",
		},
	)
	assert.Equal(t,
		ParseDSN("localhost:2379/call/s"),
		&Config{
			Addrs: "localhost:2379",
			Path:  "/call/s/",
		},
	)
	assert.Equal(t,
		ParseDSN("localhost:2379,localhost:2479"),
		&Config{
			Addrs: "localhost:2379,localhost:2479",
			Path:  "/",
		},
	)
	assert.Equal(t,
		ParseDSN("@localhost:2379,localhost:2479"),
		&Config{
			Addrs: "localhost:2379,localhost:2479",
			Path:  "/",
		},
	)
	assert.Equal(t,
		ParseDSN("wby:p:s@w@localhost:2379/"),
		&Config{
			Username: "wby",
			Password: "p:s@w",
			Addrs:    "localhost:2379",
			Path:     "/",
		},
	)
	assert.Equal(t,
		ParseDSN("wby:p:s@w@localhost:2379"),
		&Config{
			Username: "wby",
			Password: "p:s@w",
			Addrs:    "localhost:2379",
			Path:     "/",
		},
	)
}

func TestClient_Get(t *testing.T) {
	cli, _ := NewClient("localhost:2379")
	val, err := cli.Get("/test")
	t.Log(err)
	t.Log(val)
	fmt.Printf("%", val)

}

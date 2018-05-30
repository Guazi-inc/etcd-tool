package config

import (
	"os"
	"testing"

	"time"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	InitETCD("localhost:2479")
	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	//WithCustomWatch("/test", func() {
	//	fmt.Println(time.Now(), "key event")
	//}, func() {
	//	fmt.Println(time.Now(), "key event2")
	//})
	{
		var cfg map[string]string
		err := Get("/call/permission_http", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg struct {
			AnsweringTimeout int               `json:"answering_timeout"`
			Enabled          bool              `json:"enabled"`
			Gray             map[string]string `json:"gray"`
		}
		err := Get("/call/call_in", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg struct {
			AnsweringTimeout int                `json:"answering_timeout"`
			Enabled          bool               `json:"enabled"`
			Gray             *map[string]string `json:"gray"`
		}
		fmt.Println(cfg.Gray == nil)
		err := Get("/call/call_in", &cfg)
		assert.Nil(t, err)
		t.Log(cfg, cfg.Gray)
	}
	{
		var cfg struct {
			AnsweringTimeout int  `json:"answering_timeout"`
			Enabled          bool `json:"enabled"`
			Gray             *struct {
				Ids  []int  `json:"ids"`
				Type string `json:"type"`
			} `json:"gray"`
		}
		err := Get("/call/call_in", &cfg)
		assert.Nil(t, err)
		t.Log(cfg, cfg.Gray)
	}
	{
		var cfg struct {
			Ids  *[]int `json:"ids"`
			Type string `json:"type"`
		}
		err := Get("/call/call_in/gray", &cfg)
		assert.Nil(t, err)
		t.Log(cfg, cfg.Ids)
	}
	{
		var ids []int32
		err := Get("/call/call_in/gray/ids", &ids)
		assert.Nil(t, err)
		t.Log(ids)
	}
	{
		var cfg struct {
			Address string `json:"address"`
			Prefix  string `json:"prefix"`
			Timeout int32  `json:"timeout"`
		}
		err := Get("/call/redis", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}

	{
		etcdClient.DeleteWithPrefix("/test")
		etcdClient.Put("/test/1", "asd")
		etcdClient.Put("/test/2", "\"zxc\"")
		var cfg = map[int]string{}
		err := Get("/test", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}

	{
		etcdClient.DeleteWithPrefix("/test")
		etcdClient.Put("/test/1", "true")
		etcdClient.Put("/test/2", "false")
		var cfg = map[int]bool{}
		err := Get("/test", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		etcdClient.DeleteWithPrefix("/test")
		etcdClient.Put("/test/true", "true")
		etcdClient.Put("/test/false", "false")
		time.Sleep(1000 * time.Millisecond)
		var cfg = map[bool]bool{}
		err := Get("/test", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)

		etcdClient.Put("/test", "true")
		time.Sleep(1000 * time.Millisecond)
		var cfg2 string
		err = Get("/test", &cfg2)
		assert.Nil(t, err)
		t.Log(cfg2)

		etcdClient.Put("/test", "true")
		time.Sleep(1000 * time.Millisecond)
		var cfg3 bool
		err = Get("/test", &cfg3)
		assert.Nil(t, err)
		t.Log(cfg3)

		etcdClient.Put("/test", "123")
		time.Sleep(1000 * time.Millisecond)
		var cfg4 int32
		err = Get("/test", &cfg4)
		assert.Nil(t, err)
		t.Log(cfg4)
	}

	//time.Sleep(time.Minute)
}

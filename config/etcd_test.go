package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Guazi-inc/etcd-tool/client"
	"github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

type ApiConfig struct {
	Endpoint string `json:"endpoint"`
	AppKey   string `json:"app_key"`
	Secret   string `json:"secret"`
}

func TestMain(m *testing.M) {
	InitETCD("localhost:2479/call/smart_after_sale")
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
		time.Sleep(1000 * time.Millisecond)
		var cfg = map[int]string{}
		err := Get("/test", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}

	{
		etcdClient.DeleteWithPrefix("/test")
		etcdClient.Put("/test/1", "true")
		etcdClient.Put("/test/2", "false")
		time.Sleep(1000 * time.Millisecond)
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

func TestGet2(t *testing.T) {
	{
		key := "/wby/test_key"
		var val = map[string]string{
			"key":  "val",
			"key2": "val2",
			"key3": "val3",
		}
		bytes, err := jsoniter.Marshal(val)
		assert.Nil(t, err)
		etcdClient.Put(key, string(bytes))
		time.Sleep(1000 * time.Millisecond)
		var cfg map[string]string
		err = Get(key, &cfg)
		assert.Nil(t, err)
		assert.Equal(t, cfg, val)
	}
	{
		key := "/wby/test_key"
		err := etcdClient.DeleteWithPrefix(key)
		assert.Nil(t, err)
		var cfg map[string]string
		err = Get(key, &cfg)
		assert.Equal(t, err, ErrKvsEmpty)
	}

}

func TestGet3(t *testing.T) {
	cfg := client.ParseDSN("@localhost:2379")
	t.Log(cfg)
	path := strings.TrimSuffix(strings.TrimPrefix(cfg.Path, "/"), "/")
	t.Log(path)

	sp := strings.Split(cfg.Path, "/")
	t.Log(sp)
	t.Log(len(sp))
}

func TestGet4(t *testing.T) {
	SetNamespace("finance")
	type mainConfig struct {
		Env            string            `json:"env"`
		Redis          map[string]string `json:"redis"`
		PermissionHttp ApiConfig         `json:"permission_http"`
	}
	{
		var cfg mainConfig
		err := Get("/call", &cfg)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg mainConfig
		err := GetInNamespace("", &cfg, 1)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg mainConfig
		err := GetInNamespace("call", &cfg, 0)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg map[string]string
		err := GetInNamespace("email", &cfg, 1)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg map[string]string
		err := GetInNamespace("/email", &cfg, 2)
		assert.Nil(t, err)
		t.Log(cfg)
	}
	{
		var cfg map[string]string
		err := GetInNamespace("/email", &cfg, 3)
		t.Log(err)
		t.Log(cfg)
	}

}

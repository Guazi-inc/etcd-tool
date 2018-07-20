package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Guazi-inc/etcd-tool/utils"
	"github.com/coreos/etcd/clientv3"
)

type Client struct {
	*clientv3.Client
}

type Config struct {
	Addrs    string
	Username string
	Password string
	Path     string
}

// username:password@addr1,addr2/path
func ParseDSN(dsn string) *Config {
	rg, err := regexp.Compile(`^(?:(?:(.*?):(.*))?@)?(.*?)(/.*|$)`)
	if err != nil {
		return nil
	}
	ss := rg.FindStringSubmatch(dsn)
	cfg := &Config{
		Addrs:    ss[3],
		Username: ss[1],
		Password: ss[2],
		Path:     ss[4],
	}
	if cfg.Addrs == "" {
		return nil
	}
	if !strings.HasSuffix(cfg.Path, "/") {
		cfg.Path = fmt.Sprintf("%s/", cfg.Path)
	}

	return cfg
}

func NewClient(dsn string) (*Client, error) {
	cfg := ParseDSN(dsn)
	if cfg == nil || cfg.Addrs == "" {
		return nil, errors.New("invalid conf")
	}
	var cli *clientv3.Client

	if err := utils.Retries(3, 3*time.Second, func(_ int) error {
		c, e := clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(cfg.Addrs, ","),
			DialTimeout: 3 * time.Second,
			Username:    cfg.Username,
			Password:    cfg.Password,
		})
		cli = c
		return e
	}); err != nil {
		return nil, err
	}
	return &Client{
		Client: cli,
	}, nil
}

func (ec *Client) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := ec.Client.Get(ctx, key)
	cancel()
	if err != nil || len(resp.Kvs) == 0 {
		return "", err
	}
	return string(resp.Kvs[0].Value), nil
}

func (ec *Client) GetWithPrefix(key string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := ec.Client.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	var kvs = map[string]string{}
	for _, item := range resp.Kvs {
		kvs[string(item.Key)] = string(item.Value)
	}
	return kvs, nil
}

func (ec *Client) Put(key, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := ec.Client.Put(ctx, key, val)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func (ec *Client) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := ec.Client.Delete(ctx, key)
	cancel()
	return err
}

func (ec *Client) DeleteWithPrefix(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := ec.Client.Delete(ctx, key, clientv3.WithPrefix())
	cancel()
	return err
}

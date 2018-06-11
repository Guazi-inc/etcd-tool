package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/Guazi-inc/etcd-tool/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	delimiter = "/"
)

var (
	etcdClient  *client.Client
	kvsMapCache sync.Map
	kvCache     sync.Map
	watchFunc   sync.Map
	initOnce    sync.Once
)

func init() {
	dsn := os.Getenv("ETCD_ADDR")
	if dsn != "" {
		InitETCD(dsn)
	}
}

func InitETCD(addr string) {
	cli, err := client.NewClient(addr)
	if err != nil {
		panic(err)
	}
	initOnce.Do(func() {
		etcdClient = cli
		go Watch()
	})
}

func Get(key string, config interface{}) error {
	if !isValidKey(key) {
		return errors.New("invalid key")
	}

	ct := reflect.TypeOf(config).Elem()
	cv := reflect.ValueOf(config).Elem()

	switch ct.Kind() {
	case reflect.Struct, reflect.Map:
		result, err := getKvsMapWithCache(key)
		if err != nil {
			return err
		}
		return fillConfig(result, ct, cv)
	default:
		val, err := getValWithCache(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), config)
		if err != nil && ct.Kind() == reflect.String {
			*config.(*string) = val
			return nil
		}
		return err
	}
}

func getValWithCache(key string) (string, error) {
	if v, ok := kvCache.Load(key); ok {
		return v.(string), nil
	}
	val, err := etcdClient.Get(key)
	if err != nil {
		return "", err
	}
	if val != "" {
		kvCache.Store(key, val)
	}
	return val, nil
}

func getKvsMapWithCache(key string) (map[string]interface{}, error) {
	if v, ok := kvsMapCache.Load(key); ok {
		return v.(map[string]interface{}), nil
	}
	kvsMap, err := getKvsMap(key)
	if err != nil {
		return nil, err
	}
	if len(kvsMap) > 0 {
		kvsMapCache.Store(key, kvsMap)
	}
	return kvsMap, nil
}

func getKvsMap(key string) (map[string]interface{}, error) {
	if !strings.HasSuffix(key, delimiter) {
		key = key + delimiter
	}
	kvs, err := etcdClient.GetWithPrefix(key)
	if err != nil || len(kvs) == 0 {
		return nil, err
	}
	for k := range kvs {
		if !isValidKey(k) {
			delete(kvs, k)
		}
	}
	return parseKvs(key, kvs), nil
}

func isValidKey(key string) bool {
	return strings.HasPrefix(key, delimiter) && !strings.HasSuffix(key, delimiter) && !strings.Contains(key, "//")
}

func parseKvs(baseKey string, kvs map[string]string) map[string]interface{} {
	if !strings.HasSuffix(baseKey, delimiter) {
		baseKey = baseKey + delimiter
	}
	result := map[string]interface{}{}
	mapSubKvs := map[string]map[string]string{}
	for key, val := range kvs {
		if !strings.HasPrefix(key, baseKey) {
			continue
		}
		k := strings.TrimPrefix(key, baseKey)
		if splitKey := strings.Split(k, delimiter); len(splitKey) > 1 {
			if mapSubKvs[splitKey[0]] == nil {
				mapSubKvs[splitKey[0]] = map[string]string{}
			}
			mapSubKvs[splitKey[0]][key] = val
		} else {
			result[k] = val
		}
	}
	for key, val := range mapSubKvs {
		result[key] = parseKvs(fmt.Sprintf("%s%s", baseKey, key), val)
	}
	return result
}

func fillConfig(result interface{}, ct reflect.Type, cv reflect.Value) error {
	if !cv.IsValid() {
		return errors.Errorf("invalid type: %s, value: %s", ct.String(), cv.String())
	}

	if ct.Kind() == reflect.Map && cv.IsNil() {
		cv.Set(reflect.MakeMap(ct))
	}
	if ct.Kind() == reflect.Ptr {
		if cv.IsNil() {
			switch ct.Elem().Kind() {
			case reflect.Map:
				cv.Set(reflect.New(ct.Elem()))
				cv.Elem().Set(reflect.MakeMap(ct.Elem()))
			default:
				cv.Set(reflect.New(ct.Elem()))
			}
		}
		ct = ct.Elem()
		cv = cv.Elem()
	}

	switch res := result.(type) {
	case string:
		if err := json.Unmarshal([]byte(res), cv.Addr().Interface()); err != nil {
			if ct.Kind() == reflect.String {
				cv.SetString(res)
			} else {
				return err
			}
		}

	case map[string]interface{}:
		switch ct.Kind() {
		case reflect.Struct:
			for index := 0; index < ct.NumField(); index++ {
				key := strings.Split(ct.Field(index).Tag.Get("json"), ",")[0]
				val, ok := res[key]
				if key == "" || !ok {
					continue
				}
				fillConfig(val, ct.Field(index).Type, cv.Field(index))
			}
		case reflect.Map:
			for key, value := range res {
				vm := reflect.New(ct.Elem())
				switch vv := value.(type) {
				case map[string]interface{}:
					if err := fillConfig(vv, ct.Elem(), vm.Elem()); err != nil {
						return err
					}
				case string:
					if err := json.Unmarshal([]byte(vv), vm.Interface()); err != nil {
						if ct.Elem().Kind() == reflect.String {
							vm.Elem().SetString(vv)
						} else {
							return err
						}
					}
				default:
					return errors.New("unknow result type")
				}
				kt := reflect.New(ct.Key())
				if err := json.Unmarshal([]byte(key), kt.Interface()); err != nil {
					if ct.Key().Kind() == reflect.String {
						kt.Elem().SetString(key)
					} else {
						return err
					}
				}
				cv.SetMapIndex(kt.Elem(), vm.Elem())
			}
		}
	default:
		return errors.New("unknow result type")
	}

	return nil
}

func Watch() {
	wc := etcdClient.Watch(context.Background(), delimiter, clientv3.WithPrefix())
	for w := range wc {
		for _, ev := range w.Events {
			logrus.Infof("ETCD %s KEY %s", ev.Type, string(ev.Kv.Key))
			if !isValidKey(string(ev.Kv.Key)) {
				continue
			}
			if _, ok := kvCache.Load(string(ev.Kv.Key)); ok {
				kvCache.Store(string(ev.Kv.Key), string(ev.Kv.Value))
			}

			keySplit := strings.Split(strings.TrimPrefix(string(ev.Kv.Key), delimiter), delimiter)
			for i := range keySplit {
				k := fmt.Sprintf("%s%s", delimiter, strings.Join(keySplit[0:i+1], delimiter))
				if _, ok := kvsMapCache.Load(k); ok {
					if result, err := getKvsMap(k); err == nil {
						kvsMapCache.Store(k, result)
						bytes, _ := json.Marshal(result)
						if len(bytes) > 72 {
							bytes = append(bytes[0:72], byte(46), byte(46), byte(46))
						}
						logrus.Infof("Etcd cache KEY %s updated with %s", k, string(bytes))
					}
				}
			}

			go runWatchFuncs(string(ev.Kv.Key))
		}
	}
}

func WithCustomWatch(key string, fs ...func()) {
	if v, ok := watchFunc.Load(key); ok {
		if fs2, ok := v.([]func()); ok {
			fs = append(fs2, fs...)
		}
	}
	watchFunc.Store(key, fs)
	logrus.Infof("watch function registered with key: %s", key)
}

func runWatchFuncs(key string) {
	logrus.Infof("run watch funcs with key: %s", key)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("got panic when running watch function with key: %s, panic: %+v", key, r)
		}
	}()
	watchFunc.Range(func(k, v interface{}) bool {
		if strings.HasPrefix(key, k.(string)) {
			if fs, ok := v.([]func()); ok {
				for _, f := range fs {
					f()
				}
			}
		}
		return true
	})
}

func CheckKeys(keys, keysWithPrefix []string) {
	for _, k := range keys {
		if v, err := etcdClient.Get(k); err != nil || v == "" {
			panic(fmt.Sprintf("empty key: %s, Err: %+v", k, err))
		}
	}

	for _, k := range keysWithPrefix {
		if m, err := etcdClient.GetWithPrefix(k); err != nil || len(m) == 0 {
			panic(fmt.Sprintf("empty key with prefix: %s, Err: %s", k, err))
		}
	}
}

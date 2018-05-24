// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/Guazi-inc/etcd-tool/client"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put data from json file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := putArg.Run(); err != nil {
			logrus.Errorf("got err: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

var putArg = PutArg{
	Kvs: map[string]string{},
}

type PutArg struct {
	Conf     string
	Dsn      string
	Cfg      *client.Config
	DirValue string
	Dirs     []string
	Kvs      map[string]string
	DelDirs  []string
	DelKeys  []string
}

func init() {
	RootCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	putCmd.Flags().StringVarP(&putArg.Conf, "conf", "c", "", "configure file")
	putCmd.Flags().StringVarP(&putArg.Dsn, "etcd", "e", "", "etcd address")
	putCmd.Flags().StringVarP(&putArg.DirValue, "dir_value", "d", "", "dir value")

}

func (p *PutArg) Run() error {
	p.Cfg = client.ParseDSN(p.Dsn)
	if p.Cfg == nil {
		return errors.New("invalid params")
	}
	logrus.Infof("ETCD ADDRESS: %s", putArg.Cfg.Addrs)

	bytes, err := ioutil.ReadFile(p.Conf)
	if err != nil {
		return err
	}
	confMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &confMap)
	if err != nil {
		return err
	}
	if err := p.parseKeyValue(confMap, p.Cfg.Path); err != nil {
		return err
	}

	p.Dirs = append(p.Dirs, getDirs(p.Cfg.Path)...)

	cli, err := client.NewClient(p.Dsn)
	if err != nil {
		return err
	}

	var errDelKvDirs []string
	var errKvs = map[string]string{}
	for k, v := range p.Kvs {
		if err = cli.DeleteWithPrefix(k); err != nil {
			errDelKvDirs = append(errDelKvDirs, k)
		}
		if err = cli.Put(k, v); err != nil {
			errKvs[k] = v
		}
	}

	if len(errDelKvDirs) > 0 {
		logrus.Errorf("delete dir which should be key, %d fail, %+v", len(errDelKvDirs), errDelKvDirs)
		os.Exit(1)
	}
	if len(errKvs) > 0 {
		logrus.Errorf("put %d key, %d fail, %+v", len(p.Kvs), len(errKvs), errKvs)
		os.Exit(1)
	} else {
		logrus.Infof("put %d key, all success", len(p.Kvs))
	}

	var errDelDirKeys []string
	for _, d := range p.Dirs {
		if err = cli.Delete(d); err != nil {
			errDelDirKeys = append(errDelDirKeys, d)
		}
	}
	if len(errDelDirKeys) > 0 {
		logrus.Errorf("delete key which should be dir, %d fail, %+v", len(errDelDirKeys), errDelDirKeys)
		os.Exit(1)
	}
	if p.DirValue != "" {
		var errDirs []string
		for _, d := range p.Dirs {
			if err = cli.Put(d, p.DirValue); err != nil {
				errDirs = append(errDirs, d)
			}
		}
		if len(errDirs) > 0 {
			logrus.Errorf("create %d dir, %d fail, %+v", len(p.Dirs), len(errDirs), errDirs)
			os.Exit(1)
		} else {
			logrus.Infof("create %d dir, all success", len(p.Dirs))
		}
	}

	if len(p.DelDirs) > 0 {
		var errDel []string
		for _, d := range p.DelDirs {
			if err = cli.DeleteWithPrefix(d); err != nil {
				errDel = append(errDel, d)
			}
		}
		if len(errDel) > 0 {
			logrus.Errorf("delete %d dir, %d fail, %+v", len(p.DelDirs), len(errDel), errDel)
			os.Exit(1)
		} else {
			logrus.Infof("delete %d dir, all success", len(p.DelDirs))
		}
	}

	if len(p.DelKeys) > 0 {
		var errDel []string
		for _, d := range p.DelKeys {
			if err := cli.DeleteWithPrefix(d); err != nil {
				errDel = append(errDel, d)
			}
		}
		if len(errDel) > 0 {
			logrus.Errorf("delete %d key, %d fail, %+v", len(p.DelKeys), len(errDel), errDel)
			os.Exit(1)
		} else {
			logrus.Infof("delete %d key, all success", len(p.DelKeys))
		}
	}

	return nil
}

func (p *PutArg) parseKeyValue(confMap map[string]interface{}, baseKey string) error {
	if !strings.HasSuffix(baseKey, delimiter) {
		baseKey += delimiter
	}
	if len(confMap) == 0 {
		p.DelDirs = append(p.DelDirs, strings.TrimSuffix(baseKey, delimiter))
	}
	for key, value := range confMap {
		fullKey := fmt.Sprintf("%s%s", baseKey, key)
		if strings.Contains(key, delimiter) {
			return errors.Errorf("invalid key %s, contains %s", key, delimiter)
		}
		switch val := value.(type) {
		case map[string]interface{}:
			p.Dirs = append(p.Dirs, fullKey)
			if err := p.parseKeyValue(val, fullKey); err != nil {
				return err
			}
		case string:
			if val != "" {
				p.Kvs[fullKey] = val
			} else {
				p.DelKeys = append(p.DelKeys, fullKey)
			}
		default:
			buf, err := json.Marshal(val)
			if err != nil {
				return err
			}
			bufStr := string(buf)
			if bufStr != "" && bufStr != "[]" {
				p.Kvs[fullKey] = bufStr
			} else {
				p.DelKeys = append(p.DelKeys, fullKey)
			}
		}
	}
	return nil
}

func getDirs(path string) []string {
	path = strings.TrimPrefix(path, delimiter)
	path = strings.TrimSuffix(path, delimiter)
	sp := strings.Split(path, delimiter)
	var ret []string
	for idx := range sp {
		ret = append(ret, delimiter+strings.Join(sp[0:idx+1], delimiter))
	}
	return ret
}

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
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/Guazi-inc/etcd-tool/client"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copy etcd data",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := copyArg.Run(); err != nil {
			logrus.Errorf("got err: %s\n", err.Error())
		}
	},
}

var copyArg CopyArg

type CopyArg struct {
	From    string
	To      string
	FromCfg *client.Config
	ToCfg   *client.Config
}

func init() {
	RootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// copyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	copyCmd.Flags().StringVarP(&copyArg.From, "from", "f", "", "source address and dir (host:port/path)")
	copyCmd.Flags().StringVarP(&copyArg.To, "to", "t", "", "destination address and dir (host:port/path)")
}

func (c *CopyArg) Run() error {
	c.FromCfg = client.ParseDSN(c.From)
	c.ToCfg = client.ParseDSN(c.To)
	if c.FromCfg == nil || c.ToCfg == nil {
		return errors.New("invalid params")
	}

	fromCli, err := client.NewClient(c.From)
	if err != nil {
		return err
	}
	kvs, err := fromCli.GetWithPrefix(c.FromCfg.Path)
	if err != nil {
		return err
	}
	toCli, err := client.NewClient(c.To)
	if err != nil {
		return err
	}
	var errKvs []string
	for k, v := range kvs {
		nk := fmt.Sprintf("%s%s", c.ToCfg.Path, strings.TrimPrefix(k, c.FromCfg.Path))
		if err = toCli.Put(nk, v); err != nil {
			errKvs = append(errKvs, k)
		}
	}
	kvsInfo := fmt.Sprintf("copy %d key,", len(kvs))
	if len(errKvs) > 0 {
		kvsInfo += fmt.Sprintf(" %d fail", len(errKvs))
	} else {
		kvsInfo += " all success"
	}
	fmt.Println(kvsInfo)

	return nil
}

/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/version"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	VER "github.com/containernetworking/cni/pkg/version"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

const (
	postRetryTimes         = 60
	postRetryIntervalInSec = 10
)

var lockFilePath string
var lockFile *os.File

func Response4K8sNew() {
	ipv4, _ := types.ParseCIDR("1.2.3.30/24")
	routegwv4, routev4, _ := net.ParseCIDR("15.5.6.8/24")
	ipv6, _ := types.ParseCIDR("abcd:1234:ffff::cdde/64")
	routegwv6, routev6, _ := net.ParseCIDR("1111:dddd::aaaa/80")
	res := &current.Result{
		CNIVersion: "0.3.1",
		Interfaces: []*current.Interface{
			{
				Name:    "eth0",
				Mac:     "00:11:22:33:44:55",
				Sandbox: "/proc/3553/ns/net",
			},
		},
		IPs: []*current.IPConfig{
			{
				Version:   "4",
				Interface: current.Int(0),
				Address:   *ipv4,
				Gateway:   net.ParseIP("1.2.3.1"),
			},
			{
				Version:   "6",
				Interface: current.Int(0),
				Address:   *ipv6,
				Gateway:   net.ParseIP("abcd:1234:ffff::1"),
			},
		},
		Routes: []*types.Route{
			{Dst: *routev4, GW: routegwv4},
			{Dst: *routev6, GW: routegwv6},
		},
		DNS: types.DNS{
			Nameservers: []string{"1.2.3.4", "1::cafe"},
			Domain:      "acompany.com",
			Search:      []string{"somedomain.com", "otherdomain.net"},
			Options:     []string{"foo", "bar"},
		},
	}
	res.Print()
}

func SendReqToKnitterAgent(args *skel.CmdArgs, url string) error {
	defer lockLogFlush()
	klog.Infof("request URL=%v", url)
	bodyType := "application/json"
	reqJSON, err := json.Marshal(args)
	if err != nil {
		klog.Error("Marshal CNI skel.CmdArgs Error:", err)
		return err
	}
	klog.Info("URL:[", url, "]---[", string(reqJSON), "]")
	for idx := 0; idx < postRetryTimes; idx++ {
		postReader := bytes.NewReader([]byte(reqJSON))
		resp, err := http.Post(url, bodyType, postReader)
		if err != nil {
			klog.Errorf("KnitterAgent post error! -%v", err)
			if strings.Contains(err.Error(), "i/o timeout") {
				return err
			}
			time.Sleep(postRetryIntervalInSec * time.Second)
			continue
		}
		klog.Info("StatusCode:", resp.StatusCode)

		respData, _ := ioutil.ReadAll(resp.Body)
		var respMap map[string]string
		json.Unmarshal(respData, &respMap)
		if resp.StatusCode == 409 {
			klog.Errorf("SendReqToKnitterAgent error:%v", respMap["ERROR"])
			return errors.New("SendReqToKnitterAgent error:" + respMap["ERROR"])
		} else if resp.StatusCode == 200 && strings.Contains(respMap["STATUS"], "200") {
			klog.Info("SendReqToKnitterAgent success!")
			return nil
		} else {
			return errors.New("sendReqToKnitterAgent error:the error is not knitter-agent main return")
		}

	}
	return errors.New("sendReqToKnitterAgent: error: post to knitter-agent failed retry over threshold")
}

const KnitterAgent string = "http://127.0.0.1:6006/v1/pod"

func cmdAdd(args *skel.CmdArgs) error {
	defer lockLogFlush()
	defer func() {
		if err := recover(); err != nil {
			klog.Info("@@@cmdAdd panic recover start!@@@")
			klog.Error("Stack:", string(debug.Stack()))
			klog.Info("@@@cmdAdd panic recover end!@@@")
		}
	}()
	klog.Infof("cmdAdd hello!!")
	reqURL := KnitterAgent + "?operation=attach"

	record := buidRecord(args)
	klog.Infof("cmdAdd:@@@ %v", record)

	err := SendReqToKnitterAgent(args, reqURL)
	if err != nil {
		klog.Error("Attach ports to POD error:", err)
		return err
	}
	Response4K8sNew()
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	defer lockLogFlush()
	defer func() {
		if err := recover(); err != nil {
			//e = errors.New("panic")
			klog.Info("@@@cmdAdd panic recover start!@@@")
			klog.Error("Stack:", string(debug.Stack()))
			klog.Info("@@@cmdAdd panic recover end!@@@")
		}
	}()
	Response4K8sNew()
	klog.Infof("cmdDel hello!!")
	reqURL := KnitterAgent + "?operation=detach"

	record := buidRecord(args)
	klog.Infof("cmdDel:@@@ %v", record)

	err := SendReqToKnitterAgent(args, reqURL)
	if err != nil {
		klog.Error("Detach ports from POD error:", err)
	}
	return nil
}

//logFileInit creats log file for knitter,and return the filename
func logFileInit() {
	lockFile, lockFilePath, _ = klog.Create(time.Now())
	lockFile.Close()
}

//lockLogFlush lock and unlock  the file for flush, avoid muti-progress problem
func lockLogFlush() {
	klog.Flush()
}

func main() {
	klog.ConfigLog("/root/info/logs/nwnode/")
	flag.Parse()
	if version.HasVerFlag() {
		version.PrintVersion()
		return
	}

	defer lockLogFlush()

	logFileInit()
	var pluginInfo VER.PluginInfo
	pluginInfo = VER.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdDel, pluginInfo)
}

func getPodnsBy(cniparam string) string {
	var k8sns = ""
	for _, item := range strings.Split(cniparam, ";") {
		if strings.Contains(item, "K8S_POD_NAMESPACE") {
			k8sparam := strings.Split(item, "=")
			k8sns = k8sparam[1]
		}
	}
	return k8sns
}

func getPodNameBy(cniparam string) string {
	var k8sname = ""
	for _, item := range strings.Split(cniparam, ";") {
		if strings.Contains(item, "K8S_POD_NAME") {
			k8sparam := strings.Split(item, "=")
			k8sname = k8sparam[1]
		}
	}
	return k8sname
}

func buidRecord(args *skel.CmdArgs) string {
	podNs := getPodnsBy(args.Args)
	podName := getPodNameBy(args.Args)
	containerID := args.ContainerID
	timeStemp := time.Now().Local()
	record := "| time:" + fmt.Sprintf("%v", timeStemp) + "| PodNameSpace:" + fmt.Sprintf("%v", podNs) +
		"| PodName:" + fmt.Sprintf("%v", podName) + "| ContainerId:" + fmt.Sprintf("%v", containerID)
	return record
}

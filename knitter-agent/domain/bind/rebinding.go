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

package bind

import (
	//"bufio"
	//"bytes"
	"bufio"
	"bytes"
	"github.com/ZTE/Knitter/pkg/klog"
	"os/exec"
	"strings"
	"time"
)

const PCIPATH = "/sys/bus/pci"

//var PCIPATH = "/sys/bus/pci"

type Dpdknic struct {
	BusInfo    string      `json:"bus_info"`
	Name       string      `json:"name"`
	IP         string      `json:"ip"`
	Mask       string      `json:"mask"`
	Gateway    string      `json:"gateway"`
	Mac        string      `json:"mac"`
	Function   string      `json:"function"`
	Accelerate string      `json:"accelerate"`
	BusInfos   []string    `json:"bus_infos"`
	BondMode   string      `json:"bond_type"`
	Metadata   interface{} `json:"metadata"`
}

type Nicdevs struct {
	Dpdknics []Dpdknic `json:"nic_dev"`
	MTU      string    `json:"mtu"`
}

var GetLspciOutputFunc func(args []string) ([]byte, error)

func GetLspciOutput(args []string) ([]byte, error) {
	lspci, _ := exec.LookPath("lspci")
	lspciOutput, err := exec.Command(lspci, args...).CombinedOutput()

	if err != nil {
		klog.Errorf("bind-getLspciOutput:Unable to exec lspci -n, err: %v, output: %s", err, string(lspciOutput))
		return nil, err
	}

	return lspciOutput, err
}

func init() {
	GetLspciOutputFunc = GetLspciOutput
}

func SumOfVirtioNetPci() int {
	lspciArgs := []string{"-b"}
	lspciOutput, _ := GetLspciOutputFunc(lspciArgs)
	var numOfTotalPcis int = 0
	lspciScanner := bufio.NewScanner(bytes.NewBuffer(lspciOutput))
	for lspciScanner.Scan() {
		line := strings.TrimSpace(lspciScanner.Text())
		if len(line) == 0 {
			continue
		}
		if strings.Contains(line, "Inc Virtio network device") {
			numOfTotalPcis = numOfTotalPcis + 1
			//	klog.Infof("pci[%v]:%v", numOfTotalPcis, line)
		}
	}
	return numOfTotalPcis
}

const lspciCmdRetryTimes = 10

func GetPciID(busInfo string) (string, error) {
	var idx int
	var err error
	var lspciOutput []byte
	lspciArgs := []string{"-n"}
	for ; idx < lspciCmdRetryTimes; idx++ {
		lspciOutput, err = GetLspciOutputFunc(lspciArgs)
		if err != nil {
			klog.Errorf("bind-getPciId:getLspciOutputFunc failed, err: %v, just retry", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if idx == lspciCmdRetryTimes {
		klog.Errorf("bind-getPciId:getLspciOutputFunc all try failed, err: %v, exit", err)
		return "", err
	}

	pciID := ""
	lspciScanner := bufio.NewScanner(bytes.NewBuffer(lspciOutput))
	for lspciScanner.Scan() {
		line := strings.TrimSpace(lspciScanner.Text())
		if len(line) == 0 {
			continue
		}
		if strings.Contains(line, strings.TrimLeft(busInfo, "0000:")) {
			parts := strings.Split(line, " ")
			if len(parts) < 3 {
				klog.Errorf("bind-getPciId:Unexpected lspci output: %v", parts)
			}
			pciID = parts[2]
			break
		}
	}
	klog.Info("bind-getPciId:Query success, PCI ID:%s", pciID)
	return pciID, nil
}

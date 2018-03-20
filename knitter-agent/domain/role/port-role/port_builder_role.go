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

package portrole

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/pkg/klog"
	"regexp"
)

const ipReg string = "^(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|[1-9])\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)$"

type PortBuilderRole struct {
	NetworkName  string
	NetworkPlane string
	PortName     string
	VnicType     string
	Accelerate   string
	FixIP        string
	IPGroupName  string
}

func (this *PortBuilderRole) Transform(port *monitor.Port) error {
	klog.Infof("PortBuilderRole.Transform start , monitor.Port is [%v]", port)
	networkName := port.EagerAttr.NetworkName
	if networkName == "" {
		err := errors.New("Port attach network is blank")
		klog.Error("Port attach network is blank")
		return err
	}
	portName := port.EagerAttr.PortName
	if portName == "" {
		portName = "eth_" + networkName
		if len(portName) > 12 {
			rs := []rune(portName)
			portName = string(rs[0:12])
		}
	}
	if len(portName) > 12 {
		klog.Errorf("Lenth of port name is greater than 12")
		return errors.New("lenth of port name is illegal")
	}
	portFunc := port.EagerAttr.NetworkPlane

	if portFunc == "" {
		portFunc = "std"
	}
	flag := isNetworkPlane(portFunc)
	if !flag {
		portFunc = "std"
	}
	portType := port.EagerAttr.VnicType
	if portType == "" {
		portType = "normal"
	}
	flag = isVnicType(portType)
	if !flag {
		portType = "normal"
	}
	if cni.GetGlobalContext().HostType == constvalue.HostTypeVirtualMachine && portType == "physical" {
		err := errors.New("unsupport port type [physical]")
		klog.Errorf("unsupport port type [physical]")
		return err
	}
	isUseDpdk := port.EagerAttr.Accelerate
	if isUseDpdk != "true" {
		isUseDpdk = "false"
	}
	if portFunc == constvalue.NetPlaneStd {
		isUseDpdk = "false"
	}
	if cni.GetGlobalContext().HostType == constvalue.HostTypePhysicalMachine && portType == "normal" && isUseDpdk == "true" {
		portType = "direct"
	}
	ipAddress := port.EagerAttr.FixIP

	if ipAddress == "" {
		klog.Infof("Get port ip_address is blank string")
	} else {
		isLegitimate, errIPAddress := isIPLegitimate(ipAddress)
		if errIPAddress != nil {
			klog.Errorf("check ip_address failure")
			return errIPAddress
		}
		if isLegitimate {
			klog.Infof("ip_address is legitimate")
		} else {
			klog.Errorf("ip_address is illegal")
			return errors.New("ip_address is illegal")
		}
	}

	ipGroupName := port.EagerAttr.IPGroupName
	this.NetworkName = networkName
	this.NetworkPlane = portFunc
	this.PortName = portName
	this.VnicType = portType
	this.Accelerate = isUseDpdk
	this.FixIP = ipAddress
	this.IPGroupName = ipGroupName

	return nil
}

func isNetworkPlane(networkPlane string) bool {
	switch networkPlane {
	case "control":
		return true
	case "media":
		return true
	case "std":
		return true
	case "eio":
		return true
	case "oam":
		return true
	}
	return false
}

func isVnicType(vnicType string) bool {
	switch vnicType {
	case "normal":
		return true
	case "direct":
		return true
	case "physical":
		return true
	}
	return false
}

func isIPLegitimate(ipAddress string) (bool, error) {
	return regexp.MatchString(ipReg, ipAddress)
}

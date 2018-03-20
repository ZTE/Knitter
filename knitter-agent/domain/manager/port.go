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

package manager

import (
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
	"net"
)

type Port struct {
	ID               string `mapstructure:"id" json:"id"`
	NetworkID        string `mapstructure:"network_id" json:"network_id"`
	NetworkName      string `mapstructure:"network_name" json:"network_name"`
	Name             string `mapstructure:"name" json:"name"`
	MACAddress       string `mapstructure:"mac_address" json:"mac_address"`
	FixedIPs         []IP   `mapstructure:"fixed_ips" json:"fixed_ips"`
	TenantID         string `mapstructure:"tenant_id" json:"tenant_id"`
	CIDR             string `mapstructure:"cidr" json:"cidr"`
	GatewayIP        string `mapstructure:"gateway_ip" json:"gateway_ip"`
	MTU              string `mapstructure:"mtu" json:"mtu"`
	NetworkType      string `json:"neutron_network_type"`
	IsDefaultGateway bool   `json:"is_default_gateway"`
	OrgDriver        string `json:"org_driver"`
}

func (self *Port) MakeAddr() netlink.Addr {
	return netlink.Addr{
		IPNet: self.MakeIPNet(),
		Label: self.Name,
	}
}

func (self *Port) GetIPNet() *net.IPNet {
	var ipNet *net.IPNet
	//if self.Name == "eth0" {
	var ok bool
	if infra.IsEnhancedMode() {
		ok = self.NetworkName == "net_api"
	} else {
		ok = self.IsDefaultGateway
	}
	if ok {
		ipNet = MakeIPNetByCidr("0.0.0.0/0")
	} else {
		ipNet = self.MakeIPNetByCidr()
	}

	return ipNet
}

func (self *Port) MakeIPNet() *net.IPNet {
	fixedIP := self.makeIPByFixedIP0()
	cidrIP := self.MakeIPNetByCidr()

	return &net.IPNet{
		IP:   fixedIP,
		Mask: cidrIP.Mask,
	}
}

func (self *Port) makeIPByFixedIP0() net.IP {
	return net.ParseIP(self.FixedIPs[0].Address)
}

func (self *Port) MakeIPNetByCidr() *net.IPNet {
	return MakeIPNetByCidr(self.CIDR)
}

func (self *Port) GetGatewayIP() net.IP {
	return net.ParseIP(self.GatewayIP)
}

func MakeIPNetByCidr(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		klog.Errorf("AddSelectedIntfToPod:net.ParseCIDR(subnet.CIDR) error!-%v", err)
		return nil
	}
	return ipNet
}

type VLanInfo struct {
	NetworkType     string `json:"network_type"`
	VlanID          string `json:"network_id"`
	PhysicalNetwork string `json:"physical_network"`
}

type CreatePortReq struct {
	agtmgr.AgtPortReq
}

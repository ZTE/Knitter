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
	"github.com/ZTE/Knitter/knitter-agent/infra/util/base"
	"github.com/ZTE/Knitter/knitter-agent/infra/util/implement"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
	"net"
	"strconv"
)

type Link struct {
	netlink.Link
	Netlink base.NetLink
	IsVf    bool
	Pf      string
}

func NewLink(link netlink.Link) *Link {
	return &Link{Link: link, Netlink: &implement.NetLink{}}
}

func (self *Link) GetName() string {
	return self.Attrs().Name
}

func (self *Link) SetName(name string) {
	self.Attrs().Name = name
}

func (self *Link) SetDynamicName() bool {
	err := self.Netlink.LinkSetName(self.Link, self.Attrs().Name)
	if err != nil {
		klog.Errorf("bind-Link-setDynamicName:netlink.LinkSetName() name:%s error!err is : -%v", self.Attrs().Name, err)
		return false
	}
	return true
}

func (self *Link) SetUp() bool {
	err := self.Netlink.LinkSetUp(self)
	if err != nil {
		klog.Errorf("bind-Link-setUp:Link SetUp() error! err is -%v", err)
		return false
	}
	return true
}

func (self *Link) SetDown() bool {
	err := self.Netlink.LinkSetDown(self)
	if err != nil {
		klog.Errorf("bind-Link-setDown:Link SetDown() error!err is :-%v", err)
		return false
	}
	return true
}

func (self *Link) SetAddr(addr *netlink.Addr) bool {
	err := self.Netlink.AddrAdd(self, addr)
	if err != nil {
		klog.Errorf("bind-Link-setAddr:netlink.AddrAdd(self, addr) error!err is :-%v", err)
		return false
	}
	return true
}

func (self *Link) SetMtu(mtu string) bool {
	mTu, err := strconv.Atoi(mtu)
	if err != nil {
		klog.Errorf("bind-Link-setMtu:strconv.Atoi(mtu) error! err is: -%v", err)
		return false
	}
	err1 := self.Netlink.LinkSetMTU(self.Link, mTu)
	if err1 != nil {
		klog.Errorf("bind-Link-setMtu:LinkSetMtu() error!err is: -%v", err1)
		return false
	}
	return true
}

func (self *Link) SetMac(mac string) bool {
	macByte, err1 := net.ParseMAC(mac)
	if err1 != nil {
		klog.Infof("parse mac error,%v", err1)
	}
	klog.Infof("setMac mac byte is :%v", macByte)
	err := self.Netlink.LinkSetMac(self.Link, macByte)
	if err != nil {
		klog.Errorf("bind-Link-SetMac:LinkSetMac() error!err is: -%v", err)
		return false
	}
	return true
}

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

package implement

import (
	"github.com/ZTE/Knitter/knitter-agent/infra/util/base"
	"github.com/vishvananda/netlink"
	"net"
)

type NetLink struct {
	base.NetLink
}

func (self *NetLink) LinkSetNsPid(link netlink.Link, nspid int) error {
	return netlink.LinkSetNsPid(link, nspid)
}
func (self *NetLink) LinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}
func (self *NetLink) LinkSetName(link netlink.Link, name string) error {
	return netlink.LinkSetName(link, name)
}
func (self *NetLink) LinkSetUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}
func (self *NetLink) LinkSetDown(link netlink.Link) error {
	return netlink.LinkSetDown(link)
}
func (self *NetLink) AddrAdd(link netlink.Link, addr *netlink.Addr) error {
	return netlink.AddrAdd(link, addr)
}
func (self *NetLink) LinkSetMTU(link netlink.Link, mtu int) error {
	return netlink.LinkSetMTU(link, mtu)
}
func (self *NetLink) LinkSetMac(link netlink.Link, mac net.HardwareAddr) error {
	return netlink.LinkSetHardwareAddr(link, mac)
}

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

package base

import (
	"github.com/vishvananda/netlink"
	"net"
)

type NetLink interface {
	LinkSetNsPid(link netlink.Link, nspid int) error
	SetLinkListRtn(b bool)
	LinkList() ([]netlink.Link, error)
	LinkSetName(link netlink.Link, name string) error
	LinkSetUp(link netlink.Link) error
	LinkSetDown(link netlink.Link) error
	AddrAdd(link netlink.Link, addr *netlink.Addr) error
	LinkSetMTU(link netlink.Link, mtu int) error
	LinkSetMac(link netlink.Link, mac net.HardwareAddr) error
}

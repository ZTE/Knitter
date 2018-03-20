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
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Machine struct {
	NameSpace netns.NsHandle
	LinkRepo  *LinkRepository

	NetLink base.NetLink
}

func NewMachine(ns netns.NsHandle) *Machine {
	return &Machine{
		NameSpace: ns,
		LinkRepo:  NewLinkRepository(),
	}
}

func (self *Machine) UpdateLinks() {
	self.LinkRepo.Update(self.GetCurrentNsLinks())
}

func (self *Machine) FindLinkByMac(mac string) *Link {
	return self.LinkRepo.FindLinkByMac(mac)
}

func (self *Machine) FindLinkByPci(pci string) *Link {
	return self.LinkRepo.FindLinkByPci(pci)
}

func (self *Machine) FindLinkByName(name string) *Link {
	return self.LinkRepo.FindLinkByName(name)
}

func (self *Machine) FindVfLinkByName(name string) *Link {
	return self.LinkRepo.FindVfLinkByName(name)
}
func (self *Machine) GetCurrentNsLinks() []netlink.Link {
	links, err := self.NetLink.LinkList()
	if err != nil {
		klog.Errorf("bind-Machine-AttachIntfToPod:netlink.LinkList error!-%v", err)
		//return errors.New("AttachIntfToPod:netlink.LinkListerror!")
		return []netlink.Link{}
	}
	return links
}

func (self *Machine) SetNetLink(nl base.NetLink) {
	self.NetLink = nl
}

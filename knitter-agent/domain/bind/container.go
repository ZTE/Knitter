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
	"github.com/vishvananda/netns"
	//"github.com/vishvananda/netlink"
	"github.com/ZTE/Knitter/knitter-agent/infra/util/base"
	"github.com/ZTE/Knitter/pkg/klog"
	//"knitter_agent/knitter/util/implement"
)

type Container struct {
	ID        string
	Pid       int
	NameSpace netns.NsHandle
	Link      *Link

	NetLink base.NetLink
	NetNs   base.NetNs
}

func NewContainer(ID int) *Container {
	Link := &Link{}
	return &Container{Pid: ID, Link: Link}
}

func (self *Container) BindLink(link *Link) bool {
	self.Link = link
	klog.Infof("bind-container-bindLink:container Pid is: %v", self.Pid)
	var err error = nil
	self.NameSpace, err = self.NetNs.GetFromPid(self.Pid)
	if err != nil {
		klog.Errorf("bind-container-bindLink:netns.GetFromPid error!-%v", err)
	}
	defer self.NameSpace.Close()
	err = self.NetLink.LinkSetNsPid(link, self.Pid)
	if err != nil {
		klog.Errorf("bind-container-bindLink:netlink.LinkSetNsPid error!-%v", err)
		//return errors.New("AddSelectedIntfToPod:netlink.LinkSetNsPid error!")
		return false
	}
	return true
}

func (self *Container) SetNetLink(netlink base.NetLink) {
	self.NetLink = netlink
}

func (self *Container) SetNetNs(netns base.NetNs) {
	self.NetNs = netns
}

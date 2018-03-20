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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/infra/util/base"
	"github.com/ZTE/Knitter/knitter-agent/infra/util/implement"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Pod struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	K8sns   string `json:"k8sns"`
	Podtype string `json:"podtype"`

	LinkRepo *LinkRepository

	NameSpace base.NameSpace
	NetLink   base.NetLink
}

func NewPod(name string, id string, k8sns string, podType string) *Pod {
	return &Pod{
		Name:     name,
		ID:       id,
		K8sns:    k8sns,
		Podtype:  podType,
		LinkRepo: NewLinkRepository(),
	}
}

func NewTempPod() *Pod {
	return &Pod{
		Name:     "name",
		ID:       "id",
		K8sns:    "k8sns",
		Podtype:  "podType",
		LinkRepo: NewLinkRepository(),
	}
}

var AttachVethToPod = func(netNs string, port *manager.Port, vethName string) error {
	pod := NewTempPod()
	nameSpace4Pod := &implement.NameSpace{}
	pod.SetNameSpace(nameSpace4Pod)
	netLink4Pod := &implement.NetLink{}
	pod.SetNetLink(netLink4Pod)

	vmNs, _ := pod.NameSpace.Get()
	defer vmNs.Close()

	vm := NewMachine(vmNs)
	netLink4Machine := &implement.NetLink{}
	vm.SetNetLink(netLink4Machine)
	pid, errN := NetNSToPID(netNs)
	if errN != nil {
		klog.Errorf("Error moving to netns. Err: %v", errN)
		return errN
	}
	container := NewContainer(pid)
	netLink4Container := &implement.NetLink{}
	container.SetNetLink(netLink4Container)
	netNs4Container := &implement.NetNs{}
	container.SetNetNs(netNs4Container)

	err := pod.BindLinkByVethName(container, port, vm, vethName, vmNs)
	if err != nil {
		return fmt.Errorf("%v:bind-AddSelectedIntfToPod:pod.bindlink(containerID, port, vfLink)error", err)
	}

	return nil
}

func NetNSToPID(ns string) (int, error) {
	ok := strings.HasPrefix(ns, "/proc/")
	if !ok {
		return -1, fmt.Errorf("invalid nw name space: %v", ns)
	}

	elements := strings.Split(ns, "/")
	return strconv.Atoi(elements[2])
}

const MaxRetryTimesOfFindLinkByName int = 10

func (self *Pod) BindLinkByVethName(container *Container, port *manager.Port, vm *Machine, vethName string, vmNs netns.NsHandle) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var vfLink *Link = nil
	for idx := 0; idx < MaxRetryTimesOfFindLinkByName; idx++ {
		vm.UpdateLinks()
		vfLink = vm.FindLinkByName(vethName)
		if vfLink == nil {
			klog.Errorf("bind-pod-bindLink:vm.FindLinkByName(vethName:%v)[%v]times error: vfLink is null !", vethName, idx)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if vfLink == nil {
		return errors.New("bind-pod-bindLink:vm.FindLinkByName(vethName)  error: vfLink is null")
	}
	//k8s pod must contain eth0 nic
	//if port.Name == "std" {
	//	port.Name = "eth0"
	//}
	errt := self.AddLinkToContainer(container, vfLink, port, vmNs)
	if errt != nil {
		klog.Errorf("bind-pod-bindLink:addLinkToContainer error!-%v", errt)
		return fmt.Errorf("%v,bind-pod-bindLink:addLinkToContainer error", errt)
	}

	return nil
}

func (self *Pod) AddLinkToContainer(container *Container, vfLink *Link, port *manager.Port, vmNs netns.NsHandle) error {
	klog.Info("bind-Pod-addLinkToContainer:vflink.attrs().name is (origial):", vfLink.Attrs().Name)
	// Attach the VF link to pause container's namespace

	container.BindLink(vfLink)

	containerNameSpace, err := container.NetNs.GetFromPid(container.Pid)
	if err != nil {
		klog.Errorf("bind-container-bindLink:netns.GetFromPid error!-%v", err)
	}
	container.NameSpace = containerNameSpace
	defer container.NameSpace.Close()

	errSet := self.NameSpace.Set(container.NameSpace)
	if errSet != nil {
		return fmt.Errorf("%v:bind-Pod-addLinkToContainer:netns.Set error", errSet)
	}
	defer self.NameSpace.Set(vmNs)

	vfLink.SetName(port.Name)
	str, _ := json.Marshal(port)
	klog.Info("Port-Infomation:", string(str))
	if !vfLink.SetDynamicName() {
		return errors.New("bind-Pod-addLinkToContainer:netlink.LinkSetName error")
	}
	if vfLink.SetUp() == false {
		klog.Infof("bind-Pod-addLinkToContainer:setup link error!")
	}
	addr := port.MakeAddr()
	klog.Info("bind-Pod-addLinkToContainer:addr info :", addr.String())
	if !vfLink.SetAddr(&addr) {
		return errors.New("bind-Pod-addLinkToContainer:setAdd error")
	}
	if !vfLink.SetMtu(port.MTU) {
		return errors.New("pod-addLinkToContainer:netlink.LinkSetMTU error")
	}
	if !vfLink.SetMac(port.MACAddress) {
		return errors.New("pod-addLinkToContainer:netlink.SetMac error")
	}
	// add a gateway route
	klog.Info("bind-Pod-addLinkToContainer:route cidr is :", port.CIDR)

	dstIPNet := port.GetIPNet()
	if dstIPNet == nil {
		return errors.New("bind-Pod-addLinkToContainer:net.ParseCIDR error")
	}

	//route := netlink.Route{LinkIndex: vfLink.Attrs().Index, Dst: dstIp, Gw: gwip}
	err = RouteAddFunc(CreateRoute(vfLink.Attrs().Index, dstIPNet, port.GetGatewayIP()))
	klog.Infof("net[%v] gw is:%v", port.NetworkName, port.GetGatewayIP())
	if err != nil {
		klog.Errorf("bind-Pod-addLinkToContainer:netlink.RouteAdd error!-%v", err)
	}

	klog.Info("bind-pod-addLinkToContainer:success adding addr :", addr.IPNet, "|", "to link:", vfLink.Attrs().Name)

	return nil
}

func (self *Pod) DeleteLink(container *Container, portMac string) error {
	// Lock the OS Thread so we ls't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Save the current network namespace
	vmNs, _ := self.NameSpace.Get()
	defer vmNs.Close()
	defer container.NameSpace.Close()

	errSet := self.NameSpace.Set(container.NameSpace)
	if errSet != nil {
		return fmt.Errorf("%v:bind-Pod-DeleteDeviceInPod:netns.Set(ns) error", errSet)
	}
	// Get all the links
	self.UpdateLinks()
	vfLink := self.FindLinkByMac(portMac)
	if vfLink != nil {
		vfLink.SetDown()
		if vfLink.GetName() == "eth0" {
			vfLink.SetName("std")
		}
	}

	// Switch back to the original namespace
	err := self.NameSpace.Set(vmNs)
	if err != nil {
		return fmt.Errorf("%v,bind-Pod-DeleteDeviceInPod: netns.Set(origns) error", err)
	}
	return nil
}

func (self *Pod) UpdateLinks() {
	self.LinkRepo.Update(self.GetCurrentNsLinks())
}

func (self *Pod) FindLinkByMac(mac string) *Link {
	return self.LinkRepo.FindLinkByMac(mac)
}

func (self *Pod) GetCurrentNsLinks() []netlink.Link {
	links, err := self.NetLink.LinkList()
	if err != nil {
		klog.Errorf("bind-pod-getCurrentNsLinks:netlink.LinkList error!-%v", err)
		//return errors.New("AttachIntfToPod:netlink.LinkListerror!")
		return []netlink.Link{}
	}
	return links
}

func (self *Pod) SetNetLink(netlink base.NetLink) {
	self.NetLink = netlink
}

func (self *Pod) SetNameSpace(ns base.NameSpace) {
	self.NameSpace = ns
}

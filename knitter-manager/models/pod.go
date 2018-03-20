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

package models

import (
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"

	"github.com/ZTE/Knitter/pkg/db-accessor"

	"github.com/ZTE/Knitter/knitter-manager/public"
	"strings"
)

type PodIP struct {
	NetworkPlane      string `json:"network_plane"`
	NewtworkPlaneName string `json:"network_plane_name"`
	IPAddress         string `json:"ip_address"`
}

type Pod struct {
	Name       string   `json:"name"`
	PodIps     []*PodIP `json:"ips"`
	tenantUUID string
}

type EncapPod struct {
	Pod *Pod `json:"pod"`
}

type EncapPods struct {
	Pods []*Pod `json:"pods"`
}

func (self *Pod) SetTenantID(id string) error {
	self.tenantUUID = id
	return nil
}

func (self *Pod) GetFromEtcd(ns, name string) (*Pod, error) {
	key4PodSelf := dbaccessor.GetKeyOfPodSelf(self.tenantUUID, ns, name)
	klog.Info("Get-POD-From-Etcd:", key4PodSelf)
	podStr, err1 := common.GetDataBase().ReadLeaf(key4PodSelf)
	if err1 != nil {
		klog.Error("Read-Pod[", key4PodSelf, "]-data-Error:", err1.Error())
		return nil, err1
	}

	klog.Info("Pod: ", podStr)
	o, _ := jason.NewObjectFromBytes([]byte(podStr))
	self.Name, _ = o.GetString("name")
	klog.Info("pod name: ", self.Name)

	err := self.getInterfaces(ns, name)
	if err != nil {
		klog.Error("Read-NS[", ns, "]name[", name,
			"]-interface-data-Error:", err.Error())
	}
	klog.Info("Get Pod: ", self)

	klog.Info("Now out GetFromEtcd Function")
	return self, nil
}

func (self *Pod) getInterfaces(ns, name string) error {
	key := dbaccessor.GetKeyOfInterfaceGroupInPod(self.tenantUUID, ns, name)
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Error("Read interfaces in pod Error:", err)
		self.PodIps = nil
		return err
	}

	self.PodIps = make([]*PodIP, 0, len(nodes))
	for _, node := range nodes {
		portStr4Pod, err1 := common.GetDataBase().ReadLeaf(node.Value)
		if err1 != nil {
			klog.Error("Read-Pod[", node.Value, "]-data-Error:", err1.Error())
			continue
		}
		klog.Info("NS[", ns, "]name[", name, "]Port:", portStr4Pod)
		obj, err := jason.NewObjectFromBytes([]byte(portStr4Pod))
		if err != nil {
			klog.Error("NewObj from str Err:", err)
			continue
		}
		networkPlane, _ := obj.GetString("net_plane_type")
		networkPlaneName, _ := obj.GetString("net_plane_name")
		ip, _ := obj.GetString("ip")
		podIP := PodIP{NetworkPlane: networkPlane,
			IPAddress: ip, NewtworkPlaneName: networkPlaneName}
		self.PodIps = append(self.PodIps, &podIP)
	}
	return nil
}

func (self *Pod) getAllPodInNameSpace(ns string) []*Pod {
	klog.Info("Get-All-Pod-In-Name-Space:", ns)
	var pods []*Pod
	key := dbaccessor.GetKeyOfPodGroup(self.tenantUUID, ns)
	nodes, errE := common.GetDataBase().ReadDir(key)
	if errE != nil {
		klog.Error("Read NameSpace Dir from ETCD Error:", errE)
		return nil
	}

	for _, node := range nodes {
		pod := Pod{}
		pod.tenantUUID = self.tenantUUID
		namePod := strings.TrimPrefix(node.Key, key+"/")
		p, err := pod.GetFromEtcd(ns, namePod)
		if err != nil {
			klog.Error("pod.GetFromEtcd returns Error:", err)
			continue
		}
		pods = append(pods, p)
	}
	return pods
}

var PodListAll = func(pod *Pod) []*Pod {
	return pod.ListAll()
}

func (self *Pod) ListAll() []*Pod {
	klog.Info("List-All-Pod")
	var pods []*Pod
	//Get all Namespaces by user
	nsURL := dbaccessor.GetKeyOfPodNsGroup(self.tenantUUID)
	nss, errE := common.GetDataBase().ReadDir(nsURL)
	if errE != nil {
		klog.Warning("Read NameSpace info from ETCD Error:", errE)
		return nil
	}
	for _, ns := range nss {
		nsPod := strings.TrimPrefix(ns.Key, nsURL+"/")
		newPods := self.getAllPodInNameSpace(nsPod)
		if newPods != nil {
			pods = append(pods, newPods...)
		}
	}
	return pods
}

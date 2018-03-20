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

package cni

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/containernetworking/cni/pkg/skel"
	"strings"
	//"manager/models"
	"fmt"
)

type CniParam struct {
	ContainerID string `json:"ContainerId"`
	IfName      string `json:"IfName"`
	Netns       string `json:"Netns"`
	Path        string `json:"Path"`
	Args        string `json:"Args"`
	StdinData   []byte `json:"StdinData"`

	PodNs     string
	PodName   string
	Mtu       string
	HostType  string
	VnicType  string
	VMID      string
	TenantID  string
	ClusterID string

	DB             dbaccessor.DbAccessor
	Manager        manager.ManagerClient
	RemoteNetType  string
	PaasNwConfPath string

	cniArgs      *skel.CmdArgs
	serverInfo   []byte
	serverJason  *jason.Object
	k8sStdInData *jason.Object

	oseToken string
}

func (self *CniParam) init(k8s *skel.CmdArgs) error {
	self.ContainerID = k8s.ContainerID
	self.IfName = k8s.IfName
	self.Netns = k8s.Netns
	self.Path = k8s.Path
	self.Args = k8s.Args
	self.StdinData = k8s.StdinData
	paramjson, err := jason.NewObjectFromBytes([]byte(self.StdinData))
	if err != nil {
		klog.Errorf("init:jason.NewObjectFromBytes() error! -%v", err)
		return fmt.Errorf("%v:init:jason.NewObjectFromBytes() error", err)
	}
	self.k8sStdInData = paramjson
	klog.Infof("Init:cniJason[%v]", self.k8sStdInData)
	klog.Infof("Init:serverJason[%v]", self.serverJason)
	return nil
}

func (self *CniParam) AnalyzeCniParam(k8s *skel.CmdArgs) error {
	err := self.init(k8s)
	if err != nil {
		klog.Errorf("Init CNI Object ERROR!")
		return fmt.Errorf("%v:Init CNI Object ERROR", err)
	}
	self.setPodnsBy(self.Args)
	self.setPodNameBy(self.Args)
	self.Mtu = GetGlobalContext().Mtu
	klog.Infof("AnalyzeCniParam:get MTU = :", self.Mtu)

	self.HostType = GetGlobalContext().HostType
	self.VnicType = "normal"

	self.VMID = GetGlobalContext().VMID

	klog.Infof("AnalyzeCniParam: Init etcd client!------------------------------------------------")
	self.DB = GetGlobalContext().DB

	klog.Infof("AnalyzeCniParam: Init cni manager client!-------------------------------------------")
	self.Manager = GetGlobalContext().Mc
	self.ClusterID = GetGlobalContext().ClusterID

	klog.Infof("AnalyzeCniParam:Init Ose client!---------------------------------------------------")

	tenantErr := self.setTenantID()
	if tenantErr != nil {
		klog.Errorf("AnalyzeCniParam:setTenantId error! -%v", tenantErr)
		return fmt.Errorf("%v:AnalyzeCniParam:setTenantId error", tenantErr)
	}

	return nil
}

func (self *CniParam) setPodnsBy(cniparam string) error {
	var k8sns = ""
	for _, item := range strings.Split(cniparam, ";") {
		if strings.Contains(item, "K8S_POD_NAMESPACE") {
			k8sparam := strings.Split(item, "=")
			k8sns = k8sparam[1]
		}
	}
	klog.Info("setPodnsBy:New Pod k8s namespace:", k8sns)
	self.PodNs = k8sns
	return nil
}

func (self *CniParam) setPodNameBy(cniparam string) error {
	var k8sname = ""
	for _, item := range strings.Split(cniparam, ";") {
		if strings.Contains(item, "K8S_POD_NAME") {
			k8sparam := strings.Split(item, "=")
			k8sname = k8sparam[1]
		}
	}
	klog.Info("setPodNameBy:New Pod k8s name:", k8sname)
	self.PodName = k8sname
	return nil
}

func (self *CniParam) setTenantID() error {
	self.TenantID = self.PodNs
	return nil
}

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

package driver

import (
	"fmt"
	. "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"net/http"
	"sync"
)

const MaxReqForAttach int = 5

type NovaClient struct {
	VmLock    map[string]*sync.Mutex
	Channel   chan int
	AttachReq int
}

func (self *NovaClient) GetAttachReq() int {
	klog.Infof("OpenStack.GetAttachReq:%d", self.AttachReq)
	return self.AttachReq
}

func (self *NovaClient) SetAttachReq(req int) {
	if req <= 0 || req > 30 {
		req = MaxReqForAttach
	}
	self.AttachReq = req
	klog.Infof("OpenStack.SetAttachReq:%d", self.AttachReq)
}

func (self *NovaClient) getAttachPortUrl(vmId string) string {
	return getAuthSingleton().ComputeEndpoint + "servers/" + vmId + "/os-interface"
}

func (self *NovaClient) getDetachPortUrl(vmId, portId string) string {
	return self.getAttachPortUrl(vmId) + "/" + portId
}

func (self *NovaClient) makeAttachPortBody(portId string) map[string]interface{} {
	dict := make(map[string]interface{})
	dict["port_id"] = portId

	return map[string]interface{}{"interfaceAttachment": dict}
}

func (self *NovaClient) AttachPortToVM(vmId, portId string) (*Interface, error) {
	if self.Channel == nil {
		klog.Infof("make channel size:%d", self.AttachReq)
		self.Channel = make(chan int, self.AttachReq)
	}
	self.Channel <- 0

	url := self.getAttachPortUrl(vmId)
	body := self.makeAttachPortBody(portId)
	klog.Info("AttachPortToVM: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	<-self.Channel

	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("AttachPortToVM: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], error: ", err)
		return nil, fmt.Errorf("%v:%v:AttachPortToVM: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("AttachPortToVM: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:AttachPortToVM: NewObjectFromBytes parse response body error", err)
	}
	port, err := self.parseAttachInterface(rspJasObj)
	if err != nil {
		klog.Error("AttachPortToVM: parseAttachInterface error: ", err.Error())
		return nil, fmt.Errorf("%v:AttachPortToVM: parseAttachInterface error", err)
	}

	return port, nil

}

func (self *NovaClient) parseAttachInterface(obj *jason.Object) (*Interface, error) {
	id, err := obj.GetString("interfaceAttachment", "port_id")
	if err != nil {
		klog.Error("parseAttachInterface: GetString interfaceAttachment->port_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseAttachInterface: GetString interfaceAttachment->port_id error", err)
	}

	networkId, err := obj.GetString("interfaceAttachment", "net_id")
	if err != nil {
		klog.Error("parseAttachInterface: GetString interfaceAttachment->net_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseAttachInterface: GetString interfaceAttachment->net_id error", err)
	}

	mac, err := obj.GetString("interfaceAttachment", "mac_addr")
	if err != nil {
		klog.Error("parseAttachInterface: GetString interfaceAttachment->mac_addr error: ", err.Error())
		return nil, fmt.Errorf("%v:parseAttachInterface: GetString interfaceAttachment->mac_addr error", err)
	}

	status, err := obj.GetString("interfaceAttachment", "port_state")
	if err != nil {
		klog.Error("parseAttachInterface: GetString interfaceAttachment->port_state error: ", err.Error())
		return nil, fmt.Errorf("%v:parseAttachInterface: GetString interfaceAttachment->port_state error", err)
	}

	return &Interface{Id: id, NetworkId: networkId, MacAddress: mac, Status: status}, nil
}

func (self *NovaClient) DetachPortFromVM(vmId, portId string) error {
	if self.Channel == nil {
		klog.Infof("make channel size:%d", self.AttachReq)
		self.Channel = make(chan int, self.AttachReq)
	}
	self.Channel <- 0

	url := self.getDetachPortUrl(vmId, portId)
	klog.Info("DetachPortFromVM: url: ", url)
	status, err := doHttpDeleteWithReAuth(url)
	<-self.Channel

	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("DetachPortFromVM: Delete url[", url, "], status[", status, "], error: ", err)
		return fmt.Errorf("%v:%v:DetachPortFromVM: Delete request error", status, err)
	}

	return nil
}

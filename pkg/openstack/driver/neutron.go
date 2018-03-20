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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/pkg/adapter"
	. "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"net/http"
)

type NeutronClient struct {
}

type Ip struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address,omitempty"`
}

type NeutronPort struct {
	Id         string `json:"id"`
	NetworkId  string `json:"network_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	MacAddress string `json:"mac_address"`
	DeviceId   string `json:"device_id"`
	Ips        []*Ip  `json:"fixed_ips"`
}

type NeutronBulkPortsResp struct {
	Ports []*NeutronPort `json:"ports"`
}

func (self *NeutronClient) CreateNetwork(name string) (string, error) {
	url := self.getNetworksUrl()
	body := self.makeCreateNetworkBody(name)

	klog.Info("CreateNetwork: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("CreateNetwork: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], response body[", string(rspBytes), "], error: ", err)
		return "", fmt.Errorf("%v:%v:CreateNetwork: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateNetwork: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:CreateNetwork: NewObjectFromBytes parse response body error", err)
	}
	id, err := rspJasObj.GetString("network", "id")
	if err != nil {
		klog.Error("CreateNetwork: GetString network->id error: ", err.Error())
		return "", fmt.Errorf("%v:CreateNetwork: GetString network->id error", err)
	}

	return id, nil
}

func (self *NeutronClient) getNetworksUrl() string {
	return getAuthSingleton().NetworkEndpoint + "networks"
}

func (self *NeutronClient) makeCreateNetworkBody(name string) map[string]interface{} {
	dict := make(map[string]interface{})
	dict["name"] = name
	dict["admin_state_up"] = true
	dict["shared"] = false

	return map[string]interface{}{"network": dict}
}

func (self *NeutronClient) CreatePort(networkId, subnetId, portName, ip, mac, vnicType string) (*Interface, error) {
	url := self.getPortsUrl()
	body := self.makeCreatePortBody(networkId, subnetId, portName, ip, mac, vnicType)
	klog.Info("CreatePort: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("CreatePort: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], error: ", err)
		return nil, fmt.Errorf("%v:%v:CreatePort: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreatePort: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: NewObjectFromBytes parse response body error", err)
	}

	port, err := parsePortAttrs(rspJasObj)
	if err != nil {
		klog.Error("CreatePort: parseCreatePortAttrs error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: parseCreatePortAttrs error", err)
	}

	return port, nil
}

func (self *NeutronClient) getPortsUrl() string {
	return getAuthSingleton().NetworkEndpoint + "ports"
}

func (self *NeutronClient) makeCreatePortBody(networkId, subnetId, portName, ip, mac, vnicType string) map[string]interface{} {
	dictOpts := make(map[string]interface{})
	dictOpts["port"] = makeCreatePort(networkId, subnetId, portName, ip, mac, vnicType)
	return dictOpts
}

func makeCreatePort(networkId, subnetId, portName, ip, mac, vnicType string) map[string]interface{} {

	dictIp := make(map[string]interface{})
	dictIp["subnet_id"] = subnetId
	if ip != "" {
		dictIp["ip_address"] = ip
	}

	fixedIpList := []map[string]interface{}{dictIp}

	dictPort := make(map[string]interface{})
	dictPort["fixed_ips"] = fixedIpList
	dictPort["admin_state_up"] = true
	dictPort["network_id"] = networkId
	dictPort["name"] = portName
	dictPort["binding:vnic_type"] = vnicType
	if mac != "" {
		dictPort["mac_address"] = mac
	}

	return dictPort
}

func parsePortAttrs(obj *jason.Object) (*Interface, error) {
	id, err := obj.GetString("port", "id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->id error", err)
	}

	name, err := obj.GetString("port", "name")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->name error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->name error", err)
	}

	status, err := obj.GetString("port", "status")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->status error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->status error", err)
	}

	networkId, err := obj.GetString("port", "network_id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->network_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->network_id error", err)
	}

	mac, err := obj.GetString("port", "mac_address")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->mac_address error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->mac_address error", err)
	}

	deviceId, err := obj.GetString("port", "device_id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->device_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->device_id error", err)
	}

	fixedIps, err := obj.GetObjectArray("port", "fixed_ips")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->fixed_ips error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->fixed_ips error", err)
	}

	if len(fixedIps) < 1 {
		klog.Error("parseCreatePortAttrs: fixed_ips has no content")
		return nil, errors.New("parseCreatePortAttrs: fixed_ips has no content")
	}

	// now wo only use first element of fixed_ips array
	subnetId, err := fixedIps[0].GetString("subnet_id")
	if err != nil {
		klog.Error("CreatePort: GetString port->fixed_ips->subnet_id error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: GetString port->fixed_ips->subnet_id error", err)
	}

	ip, err := fixedIps[0].GetString("ip_address")
	if err != nil {
		klog.Error("CreatePort: GetString port->fixed_ips->ip_address error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: GetString port->fixed_ips->ip_address error", err)
	}

	return &Interface{Id: id, Name: name, Status: status, NetworkId: networkId, MacAddress: mac, DeviceId: deviceId, SubnetId: subnetId, Ip: ip}, nil
}

func parsePortsAttrs(obj *jason.Object) ([]*Interface, error) {
	portsObj, err := obj.GetObjectArray("ports")
	if err != nil {
		klog.Error("parsePortsAttrs: GetObjectArray ports error: ", err.Error())
		return nil, fmt.Errorf("%v:parsePortsAttrs: GetObjectArray ports error", err)
	}

	interfaces := []*Interface{}
	for _, portObj := range portsObj {
		id, err := portObj.GetString("id")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->id error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->id error", err)
		}

		name, err := portObj.GetString("name")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->name error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->name error", err)
		}

		status, err := portObj.GetString("status")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->status error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->status error", err)
		}

		networkId, err := portObj.GetString("network_id")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->network_id error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->network_id error", err)
		}

		mac, err := portObj.GetString("mac_address")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->mac_address error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->mac_address error", err)
		}

		deviceId, err := portObj.GetString("device_id")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->device_id error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->device_id error", err)
		}

		fixedIps, err := portObj.GetObjectArray("fixed_ips")
		if err != nil {
			klog.Error("parsePortsAttrs: GetString port->fixed_ips error: ", err.Error())
			return nil, fmt.Errorf("%v:parsePortsAttrs: GetString port->fixed_ips error", err)
		}

		if len(fixedIps) < 1 {
			klog.Error("parsePortsAttrs: fixed_ips has no content")
			return nil, errors.New("parsePortsAttrs: fixed_ips has no content")
		}

		// now wo only use first element of fixed_ips array
		subnetId, err := fixedIps[0].GetString("subnet_id")
		if err != nil {
			klog.Error("CreatePort: GetString port->fixed_ips->subnet_id error: ", err.Error())
			return nil, fmt.Errorf("%v:CreatePort: GetString port->fixed_ips->subnet_id error", err)
		}

		ip, err := fixedIps[0].GetString("ip_address")
		if err != nil {
			klog.Error("CreatePort: GetString port->fixed_ips->ip_address error: ", err.Error())
			return nil, fmt.Errorf("%v:CreatePort: GetString port->fixed_ips->ip_address error", err)
		}

		interfaces = append(interfaces, &Interface{Id: id, Name: name, Status: status, NetworkId: networkId, MacAddress: mac, DeviceId: deviceId, SubnetId: subnetId, Ip: ip})
	}

	return interfaces, nil

}

const LogicalPortDefaultVnicType = "normal"

func (self *NeutronClient) CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*Interface, error) {
	url := self.getPortsUrl()
	opts := make([]map[string]interface{}, 0)
	for _, reqPort := range req.Ports {
		opt := makeCreatePort(reqPort.NetworkId, reqPort.SubnetId, reqPort.PortName, reqPort.FixIP, "", LogicalPortDefaultVnicType)
		opts = append(opts, opt)
	}
	m := make(map[string]interface{})
	m["ports"] = opts

	klog.Info("CreateBulkPorts: url: ", url, " http body: ", m)
	status, rspBytes, err := doHttpPostWithReAuth(url, m)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("CreateBulkPorts: Post url[", url, "], body[", m, "], status[", status, "], response body[", string(rspBytes), "], response body[", string(rspBytes), "], error: ", err)
		return nil, fmt.Errorf("%v:%v:CreateBulkPorts: Post request error", status, err)
	}

	resp := NeutronBulkPortsResp{}
	err = adapter.Unmarshal(rspBytes, &resp)
	if err != nil {
		return nil, errobj.ErrUnmarshalFailed
	}
	inters := make([]*Interface, 0)
	for _, port := range resp.Ports {
		inter := &Interface{
			Id:         port.Id,
			Name:       port.Name,
			NetworkId:  port.NetworkId,
			Status:     port.Status,
			MacAddress: port.MacAddress,
			DeviceId:   port.DeviceId,
			SubnetId:   port.Ips[0].SubnetId,
			Ip:         port.Ips[0].IpAddress,
		}
		inters = append(inters, inter)
	}
	return inters, nil
}

func (self *NeutronClient) GetPort(id string) (*Interface, error) {
	url := self.getPortIdUrl(id)
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("GetPort: Get url[", url, "], status[", status, "], error: ", err)
		return nil, fmt.Errorf("%v:%v:GetPort: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetPort: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetPort: NewObjectFromBytes parse response body error", err)
	}

	port, err := parsePortAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetPort: parseCreatePortAttrs error: ", err.Error())
		return nil, fmt.Errorf("%v:GetPort: parseCreatePortAttrs error", err)
	}

	return port, nil
}

func (self *NeutronClient) DeletePort(id string) error {
	url := self.getPortIdUrl(id)
	klog.Info("DeletePort: url: ", url)
	status, err := doHttpDeleteWithReAuth(url)
	if status < http.StatusOK || (status > http.StatusMultipleChoices && status != http.StatusNotFound) || err != nil {
		klog.Error("DeletePort: Delete url[", url, "], status[", status, "], error: ", err)
		return fmt.Errorf("%v:%v:DeletePort: Delete request error", status, err)
	}

	return nil
}

func (self *NeutronClient) getPortIdUrl(id string) string {
	return self.getPortsUrl() + "/" + id
}

func (self *NeutronClient) CreateProviderNetwork(name, nwType, phyNet, sId string) (string, error) {
	url := self.getNetworksUrl()
	body := self.makeCreateProviderNetworkOpts(name, nwType, phyNet, sId)
	klog.Info("CreateProviderNetwork: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("CreateProviderNetwork: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], error: ", err)
		return "", fmt.Errorf("%v:%v:CreateProviderNetwork: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateProviderNetwork: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:CreateNetwork: NewObjectFromBytes parse response body error", err)
	}
	id, err := rspJasObj.GetString("network", "id")
	if err != nil {
		klog.Error("CreateProviderNetwork: GetString network->id error: ", err.Error())
		return "", fmt.Errorf("%v:CreateNetwork: GetString network->id error", err)
	}

	return id, nil
}

func (self *NeutronClient) makeCreateProviderNetworkOpts(name, netType, physNet, segId string) map[string]interface{} {
	dictNetwork := make(map[string]interface{})
	dictNetwork["name"] = name
	dictNetwork["admin_state_up"] = true
	dictNetwork["shared"] = false
	dictNetwork["provider:network_type"] = netType
	dictNetwork["provider:physical_network"] = physNet
	dictNetwork["provider:segmentation_id"] = segId

	dictOpts := make(map[string]interface{})
	dictOpts["network"] = dictNetwork

	return dictOpts
}

func (self *NeutronClient) makeRouterInterfaceOpts(subnetId string) map[string]interface{} {
	dictRouter := make(map[string]interface{})
	dictRouter["subnet_id"] = subnetId

	return dictRouter
}

func (self *NeutronClient) DeleteNetwork(id string) error {
	url := self.getNetworkIdUrl(id)
	klog.Info("DeleteNetwork: url: ", url)
	status, err := doHttpDeleteWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("DeleteNetwork: Delete url[", url, "], status[", status, "], error: ", err)
		return fmt.Errorf("%v:%v:DeleteNetwork: Delete request error", status, err)
	}
	return nil
}

func (self *NeutronClient) getNetworkIdUrl(id string) string {
	return NormalizeURL(self.getNetworksUrl()) + id
}

func (self *NeutronClient) GetNetworkID(networkName string) (string, error) {
	url := self.getNetworksUrl()
	if networkName != "" {
		url += "?name=" + networkName
	}
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("GetNetworkID: Get url[", url, "], status[", status, "], error: ", err)
		return "", fmt.Errorf("%v:%v:GetNetworkID: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetNetworkID: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:GetNetworkID: NewObjectFromBytes parse response body error", err)
	}

	nwArray, err := rspJasObj.GetObjectArray("networks")
	if err != nil {
		klog.Error("GetNetworkID: GetObjectArray error: ", err.Error())
		return "", fmt.Errorf("%v:GetNetworkID: GetObjectArray parse response body array error", err)
	}

	if len(nwArray) == 0 {
		klog.Error("GetNetworkID: network not find")
		return "", fmt.Errorf("GetNetworkID: network %v not find", networkName)
	}

	id, err := nwArray[0].GetString("id")
	if err != nil {
		klog.Error("GetNetworkID: GetString not find key: id")
		return "", fmt.Errorf("%v:GetNetworkID: GetString not find key: id", err)
	}
	return id, nil
}

func (self *NeutronClient) matchNetwork(objs []*jason.Object, name string) (*jason.Object, error) {
	for _, obj := range objs {
		nwName, err := obj.GetString("name")
		if err != nil {
			klog.Error("matchNetwork: jason.GetString: name error: ", err.Error())
			continue
		}

		if nwName == name {
			klog.Trace("matchNetwork: find network [name: ", name, "]")
			return obj, nil
		}
	}

	klog.Error("matchNetwork: not found the network named: ", name, ", error")
	return nil, errors.New("matchNetwork: not found the network")
}

func (self *NeutronClient) GetNetwork(id string) (*Network, error) {
	url := self.getNetworkIdUrl(id)
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || (status > http.StatusMultipleChoices && status != http.StatusNotFound) || err != nil {
		klog.Error("GetNetwork: Get url[", url, "], status[", status, "], error: ", err)
		return nil, fmt.Errorf("%v:%v:GetNetwork: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetNetwork: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetwork: NewObjectFromBytes parse response body error", err)
	}

	netAttrs, err := self.parseNetworkAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetNetwork: parseNetworkAttrs error")
		return nil, fmt.Errorf("%v:GetNetwork: parseNetworkAttrs error", err)
	}
	klog.Info("######-OpenStack GetNetwork-end-###")
	return netAttrs, nil
}

func (self *NeutronClient) parseNetworkAttrs(obj *jason.Object) (*Network, error) {
	id, err := obj.GetString("network", "id")
	if err != nil {
		klog.Error("GetNetwork: GetString network->id error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetwork: GetString network->id error", err)
	}

	name, err := obj.GetString("network", "name")
	if err != nil {
		klog.Error("GetNetwork: GetString network->name error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetwork: GetString network->name error", err)
	}

	return &Network{Id: id, Name: name}, nil
}

func (self *NeutronClient) GetNetworkExtenAttrs(id string) (*NetworkExtenAttrs, error) {
	url := self.getNetworkIdUrl(id)
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("GetNetworkExtenAttrs: Get url[", url, "], status[", status, "], error: ", err)
		return nil, fmt.Errorf("%v:%v:GetNetworkExtenAttrs: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetNetworkExtenAttrs: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetworkExtenAttrs: NewObjectFromBytes parse response body error", err)
	}

	extAttrs, err := self.parseNetworkExtAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetNetworkExtenAttrs: parseNetworkExtAttrs error")
		return nil, fmt.Errorf("%v:GetNetworkExtenAttrs: parseNetworkExtAttrs error", err)
	}

	return extAttrs, nil
}

func (self *NeutronClient) parseNetworkExtAttrs(obj *jason.Object) (*NetworkExtenAttrs, error) {
	id, err := obj.GetString("network", "id")
	if err != nil {
		klog.Error("getNetworkExtAttrs: GetString network->id error: ", err.Error())
		return nil, fmt.Errorf("%v:getNetworkExtAttrs: GetString network->id error", err)
	}

	name, err := obj.GetString("network", "name")
	if err != nil {
		klog.Error("getNetworkExtAttrs: GetString network->name error: ", err.Error())
		return nil, fmt.Errorf("%v:getNetworkExtAttrs: GetString network->name error", err)
	}

	nwType, err := obj.GetString("network", "provider:network_type")
	if err != nil {
		klog.Error("getNetworkExtAttrs: GetString network->provider:network_type error: ", err.Error())
		return nil, fmt.Errorf("%v:getNetworkExtAttrs: GetString network->provider:network_type error", err)
	}

	physNet, err := obj.GetString("network", "provider:physical_network")
	if err != nil {
		klog.Error("getNetworkExtAttrs: GetString network->provider:physical_network error: ", err.Error())
		return nil, fmt.Errorf("%v:getNetworkExtAttrs: GetString network->provider:physical_network error", err)
	}
	var segId json.Number
	if nwType != "flat" {
		segId, err = obj.GetNumber("network", "provider:segmentation_id")
		if err != nil {
			klog.Error("getNetworkExtAttrs: GetString network->provider:segmentation_id error: ", err.Error())
			return nil, fmt.Errorf("%v:getNetworkExtAttrs: GetString network->provider:segmentation_id error", err)
		}
	} else {
		segId = ""
	}
	return &NetworkExtenAttrs{Id: id, Name: name, NetworkType: nwType, PhysicalNetwork: physNet, SegmentationID: string(segId)}, nil
}

func (self *NeutronClient) CreateSubnet(id, cidr, gw string, allocationPools []subnets.AllocationPool) (string, error) {
	url := self.getSubnetsUrl()
	body := self.makeCreateSubnetOpts(id, cidr, gw, allocationPools)
	klog.Info("CreateSubnet: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("CreateBulkPorts: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], error: ", err)
		return "", fmt.Errorf("%v:%v:CreateBulkPorts: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateSubnet: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:CreateSubnet: NewObjectFromBytes parse response body error", err)
	}
	subnetId, err := rspJasObj.GetString("subnet", "id")
	if err != nil {
		klog.Error("CreateSubnet: GetString subnet->id error: ", err.Error())
		return "", fmt.Errorf("%v:CreateSubnet: GetString subnet->id error", err)
	}

	return subnetId, nil
}

func (self *NeutronClient) makeCreateSubnetOpts(id, cidr, gw string, allocationPools []subnets.AllocationPool) map[string]interface{} {
	subnet := make(map[string]interface{})
	subnet["network_id"] = id
	subnet["ip_version"] = 4
	subnet["cidr"] = cidr
	if gw != "" {
		subnet["gateway_ip"] = gw
	}
	if len(allocationPools) != 0 {
		subnet["allocation_pools"] = allocationPools
	}

	return map[string]interface{}{"subnet": subnet}
}

func (self *NeutronClient) getSubnetsUrl() string {
	return getAuthSingleton().NetworkEndpoint + "subnets"
}

func (self *NeutronClient) DeleteSubnet(id string) error {
	url := self.getSubnetIdUrl(id)
	klog.Info("DeleteSubnet: url: ", url)
	status, err := doHttpDeleteWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("DeleteSubnet: Delete url[", url, "], status[", status, "], error: ", err)
		return fmt.Errorf("%v:%v:DeleteSubnet: Delete request error", status, err)
	}
	return nil
}

func (self *NeutronClient) getAddRouterInterfaceUrl(routerId string) string {
	return getAuthSingleton().NetworkEndpoint + "routers/" + routerId + "/add_router_interface"
}

func (self *NeutronClient) getRemoveRouterInterfaceUrl(routerId string) string {
	return getAuthSingleton().NetworkEndpoint + "routers/" + routerId + "/remove_router_interface"
}

func (self *NeutronClient) getSubnetIdUrl(id string) string {
	return NormalizeURL(self.getSubnetsUrl()) + id
}

func (self *NeutronClient) GetSubnetID(networkId string) (string, error) {
	url := self.getNetworkIdUrl(networkId)
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("GetSubnetID: Get url[", url, "], status[", status, "], error: ", err)
		return "", fmt.Errorf("%v:%v:GetSubnetID: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetSubnetID: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:GetSubnetID: NewObjectFromBytes parse response body error", err)
	}

	subnets, err := rspJasObj.GetStringArray("network", "subnets")
	if err != nil {
		klog.Error("GetSubnetID: GetStringArray subnets error: ", err.Error())
		return "", fmt.Errorf("%v:GetSubnetID: GetStringArray subnets error", err)
	}

	if len(subnets) < 1 {
		klog.Error("GetSubnetID: there is no subnet in network[id: ", networkId, "]")
		return "", errors.New("GetSubnetID: there is no subnet in network")
	}

	klog.Trace("GetSubnetID: found subnet[id: ", subnets[0], "] in network[id: ", networkId, "]")
	return subnets[0], nil
}

func (self *NeutronClient) GetSubnet(id string) (*Subnet, error) {
	url := self.getSubnetIdUrl(id)
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("GetSubnet: Get url[", url, "], status[", status, "], error: ", err)
		return nil, fmt.Errorf("%v:%v:GetSubnet: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetSubnet: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetSubnet: NewObjectFromBytes parse response body error", err)
	}

	attrs, err := self.parseSubnetAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetSubnet: parseSubnetAttrs error")
		return nil, fmt.Errorf("%v:GetSubnet: parseSubnetAttrs error", err)
	}

	return attrs, nil
}

func (self *NeutronClient) parseSubnetAttrs(obj *jason.Object) (*Subnet, error) {
	id, err := obj.GetString("subnet", "id")
	if err != nil {
		klog.Error("parseSubnetAttrs: GetString subnet->id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseSubnetAttrs: GetString subnet->id error", err)
	}

	name, err := obj.GetString("subnet", "name")
	if err != nil {
		klog.Error("parseSubnetAttrs: GetString subnet->name error: ", err.Error())
		return nil, fmt.Errorf("%v:parseSubnetAttrs: GetString subnet->name error", err)
	}

	networkId, err := obj.GetString("subnet", "network_id")
	if err != nil {
		klog.Error("parseSubnetAttrs: GetString subnet->network_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseSubnetAttrs: GetString subnet->network_id error", err)
	}

	cidr, err := obj.GetString("subnet", "cidr")
	if err != nil {
		klog.Error("parseSubnetAttrs: GetString subnet->cidr error: ", err.Error())
		return nil, fmt.Errorf("%v:parseSubnetAttrs: GetString subnet->cidr error", err)
	}

	gatewayIp, err := obj.GetString("subnet", "gateway_ip")
	if err != nil {
		klog.Info("parseSubnetAttrs: GetString subnet->gateway_ip error: ", err.Error())
	}

	tenantId, err := obj.GetString("subnet", "tenant_id")
	if err != nil {
		klog.Error("parseSubnetAttrs: GetString subnet->tenant_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseSubnetAttrs: GetString subnet->tenant_id error", err)
	}
	var pools []subnets.AllocationPool
	allocationPools, err := obj.GetObjectArray("subnet", "allocation_pools")
	if err == nil && len(allocationPools) != 0 {
		for _, pool := range allocationPools {
			startIP, _ := pool.GetString("start")
			endIP, _ := pool.GetString("end")
			allocationPool := subnets.AllocationPool{
				Start: startIP,
				End:   endIP,
			}
			pools = append(pools, allocationPool)
		}
	}
	return &Subnet{Id: id, Name: name, NetworkId: networkId, Cidr: cidr,
		GatewayIp: gatewayIp, TenantId: tenantId, AllocationPools: pools}, nil
}

func (self *NeutronClient) ListPorts(networkID string) ([]*Interface, error) {
	url := self.getPortsUrl()
	url += "?network_id=" + networkID
	status, rspBytes, err := doHttpGetWithReAuth(url)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("ListPorts: Get url[", url, "], status[", status, "], error: ", err)
		return nil, fmt.Errorf("%v:%v:ListPorts: Get request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("ListPorts: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:ListPorts: NewObjectFromBytes parse response body error", err)
	}

	ports, err := parsePortsAttrs(rspJasObj)
	if err != nil {
		klog.Error("ListPorts: parseCreatePortAttrs error: ", err.Error())
		return nil, fmt.Errorf("%v:ListPorts: parseCreatePortAttrs error", err)
	}

	return ports, nil
}

func (self *NeutronClient) CreateRouter(name, extNetId string) (string, error) {
	return "", nil
}

func (self *NeutronClient) UpdateRouter(id, name, extNetID string) error {
	return nil
}

func (self *NeutronClient) GetRouter(id string) (*Router, error) {
	return nil, nil
}

func (self *NeutronClient) DeleteRouter(id string) error {
	return nil
}

func (self *NeutronClient) AttachNetToRouter(routerId, subNetId string) (string, error) {
	subnet, err := self.GetSubnet(subNetId)
	if err != nil {
		klog.Error("AttachNetToRouter call GetSubnet Error:", err)
		return "", err
	}

	url := self.getAddRouterInterfaceUrl(routerId)
	body := self.makeRouterInterfaceOpts(subnet.Id)
	klog.Info("AttachNetToRouter: attach router interface url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("AttachNetToRouter: Post url[", url, "], body[", body, "], status[", status, "], error: ", err)
		return "", fmt.Errorf("%v:%v:AttachNetToRouter: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("AttachNetToRouter: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:AttachNetToRouter: NewObjectFromBytes parse response body error", err)
	}

	portId, err := rspJasObj.GetString("port_id")
	if err != nil {
		klog.Error("AttachNetToRouter: GetString error: ", err.Error())
		return "", fmt.Errorf("%v:AttachNetToRouter: GetString error", err)
	}

	klog.Info("AttachNetToRouter Net:[", subNetId, "]to Router:[", routerId, "]OK:", portId)
	return portId, nil
}

func (self *NeutronClient) DetachNetFromRouter(routerId, subNetId string) (string, error) {
	subnet, err := self.GetSubnet(subNetId)
	if err != nil {
		klog.Error("AttachNetToRouter call GetSubnet Error:", err)
		return "", err
	}

	url := self.getRemoveRouterInterfaceUrl(routerId)
	body := self.makeRouterInterfaceOpts(subnet.Id)

	klog.Info("DetachNetFromRouter: url: ", url, " http body: ", body)
	status, rspBytes, err := doHttpPostWithReAuth(url, body)
	if status < http.StatusOK || status > http.StatusMultipleChoices || err != nil {
		klog.Error("DetachNetFromRouter: Post url[", url, "], body[", body, "], status[", status, "], response body[", string(rspBytes), "], error: ", err)
		return "", fmt.Errorf("%v:%v:DetachNetFromRouter: Post request error", status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("DetachNetFromRouter: NewObjectFromBytes error: ", err.Error())
		return "", fmt.Errorf("%v:DetachNetFromRouter: NewObjectFromBytes parse response body error", err)
	}

	portId, err := rspJasObj.GetString("port_id")
	if err != nil {
		klog.Error("DetachNetFromRouter: GetString error: ", err.Error())
		return "", fmt.Errorf("%v:DetachNetFromRouter: GetString error", err)
	}

	klog.Info("DetachNetFromRouter Net:[", subNetId, "]from Router:[", routerId, "]OK:", portId)
	return portId, nil
}

func doHttpPostWithReAuth(url string, body map[string]interface{}) (int, []byte, error) {
	header := make(map[string]string)
	header["X-Auth-Token"] = getAuthSingleton().TokenID
	status, rspBytes, err := adapter.DoHttpPost(url, body, header)
	//reauth
	if status == http.StatusUnauthorized && getAuthSingleton().AllowReauth {
		klog.Warning("doHttpPostWithReAuth, url:[", url, "]")
		getAuthSingleton().auth()
		header["X-Auth-Token"] = getAuthSingleton().TokenID
		status, rspBytes, err = adapter.DoHttpPost(url, body, header)
	}

	return status, rspBytes, err
}

func doHttpGetWithReAuth(url string) (int, []byte, error) {
	header := make(map[string]string)
	header["X-Auth-Token"] = getAuthSingleton().TokenID
	status, rspBytes, err := adapter.DoHttpGet(url, header)
	//reauth
	if status == http.StatusUnauthorized && getAuthSingleton().AllowReauth {
		klog.Warning("doHttpGetWithReAuth, url:[", url, "]")
		getAuthSingleton().auth()
		header["X-Auth-Token"] = getAuthSingleton().TokenID
		status, rspBytes, err = adapter.DoHttpGet(url, header)
	}

	return status, rspBytes, err
}

func doHttpDeleteWithReAuth(url string) (int, error) {
	header := make(map[string]string)
	header["X-Auth-Token"] = getAuthSingleton().TokenID
	status, err := adapter.DoHttpDelete(url, header)
	//reauth
	if status == http.StatusUnauthorized && getAuthSingleton().AllowReauth {
		klog.Warning("doHttpDeleteWithReAuth, url:[", url, "]")
		getAuthSingleton().auth()
		header["X-Auth-Token"] = getAuthSingleton().TokenID
		status, err = adapter.DoHttpDelete(url, header)
	}

	return status, err
}

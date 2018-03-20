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

package noauth_openstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/pkg/adapter"
	"github.com/ZTE/Knitter/pkg/http"
	. "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	NETHTTP "net/http"
	"strconv"
	"strings"
)

const (
	DefaultNoauthOpenStackTenantId = "knitter-blank-openstack-tenant-id"
	DefaultOpenStackIpVersion      = 4
)

type NoauthOpenStackConf struct {
	IP       string `json:"ip"`
	Tenantid string `json:"tenant_id"`
	URL      string `json:"url"`
}

type DefaultProviderConf struct {
	PhyscialNetwork string
	NetworkType     string
}

type NoauthNeutronConf struct {
	Port         string
	ApiVer       string
	ProviderConf DefaultProviderConf
}

type NoauthOpenStack struct {
	NoauthOpenStackConf
	NeutronConf NoauthNeutronConf
}

type Ip struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address,omitempty"`
}

type NoauthNeutronPort struct {
	Id         string `json:"id"`
	NetworkId  string `json:"network_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	MacAddress string `json:"mac_address"`
	DeviceId   string `json:"device_id"`
	Ips        []*Ip  `json:"fixed_ips"`
}

type NoAuthNeutronBulkPortsResp struct {
	Ports []*NoauthNeutronPort `json:"ports"`
}

func NewNoauthOpenStack(conf NoauthOpenStackConf, noauthNeutronConf NoauthNeutronConf) *NoauthOpenStack {
	op := NoauthOpenStack{NoauthOpenStackConf: conf, NeutronConf: noauthNeutronConf}
	return &op
}

func (self *NoauthOpenStack) getUrlCommonSeg() string {
	return self.URL
}

func (self *NoauthOpenStack) getNetworksUrl() string {
	return self.getUrlCommonSeg() + "/networks"
}

func (self *NoauthOpenStack) getNetworkIdUrl(id string) string {
	return self.getNetworksUrl() + "/" + id
}

func (self *NoauthOpenStack) getSubnetsUrl() string {
	return self.getUrlCommonSeg() + "/subnets"
}

func (self *NoauthOpenStack) getSubnetIdUrl(id string) string {
	return self.getSubnetsUrl() + "/" + id
}

func (self *NoauthOpenStack) getPortsUrl() string {
	return self.getUrlCommonSeg() + "/ports"
}

func (self *NoauthOpenStack) getPortIdUrl(id string) string {
	return self.getPortsUrl() + "/" + id
}

func (self *NoauthOpenStack) listPortsUrl(networkID string) string {
	return self.getUrlCommonSeg() + "/ports" + "?network_id=" + networkID
}

func (self *NoauthOpenStack) GetTenantUUID(cfg string) (string, error) {
	return self.Tenantid, nil
}

func (self *NoauthOpenStack) Auth() error {
	return nil
}

func (self *NoauthOpenStack) GetType() string {
	return "vNM"
}

func makeCreatePortOpts(networkId, subnetId, networkPlane, ip, mac, vnicType string) map[string]interface{} {

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
	dictPort["name"] = networkPlane
	dictPort["tenant_id"] = DefaultNoauthOpenStackTenantId
	dictPort["binding:vnic_type"] = vnicType
	if mac != "" {
		dictPort["mac_address"] = mac
	}

	dictOpts := make(map[string]interface{})
	dictOpts["port"] = dictPort

	return dictOpts
}

func makeCreatePort(networkId, subnetId, networkPlane, ip, mac, vnicType string) map[string]interface{} {

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
	dictPort["name"] = networkPlane
	dictPort["tenant_id"] = DefaultNoauthOpenStackTenantId
	dictPort["binding:vnic_type"] = vnicType
	if mac != "" {
		dictPort["mac_address"] = mac
	}

	return dictPort
}

func parsePortAttrs(obj *jason.Object) (*Interface, error) {
	id, err := obj.GetString("id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->id error", err)
	}

	name, err := obj.GetString("name")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->name error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->name error", err)
	}

	status, err := obj.GetString("status")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->status error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->status error", err)
	}

	networkId, err := obj.GetString("network_id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->network_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->network_id error", err)
	}

	mac, err := obj.GetString("mac_address")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->mac_address error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->mac_address error", err)
	}

	deviceId, err := obj.GetString("device_id")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->device_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->device_id error", err)
	}

	fixedIps, err := obj.GetObjectArray("fixed_ips")
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
		klog.Error("parseCreatePortAttrs: GetString port->fixed_ips->subnet_id error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->fixed_ips->subnet_id error", err)
	}

	ip, err := fixedIps[0].GetString("ip_address")
	if err != nil {
		klog.Error("parseCreatePortAttrs: GetString port->fixed_ips->ip_address error: ", err.Error())
		return nil, fmt.Errorf("%v:parseCreatePortAttrs: GetString port->fixed_ips->ip_address error", err)
	}

	return &Interface{Id: id, Name: name, Status: status, NetworkId: networkId, MacAddress: mac, DeviceId: deviceId, SubnetId: subnetId, Ip: ip}, nil
}

func parsePort(obj *jason.Object) (*Interface, error) {
	attrObj, err := obj.GetObject("port")
	if err != nil {
		klog.Error("parsePort: GetObject port error: ", err.Error())
		return nil, fmt.Errorf("%v:parsePort: GetObject port error", err)
	}

	return parsePortAttrs(attrObj)
}

func (self *NoauthOpenStack) CreatePort(networkId, subnetId, networkPlane, ip, mac, vnicType string) (*Interface, error) {
	url := self.getPortsUrl()
	body := makeCreatePortOpts(networkId, subnetId, networkPlane, ip, mac, vnicType)
	rspBytes, err := http.GetHTTPClientObj().Post(url, body)
	if err != nil {
		klog.Error("CreatePort: Post url[", url, "], body[", body, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: Post error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreatePort: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: NewObjectFromBytes parse response body error", err)
	}

	port, err := parsePort(rspJasObj)
	if err != nil {
		klog.Error("CreatePort: parseCreatePortAttrs error: ", err.Error())
		return nil, fmt.Errorf("%v:CreatePort: parseCreatePortAttrs error", err)
	}

	return port, nil
}

func (n *NoauthOpenStack) CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*Interface, error) {
	url := n.getPortsUrl()
	opts := make([]map[string]interface{}, 0)
	for _, reqPort := range req.Ports {
		opt := makeCreatePort(reqPort.NetworkId, reqPort.SubnetId,
			reqPort.PortName, reqPort.FixIP, "", reqPort.VnicType)
		opts = append(opts, opt)
	}
	m := make(map[string]interface{})
	m["ports"] = opts
	klog.Infof("CreateBulkPorts: Create bulk ports [http]: %v", opts)
	rspBytes, err := http.GetHTTPClientObj().Post(url, m)
	if err != nil {
		klog.Errorf("Create bulk ports error[http]: %v", err)
		return nil, err
	}
	resp := NoAuthNeutronBulkPortsResp{}
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

func (self *NoauthOpenStack) GetPort(id string) (*Interface, error) {
	url := self.getPortIdUrl(id)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("GetPort: Get url[", url, "] error: ", err.Error())
		return nil, errors.New("GetPort: Get error")
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetPort: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetPort: NewObjectFromBytes parse response body error", err)
	}

	port, err := parsePort(rspJasObj)
	if err != nil {
		klog.Error("GetPort: parseCreatePortAttrs error: ", err.Error())
		return nil, fmt.Errorf("%v:GetPort: parseCreatePortAttrs error", err)
	}

	return port, nil
}

func (self *NoauthOpenStack) DeletePort(id string) error {
	url := self.getPortIdUrl(id)
	err, statusCode, respBodyStr := http.GetHTTPClientObj().Delete(url)
	if err != nil {
		klog.Errorf("DeletePort: Delete port[id: %v] error: %v", id, err.Error())
		return fmt.Errorf("%v:DeletePort: Delete port error", err)
	}
	if (statusCode < 200 || statusCode >= 300) && statusCode != 404 {
		klog.Errorf("DeletePort: Delete port[id: %v] error, Status Code: %v", id, statusCode)
		return fmt.Errorf("Delete port[id: %v] error[%v], Status Code: %v", id, respBodyStr, statusCode)
	}
	return nil
}

func parsePortArray(arrayObj *jason.Object) ([]*Interface, error) {
	portObjs, err := arrayObj.GetObjectArray("ports")
	if err != nil {
		klog.Errorf("parsePortArray: GetObjectArray error: %v", err)
		return nil, err
	}

	ports := make([]*Interface, 0)
	for _, portObj := range portObjs {
		port, err := parsePortAttrs(portObj)
		if err != nil {
			klog.Errorf("parsePortArray: GetObjectArray error: %v", err)
			return nil, err
		}

		klog.Tracef("parse port: %v", port)
		ports = append(ports, port)
	}
	klog.Tracef("parsePortArray: all ports: %v, len: %d ", ports, len(ports))
	return ports, nil
}

func (self *NoauthOpenStack) ListPorts(networkID string) ([]*Interface, error) {
	url := self.listPortsUrl(networkID)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("ListPorts: Get url[", url, "] error: ", err.Error())
		return nil, errors.New("ListPorts: Get error")
	}

	jasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("ListPorts: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:ListPorts: NewObjectFromBytes parse response body error", err)
	}

	ports, err := parsePortArray(jasObj)
	if err != nil {
		klog.Error("ListPorts: parsePortArray error: ", err.Error())
		return nil, fmt.Errorf("%v:ListPorts: parsePortArray error", err)
	}

	return ports, nil
}

func makeCreateNetworkOpts(name, netType, physNet string) map[string]interface{} {
	dict := make(map[string]interface{})
	dict["name"] = name
	dict["admin_state_up"] = true
	dict["shared"] = false
	dict["tenant_id"] = DefaultNoauthOpenStackTenantId
	//dict["provider:network_type"] = netType
	if physNet != "" {
		dict["provider:physical_network"] = physNet
		dict["provider:network_type"] = netType
	}

	return map[string]interface{}{"network": dict}
}

func (self *NoauthOpenStack) CreateNetwork(name string) (*Network, error) {
	url := self.getNetworksUrl()
	body := makeCreateNetworkOpts(name,
		self.NeutronConf.ProviderConf.NetworkType,
		self.NeutronConf.ProviderConf.PhyscialNetwork)
	klog.Info("CreateNetwork: create network http body: ", body)
	rspBytes, err := http.GetHTTPClientObj().Post(url, body)
	if err != nil {
		klog.Error("CreateNetwork: Post url[", url, "], body[", body, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateNetwork: Post request error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateNetwork: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateNetwork: NewObjectFromBytes parse response body error", err)
	}
	id, err := rspJasObj.GetString("network", "id")
	if err != nil {
		klog.Error("CreateNetwork: GetString network->id error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateNetwork: GetString network->id error", err)
	}

	return &Network{Name: name, Id: id}, nil
}

func makeCreateProviderNetworkOpts(name, netType, physNet, segId string, vlanTransparent bool) map[string]interface{} {
	dictNetwork := make(map[string]interface{})
	dictNetwork["name"] = name
	dictNetwork["admin_state_up"] = true
	dictNetwork["shared"] = false
	dictNetwork["tenant_id"] = DefaultNoauthOpenStackTenantId
	if netType != "" {
		dictNetwork["provider:network_type"] = netType
	}
	if physNet != "" {
		dictNetwork["provider:physical_network"] = physNet
	}
	if vlanTransparent {
		dictNetwork["vlan_transparent"] = vlanTransparent
	}
	segmentID, err := strconv.Atoi(segId)
	if err == nil {
		dictNetwork["provider:segmentation_id"] = segmentID
	}

	dictOpts := make(map[string]interface{})
	dictOpts["network"] = dictNetwork

	return dictOpts
}

func (self *NoauthOpenStack) CreateProviderNetwork(name, netType, physNet, segId string,
	vlanTransparent bool) (*Network, error) {
	url := self.getNetworksUrl()
	body := makeCreateProviderNetworkOpts(name, netType, physNet, segId, vlanTransparent)
	klog.Info("CreateProviderNetwork: create provider network http body: ", body)
	rspBytes, err := http.GetHTTPClientObj().Post(url, body)
	if err != nil {
		klog.Error("CreateProviderNetwork: Post url[", url, "], body[", body, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateProviderNetwork: Post error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateProviderNetwork: NewObjectFromBytes[", rspBytes, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateProviderNetwork: NewObjectFromBytes error", err)
	}

	id, err := rspJasObj.GetString("network", "id")
	if err != nil {
		klog.Error("CreateProviderNetwork: GetString network->id error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateProviderNetwork: GetString error", err)
	}

	return &Network{Name: name, Id: id}, nil
}

func (self *NoauthOpenStack) DeleteNetwork(id string) error {
	klog.Infof("START noauth DeleteNetwork[%v]", id)
	url := self.getNetworkIdUrl(id)
	err, respCode, respBodyStr := http.GetHTTPClientObj().Delete(url)
	klog.Infof("http DeleteNetwork: id[%v] code[%v] error[%v]", id, respCode, err)
	if err != nil {
		klog.Errorf("http DeleteNetwork: Delete network[%v] error: %v", id, err)
		return fmt.Errorf("%v:Delete network error", err)
	}

	if respCode >= NETHTTP.StatusBadRequest && respCode != NETHTTP.StatusNotFound {
		klog.Errorf("http DeleteNetwork: Delete network[%v] respCode: %v", id, respCode)
		return fmt.Errorf("Return code %v:Delete network error[%v]", respCode, respBodyStr)
	}
	klog.Infof("END noauth DeleteNetwork[%v]", id)
	return nil
}

func matchNetwork(opn *NoauthOpenStack, objs []*jason.Object, name string) (*jason.Object, error) {
	for _, obj := range objs {
		nwname, err := obj.GetString("name")
		if err != nil {
			klog.Error("matchNetwork: jason.GetString: name error: ", err.Error())
			continue
		}

		if nwname == name {
			klog.Trace("matchNetwork: find network [name: ", name, "]")
			return obj, nil
		}
	}

	klog.Error("matchNetwork: not found the network named: ", name, ", error")
	return nil, errors.New("matchNetwork: not found the network")
}

func (self *NoauthOpenStack) GetNetworkID(name string) (string, error) {
	url := self.getNetworksUrl()
	klog.Info("GetNetworkID: get network uuid")
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("GetNetworkID: Get url[", url, "] error: ", err.Error())
		return "", fmt.Errorf("%v:GetNetworkID: Get request error", err)
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

	nw, err := matchNetwork(self, nwArray, name)
	if err != nil {
		klog.Error("GetNetworkID: findNetwork not find the network named: ", name)
		return "", fmt.Errorf("%v:GetNetworkID: findNetwork not find the network", err)
	}

	id, err := nw.GetString("id")
	if err != nil {
		klog.Error("GetNetworkID: GetString not find key: id")
		return "", fmt.Errorf("%v:GetNetworkID: GetString not find key: id", err)
	}

	return id, nil
}

func parseNetworkAttrs(obj *jason.Object) (*Network, error) {
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

func (self *NoauthOpenStack) GetNetwork(id string) (*Network, error) {
	klog.Info("######-NoauthOpenStack GetNetwork-begin###")
	url := self.getNetworkIdUrl(id)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Warningf("GetNetwork: Get url[", url, "] error: ", err.Error())
		if strings.Contains(err.Error(), "404") { //todo
			return nil, err
		}
		return nil, fmt.Errorf("%v:GetNetwork: socket-error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetNetwork: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetwork: NewObjectFromBytes parse response body error", err)
	}

	netAttrs, err := parseNetworkAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetNetwork: parseNetworkAttrs error")
		return nil, fmt.Errorf("%v:GetNetwork: parseNetworkAttrs error", err)
	}
	klog.Info("######-NoauthOpenStack GetNetwork-end-###")
	return netAttrs, nil
}

func parseNetworkExtAttrs(obj *jason.Object) (*NetworkExtenAttrs, error) {
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
	vlanTransparent, err := obj.GetBoolean("network", "vlan_transparent")
	if err != nil {
		klog.Warningf("getNetworkExtAttrs: GetString network->vlan_transparent fail: %v, use default false ", err)
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
	extAttrs := &NetworkExtenAttrs{
		Id:              id,
		Name:            name,
		NetworkType:     nwType,
		PhysicalNetwork: physNet,
		SegmentationID:  string(segId),
		VlanTransparent: vlanTransparent}
	return extAttrs, nil
}

func (self *NoauthOpenStack) GetNetworkExtenAttrs(id string) (*NetworkExtenAttrs, error) {
	url := self.getNetworkIdUrl(id)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("GetNetworkExtenAttrs: Get url[", url, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetworkExtenAttrs: Get error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetNetworkExtenAttrs: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetNetworkExtenAttrs: NewObjectFromBytes parse response body error", err)
	}

	extAttrs, err := parseNetworkExtAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetNetworkExtenAttrs: parseNetworkExtAttrs error")
		return nil, fmt.Errorf("%v:GetNetworkExtenAttrs: parseNetworkExtAttrs error", err)
	}

	return extAttrs, nil
}

func makeCreateSubnetOpts(id, cidr, gw string, allocationPools []subnets.AllocationPool) map[string]interface{} {
	subnet := make(map[string]interface{})
	subnet["network_id"] = id
	subnet["ip_version"] = DefaultOpenStackIpVersion
	subnet["cidr"] = cidr
	if gw != "" {
		subnet["gateway_ip"] = gw
	}
	subnet["tenant_id"] = DefaultNoauthOpenStackTenantId
	if len(allocationPools) != 0 {
		subnet["allocation_pools"] = allocationPools
	}
	//subnet["admin_state_up"] = true
	//subnet["shared"] = false

	return map[string]interface{}{"subnet": subnet}
}

func (self *NoauthOpenStack) CreateSubnet(id, cidr, gw string, allocationPools []subnets.AllocationPool) (*Subnet, error) {
	url := self.getSubnetsUrl()
	body := makeCreateSubnetOpts(id, cidr, gw, allocationPools)
	klog.Info("CreateSubnet: create subnet http body: ", body)
	rspBytes, err := http.GetHTTPClientObj().Post(url, body)
	if err != nil {
		klog.Error("CreateSubnet: Post url[", url, "] body[", body, "], error: ", err.Error(), ", response content is: ", string(rspBytes))
		return nil, fmt.Errorf("%v:CreateSubnet: Post error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("CreateSubnet: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:NewObjectFromBytes parse response body error", err)
	}

	subnet, err := parseSubnetAttrs(rspJasObj)
	if err != nil {
		klog.Error("CreateSubnet: GetString subnet->id error: ", err.Error())
		return nil, fmt.Errorf("%v:CreateSubnet: GetString subnet->id error", err)
	}

	return subnet, nil
}

func (self *NoauthOpenStack) DeleteSubnet(id string) error {
	url := self.getSubnetIdUrl(id)
	err, statusCode, respBodyStr := http.GetHTTPClientObj().Delete(url)
	if err != nil {
		beego.Error("DeleteSubnet: Delete subnet[id: ", id, "] error: ", err.Error())
		return fmt.Errorf("%v:Delete subnet error", err)
	}
	if (statusCode >= 300 || statusCode < 200) && statusCode != 404 {
		beego.Error("DeleteSubnet: Delete subNet[id: ", id, "] code: ", statusCode)
		return fmt.Errorf("Return code %v:Delete subNet error[%v]", statusCode, respBodyStr)
	}

	return nil
}

func (self *NoauthOpenStack) GetSubnetID(networkId string) (string, error) {
	url := self.getNetworkIdUrl(networkId)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("GetSubnetID: Get url[", url, "] error: ", err.Error())
		return "", fmt.Errorf("%v:GetSubnetID: Get error", err)
	}

	klog.Error("GetSubnetID: rspBytes is : ", string(rspBytes))

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

func parseSubnetAttrs(obj *jason.Object) (*Subnet, error) {
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

func (self *NoauthOpenStack) GetSubnet(id string) (*Subnet, error) {
	url := self.getSubnetIdUrl(id)
	rspBytes, err := http.GetHTTPClientObj().Get(url)
	if err != nil {
		klog.Error("GetSubnet: http Get[", url, "] error: ", err.Error())
		return nil, fmt.Errorf("%v:GetSubnet: http Get error", err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("GetSubnet: NewObjectFromBytes error: ", err.Error())
		return nil, fmt.Errorf("%v:GetSubnet: NewObjectFromBytes parse response body error", err)
	}

	attrs, err := parseSubnetAttrs(rspJasObj)
	if err != nil {
		klog.Error("GetSubnet: parseSubnetAttrs error")
		return nil, fmt.Errorf("%v:GetSubnet: parseSubnetAttrs error", err)
	}

	return attrs, nil
}

func (self *NoauthOpenStack) CreateRouter(name, extNetId string) (string, error) {
	return "", errors.New("Noauth-OpenStack unsupported operation: CreateRouter")
}

func (self *NoauthOpenStack) UpdateRouter(id, name, extNetID string) error {
	return errors.New("Noauth-OpenStack unsupported operation: UpdateRouter")
}

func (self *NoauthOpenStack) GetRouter(id string) (*Router, error) {
	return nil, errors.New("Noauth-OpenStack unsupported operation: GetRouter")
}

func (self *NoauthOpenStack) DeleteRouter(id string) error {
	return errors.New("Noauth-OpenStack unsupported operation: DeleteRouter")
}

func (self *NoauthOpenStack) AttachPortToVM(vmId, portId string) (*Interface, error) {
	return nil, errors.New("Noauth-OpenStack unsupported operation: AttachPortToVM")
}

func (self *NoauthOpenStack) DetachPortFromVM(vmId, portId string) error {
	return errors.New("Noauth-OpenStack unsupported operation: DetachPortFromVM")
}

func (self *NoauthOpenStack) AttachNetToRouter(routerId, subNetId string) (string, error) {
	return "", errors.New("Noauth-OpenStack unsupported operation: AttachNetToRouter")
}

func (self *NoauthOpenStack) DetachNetFromRouter(routerId, netId string) (string, error) {
	return "", errors.New("Noauth-OpenStack unsupported operation: DetachNetFromRouter")
}
func (self *NoauthOpenStack) GetAttachReq() int {
	return 0
}
func (self *NoauthOpenStack) SetAttachReq(req int) {
}

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

package openstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/adapter"
	. "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/provider"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/rackspace/gophercloud/pagination"
	"regexp"
	"strings"
	"sync"
)

const MaxReqForAttach int = 5

var InitLock sync.Mutex

type OpenStack struct {
	neutronClient *gophercloud.ServiceClient
	novaClient    *gophercloud.ServiceClient
	provider      *gophercloud.ProviderClient
	config        gophercloud.AuthOptions
	VmLock        map[string]*sync.Mutex
	Channel       chan int
	AttachReq     int
}

type OpenStackConf struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Url        string `json:"url"`
	Tenantid   string `json:"tenantid"`
	TenantName string `json:"tenantname"`
}

func NewOpenstack() *OpenStack {
	op := OpenStack{AttachReq: MaxReqForAttach}
	return &op
}

func (self *OpenStack) isUuid(str string) bool {
	const UUID string = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	rxUUID := regexp.MustCompile(UUID)
	return rxUUID.MatchString(str)
}

func (self *OpenStack) GetTenantUUID(cfgStr string) (string, error) {
	config := OpenStackConf{}
	err := json.Unmarshal([]byte(cfgStr), &config)
	if err != nil {
		return "", err
	}
	return config.Tenantid, nil
}

func (self *OpenStack) SetOpenstackConfig(cfgStr string) error {
	config := OpenStackConf{}
	err := json.Unmarshal([]byte(cfgStr), &config)
	if err != nil {
		self.config = gophercloud.AuthOptions{}
		return err
	}
	conf := gophercloud.AuthOptions{
		IdentityEndpoint: config.Url + "/tokens",
		Username:         config.Username,
		Password:         config.Password,
		TenantID:         config.Tenantid,
		AllowReauth:      true,
	}
	self.config = conf
	return nil
}

func (self *OpenStack) SetConfig(conf gophercloud.AuthOptions) {
	self.config = conf
}

func (self *OpenStack) GetType() string {
	return "TECS"
}

func (self *OpenStack) Auth() error {
	err := self.setProvider()
	if err != nil {
		klog.Error("auth call setProvider", err)
		return err
	}
	err = self.setNeutronClient()
	if err != nil {
		klog.Error("auth call setNeutronClient", err)
		return err
	}
	err = self.setNovaClient()
	if err != nil {
		klog.Error("auth call setNovaClient", err)
		return err
	}
	klog.Info("auth OK")
	return nil
}

func AuthenticatedClientV2(options gophercloud.AuthOptions) (*gophercloud.ProviderClient, error) {
	client, err := openstack.NewClient(options.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	err = openstack.AuthenticateV2(client, options)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (self *OpenStack) setProvider() error {
	provider, err := AuthenticatedClientV2(self.config)
	if err != nil {
		klog.Error("AuthenticatedClientV2 ERROR:", err.Error())
		return err
	}
	self.provider = provider
	return nil
}

func (self *OpenStack) setNeutronClient() error {
	client, err := openstack.NewNetworkV2(self.provider,
		gophercloud.EndpointOpts{Name: "neutron"})
	if err != nil {
		klog.Error("setNeutronClient failed!", err)
		return err
	}
	self.neutronClient = client
	return nil
}

func (self *OpenStack) setNovaClient() error {
	client, err := openstack.NewComputeV2(self.provider,
		gophercloud.EndpointOpts{Name: "nova"})
	if err != nil {
		klog.Error("setNovaClient failed!", err)
		return err
	}
	self.novaClient = client
	return nil
}

func (self *OpenStack) getPortIpAddr(ips []ports.IP) string {
	for _, ip := range ips {
		return ip.IPAddress
	}
	return ""
}

func (self *OpenStack) makeCreatePortOps(networkId, subnetId, portName, ip, mac, vnicType string) (*ports.CreateOpts, error) {
	ops := ports.CreateOpts{
		NetworkID:    networkId,
		Name:         portName,
		AdminStateUp: ports.Up,
		FixedIPs:     []ports.IP{{SubnetID: subnetId, IPAddress: ip}},
		MACAddress:   mac,
		VnicType:     vnicType,
	}
	return &ops, nil
}

func (self *OpenStack) CreatePort(networkId, subnetId, portName, ip, mac, vnicType string) (*Interface, error) {
	ops, err := self.makeCreatePortOps(networkId, subnetId, portName, ip, mac, vnicType)
	if err != nil {
		klog.Error("CreatePort call makeCreatePortOps error :", err)
		return nil, err
	}
	newport, err := ports.Create(self.neutronClient, ops).Extract()
	if err != nil {
		klog.Error("CreatePort call Create error :", err)
		return nil, fmt.Errorf("%v:CreatePort:ports.Create error", err)
	}
	port := Interface{Id: newport.ID, Name: newport.Name, Status: newport.Status,
		Ip: newport.FixedIPs[0].IPAddress, MacAddress: newport.MACAddress,
		NetworkId: newport.NetworkID, DeviceId: newport.DeviceID,
		SubnetId: newport.FixedIPs[0].SubnetID}
	klog.Info("CreatePort OK :", port)
	return &port, nil
}

func (o *OpenStack) CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*Interface, error) {
	opts := ports.BulkPorts{Opts: make([]*ports.CreateOpts, 0)}
	for _, reqPort := range req.Ports {
		opt, _ := o.makeCreatePortOps(reqPort.NetworkId, reqPort.SubnetId, reqPort.PortName,
			reqPort.FixIP, "", reqPort.VnicType)
		opts.Opts = append(opts.Opts, opt)
	}

	newPorts, err := adapter.CreateBulkPorts(o.neutronClient, opts)
	if err != nil {
		klog.Errorf("CreateBulkPorts error : %v", err)
		return nil, err
	}
	ports := make([]*Interface, 0)
	for _, newPort := range newPorts {
		inter := &Interface{
			Id: newPort.ID, Name: newPort.Name, Status: newPort.Status,
			Ip: newPort.FixedIPs[0].IPAddress, MacAddress: newPort.MACAddress,
			NetworkId: newPort.NetworkID, DeviceId: newPort.DeviceID,
			SubnetId: newPort.FixedIPs[0].SubnetID,
		}
		ports = append(ports, inter)
	}
	return ports, nil
}

func (self *OpenStack) GetPort(id string) (*Interface, error) {
	pt, err := ports.Get(self.neutronClient, id).Extract()
	if err != nil {
		klog.Error("GetPort call Get error :", err)
		return nil, err
	}
	klog.Infof("GetPort: port[id: %s] info details: %v", id, pt)
	rp := Interface{Name: pt.Name, Status: pt.Status, Id: pt.ID,
		Ip: self.getPortIpAddr(pt.FixedIPs), MacAddress: pt.MACAddress,
		NetworkId: pt.NetworkID, DeviceId: pt.DeviceID, SubnetId: pt.FixedIPs[0].SubnetID}
	klog.Info("GetPort OK :", rp)
	return &rp, nil
}

func (self *OpenStack) DeletePort(portId string) error {
	res := ports.Delete(self.neutronClient, portId)
	if res.Err != nil && !strings.Contains(res.Err.Error(), "got 404 instead") {
		klog.Error("DeletePort call Delete error :", res.Err)
		return res.Err
	}
	klog.Info("DeletePort OK:", res)
	return nil
}

func (self *OpenStack) ListPorts(networkID string) ([]*Interface, error) {
	pager := ports.List(self.neutronClient, ports.ListOpts{NetworkID: networkID})
	err := pager.Err
	if err != nil {
		klog.Errorf("ListPorts call ports.List for networkID: %s error: %v", networkID, err)
		return nil, err
	}

	portPage, err := pager.AllPages()
	if err != nil {
		klog.Errorf("ListPorts call pager.AllPages for networkID: %s error: %v", networkID, err)
		return nil, err
	}

	ports, err := ports.ExtractPorts(portPage)
	if err != nil {
		klog.Errorf("ListPorts call ports.ExtractPorts for networkID: %s, portPage: %v error: %v",
			networkID, portPage, err)
		return nil, err
	}

	nifs := make([]*Interface, len(ports))
	for idx, port := range ports {
		nif := Interface{}
		nif.Id = port.ID
		nif.DeviceId = port.DeviceID
		nif.NetworkId = port.NetworkID
		nifs[idx] = &nif
	}
	klog.Tracef("ListPorts result: %v SUCC", nifs)
	return nifs, nil
}

func (self *OpenStack) CreateNetwork(name string) (*Network, error) {
	iTrue := true
	iFalse := false
	opts4net := networks.CreateOpts{Name: name,
		AdminStateUp: &iTrue, Shared: &iFalse}
	klog.Info(self, "----------", opts4net)
	rsp4net, err := networks.Create(self.neutronClient, opts4net).Extract()
	if err != nil {
		klog.Error("CreateNetwork call Create error :", err)
		return nil, err
	}
	klog.Info("CreateNetwork OK:", rsp4net)

	return &Network{Name: rsp4net.Name, Id: rsp4net.ID}, nil
}

func (self *OpenStack) CreateProviderNetwork(name, nwType, phyNet, sId string, vlanTransparent bool) (*Network, error) {
	iTrue := true
	iFalse := false
	opts4net := provider.CreateNetOpts{Name: name,
		AdminStateUp: &iTrue, Shared: &iFalse,
		NetworkType: nwType, PhysicalNetwork: phyNet, SegmentationID: sId}
	klog.Info(self, "----------", opts4net)
	rsp4net, err := networks.Create(self.neutronClient, opts4net).Extract()
	if err != nil {
		klog.Error("CreateNetwork call Create error :", err)
		return nil, err
	}
	klog.Info("CreateNetwork OK:", rsp4net)
	return &Network{Name: rsp4net.Name, Id: rsp4net.ID}, nil
}

func (self *OpenStack) GetNetwork(id string) (*Network, error) {
	net, err := networks.Get(self.neutronClient, id).Extract()
	if err != nil {
		klog.Error("GetNetwork call Get error :", err)
		if net == nil {
			if strings.Contains(err.Error(), "but got 404 instead") && id == "this-is-a-error-uuid-for-auth" {
				return nil, nil
			}
			return nil, fmt.Errorf("%v:GetNetwork: socket-error", err)
		}
		return nil, err
	}
	klog.Info("GetNetwork OK :", net)
	network := Network{Name: net.Name, Id: net.ID}
	return &network, nil
}

func (self *OpenStack) GetNetworkExtenAttrs(id string) (*NetworkExtenAttrs, error) {
	getResult := networks.Get(self.neutronClient, id)
	if getResult.Err != nil {
		klog.Errorf("GetNetworkExtAttrs call Get error: %v", getResult.Err)
		return nil, getResult.Err
	}

	networkExt, err := provider.ExtractGet(getResult)
	if err != nil {
		klog.Errorf("GetNetworkExtAttrs call ExtractGet error: %v", err)
		return nil, err
	}

	klog.Info("GetNetworkExtAttrs OK :", networkExt)
	netExtRsp := NetworkExtenAttrs{Name: networkExt.Name, Id: networkExt.ID, NetworkType: networkExt.NetworkType,
		PhysicalNetwork: networkExt.PhysicalNetwork, SegmentationID: networkExt.SegmentationID}

	return &netExtRsp, nil
}

func (self *OpenStack) DeleteNetwork(id string) error {
	rsp := networks.Delete(self.neutronClient, id)
	klog.Info("DeleteNetwork:", rsp)
	return rsp.Err
}

func (self *OpenStack) CreateSubnet(id, cidr, gw string, allocationPools []subnets.AllocationPool) (*Subnet, error) {
	opts4sub := subnets.CreateOpts{
		NetworkID: id,
		IPVersion: 4,
		CIDR:      cidr,
		GatewayIP: gw,
	}
	if len(allocationPools) != 0 {
		opts4sub.AllocationPools = allocationPools
	}

	rsp4sub, err := subnets.Create(self.neutronClient, opts4sub).Extract()
	if err != nil {
		klog.Error("CreateSubnet call Create error :", err)
		return nil, err
	}
	subNet := Subnet{
		Id:              rsp4sub.ID,
		Name:            rsp4sub.Name,
		NetworkId:       rsp4sub.NetworkID,
		Cidr:            rsp4sub.CIDR,
		GatewayIp:       rsp4sub.GatewayIP,
		TenantId:        rsp4sub.TenantID,
		AllocationPools: rsp4sub.AllocationPools}
	klog.Tracef("CreateSubnet OK:", rsp4sub)
	return &subNet, nil
}

func (self *OpenStack) GetSubnet(id string) (*Subnet, error) {
	sub, err := subnets.Get(self.neutronClient, id).Extract()
	if err != nil {
		klog.Warning("GetSubnet call Get error!", err)
		return nil, err
	}
	subNet := Subnet{Id: sub.ID, Name: sub.Name, NetworkId: sub.NetworkID,
		Cidr: sub.CIDR, GatewayIp: sub.GatewayIP, TenantId: sub.TenantID,
		AllocationPools: sub.AllocationPools}
	klog.Info("GetSubnet OK:", subNet)
	return &subNet, nil
}

func (self *OpenStack) DeleteSubnet(id string) error {
	rsp := subnets.Delete(self.neutronClient, id)
	klog.Info("DeleteSubnet:", rsp)
	return rsp.Err
}

func (self *OpenStack) makeCreateRouterOps(name, extNetId string) (*routers.CreateOpts, error) {
	var options routers.CreateOpts
	if self.isUuid(extNetId) {
		gwi := routers.GatewayInfo{NetworkID: extNetId}
		options = routers.CreateOpts{GatewayInfo: &gwi}
	}
	options.Name = name
	asu := true
	options.AdminStateUp = &asu
	return &options, nil
}

func (self *OpenStack) CreateRouter(name, extNetId string) (string, error) {
	options, _ := self.makeCreateRouterOps(name, extNetId)
	rsp, err := routers.Create(self.neutronClient, *options).Extract()
	if err != nil {
		klog.Error("CreateRouter call Create error :", err)
		return "", err
	}
	klog.Info("CreateRouter OK:", rsp)
	return rsp.ID, nil
}

func (self *OpenStack) makeUpdateRouterOps(name, extNetId string) (*routers.UpdateOpts, error) {
	var options routers.UpdateOpts
	if self.isUuid(extNetId) {
		gwi := routers.GatewayInfo{NetworkID: extNetId}
		options = routers.UpdateOpts{GatewayInfo: &gwi}
	}
	options.Name = name
	asu := true
	options.AdminStateUp = &asu
	return &options, nil
}

func (self *OpenStack) UpdateRouter(routerID, name, extNetId string) error {
	options, _ := self.makeUpdateRouterOps(name, extNetId)
	rsp, err := routers.Update(self.neutronClient, routerID, *options).Extract()
	if err != nil {
		klog.Error("UpdateRouter call Update error :", err)
		return err
	}
	klog.Info("UpdateRouter OK:", rsp)
	return nil
}

func (self *OpenStack) GetRouter(id string) (*Router, error) {
	rsp, err := routers.Get(self.neutronClient, id).Extract()
	if err != nil {
		klog.Error("GetRouter call Get error :", err)
		return nil, err
	}
	rt := Router{Name: rsp.Name, Id: rsp.ID, ExtNetId: rsp.GatewayInfo.NetworkID}
	klog.Info("GetRouter OK:", rt)
	return &rt, nil
}

func (self *OpenStack) DeleteRouter(id string) error {
	err := routers.Delete(self.neutronClient, id).ExtractErr()
	if err != nil {
		klog.Error("DeleteRouter call Delete Error :", err)
		return err
	}
	klog.Info("DeleteRouter OK:", id)
	return nil
}

func (self *OpenStack) VmLockInit(vmId string) {
	InitLock.Lock()
	defer InitLock.Unlock()

	if self.VmLock == nil {
		self.VmLock = make(map[string]*sync.Mutex)
	}
	if self.VmLock[vmId] == nil {
		self.VmLock[vmId] = &sync.Mutex{}
	}
	if self.Channel == nil {
		klog.Infof("make channel size:%d", self.AttachReq)
		self.Channel = make(chan int, self.AttachReq)
	}
}

func (self *OpenStack) Lock(vmId string) {
	self.VmLockInit(vmId)
	self.VmLock[vmId].Lock()
	self.Channel <- 0
}

func (self *OpenStack) Unlock(vmId string) {
	<-self.Channel
	self.VmLock[vmId].Unlock()
}

func (self *OpenStack) AttachPortToVM(vmId, portId string) (*Interface, error) {
	ops := servers.AttachOpts{PortID: portId}
	if self.Channel == nil {
		klog.Infof("make channel size:%d", self.AttachReq)
		self.Channel = make(chan int, self.AttachReq)
	}
	self.Channel <- 0
	attachResult := servers.Attach(self.novaClient, vmId, ops)
	<-self.Channel
	if attachResult.Err != nil {
		klog.Errorf("AttachPortToVM call Attach port[id: %s] failed, error: %v!", portId, attachResult.Err)
		return nil, fmt.Errorf("%v:AttachPort:servers.Attach error", attachResult.Err)
	}

	klog.Errorf("AttachPortToVM: port[id: %s] attach result finish, http body is : %s", portId, attachResult.PrettyPrintJSON())

	IntfAtt, err := attachResult.ExtractInterface()
	if err != nil {
		klog.Errorf("AttachPortToVM: attachResult.ExtractInterface port[id: %s] failed, error: %v", portId, err)
	}
	klog.Infof("AttachPortToVM: port[id: %s] succeed, attach result: %v", portId, IntfAtt)
	intf := Interface{Id: IntfAtt.PortID, Status: IntfAtt.PortState,
		MacAddress: IntfAtt.MacAddr, NetworkId: IntfAtt.NetID}
	return &intf, nil
}

func (self *OpenStack) DetachPortFromVM(vmId, portId string) error {
	if self.Channel == nil {
		klog.Infof("make channel size:%d", self.AttachReq)
		self.Channel = make(chan int, self.AttachReq)
	}
	self.Channel <- 0
	result := servers.Detach(self.novaClient, vmId, portId)
	<-self.Channel
	if result.Err != nil {
		klog.Error("DetachPortFromVM call Detach error!", result.Err)
		return result.Err
	}
	klog.Info("DetachPortFromVM OK!", "VM[", vmId, "]PORT[", portId, "].")
	return nil
}

func (self *OpenStack) AttachNetToRouter(routerId, subNetId string) (string, error) {
	subnet, err := self.GetSubnet(subNetId)
	if err != nil {
		klog.Error("AttachNetToRouter call GetSubnet Error:", err)
		return "", err
	}
	opts := routers.InterfaceOpts{SubnetID: subnet.Id}
	res, err := routers.AddInterface(self.neutronClient, routerId, opts).Extract()
	if err != nil {
		klog.Error("AttachNetToRouter Net:[", subNetId, "]to Router:[", routerId, "]Error:", err)
		return "", err
	}
	klog.Info("AttachNetToRouter Net:[", subNetId, "]to Router:[", routerId, "]OK:", res)
	return res.PortID, nil
}

func (self *OpenStack) DetachNetFromRouter(routerId, subNetId string) (string, error) {
	subnet, err := self.GetSubnet(subNetId)
	if err != nil {
		klog.Error("DetachNetFromRouter call GetSubnet Error:", err)
		return "", err
	}
	opts := routers.InterfaceOpts{SubnetID: subnet.Id}
	res, err := routers.RemoveInterface(self.neutronClient, routerId, opts).Extract()
	if err != nil {
		klog.Error("DetachNetFromRouter Net:[", subNetId, "]to Router:[", routerId, "]Error:", err)
		return "", err
	}
	klog.Info("DetachNetFromRouter Net:[", subNetId, "] from Router:[", routerId, "] OK:", res)
	return res.PortID, nil
}

func (self *OpenStack) GetNetworkID(networkName string) (string, error) {
	listops := networks.ListOpts{Name: networkName}
	var networkId = ""
	results := networks.List(self.neutronClient, listops)
	klog.Infof("GetNetworkID: neutron network list result Headers is: %v", results.Headers)
	err := results.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			klog.Errorf("GetNetworkIDBy:ExtractNetworks pages error! -%v", err)
			return false, fmt.Errorf("%v:GetNetworkIDBy:ExtractNetworks pages error!", err)
		}
		//networkId = networkList[0].ID

		for _, net := range networkList {
			networkId = net.ID
			klog.Infof("GetNetworkID: get network id: %s", networkId)
			return true, nil
		}
		return false, errors.New("GetNetworkID: network not found")
	})
	if err != nil {
		klog.Errorf("GetNetworkID: get all pages error: %v", err)
		return "", fmt.Errorf("%v:GetNetworkID: get all pages error", err)
	}
	return networkId, nil
}

func (self *OpenStack) GetSubnetID(networkId string) (string, error) {
	network, err := networks.Get(self.neutronClient, networkId).Extract()
	if err != nil {
		klog.Errorf("GetSubnetID: get network[id: %s] subnetId error: %v", networkId, err)
		return "", fmt.Errorf("%v:GetSubnetID: get network error", err)
	}

	for _, subnetId := range network.Subnets {
		klog.Infof("GetSubnetID: get subnet id: %s", subnetId)
		return subnetId, nil
	}
	klog.Errorf("GetSubnetID: network[id: %s] get subnet failed, subnet not found", networkId)
	return "", errors.New("GetSubnetID: network[id: %s] get subnet failed, subnet not found")
}

func (self *OpenStack) GetAttachReq() int {
	klog.Infof("OpenStack.GetAttachReq:%d", self.AttachReq)
	return self.AttachReq
}

func (self *OpenStack) SetAttachReq(req int) {
	if req <= 0 || req > 30 {
		req = MaxReqForAttach
	}
	self.AttachReq = req
	klog.Infof("OpenStack.SetAttachReq:%d", self.AttachReq)
}

func (self *OpenStack) GetTenantID() string {
	return self.provider.TenantID
}

func (self *OpenStack) GetTenantName() string {
	return self.provider.TenantName
}

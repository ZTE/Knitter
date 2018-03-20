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
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/ZTE/Knitter/pkg/klog"

	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"

	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

const (
	NetworkStatActive   = "ACTIVE"
	NetworkStatDown     = "DOWN"
	DefaultQuotaNoAdmin = 10
	DefaultQuotaAdmin   = 100
	MaxQuota            = 1000
	MAXIPSEG            = 4
	IPREG               = "^(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|[1-9])\\." +
		"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
		"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
		"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)$"
)

var (
	QuotaNoAdmin int = DefaultQuotaNoAdmin
	QuotaAdmin   int = DefaultQuotaAdmin
)

type Net struct {
	Network         iaasaccessor.Network
	Subnet          iaasaccessor.Subnet
	VlanTransparent bool
	Provider        iaasaccessor.NetworkExtenAttrs
	TenantUUID      string
	Public          bool
	ExternalNet     bool
	CreateTime      string `json:"create_time"`
	Status          string `json:"state"`
	Description     string `json:"description"`
}

type PaasNetwork struct {
	Name            string                         `json:"name"`
	ID              string                         `json:"network_id"`
	GateWay         string                         `json:"gateway"`
	Cidr            string                         `json:"cidr"`
	CreateTime      string                         `json:"create_time"`
	Status          string                         `json:"state"`
	Public          bool                           `json:"public"`
	ExternalNet     bool                           `json:"external"`
	Owner           string                         `json:"owner"`
	Description     string                         `json:"description"`
	SubnetID        string                         `json:"subnet_id"`
	Provider        iaasaccessor.NetworkExtenAttrs `json:"provider"`
	AllocationPools []subnets.AllocationPool       `json:"allocation_pools"`
}

type EncapPaasNetwork struct {
	Network *PaasNetwork `json:"network"`
}

type EncapPaasNetworks struct {
	Networks []*PaasNetwork `json:"networks"`
}

type CreateProviderNetwork struct {
	Name            string                   `json:"name"`
	ID              string                   `json:"network_id"`
	GateWay         string                   `json:"gateway"`
	Cidr            string                   `json:"cidr"`
	NetworkType     string                   `json:"provider:network_type"`
	PhysicalNetwork string                   `json:"provider:physical_network"`
	SegmentationID  string                   `json:"provider:segmentation_id"`
	VlanTransparent bool                     `json:"vlan_transparent"`
	AllocationPools []subnets.AllocationPool `json:"allocation_pools"`
}

type EncapCreateProviderNetwork struct {
	Network *CreateProviderNetwork `json:"provider_network"`
}

type EncapNetworkExten struct {
	Network *iaasaccessor.NetworkExtenAttrs `json:"network_exten"`
}

func (self *Net) getNetworkOwner(key string) string {
	var owner string = "unkown-owner"
	strList := strings.Split(key, "/")
	if len(strList) > 4 {
		owner = strList[3]
	}
	klog.Info("Network-owner is :", owner)
	return owner
}

func transNetObjsTOPaasNetworks(netObjs []*NetworkObject) ([]*PaasNetwork, error) {
	pnetList := make([]*PaasNetwork, 0)
	for _, netObj := range netObjs {
		subnetObj, err := GetSubnetObjRepoSingleton().Get(netObj.SubnetID)
		if err != nil {
			klog.Errorf("transNetObjsTOPaasNetworks: GetSubnetObj for subnetID: %s FAIL, erorr: %v",
				netObj.SubnetID, err)
			return nil, err
		}
		pnet := transNetObjToPaasNetwork(netObj, subnetObj)
		pnetList = append(pnetList, pnet)
	}
	return pnetList, nil
}

func transNetObjToPaasNetwork(netObj *NetworkObject, subnetObj *SubnetObject) *PaasNetwork {
	allocPool := make([]subnets.AllocationPool, 0)
	for _, ap := range subnetObj.AllocPools {
		allocPool = append(allocPool, subnets.AllocationPool{Start: ap.Start, End: ap.End})
	}
	return &PaasNetwork{
		Name:        netObj.Name,
		ID:          netObj.ID,
		GateWay:     subnetObj.GatewayIP,
		Cidr:        subnetObj.CIDR,
		CreateTime:  netObj.CreateTime,
		Status:      constvalue.NetworkStatActive,
		Public:      netObj.IsPublic,
		ExternalNet: netObj.IsExternal,
		Owner:       netObj.TenantID,
		Description: netObj.Description,
		SubnetID:    subnetObj.ID,
		Provider: iaasaccessor.NetworkExtenAttrs{
			Id:              netObj.ID,
			Name:            netObj.Name,
			NetworkType:     netObj.ExtAttrs.NetworkType,
			PhysicalNetwork: netObj.ExtAttrs.PhysicalNetwork,
			SegmentationID:  netObj.ExtAttrs.SegmentationID,
		},
		AllocationPools: allocPool,
	}
}

func GetTenantOwnedNetworks(tenantID string) ([]*PaasNetwork, error) {
	netObjs, err := GetNetObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		klog.Errorf("GetTenantOwnedNetworks: GetNetObjRepoSingleton().ListByTenantID(%s) FAIL, error: %v",
			tenantID, err)
		return nil, err
	}

	var nwList []*PaasNetwork
	for _, netObj := range netObjs {
		subnetObj, err := GetSubnetObjRepoSingleton().Get(netObj.SubnetID)
		if err != nil {
			klog.Errorf("GetTenantOwnedNetworks: GetSubnetObjRepoSingleton().Get(%s) FAIL, error: %v",
				netObj.SubnetID, err)
			return nil, err
		}
		nw := transNetObjToPaasNetwork(netObj, subnetObj)
		nwList = append(nwList, nw)
	}
	return nwList, nil
}

func isTenantNetworkPublic(networkID string) bool {
	netObj, err := GetNetObjRepoSingleton().Get(networkID)
	if err != nil || !netObj.IsPublic {
		klog.Tracef("isTenantNetworkPublic: GetNetObjRepoSingleton().Get(networkID: %s) erorr code: %v, return false",
			networkID, err)
		return false
	}
	klog.Tracef("isTenantNetworkPublic: networkID: %s is public network", networkID, err)
	return true
}

func GetAllPublicNetworks() ([]*PaasNetwork, error) {
	netObjs, err := GetNetObjRepoSingleton().ListByIsPublic(constvalue.StrConstTrue)
	if err != nil {
		return nil, err
	}

	var paasNets = []*PaasNetwork{}
	for _, netObj := range netObjs {
		subnetObj, err := GetSubnetObjRepoSingleton().Get(netObj.SubnetID)
		if err != nil {
			klog.Errorf("GetAllPublicNetworks: Get subnet(id: %s) FAIL, error: %v", netObj.SubnetID, err)
			return nil, err
		}
		pnet := transNetObjToPaasNetwork(netObj, subnetObj)
		paasNets = append(paasNets, pnet)
	}
	return paasNets, nil
}

func (self *Net) GetByID() (*iaasaccessor.Network, error) {
	klog.Info("Now in GetNetworkById Function")
	nw, err := iaas.GetIaaS(self.TenantUUID).GetNetwork(self.Network.Id)
	klog.Info("Now out GetNetworkById Function")
	return nw, err
}

func (self *Net) GetExtenByID() (*iaasaccessor.NetworkExtenAttrs, error) {
	nwe, err := iaas.GetIaaS(self.TenantUUID).GetNetworkExtenAttrs(self.Network.Id)
	if err == nil && nwe.NetworkType == "flat" {
		nwe.SegmentationID = iaasaccessor.FLAT_DEFAULT_ID
	}
	if nwe.NetworkType == constvalue.ProviderNetworkTypeVlan &&
		nwe.SegmentationID == constvalue.VlanTransparentSegmentationID {
		nwe.VlanTransparent = true
	} else {
		nwe.VlanTransparent = false
	}
	klog.Info(" GetExtenById:" + self.Network.Id + "OK!")
	return nwe, err
}

func (self *Net) GetSubnetByID(id string) (*iaasaccessor.Subnet, error) {
	klog.Info("Now in GetSubnetById Function")
	sub, err := iaas.GetIaaS(self.TenantUUID).GetSubnet(id)
	if sub != nil && len(sub.AllocationPools) == 0 {
		sub.AllocationPools = GetAllocationPoolsByCidr(sub.Cidr)
	}
	klog.Info("Now out GetSubnetById Function")
	return sub, err
}

func IsNetworkExist(id string) bool {
	_, err := GetNetObjRepoSingleton().Get(id)
	if err == nil {
		return true
	}
	return false
}

func GetTenantAllNetworks(tenantID string) ([]*PaasNetwork, error) {
	privNets, err := GetNetObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		return nil, err
	}

	klog.Infof("GetTenantAllNetworks: ListByTenantID result: %v", privNets)

	pubNets, err := GetNetObjRepoSingleton().ListByIsPublic(constvalue.StrConstTrue)
	if err != nil {
		return nil, err
	}
	klog.Infof("GetTenantAllNetworks: ListByIsPublic result: %v", pubNets)

	netObjList := make([]*NetworkObject, 0)
	netObjList = append(netObjList, privNets...)
	for _, pubNet := range pubNets {
		var isDup bool
		for _, netObj := range netObjList {
			if pubNet.ID == netObj.ID {
				isDup = true
				break
			}
		}
		if !isDup {
			netObjList = append(netObjList, pubNet)
		}
	}

	pnetList, err := transNetObjsTOPaasNetworks(netObjList)
	if err != nil {
		klog.Errorf("GetTenantNetworks: transNetObjsTOPaasNetworks FAIL, error: %v", err)
		return nil, err
	}
	return pnetList, nil
}

func GetNetworkByName(tenantID, netName string) (*PaasNetwork, error) {
	ownedNets, err := GetTenantOwnedNetworks(tenantID)
	if err != nil {
		klog.Errorf("GetNetworkByName: GetTenantOwnedNetworks[tenantID: %s] FAIL, error: %v", tenantID, err)
		return nil, err
	}
	for _, net := range ownedNets {
		if net.Name == netName {
			klog.Infof("GetNetworkByName: find paasnet[%+v] match [tid: %s, netName: %s]", net, tenantID, netName)
			return net, nil
		}
	}
	klog.Infof("GetNetworkByName: not find any one match [tid: %s, netName: %s], try public networks",
		tenantID, netName)

	pubNets, err := GetAllPublicNetworks()
	if err != nil {
		klog.Errorf("GetNetworkByName: GetAllPublicNetworks FAIL, error: %v", err)
		return nil, err
	}
	for _, net := range pubNets {
		if net.Name == netName {
			klog.Infof("GetNetworkByName: find public paasnet[%+v] match [tid: %s, netName: %s]",
				net, tenantID, netName)
			return net, nil
		}
	}

	klog.Errorf("GetNetworkByName: not find any network(public and private) match [tid: %s, netName: %s]",
		tenantID, netName)
	return nil, errobj.ErrNetworkNotExist
}

var GetNetwork = func(net *Net) (*PaasNetwork, error) {
	return GetNetworkByName(net.TenantUUID, net.Network.Name)
}

func isErrNetworkNotFound(err error) bool {
	if err != nil && strings.Contains(err.Error(), "404") {
		return true
	}
	return false
}

var DeleteNetwork = func(id string) error {
	netObj, err := GetNetObjRepoSingleton().Get(id)
	if err != nil {
		klog.Errorf("DelNetwork: Get NetworkObject [networkID: %s] from repo FAIL, error: %v", id, err)
		return err
	}

	if IsNetworkUsedByIPGroup(id, netObj.TenantID) {
		klog.Errorf("DelNetwork: network[id: %s] has ip group in use", netObj.ID)
		return errobj.ErrNetworkHasIGsInUse
	}

	if IsNetworkInUse(id) {
		klog.Errorf("DelNetwork: network[id: %s] has port in use", netObj.ID)
		return errobj.ErrNetworkHasPortsInUse
	}

	if !netObj.IsExternal {
		err = forceDeleteNetwork(netObj.TenantID, id)
		if err != nil && !isErrNetworkNotFound(err) {
			klog.Errorf("DelNetwork: delete forceDeleteNetwork(id: %s) from iaas FAIL, error: %v", id, err)
			return err
		}
		klog.Infof("DelNetwork: forceDeleteNetwork[id: %s] SUCC", netObj.ID)
	}

	subnets, err := GetSubnetObjRepoSingleton().ListByNetworID(id)
	if err != nil {
		klog.Errorf("DelNetwork: ListByNetworID [networkID: %s] from repo FAIL, error: %v", id, err)
		return err
	}

	for _, subnet := range subnets {
		err = DelSubnet(subnet.ID)
		if err != nil {
			klog.Errorf("DelNetwork: DelSubnet[id: %s] from DB FAIL, error: %v", subnet.ID, err)
			return err
		}
		err = GetSubnetObjRepoSingleton().Del(subnet.ID)
		if err != nil {
			klog.Errorf("DelNetwork: del subnetObject[id: %s] from repo FAIL, error: %v", subnet.ID, err)
			return err
		}
	}

	err = DelNetwork(id)
	if err != nil {
		klog.Errorf("DelNetwork: DelNetwork[id: %s] from DB FAIL, error: %v", id, err)
		return err
	}

	err = GetNetObjRepoSingleton().Del(id)
	if err != nil {
		klog.Errorf("DelNetwork: delete NetworkObject[id: %s] from repo FAIL, error: %v", id, err)
		return err
	}

	err = Net{TenantUUID: netObj.TenantID}.SaveQuota()
	if err != nil {
		klog.Errorf("DelNetwork: SaveQuota[tenantID: %s] FAIL, error: %v", netObj.TenantID, err)
		return err
	}

	klog.Tracef("DelNetwork: delete network[id: %s] SUCC", id)
	return nil
}

// todo: get network info from iaas call directly and for future store in DB
func (self *Net) getNetWorkInfoFromIaas(subID string) error {
	nwt, err := self.GetByID()
	if err != nil {
		klog.Error("Net Create call self.GetById ERROR:", err)
		return err
	}
	self.Network = *nwt

	sub, err := self.GetSubnetByID(subID)
	if err != nil {
		klog.Error("Net Create call self.GetSubnetById ERROR:", err)
		return err
	}
	self.Subnet = *sub

	netExt, err := self.GetExtenByID()
	if err != nil {
		klog.Errorf("Net Create call self.GetExtenByID ERROR:%v", err)
		return err
	}
	self.Provider = *netExt

	return nil
}

func (self *Net) Create() error {
	klog.Info("Now in CreateNetWork Function")
	var err error = nil
	var network *iaasaccessor.Network
	iaasNetworkName := self.TenantUUID + "_" + self.Network.Name
	if self.Provider.NetworkType != "" || self.Provider.PhysicalNetwork != "" {
		network, err = iaas.GetIaaS(self.TenantUUID).CreateProviderNetwork(
			iaasNetworkName,
			self.Provider.NetworkType,
			self.Provider.PhysicalNetwork,
			self.Provider.SegmentationID,
			self.VlanTransparent)
	} else {
		network, err = iaas.GetIaaS(self.TenantUUID).CreateNetwork(iaasNetworkName)
	}
	if err != nil {
		klog.Error("Net Create call GetIaaS().CreateNetwork ERROR:", err)
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	self.Network.Id = network.Id
	subnet, err := iaas.GetIaaS(self.TenantUUID).CreateSubnet(
		self.Network.Id,
		self.Subnet.Cidr,
		self.Subnet.GatewayIp,
		self.Subnet.AllocationPools)
	if err != nil {
		klog.Error("Net Create call GetIaaS().CreateSubnet ERROR:", err)
		iaas.GetIaaS(self.TenantUUID).DeleteNetwork(self.Network.Id)
		self.Network.Id = ""
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	extInfo, err := self.GetExtenByID()
	if err != nil {
		klog.Errorf("Net Create call GetExtenByID FAIL, error: %v", err)
		iaas.GetIaaS(self.TenantUUID).DeleteNetwork(self.Network.Id)
		self.Network.Id = ""
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	self.Status = NetworkStatActive
	self.ExternalNet = false
	self.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	//err = self.saveNetworkToEtcd()
	err = saveNetwork(self, network, subnet, extInfo)
	if err != nil {
		klog.Errorf("Create:saveNetworkToEtcd error,self.Network.Id:[%v],ERROR:[%v]", self.Network.Id, err)
		err1 := iaas.GetIaaS(self.TenantUUID).DeleteNetwork(self.Network.Id)
		if err1 != nil {
			klog.Warningf("Create:common.GetIaaS().DelNetwork(%v)", self.Network.Id)
		}
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	klog.Info("Create:Net Create all OK:", self)
	return nil
}

// todo: temporary function, will be replaced by NetworkManager.Save() object method in future
func saveNetwork(net *Net,
	iaasNet *iaasaccessor.Network,
	iaasSubnet *iaasaccessor.Subnet,
	extAttrs *iaasaccessor.NetworkExtenAttrs) error {
	network := &Network{
		Name:     net.Network.Name,
		ID:       iaasNet.Id,
		SubnetID: iaasSubnet.Id,
		ExtAttrs: ExtenAttrs{
			NetworkType:     extAttrs.NetworkType,
			PhysicalNetwork: extAttrs.PhysicalNetwork,
			SegmentationID:  extAttrs.SegmentationID,
			VlanTransparent: extAttrs.VlanTransparent,
		},
		TenantID:    net.TenantUUID,
		IsPublic:    net.Public,
		IsExternal:  net.ExternalNet,
		CreateTime:  net.CreateTime,
		Description: net.Description,
	}
	err := SaveNetwork(network)
	if err != nil {
		klog.Errorf("saveNetwork: save network: %v FAIL, error: %v", network, err)
		return err
	}

	netObj := TransNetworkToNetworkObject(network)
	err = GetNetObjRepoSingleton().Add(netObj)
	if err != nil {
		klog.Errorf("saveNetwork: save netObj: %v FAIL, error: %v", netObj, err)
		return err
	}

	allocPool := make([]AllocationPool, 0)
	for _, ap := range iaasSubnet.AllocationPools {
		allocPool = append(allocPool, AllocationPool{Start: ap.Start, End: ap.End})
	}
	subnet := &Subnet{
		ID:         iaasSubnet.Id,
		NetworkID:  iaasNet.Id,
		Name:       iaasSubnet.Name,
		CIDR:       iaasSubnet.Cidr,
		GatewayIP:  iaasSubnet.GatewayIp,
		TenantID:   iaasSubnet.TenantId,
		AllocPools: allocPool,
	}
	err = SaveSubnet(subnet)
	if err != nil {
		klog.Errorf("saveNetwork: save subnet: %v FAIL, error: %v", subnet, err)
		return err
	}

	subnetObj := TransSubnetToSubnetObject(subnet)
	err = GetSubnetObjRepoSingleton().Add(subnetObj)
	if err != nil {
		klog.Errorf("saveNetwork: save SubnetObj: %v FAIL, error: %v", subnetObj, err)
		return err
	}
	return nil
}

func IsNetworkUsedByIPGroup(networkID, tenantID string) bool {
	ig := IPGroup{TenantID: tenantID, NetworkID: networkID}
	igs, err := ig.GetIGsByNet()
	if err != nil {
		klog.Errorf("IsNetworkUsedByIPGroup: List ip groups ByNetworkID(id: %s) FAIL, error: %v", networkID, err)
		return false
	}

	if len(igs) != 0 {
		return true
	}

	return false
}

func IsNetworkInUse(id string) bool {
	ports, err := GetPortObjRepoSingleton().ListByNetworkID(id)
	if err != nil {
		klog.Errorf("IsNetworkInUse: List ports ByNetworkID(id: %s) FAIL, error: %v", id, err)
		return false
	}
	if len(ports) != 0 {
		return true
	}

	physPorts, err := GetPhysPortObjRepoSingleton().ListByNetworkID(id)
	if err != nil {
		klog.Errorf("IsNetworkInUse: List Physical port ByNetworkID(id: %s) FAIL, error: %v", id, err)
		return false
	}
	if len(physPorts) != 0 {
		return true
	}
	return false
}

func RegisterNetwork(user, id, subid string, public bool) error {
	var this *Net = &Net{}
	this.TenantUUID = user
	this.Network.Id = id
	this.Public = public
	this.ExternalNet = true

	klog.Info("User-Name:", this.TenantUUID, "<--->Network-ID:", id)
	//check whether net exits
	if IsNetworkExist(id) {
		return errors.New(strconv.Itoa(http.StatusConflict) +
			"::net is already exist. Please delete this net if to register again")
	}

	err := this.CheckQuota()
	if err != nil {
		return err
	}
	//get subid from iaas when subid is none
	if subid == "" {
		subidFromIaas, err := iaas.GetIaaS(this.TenantUUID).GetSubnetID(id)
		if err != nil {
			klog.Error("RegisterNetwork Get Subnetwork ID ERR:", err.Error())
			return BuildErrWithCode(http.StatusNotFound, err)
		}
		subid = subidFromIaas
	}

	klog.Info("Subnet-uuit is:", subid)
	err = this.getNetWorkInfoFromIaas(subid)
	if err != nil {
		klog.Error("RegisterNetwork Get Subnetwork by ID ERR:", err.Error())
		return BuildErrWithCode(http.StatusNotFound, err)
	}

	this.Status = NetworkStatActive
	this.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	//err = this.saveNetworkToEtcd()
	err = saveNetwork(this, &this.Network, &this.Subnet, &this.Provider)
	if err != nil {
		klog.Error("RegisterNetwork Save network info ERR:", err.Error())
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}
	this.SaveQuota()

	klog.Info("Network register OK:", this)
	return nil
}

func ListAllUser() ([]string, error) {
	rootURL := dbaccessor.GetKeyOfTenants()
	nodes, err := common.GetDataBase().ReadDir(rootURL)
	if err != nil {
		klog.Warning("Read Network dir from ETCD Error:", err)
		return nil, err
	}
	tids := make([]string, 0)
	for _, node := range nodes {
		tid := strings.TrimPrefix(node.Key, rootURL+"/")
		tids = append(tids, tid)
	}
	return tids, err
}

func (self *Net) CheckQuota() error {
	tenant := &Tenant{Quota: 0}
	NetNum := 0
	if self.TenantUUID != "admin" {
		tenantKey := dbaccessor.GetKeyOfTenantSelf(self.TenantUUID)
		tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
		if err == nil {
			err = json.Unmarshal([]byte(tenantValue), tenant)
			NetNum = GetNetNumOfTenant(self.TenantUUID)
			if err == nil && NetNum < tenant.Quota {
				klog.Infof("tenant:%v,netname:%v,Quota:%v,NetNum:%v check quota successful",
					self.TenantUUID, self.Network.Name, tenant.Quota, NetNum)
				return nil
			}
		}
		klog.Errorf("tenant:%v,netname:%v,Quota:%v,NetNum:%v check quota failed err:%v",
			self.TenantUUID, self.Network.Name, tenant.Quota, NetNum, err)
		return fmt.Errorf(strconv.Itoa(http.StatusInternalServerError)+"::tenant:%v,netname:%v,checkquota failed",
			self.TenantUUID, self.Network.Name)
	}
	NetNum = GetNetNumOfTenant(self.TenantUUID)
	if NetNum < QuotaAdmin {
		klog.Infof("tenant:admin,netname:%v,Quota:%v,NetNum:%v check quota successful",
			self.Network.Name, QuotaAdmin, NetNum)
		return nil
	}
	klog.Errorf("tenant:admin,netname:%v,Quota:%v,NetNum:%v,check quota failed",
		self.Network.Name, QuotaAdmin, NetNum)
	return fmt.Errorf(strconv.Itoa(http.StatusInternalServerError)+"::tenant:admin,netname:%v,checkquota failed",
		self.Network.Name)
}

func (self Net) SaveQuota() error {
	// todo: need refactor in future, because only Net.tenantUUID is used in Net struct,
	// change it to a function not a method of Net
	tenantID := self.TenantUUID
	tenant := &Tenant{Quota: 0}
	if tenantID == constvalue.PaaSTenantAdminDefaultUUID {
		return nil
	}

	tenantKey := dbaccessor.GetKeyOfTenantSelf(tenantID)
	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil {
		klog.Errorf("Net.SaveQuota: DB.ReadLeaf(key: %s) FAIL, error: %v", tenantKey, err)
		return fmt.Errorf("%v: Read TenantSelf Error", err)
	}

	err = json.Unmarshal([]byte(tenantValue), tenant)
	if err != nil {
		klog.Errorf("Net.SaveQuota: json.Unmarshal(%s) FAIL, error: %v", tenantValue, err)
		return fmt.Errorf("%v: Unmarshal Error", err)
	}

	tenant.NetNum = GetNetNumOfTenant(tenantID)
	value, _ := json.Marshal(tenant)
	err = common.GetDataBase().SaveLeaf(tenantKey, string(value))
	if err != nil {
		klog.Errorf("Net.SaveQuota: DB.SaveLeaf(key: %s, vlaue: %s) FAIL, error: %v", tenantKey, string(value), err)
		return fmt.Errorf("%v: SaveLeaf Error", err)
	}

	klog.Info("SaveQuota Successful")
	return nil
}

var GetIPSeg2Range = func(ipSegs []string, maskLen int) (int, int) {
	if maskLen > 16 {
		segIP, _ := strconv.Atoi(ipSegs[1])
		return segIP, segIP
	}
	ipSeg, _ := strconv.Atoi(ipSegs[1])
	return GetIPSegRange(uint8(ipSeg), uint8(16-maskLen))
}

var GetIPSeg3Range = func(ipSegs []string, maskLen int) (int, int) {
	if maskLen > 24 {
		segIP, _ := strconv.Atoi(ipSegs[2])
		return segIP, segIP
	}
	ipSeg, _ := strconv.Atoi(ipSegs[2])
	return GetIPSegRange(uint8(ipSeg), uint8(24-maskLen))
}

var GetIPSeg4Range = func(ipSegs []string, maskLen int) (int, int) {
	ipSeg, _ := strconv.Atoi(ipSegs[3])
	segMinIP, segMaxIP := GetIPSegRange(uint8(ipSeg), uint8(32-maskLen))
	return segMinIP + 1, segMaxIP - 1
}

var GetIPSegRange = func(userSegIP, offset uint8) (int, int) {
	var ipSegMax uint8 = 255
	netSegIP := ipSegMax << offset
	segMinIP := netSegIP & userSegIP
	segMaxIP := userSegIP&(255<<offset) | ^(255 << offset)
	return int(segMinIP), int(segMaxIP)
}

var GetCidrIPRange = func(cidr string) (string, string) {
	if cidr == "" {
		return "", ""
	}
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	seg2MinIP, seg2MaxIP := GetIPSeg2Range(ipSegs, maskLen)
	seg3MinIP, seg3MaxIP := GetIPSeg3Range(ipSegs, maskLen)
	seg4MinIP, seg4MaxIP := GetIPSeg4Range(ipSegs, maskLen)
	ipPrefix := ipSegs[0] + "."
	return ipPrefix + strconv.Itoa(seg2MinIP) + "." + strconv.Itoa(seg3MinIP) + "." + strconv.Itoa(seg4MinIP),
		ipPrefix + strconv.Itoa(seg2MaxIP) + "." + strconv.Itoa(seg3MaxIP) + "." + strconv.Itoa(seg4MaxIP)
}

var IsAllocationPoolsInCidr = func(allocationPools []subnets.AllocationPool, cidr string) bool {
	for _, pool := range allocationPools {
		if !IsFixIPInCidr(pool.Start, cidr) || !IsFixIPInCidr(pool.End, cidr) {
			return false
		}
	}
	return true
}

var IsFixIPInIPRange = func(ip string, allocationPool subnets.AllocationPool) bool {
	if ip == "" || allocationPool.Start == "" || allocationPool.End == "" {
		return false
	}

	startIPByte := net.ParseIP(allocationPool.Start)
	endIPByte := net.ParseIP(allocationPool.End)
	ipByte := net.ParseIP(ip)
	if bytes.Compare(ipByte, startIPByte) >= 0 && bytes.Compare(ipByte, endIPByte) <= 0 {
		return true
	}

	return false
}

var IsAllocationPoolsCoverd = func(allocationPools []subnets.AllocationPool) bool {
	if len(allocationPools) == 0 {
		return true
	}
	for i := 0; i < len(allocationPools); i++ {
		for j := 0; j < len(allocationPools); j++ {
			if i == j {
				continue
			}
			if IsFixIPInIPRange(allocationPools[i].Start, allocationPools[j]) {
				return true
			}
		}
	}
	return false
}

var CheckAllocationPools = func(allocationPools []subnets.AllocationPool, cidr string) error {
	if len(allocationPools) == 0 || cidr == "" {
		return errobj.ErrCheckAllocationPools
	}
	if IsAllocationPoolsInCidr(allocationPools, cidr) && !IsAllocationPoolsCoverd(allocationPools) {
		return nil
	}
	return errobj.ErrCheckAllocationPools
}

var IsAllocationPoolsLegal = func(allocationPools []subnets.AllocationPool, cidr string) bool {
	if len(allocationPools) == 0 || cidr == "" {
		return false
	}
	for _, pool := range allocationPools {
		startIPByte := net.ParseIP(pool.Start)
		endIPByte := net.ParseIP(pool.End)
		if bytes.Compare(startIPByte, endIPByte) > 0 {
			return false
		}
	}
	err := CheckAllocationPools(allocationPools, cidr)
	if err != nil {
		return false
	}
	return true
}

var IsFixIPInCidr = func(ip, cidr string) bool {
	_, ipNet, _ := net.ParseCIDR(cidr)
	tmpIP := net.ParseIP(ip)
	if ipNet.Contains(tmpIP) {
		return true
	}
	return false
}

var GetAllocationPoolsByCidr = func(cidr string) []subnets.AllocationPool {
	if !IsCidrLegal(cidr) {
		return []subnets.AllocationPool{}
	}
	MinIP, MaxIP := GetCidrIPRange(cidr)
	allocationPools := []subnets.AllocationPool{
		{
			Start: MinIP,
			End:   MaxIP,
		},
	}
	return allocationPools
}

var GetAllocationPools = func(allocationPools []*jason.Object, cidr, gw string) ([]subnets.AllocationPool, error) {
	var pools []subnets.AllocationPool
	if len(allocationPools) == 0 {
		return []subnets.AllocationPool{}, nil
	}
	if !IsCidrLegal(cidr) {
		return []subnets.AllocationPool{}, errobj.ErrCheckAllocationPools
	}
	for _, pool := range allocationPools {
		startIP, _ := pool.GetString("start")
		endIP, _ := pool.GetString("end")
		allocationPool := subnets.AllocationPool{
			Start: startIP,
			End:   endIP,
		}
		if IsFixIPInIPRange(gw, allocationPool) {
			return []subnets.AllocationPool{}, errobj.ErrCheckAllocationPools
		}
		pools = append(pools, allocationPool)
	}
	if IsAllocationPoolsLegal(pools, cidr) {
		return pools, nil
	}
	return []subnets.AllocationPool{}, errobj.ErrCheckAllocationPools
}

var IsCidrLegal = func(cidr string) bool {
	cidrInfo := strings.Split(cidr, "/")
	if len(cidrInfo) != 2 {
		return false
	}
	bl, _ := isIPLegitimate(cidrInfo[0])
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	if maskLen < 8 || !bl {
		return false
	}
	return true
}

func isIPLegitimate(ipAddress string) (bool, error) {
	return regexp.MatchString(IPREG, ipAddress)
}

// todo: migration network/subnet to new database model
func migrateNetworksToNewDBModel() error {
	klog.Infof("migrateNetworksToNewDBModel: START")
	defer klog.Infof("migrateNetworksToNewDBModel: END")
	nets, err := getAllNetworks()
	if err != nil {
		klog.Errorf("migrateNetworksToNewDBModel: getAllNetworks FAIL, error: %v", err)
		return err
	}

	err = transAllNetworksToNewModel(nets)
	if err != nil {
		klog.Errorf("migrateNetworksToNewDBModel: transAllNetworksToNewModel FAIL, error: %v", err)
		return err
	}

	return nil
}

func transAllNetworksToNewModel(nets []*Net) error {
	klog.Infof("transAllNetworksToNewModel: START")
	defer klog.Infof("transAllNetworksToNewModel: END")
	for _, net := range nets {
		tmpNet, tmpSubnet := TransNetToNetwork(net)
		// add subnet to db
		err := SaveSubnet(tmpSubnet)
		if err != nil {
			klog.Errorf("transAllNetworksToNewModel: SaveSubnet(%+v) FAIL, error: %v", tmpSubnet, err)
			return err
		}

		klog.Infof("transAllNetworksToNewModel: add subnet(%+v) to db SUCC", tmpSubnet)
		// add network to db
		err = SaveNetwork(tmpNet)
		if err != nil {
			klog.Errorf("transAllNetworksToNewModel: SaveNetwork(%+v) FAIL, error: %v", tmpNet, err)
			return err
		}
		klog.Infof("transAllNetworksToNewModel: add network(%+v) to db SUCC", tmpNet)
	}

	return nil
}

func TransNetToNetwork(net *Net) (*Network, *Subnet) {
	paasNetName := strings.TrimPrefix(net.Network.Name, net.TenantUUID+"_")
	tmpNet := &Network{
		Name:     paasNetName,
		ID:       net.Network.Id,
		SubnetID: net.Subnet.Id,
		ExtAttrs: ExtenAttrs{
			NetworkType:     net.Provider.NetworkType,
			PhysicalNetwork: net.Provider.PhysicalNetwork,
			SegmentationID:  net.Provider.SegmentationID,
		},
		TenantID:    net.TenantUUID,
		IsPublic:    net.Public,
		IsExternal:  net.ExternalNet,
		CreateTime:  net.CreateTime,
		Description: net.Description,
	}

	tmpSubnet := &Subnet{
		ID:        net.Subnet.Id,
		NetworkID: net.Network.Id,
		Name:      net.Subnet.Name,
		CIDR:      net.Subnet.Cidr,
		GatewayIP: net.Subnet.GatewayIp,
		TenantID:  net.TenantUUID,
	}

	return tmpNet, tmpSubnet
}

func getAllNetworks() ([]*Net, error) {
	tids, err := getAllTenantIds()
	if err != nil {
		klog.Errorf("getAllNetworks: getAllTenantIds FAIL, error: %v", err)
		return nil, err
	}

	var allNets = []*Net{}
	for _, tid := range tids {
		nets, err := getTenantNetworks(tid)
		if err != nil && !IsKeyNotFoundError(err) {
			klog.Errorf("getAllNetworks: getTenantNetworks(tid: %s) FAIL, error: %v", tid, err)
			return nil, err
		}
		if len(nets) == 0 {
			continue
		}

		allNets = append(allNets, nets...)
	}

	klog.Infof("getAllNetworks: len: %d, detail: %v", len(allNets), allNets)
	return allNets, nil
}

func getTenantNetworks(tid string) ([]*Net, error) {
	netsKeys := dbaccessor.GetKeyOfNetworkGroup(tid)
	nodes, err := common.GetDataBase().ReadDir(netsKeys)
	if err != nil {
		klog.Errorf("getTenantNetworks: ReadDir(%s) FAIL, error: %v", netsKeys, err)
		return nil, err
	}

	var nets = []*Net{}
	for _, node := range nodes {
		klog.Infof("getTenantNetworks: process net key: %s", node.Key)
		netID := strings.TrimPrefix(node.Key, netsKeys+"/")
		netKey := dbaccessor.GetKeyOfNetworkSelf(tid, netID)
		netJSON, err := common.GetDataBase().ReadLeaf(netKey)
		if err != nil {
			if !IsKeyNotFoundError(err) {
				klog.Errorf("getTenantNetworks: ReadLeaf(%s) FAIL, error: %v", netKey, err)
				return nil, err
			}
			klog.Infof("getTenantNetworks: ReadLeaf(%s) return key not found, public net: %s", netKey, netID)
			continue
		}

		netContent := &Net{}
		err = json.Unmarshal([]byte(netJSON), netContent)
		if err != nil {
			klog.Errorf("getTenantNetworks: Unmarshal(%s) FAIL, error: %v", netJSON, err)
			return nil, err
		}

		nets = append(nets, netContent)
	}

	klog.Infof("getTenantNetworks: result: %v", nets)
	return nets, nil
}

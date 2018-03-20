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

package controllers

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
)

// Operations about network
type NetworkController struct {
	beego.Controller
}

type Network struct {
	Name            string                   `json:"name"`
	ID              string                   `json:"network_id"`
	GateWay         string                   `json:"gateway"`
	Cidr            string                   `json:"cidr"`
	CreateTime      string                   `json:"create_time"`
	Status          string                   `json:"state"`
	Public          bool                     `json:"public"`
	ExternalNet     bool                     `json:"external"`
	Owner           string                   `json:"owner"`
	Description     string                   `json:"description"`
	SubnetID        string                   `json:"subnet_id"`
	AllocationPools []subnets.AllocationPool `json:"allocation_pools"`
}

type EncapPaasNetwork struct {
	Network *Network `json:"network"`
}

type EncapPaasNetworks struct {
	Networks []*Network `json:"networks"`
}

type EncapCreateProviderNetwork struct {
	Network *models.CreateProviderNetwork `json:"provider_network"`
}

type EncapNetworkExten struct {
	Network *iaasaccessor.NetworkExtenAttrs `json:"network_exten"`
}

func isGatewayValid(gwip, cidr string) bool {
	ipad := net.ParseIP(gwip)
	if ipad == nil {
		return false
	}
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ip.Equal(ipad.Mask(network.Mask))
}

func isNetworkPublicNotPermitted(isPublic bool, netOwnerID string) bool {
	if isPublic == true && netOwnerID != constvalue.PaaSTenantAdminDefaultUUID {
		return true
	}
	return false
}

//todo refactor
func CreateDefaultNetwork() error {

	net := models.Net{}
	net.Network.Name = constvalue.DefaultPaaSNetwork
	net.Subnet.Cidr = constvalue.DefaultPaaSCidr
	net.Public = true
	net.TenantUUID = constvalue.PaaSTenantAdminDefaultUUID
	netsInTenant, err := models.GetTenantOwnedNetworks(net.TenantUUID)
	if err != nil {
		klog.Errorf("CreateDefaulNetwork:  models.GetTenantNetworks(tenantID: %s) FAIL, error: %v", net.TenantUUID, err)
		return err
	}
	for _, netDetail := range netsInTenant {
		if netDetail.Name == net.Network.Name {
			klog.Warningf("CreateDefaulNetwork: [%v] network exist", constvalue.DefaultPaaSNetwork)
			return nil
		}
	}
	err = net.CheckQuota()
	if err != nil {
		return err
	}
	err = net.Create()
	if err != nil {
		return err
	}
	net.SaveQuota()
	return nil
}

func (self *NetworkController) CreateNetwork(req *jason.Object) error {
	net := models.Net{}
	net.Network.Name, _ = req.GetString("name")
	net.Subnet.Cidr, _ = req.GetString("cidr")
	gw, err := req.GetString("gateway")
	if err != nil || gw == "" {
		klog.Info("net gateway is blank string, omit it")
	} else if !isGatewayValid(gw, net.Subnet.Cidr) {
		return models.BuildErrWithCode(http.StatusBadRequest, errors.New("invalid gateway"))
	}
	net.Subnet.GatewayIp = gw

	allocationPools, errAllocationPools := req.GetObjectArray("allocation_pools")
	if errAllocationPools != nil {
		klog.Info("Allocation pools by default iaas")
	} else {
		pools, err := models.GetAllocationPools(allocationPools, net.Subnet.Cidr, net.Subnet.GatewayIp)
		if err != nil {
			return models.BuildErrWithCode(http.StatusBadRequest, errors.New("invalid allocation_pools"))
		}
		net.Subnet.AllocationPools = pools
	}
	net.Public, _ = req.GetBoolean("public")
	net.TenantUUID = self.GetString(":user")
	if isNetworkPublicNotPermitted(net.Public, net.TenantUUID) {
		klog.Error("NetworkController.CreateNetwork: isNetworkPublicNotPermitted() return true")
		return models.BuildErrWithCode(http.StatusForbidden, errobj.ErrRequestNeedAdminPermission)
	}

	netDescription, err := req.GetString("description")
	if err != nil {
		klog.Info("network description is blank string, omit it")
	}
	net.Description = netDescription

	netsInTenant, err := models.GetTenantOwnedNetworks(net.TenantUUID)
	if err != nil {
		klog.Errorf("CreateNetwork:  models.GetTenantNetworks(tenantID: %s) FAIL, error: %v", net.TenantUUID, err)
		return err
	}
	for _, netDetail := range netsInTenant {
		if netDetail.Name == net.Network.Name {
			return models.BuildErrWithCode(http.StatusConflict, errors.New("can not create the same name"))
		}
	}
	err = net.CheckQuota()
	if err != nil {
		return err
	}
	err = net.Create()
	if err != nil {
		return err
	}
	net.SaveQuota()

	cnw := Network{Name: net.Network.Name,
		ID: net.Network.Id, Cidr: net.Subnet.Cidr,
		GateWay: net.Subnet.GatewayIp, Public: net.Public,
		Owner: net.TenantUUID, CreateTime: net.CreateTime,
		Status:      constvalue.NetworkStatActive,
		Description: net.Description, SubnetID: net.Subnet.Id,
		AllocationPools: net.Subnet.AllocationPools}

	self.Data["json"] = EncapPaasNetwork{Network: &cnw}
	self.ServeJSON()
	return nil
}

func checkSegmentationID(networkType, segmentationID string) bool {
	if networkType == constvalue.ProviderNetworkTypeVlan {
		segID, err := strconv.Atoi(segmentationID)
		if err != nil {
			klog.Errorf("checkSegmentationID: trans(segmentationID: %s) to integer error: %v",
				segmentationID, err)
			return false
		}
		if segID < constvalue.MinVlanID || segID > constvalue.MaxVlanID {
			klog.Errorf("checkSegmentationID: segmentationID: %d is invalid vlanID", segID)
			return false
		}
	}
	return true
}

func (self *NetworkController) CreateProviderNetwork(req *jason.Object) (err error) {
	net := models.Net{}
	net.Network.Name, _ = req.GetString("name")
	net.Subnet.Cidr, _ = req.GetString("cidr")
	net.Subnet.GatewayIp, _ = req.GetString("gateway")
	net.Provider.NetworkType, _ = req.GetString("provider:network_type")
	net.Provider.PhysicalNetwork, _ = req.GetString("provider:physical_network")
	net.VlanTransparent, _ = req.GetBoolean("vlan_transparent")

	segID, err := req.GetString("provider:segmentation_id")
	if net.VlanTransparent {
		if err == nil {
			return models.BuildErrWithCode(http.StatusBadRequest, errobj.ErrVlanTransparentConflictArgs)
		}
		segID = constvalue.VlanTransparentSegmentationID
	} else {
		if err == nil && !checkSegmentationID(net.Provider.NetworkType, segID) {
			return models.BuildErrWithCode(http.StatusBadRequest, errobj.ErrInvalidVlanID)
		}
	}
	net.Provider.SegmentationID = segID

	net.Public, _ = req.GetBoolean("public")
	net.TenantUUID = self.GetString(":user")
	if isNetworkPublicNotPermitted(net.Public, net.TenantUUID) {
		klog.Error("NetworkController.CreateProviderNetwork: isNetworkPublicNotPermitted() return true")
		return models.BuildErrWithCode(http.StatusForbidden, errobj.ErrRequestNeedAdminPermission)
	}

	allocationPools, errAllocationPools := req.GetObjectArray("allocation_pools")
	if errAllocationPools != nil {
		klog.Info("Allocation pools by default iaas")
	} else {
		pools, err := models.GetAllocationPools(allocationPools, net.Subnet.Cidr, net.Subnet.GatewayIp)
		if err != nil {
			return models.BuildErrWithCode(http.StatusBadRequest, errors.New("invalid allocation_pools"))
		}
		net.Subnet.AllocationPools = pools
	}
	err = net.CheckQuota()
	if err != nil {
		return err
	}
	err = net.Create()
	if err != nil {
		return err
	}
	net.SaveQuota()

	cnw := models.CreateProviderNetwork{
		Name: net.Network.Name,
		ID:   net.Network.Id,
		Cidr: net.Subnet.Cidr, GateWay: net.Subnet.GatewayIp,
		NetworkType:     net.Provider.NetworkType,
		PhysicalNetwork: net.Provider.PhysicalNetwork,
		SegmentationID:  net.Provider.SegmentationID,
		AllocationPools: net.Subnet.AllocationPools}
	self.Data["json"] = EncapCreateProviderNetwork{Network: &cnw}
	self.ServeJSON()
	return nil
}

func (self *NetworkController) RegisterNetwork(extNetwork *jason.Object) error {
	var err error
	id, _ := extNetwork.GetString("id")
	name, _ := extNetwork.GetString("name")
	subid, _ := extNetwork.GetString("subid")
	public, _ := extNetwork.GetBoolean("public")
	paasTenantID := self.GetString(":user")
	klog.Info("register extenal_network --->", id, "-----name", name, "---public:", public)
	if isNetworkPublicNotPermitted(public, paasTenantID) {
		klog.Error("NetworkController.Post:isNetworkPublicNotPermitted(): true")
		return models.BuildErrWithCode(http.StatusForbidden, errobj.ErrRequestNeedAdminPermission)
	}

	if (id == "") && (name != "") {
		id, err = iaas.GetIaaS(paasTenantID).GetNetworkID(name)
		if err != nil {
			klog.Warning("GetNetworkID ERROR:", err)
		}
	}
	if id == "" {
		err = errors.New("register network input error")
		return models.BuildErrWithCode(http.StatusNotFound, err)
	}

	err = models.RegisterNetwork(paasTenantID, id, subid, public)
	if err != nil {
		return models.BuildErrWithCode(http.StatusInternalServerError, err)
	}

	self.Data["json"] = map[string]string{"RegisterNetwork": "OK"}
	self.ServeJSON()
	return nil
}

// @Title create
// @Description create network
// @Param	body		body 	models.EncapNetwork	true		"configration for network"
// @Success 200 {string} models.Network.Id
// @Failure 403 invalid request body
// @Failure 406 create network error
// @router / [post]
func (self *NetworkController) Post() {
	klog.Infof("@@@Create network START")
	defer klog.Infof("@@@Create network END")
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	var err error

	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info(string(body))

	networkCfg, err := jason.NewObjectFromBytes(body)
	if err != nil {
		Err400(&self.Controller, err)
		return
	}

	network, err := networkCfg.GetObject("network")
	if err == nil && network != nil {
		err = self.CreateNetwork(network)
		if err != nil {
			HandleErr(&self.Controller, err)
			return
		}
		return
	}

	provider, err := networkCfg.GetObject("provider_network")
	if err == nil && provider != nil {
		err = self.CreateProviderNetwork(provider)
		if err != nil {
			HandleErr(&self.Controller, err)
			return
		}
		return
	}

	var extNetwork *jason.Object
	extNetwork, err = networkCfg.GetObject("external_network")
	if err != nil {
		extNetwork, err = networkCfg.GetObject("extenal_network")
	}
	if err == nil && extNetwork != nil {
		err = self.RegisterNetwork(extNetwork)
		if err != nil {
			HandleErr(&self.Controller, err)
			return
		}
		return
	}

	klog.Warning("NETWORK-CREATE-REQ-UNKNOW:" + string(body))
	ErrorRequstRsp400(&self.Controller, string(body))
	return
}

func makeNetworkInfo(net *models.PaasNetwork) *Network {
	nw := Network{
		Name:            net.Name,
		ID:              net.ID,
		SubnetID:        net.SubnetID,
		Cidr:            net.Cidr,
		GateWay:         net.GateWay,
		Description:     net.Description,
		Status:          constvalue.NetworkStatActive,
		Public:          net.Public,
		Owner:           net.Owner,
		ExternalNet:     net.ExternalNet,
		CreateTime:      net.CreateTime,
		AllocationPools: net.AllocationPools}
	return &nw
}

func transNetObjToNetwork(netObj *models.NetworkObject, subnetObj *models.SubnetObject) *Network {
	allocPool := make([]subnets.AllocationPool, 0)
	for _, ap := range subnetObj.AllocPools {
		allocPool = append(allocPool, subnets.AllocationPool{Start: ap.Start, End: ap.End})
	}
	nw := Network{
		Name:            netObj.Name,
		ID:              netObj.ID,
		SubnetID:        netObj.SubnetID,
		Cidr:            subnetObj.CIDR,
		GateWay:         subnetObj.GatewayIP,
		Description:     netObj.Description,
		Public:          netObj.IsPublic,
		Owner:           netObj.TenantID,
		ExternalNet:     netObj.IsExternal,
		CreateTime:      netObj.CreateTime,
		Status:          constvalue.NetworkStatActive,
		AllocationPools: allocPool}
	return &nw
}

func makeNetworkInfoList(nets []*models.PaasNetwork) []*Network {
	var nl []*Network
	for _, net := range nets {
		nw := makeNetworkInfo(net)
		nl = append(nl, nw)
	}
	return nl
}

func getNetworkByID(id string) (*Network, error) {
	netObj, err := models.GetNetObjRepoSingleton().Get(id)
	if err != nil {
		klog.Errorf("getNetworkByID: get NetworkObject(id: %s) FAIL, error: %v", id, err)
		return nil, err
	}
	subnetObj, err := models.GetSubnetObjRepoSingleton().Get(netObj.SubnetID)
	if err != nil {
		klog.Errorf("getNetworkByID: get SubnetObject(id: %s) FAIL, error: %v", netObj.SubnetID, err)
		return nil, err
	}
	net := transNetObjToNetwork(netObj, subnetObj)
	klog.Infof("getNetworkByID: get Network response: %v", net)
	return net, nil
}

func (self *NetworkController) GetNetworkInfo(id string) {
	net, err := getNetworkByID(id)
	if err != nil {
		klog.Errorf("GetNetworkInfo: getNetworkByID(id: %s) FAIL, error: %v", id, err)
		NotfoundErr404(&self.Controller, err)
		return
	}

	self.Data["json"] = EncapPaasNetwork{Network: net}
	self.ServeJSON()
}

func (self *NetworkController) GetNetworkExtenInfo(id string) {
	net := iaasaccessor.Network{Id: id}
	nw := &models.Net{Network: net}
	nw.TenantUUID = self.GetString(":user")
	netInfo, err := getNetworkByID(id)
	if err != nil {
		NotfoundErr404(&self.Controller, err)
		return
	}
	extInfo, err := nw.GetExtenByID()
	if err != nil {
		NotfoundErr404(&self.Controller, err)
		return
	}

	network := models.PaasNetwork{
		Name:        netInfo.Name,
		ID:          netInfo.ID,
		GateWay:     netInfo.GateWay,
		Cidr:        netInfo.Cidr,
		CreateTime:  netInfo.CreateTime,
		Status:      netInfo.Status,
		Public:      netInfo.Public,
		ExternalNet: netInfo.ExternalNet,
		Owner:       netInfo.Owner,
		Description: netInfo.Description,
		SubnetID:    netInfo.SubnetID,
		Provider:    *extInfo,
	}
	self.Data["json"] = models.EncapPaasNetwork{Network: &network}
	self.ServeJSON()
}

// @Title Get
// @Description find network by network_id
// @Param	network_id		path 	string	true		"the network_id you want to get"
// @Success 200 {object} models.EncapNetwork
// @Failure 404 : Network Not Exist
// @router /:network_id [get]
func (self *NetworkController) Get() {
	klog.Infof("@@@Get network START")
	defer klog.Infof("@@@Get network END")
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	id := self.Ctx.Input.Param(":network_id")
	klog.Info("Request network_id: ", id)

	exten, err := self.GetBool("extension")
	if err == nil && exten == true {
		self.GetNetworkExtenInfo(id)
		return
	}

	self.GetNetworkInfo(id)
	return

}

func (self *NetworkController) GetAllNetworks() {
	tids, err := models.ListAllUser()
	if err != nil {
		var nwl []*Network
		self.Data["json"] = EncapPaasNetworks{Networks: nwl}
		self.ServeJSON()
		return
	}
	var networkList []*models.PaasNetwork
	for _, tid := range tids {
		obs, err := models.GetTenantOwnedNetworks(tid)
		if err != nil {
			klog.Errorf("NetworkController: GetAllNetworks(tenantID: %s) FAIL, error: %v",
				tid, err)
		}

		networkList = append(networkList, obs...)
	}
	self.Data["json"] = EncapPaasNetworks{
		Networks: makeNetworkInfoList(networkList)}
	self.ServeJSON()
}

func (self *NetworkController) GetAllPublicNetworks() {
	obs, err := models.GetAllPublicNetworks()
	if err != nil {
		klog.Errorf("GetAllPublicNetworks:  models.GetAllPublicNetworks FAIL, error: %v", err)
	}

	self.Data["json"] = EncapPaasNetworks{
		Networks: makeNetworkInfoList(obs)}
	self.ServeJSON()
	return
}

func (self *NetworkController) GetUserAllNetworks(tenantID, networkName string) {
	klog.Infof("Request tenant[ID: %s]'s all networks START, network name[%s]", tenantID, networkName)

	obs := make([]*models.PaasNetwork, 0)
	if networkName == "" {
		paasNets, err := models.GetTenantAllNetworks(tenantID)
		if err != nil {
			klog.Errorf("GetUserAllNetworks: GetTenantNetworks FAIL, error: %v", err)
			Err400(&self.Controller, err)
			return
		}

		obs = append(obs, paasNets...)
	} else {
		paasNet, err := models.GetNetworkByName(tenantID, networkName)
		if err != nil {
			klog.Errorf("GetUserAllNetworks: GetNetworkByName FAIL, error: %v", err)
			Err400(&self.Controller, err)
			return
		}

		obs = append(obs, paasNet)
	}

	self.Data["json"] = EncapPaasNetworks{
		Networks: makeNetworkInfoList(obs)}
	self.ServeJSON()
	return
}

// @Title GetAll
// @Description get all network
// @Success 200 {object} models.EncapNetworks
// @router / [get]
func (self *NetworkController) GetAll() {
	klog.Infof("@@@GetAll networks START")
	defer klog.Infof("@@@GetAll networks END")
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}

	all, err := self.GetBool("all")
	if err == nil && all == true {
		self.GetAllNetworks()
		return
	}

	public, err := self.GetBool("public")
	if err == nil && public == true {
		self.GetAllPublicNetworks()
		return
	}

	name := self.GetString("name")
	self.GetUserAllNetworks(paasTenantID, name)
	return
}

// @Title delete
// @Description delete network by network_id
// @Param	network_id		path 	string	true		"The network_id you want to delete"
// @Success 200 {string} delete success!
// @Failure 404 Network not Exist
// @router /:network_id [delete]
func (self *NetworkController) Delete() {
	klog.Infof("@@@Delete network START")
	defer klog.Infof("@@@Delete network END")
	var err error

	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	id := self.Ctx.Input.Param(":network_id")
	err = models.DeleteNetwork(id)
	if err != nil {
		if err == errobj.ErrRecordNotExist {
			NotfoundErr404(&self.Controller, err)
		} else {
			HandleErr(&self.Controller, err)
		}
		return
	}
	self.Data["json"] = ""
	self.ServeJSON()
}

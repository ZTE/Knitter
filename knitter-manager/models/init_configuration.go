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
	"context"
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"net"
	"net/http"
	"time"
)

var Scene string = ""
var DeferDelNetworks []string = make([]string, 0)
var RecyleTmOutFlag bool = false
var RecyleInitConfFlag bool = false
var RecyleInitNetsFlag bool = false

type InitCfg struct {
	Configuration InitConfiguration `json:"configuration"`
	Networks      InitNetworks      `json:"networks"`
}

type InitConfiguration struct {
	EndPoint   string `json:"endpoint"`
	User       string `json:"user"`
	Password   string `json:"password"`
	TenantName string `json:"tenant_name"`
	TenantID   string `json:"tenant_id"`
}

type InitNetworks struct {
	InitRegNetworks    []*InitRNetwork `json:"registered_networks"`
	InitCreateNetworks []*InitCNetwork `json:"created_networks"`
}

type InitRNetwork struct {
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
	Desc   string `json:"desc"`
	Public bool   `json:"public"`
}

type InitCNetwork struct {
	Name            string                   `json:"name"`
	Cidr            string                   `json:"cidr"`
	Desc            string                   `json:"desc"`
	Public          bool                     `json:"public"`
	Gw              string                   `json:"gw"`
	AllocationPool  []subnets.AllocationPool `json:"allocation_pool"`
	NetworksType    string                   `json:"provider:network_type"`
	PhysicalNetwork string                   `json:"provider:physical_network"`
	SegmentationID  string                   `json:"provider:segmentation_id"`
}

var CfgInit = func(cfgInput *jason.Object, ctx context.Context) error {
	RecyleInitConfFlag = false
	RecyleTmOutFlag = false
	DeferDelNetworks = make([]string, 0)
	opnCfg := openstack.OpenStackConf{
		Tenantid:   constvalue.DefaultIaasTenantID,
		TenantName: constvalue.DefaultIaasTenantName,
	}
	defer func() {
		if p := recover(); p != nil {
			klog.Errorf("CfgInitConf panic")
		}
		if RecyleInitConfFlag == true {
			RollBackInitConf(opnCfg.Tenantid)
		}
		if RecyleTmOutFlag == true {
			RollBackInitNets(DeferDelNetworks)
			RollBackInitConf(opnCfg.Tenantid)
		}
	}()
	select {
	case <-ctx.Done():
		RecyleTmOutFlag = true
		errBud := BuildErrWithCode(http.StatusRequestTimeout, errobj.ErrTmOut)
		return errBud
	default:
	}
	cfg, err := cfgInput.GetObject("init_configuration")
	if err != nil {
		klog.Errorf("GetObject[init_configuration] Error: %v", err)
		errBud := BuildErrWithCode(http.StatusUnsupportedMediaType, err)
		return errBud
	}

	klog.Infof("Scene is %v", Scene)
	if Scene == constvalue.TECS {
		initConfiguration, errGetCfg := GetInitConfiguraton(cfg)
		if errGetCfg != nil {
			klog.Errorf("GetInitConfiguraton Error: %v", errGetCfg)
			errBud := BuildErrWithCode(http.StatusUnauthorized, errGetCfg)
			return errBud
		}
		opnCfg = openstack.OpenStackConf{
			Username:   initConfiguration.User,
			Password:   initConfiguration.Password,
			Url:        initConfiguration.EndPoint,
			Tenantid:   initConfiguration.TenantID,
			TenantName: initConfiguration.TenantName,
		}
		value, _ := json.Marshal(opnCfg)
		errS := SaveOpenStackConfg(value)
		if errS != nil {
			klog.Errorf("SaveOpenStackConfg Error: %v", errS)
			errBud := BuildErrWithCode(http.StatusInternalServerError, err)
			return errBud
		}
		var req int = 0
		if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) != nil {
			req = iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetAttachReq()
		}
		errI := iaas.InitIaaS()
		if errI != nil {
			klog.Errorf("InitIaaS Error: %v", errI)
			RecyleInitConfFlag = true
			errBud := BuildErrWithCode(http.StatusInternalServerError, errI)
			return errBud
		}
		if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) != nil {
			iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).SetAttachReq(req)
		}
		err := iaas.SaveIaasTenantInfoToDB(&opnCfg)
		if err != nil {
			klog.Errorf("SaveIaasTenantInfoToDB err: %v", err)
			RecyleInitConfFlag = true
			return err
		}
	}
	if Scene == constvalue.VNM {
		defalutPhysnet, _ := GetDefaultProviderNetworkByInitConfig(cfg)
		err := UpdateDefaultPhysnet(defalutPhysnet)
		if err != nil {
			return err
		}
	}
	err = SaveAdminTenantInfoToDB(opnCfg.Tenantid, opnCfg.TenantName)
	if err != nil {
		klog.Errorf("SaveAdminTenantInfoToDB err: %v", err)
		RecyleInitConfFlag = true
		return err
	}
	errHandleInitNetworks := HandleInitNetworks(cfg)
	if errHandleInitNetworks != nil {
		klog.Errorf("HandleInitNetworks Error: %v", errHandleInitNetworks)
		RecyleInitConfFlag = true
		errBud := BuildErrWithCode(http.StatusInternalServerError, errHandleInitNetworks)
		return errBud
	}
	cfgByte, _ := cfgInput.Marshal()
	SaveInitConf(cfgByte)
	return nil
}

var GetInitConfiguraton = func(cfg *jason.Object) (*InitConfiguration, error) {
	config, errConfiguration := cfg.GetObject("configuration")
	if errConfiguration != nil || config == nil {
		klog.Errorf("GetObject[configuration] Error: %v", errConfiguration)
		return nil, errors.New("get configuration error, bad json[configuration]")
	}
	configurationObj, errAna := AnalyseInitConfiguration(config)
	if errAna != nil {
		klog.Errorf("AnalyseConfiguration Error: %v", errAna)
		return nil, errors.New("analyse configuration error, " + errAna.Error())
	}
	confAfterAuth, errAuth := AuthInitCfg(configurationObj)
	if errAuth != nil {
		klog.Errorf("AuthInitCfg Error: %v", errAuth)
		return nil, errobj.ErrAuth
	}
	return confAfterAuth, nil
}

var AnalyseInitConfiguration = func(cfg *jason.Object) (*InitConfiguration, error) {
	edp, errEndpoint := cfg.GetString("endpoint")
	if errEndpoint != nil || edp == "" {
		klog.Errorf("GetObject[endpoint] Error: %v", errEndpoint)
		return nil, errors.New("bad json[endpoint]")
	}
	user, errUser := cfg.GetString("user")
	if errUser != nil || user == "" {
		klog.Errorf("GetObject[user] Error: %v", errUser)
		return nil, errors.New("bad json[user]")
	}
	paasword, errPaasword := cfg.GetString("password")
	if errPaasword != nil || paasword == "" {
		klog.Errorf("GetObject[password] Error: %v", errPaasword)
		return nil, errors.New("bad json[password]")
	}
	tenantName, _ := cfg.GetString("tenant_name")
	tenantID, _ := cfg.GetString("tenant_id")
	configurationObj := &InitConfiguration{
		EndPoint:   edp,
		User:       user,
		Password:   paasword,
		TenantName: tenantName,
		TenantID:   tenantID,
	}
	return configurationObj, nil
}

var AuthInitCfg = func(configurationObj *InitConfiguration) (*InitConfiguration, error) {
	opnCfg := openstack.OpenStackConf{
		Username:   configurationObj.User,
		Password:   configurationObj.Password,
		Url:        configurationObj.EndPoint,
		Tenantid:   configurationObj.TenantID,
		TenantName: configurationObj.TenantName,
	}
	client, errCheck := iaas.CheckOpenstackConfig(&opnCfg)
	if errCheck != nil {
		klog.Errorf("CheckOpenstackConfig Error: %v", errCheck)
		return nil, errobj.ErrAuth
	}
	configurationObj.TenantID = client.TenantID
	configurationObj.TenantName = client.TenantName
	return configurationObj, nil
}

var HandleInitNetworks = func(cfg *jason.Object) error {
	DeferDelNetworks = make([]string, 0)
	RecyleInitNetsFlag = false
	defer func() {
		if p := recover(); p != nil {
			klog.Errorf("HandleInitNetworks panic")
		}
		if RecyleInitNetsFlag == true {
			RollBackInitNets(DeferDelNetworks)
		}
	}()
	var err error

	klog.Infof("Scene is %v", Scene)
	networks, errGetObject := cfg.GetObject("networks")
	if errGetObject != nil || networks == nil {
		klog.Errorf("GetObject[networks] Error: %v", errGetObject)
		return errors.New("bad json[networks]")
	}
	err = InitCNetworks(networks)
	if err != nil {
		klog.Errorf("Init register network error: %v", err)
		RecyleInitNetsFlag = true
		return err
	}
	return nil
}

var InitRNetworks = func(networks *jason.Object) error {
	rNetworks, err := networks.GetObjectArray("registered_networks")
	if err != nil || len(rNetworks) == 0 {
		klog.Errorf("GetObject[registered_networks] Error: %v", err)
		return errors.New("bad json[registered_networks]")
	}
	regNets, err := AnalyseRegNets(rNetworks)
	if err != nil {
		klog.Errorf("Analyse register networks failed! error: %v", err)
		return err
	}
	err = CheckRegNets(regNets)
	if err != nil {
		klog.Errorf("Check register networks failed! error: %v", err)
		return err
	}
	err = RegisterInitNetworks(regNets)
	if err != nil {
		klog.Errorf("Register networks failed! error: %v", err)
		return err
	}
	return nil
}

var AnalyseRegNets = func(rNets []*jason.Object) ([]*InitRNetwork, error) {
	var regNets []*InitRNetwork
	for _, net := range rNets {
		netName, _ := net.GetString("name")
		netID, err := net.GetString("uuid")
		if err != nil || netID == "" {
			klog.Errorf("AnalyseRegNets[%v] Error: %v", netName, err)
			return nil, errors.New("bad json[uuid]")
		}
		netDesc, _ := net.GetString("desc")
		netPub, _ := net.GetBoolean("public")
		netObj := &InitRNetwork{
			Name:   netName,
			UUID:   netID,
			Desc:   netDesc,
			Public: netPub,
		}
		regNets = append(regNets, netObj)
	}
	return regNets, nil
}

var CheckRegNets = func(regNets []*InitRNetwork) error {
	var cheInt int = 0
	for _, net := range regNets {
		if net.Name == "net_api" || net.Name == "net_mgt" {
			cheInt++
		}
	}
	if cheInt != 2 {
		klog.Errorf("CheckApiMgtNet Error: Api Mgt is not exist")
		return errors.New("api Mgt is not exist")
	}
	return nil
}

var RegisterInitNetworks = func(regNets []*InitRNetwork) error {
	var needRegNets []*InitRNetwork
	netsAdmin, err := GetTenantOwnedNetworks(constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		klog.Errorf("RegisterInitNetworks: GetTenantOwnedNetworks for tenantID: %s FAIL, error: %v",
			constvalue.PaaSTenantAdminDefaultUUID, err)
		return err
	}

	if netsAdmin == nil {
		needRegNets = regNets
	} else {
		needRegNets = GetNeedRegNets(regNets, netsAdmin)
	}
	for _, regNet := range needRegNets {
		netID, err := RegInitNetwork(*regNet)
		if err != nil {
			klog.Errorf("Register init networks failed. error: %v", err)
			return err
		}
		DeferDelNetworks = append(DeferDelNetworks, netID)
	}
	return nil
}

var GetNeedRegNets = func(regNets []*InitRNetwork, netsAdmin []*PaasNetwork) []*InitRNetwork {
	needRegNets := make([]*InitRNetwork, 0)
	for _, regNet := range regNets {
		flag := false
		for _, netAdmin := range netsAdmin {
			if regNet.UUID == netAdmin.ID && netAdmin.ExternalNet {
				flag = true
				break
			}
		}
		if !flag {
			needRegNets = append(needRegNets, regNet)
		}
	}
	return needRegNets
}

var RegInitNetwork = func(regNet InitRNetwork) (string, error) {
	var net *Net = &Net{}
	net.TenantUUID = constvalue.PaaSTenantAdminDefaultUUID
	net.Network.Id = regNet.UUID
	net.Public = regNet.Public
	net.ExternalNet = true
	err := net.CheckQuota()
	if err != nil {
		klog.Errorf("CheckQuota Error: %v", err)
		return "", errors.New("check admin quota error")
	}
	subidFromIaas, err := iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetSubnetID(regNet.UUID)
	if err != nil {
		klog.Errorf("Get subnetwork ID error: %v", err)
		return "", err
	}
	net.Subnet.Id = subidFromIaas
	err = net.getNetWorkInfoFromIaas(net.Subnet.Id)
	if err != nil {
		klog.Errorf("Get subnetwork by ID error:%v", err)
		return "", err
	}
	net.Status = NetworkStatActive
	net.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	err = saveNetwork(net, &net.Network, &net.Subnet, &net.Provider)
	if err != nil {
		klog.Errorf("Save network info error: %v", err)
		return "", errors.New("save network[" + net.Network.Name + "] error")
	}
	net.SaveQuota()
	return regNet.UUID, nil
}

var InitCNetworks = func(networks *jason.Object) error {
	cNetworks, err := networks.GetObjectArray("created_networks")
	if err != nil || len(cNetworks) == 0 {
		klog.Warningf("Networks for created are nil")
		return nil
	}
	crtNets, err := AnalyseCrtNets(cNetworks)
	if err != nil {
		klog.Errorf("Analyse create networks failed! error: %v", err)
		return err
	}
	netsCrtFinal, err := CheckCrtNets(crtNets)
	if err != nil {
		klog.Errorf("Analyse register networks failed! error: %v", err)
		return err
	}
	err = CreateInitNetworks(netsCrtFinal)
	if err != nil {
		klog.Errorf("Create networks failed! error: %v", err)
		return err
	}
	return nil
}

var AnalyseCrtNets = func(cNets []*jason.Object) ([]*InitCNetwork, error) {
	var crtNets = make([]*InitCNetwork, 0)
	for _, net := range cNets {
		netName, err := net.GetString("name")
		if err != nil || netName == "" {
			klog.Errorf("AnalyseCrtNets name Error: %v ", err)
			return nil, errors.New("bad json[created_networks name]")
		}
		netCidr, err := net.GetString("cidr")
		if err != nil || netCidr == "" {
			klog.Errorf("AnalyseCrtNets[%v] Error: %v", netName, err)
			return nil, errors.New("bad json[cidr]")
		}
		netDesc, _ := net.GetString("desc")
		netPub, _ := net.GetBoolean("public")
		netGw, _ := net.GetString("gw")
		netAlcp := make([]subnets.AllocationPool, 0)
		netAllocation, errAlcp := net.GetObjectArray("allocation_pool")
		if errAlcp == nil {
			for _, alcp := range netAllocation {
				start, err := alcp.GetString("start")
				if err != nil || start == "" {
					klog.Errorf("GetString[start] Error: %v", err)
					return nil, errors.New("bad json[start]")
				}
				end, err := alcp.GetString("end")
				if err != nil || end == "" {
					klog.Errorf("GetString[end] Error: %v", err)
					return nil, errors.New("bad json[end]")
				}
				alcpObj := subnets.AllocationPool{
					Start: start,
					End:   end,
				}
				netAlcp = append(netAlcp, alcpObj)
			}
		}
		netType, _ := net.GetString("provider:network_type")
		phsicalNet, _ := net.GetString("provider:physical_network")
		netSegID, _ := net.GetString("provider:segmentation_id")
		netObj := &InitCNetwork{
			Name:            netName,
			Cidr:            netCidr,
			Desc:            netDesc,
			Public:          netPub,
			Gw:              netGw,
			AllocationPool:  netAlcp,
			NetworksType:    netType,
			PhysicalNetwork: phsicalNet,
			SegmentationID:  netSegID,
		}
		crtNets = append(crtNets, netObj)
	}

	return crtNets, nil
}

var CheckCrtNets = func(crtNets []*InitCNetwork) ([]*InitCNetwork, error) {
	var cheInt int = 0
	netsFinal := make([]*InitCNetwork, 0)
	if Scene == constvalue.TECS {
		for _, net := range crtNets {
			net.SegmentationID = ""
			net.PhysicalNetwork = ""
			net.NetworksType = ""
			netsFinal = append(netsFinal, net)
		}
	} else if Scene == constvalue.VNM {
		for _, net := range crtNets {
			if net.NetworksType == "" {
				net.PhysicalNetwork = ""
				net.SegmentationID = ""
			}
			netsFinal = append(netsFinal, net)
		}
	} else if Scene == constvalue.EMBEDDED {
		for _, net := range crtNets {
			if net.Name == "net_api" {
				cheInt++
			}
			net.SegmentationID = ""
			net.PhysicalNetwork = ""
			net.NetworksType = ""
			netsFinal = append(netsFinal, net)
		}
		if cheInt != 1 {
			klog.Errorf("CheckCrtNets Error: control  media net_api not exist")
			return make([]*InitCNetwork, 0), errors.New("control media net_api is not exist")
		}
	} else {
		klog.Errorf("CheckCrtNets Error: unsupported scene")
		return make([]*InitCNetwork, 0), errors.New("unsupported scene")
	}
	return netsFinal, nil
}

var CreateInitNetworks = func(crtNets []*InitCNetwork) error {
	var needCrtNets []*InitCNetwork
	netsAdmin, err := GetTenantOwnedNetworks(constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		klog.Errorf("CreateInitNetworks: GetTenantOwnedNetworks for tenantID: %s FAIL, error: %v",
			constvalue.PaaSTenantAdminDefaultUUID, err)
		return err
	}

	if netsAdmin == nil {
		needCrtNets = crtNets
	} else {
		needCrtNets = GetNeedCrtNets(crtNets, netsAdmin)
	}
	for _, crtNet := range needCrtNets {
		netID, err := CrtInitNetwork(*crtNet)
		if err != nil {
			klog.Errorf("Create init networks failed. error: %v", err)
			return err
		}
		DeferDelNetworks = append(DeferDelNetworks, netID)
	}
	return nil
}

var GetNeedCrtNets = func(crtNets []*InitCNetwork, netsAdmin []*PaasNetwork) []*InitCNetwork {
	needCrtNets := make([]*InitCNetwork, 0)
	for _, crtNet := range crtNets {
		flag := false
		for _, netAdmin := range netsAdmin {
			if !netAdmin.ExternalNet && crtNet.Name == netAdmin.Name && crtNet.Cidr == netAdmin.Cidr {
				flag = true
				break
			}
		}
		if !flag {
			needCrtNets = append(needCrtNets, crtNet)
		}
	}
	return needCrtNets
}

var CrtInitNetwork = func(crtNet InitCNetwork) (string, error) {
	net := Net{}
	net.Network.Name = crtNet.Name
	net.Subnet.Cidr = crtNet.Cidr
	if crtNet.Gw != "" && !IsGatewayValid(crtNet.Gw, net.Subnet.Cidr) {
		klog.Errorf("CrtInitNetwork[%v] Error: invalid gateway[%v]", crtNet.Name, crtNet.Gw)
		return "", errors.New("invalid gateway")
	}
	net.Subnet.GatewayIp = crtNet.Gw
	if !IsLegalInitAllocationPools(crtNet.AllocationPool, net.Subnet.Cidr, net.Subnet.GatewayIp) {
		return "", errors.New("invalid allocation_pool")
	}
	net.Subnet.AllocationPools = crtNet.AllocationPool
	net.Public = crtNet.Public
	net.TenantUUID = constvalue.PaaSTenantAdminDefaultUUID
	net.Description = crtNet.Desc
	net.Provider.NetworkType = crtNet.NetworksType
	net.Provider.PhysicalNetwork = crtNet.PhysicalNetwork
	net.Provider.SegmentationID = crtNet.SegmentationID
	err := net.CheckQuota()
	if err != nil {
		return "", errors.New("check admin quota error")
	}
	err = net.Create()
	if err != nil {
		return "", err
	}
	net.SaveQuota()

	return net.Network.Id, nil
}

var RollBackInitConf = func(tenantID string) error {
	DelAdminTenantInfoFromDB()
	iaas.DelIaasTenantInfoFromDB(tenantID)
	DelOpenStackConfg()
	iaas.SetIaaS(tenantID, nil)
	return nil
}

var RollBackInitNets = func(netList []string) error {
	for _, id := range netList {
		net := iaasaccessor.Network{Id: id}
		nw := Net{Network: net}
		nw.TenantUUID = constvalue.PaaSTenantAdminDefaultUUID
		DeleteNetwork(id)
	}
	return nil
}

var IsGatewayValid = func(gwip, cidr string) bool {
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

var IsLegalInitAllocationPools = func(allocationPools []subnets.AllocationPool, cidr, gw string) bool {
	if len(allocationPools) == 0 {
		klog.Warningf("IsLegalInitAllocationPools Warning: len allocationPools is 0")
		return true
	}
	if !IsCidrLegal(cidr) {
		klog.Errorf("IsLegalInitAllocationPools Error: cidr is not right")
		return false
	}
	for _, pool := range allocationPools {
		if IsFixIPInIPRange(gw, pool) {
			klog.Errorf("IsLegalInitAllocationPools Error: gw is in IP range")
			return false
		}
	}
	if !IsAllocationPoolsLegal(allocationPools, cidr) {
		klog.Errorf("IsLegalInitAllocationPools Error: pools[start or end] is not right")
		return false
	}
	return true
}

var SaveInitConf = func(value []byte) error {
	key := dbaccessor.GetKeyOfInitConf()
	err := common.GetDataBase().SaveLeaf(key, string(value))
	if err != nil {
		klog.Errorf("SaveInitConf Error: %v", err)
		return err
	}
	return nil
}

var ReadInitConf = func() (string, error) {
	key := dbaccessor.GetKeyOfInitConf()
	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("ReadInitConf Error: %v", err)
		return "", err
	}
	return value, nil
}

var RecoverInitNetwork = func() error {
	value, err := ReadInitConf()
	if err != nil {
		klog.Errorf("RecoverInitNetwork Error: ReadInitConf err[%v]", err)
		return err
	}
	cfg, err := jason.NewObjectFromBytes([]byte(value))
	if err != nil {
		klog.Errorf("RecoverInitNetwork Error: NewObjectFromBytes err[%v]", err)
		return err
	}
	cfgNet, err := cfg.GetObject("init_configuration")
	if err != nil {
		klog.Errorf("GetObject[init_configuration] Error: %v", err)
		return err
	}
	errHandle := HandleInitNetworks(cfgNet)
	if errHandle != nil {
		klog.Errorf("RecoverInitNetwork Error: HandleInitNetworks err[%v]", errHandle)
		return errHandle
	}
	return nil
}

func GetDefaultProviderNetworkByInitConfig(cfg *jason.Object) (string, error) {
	defaultPhysnet, err := cfg.GetString("configuration", "default_physnet")
	if err != nil {
		klog.Errorf("GetString[default_physnet] error: %v", err)
		return "", errors.New("get default_physnet error")
	}
	return defaultPhysnet, nil
}

func SaveAdminTenantInfoToDB(iaasID, iaasName string) error {
	adminTenant := Tenant{
		TenantName:     constvalue.PaaSTenantAdminDefaultUUID,
		TenantUUID:     constvalue.PaaSTenantAdminDefaultUUID,
		Networks:       dbaccessor.GetKeyOfNetworkGroup(constvalue.PaaSTenantAdminDefaultUUID),
		Interfaces:     dbaccessor.GetKeyOfInterfaceGroup(constvalue.PaaSTenantAdminDefaultUUID),
		Quota:          DefaultQuotaAdmin,
		IsCancelling:   false,
		CreateTime:     time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		IaasTenantID:   iaasID,
		IaasTenantName: iaasName,
	}
	return adminTenant.SaveTenantToEtcd()
}

func DelAdminTenantInfoFromDB() error {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(constvalue.PaaSTenantAdminDefaultUUID)
	return common.GetDataBase().DeleteLeaf(tenantKey)
}

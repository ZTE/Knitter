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

package iaas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/embedded"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/noauth_openstack"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/ZTE/Knitter/pkg/uuid"
	"github.com/ZTE/Knitter/pkg/version"
	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud"
	"strings"
	"sync"
	"time"
)

const authIntenal int64 = 5

func SetIaaS(tenantID string, i iaasaccessor.IaaS) error {
	GetIaasObjMgrSingleton().Add(tenantID, i)
	return nil
}

var GetIaaS = func(paasTenantID string) iaasaccessor.IaaS {
	iaasTenantID, _ := GetIaasTenantIDByPaasTenantID(paasTenantID)
	if iaasTenantID == "" {
		iaasObj, err := InitIaasForDefault(paasTenantID)
		if err != nil {
			return nil
		}
		return iaasObj.IaaSInterface
	}
	timeNow := time.Now().Unix()
	iaasObj, err := GetIaasObjMgrSingleton().Get(iaasTenantID)
	if err == nil {
		return iaasObj.IaaSInterface
	}
	isInitIaas := (err.Error() == constvalue.ErrOfIaasTenantNotExist) ||
		(err.Error() == constvalue.ErrOfIaasInterfaceNil && timeNow-iaasObj.TimeLastGetIaas >= authIntenal)
	if isInitIaas {
		err := InitIaasByKnitterJSON(iaasTenantID)
		if err != nil {
			return nil
		}
		iaasObjAfterInit, _ := GetIaasObjMgrSingleton().Get(iaasTenantID)
		return iaasObjAfterInit.IaaSInterface
	}
	return nil
}

func GetExternalPortName(iaasID, paasID, vmID, podID, portName string) string {
	klog.Infof("GetExternalPortName IaaS-ID[%v]PaaS-ID[%v]VM-ID[%v]POD-ID[%v]portName[%v]",
		iaasID, paasID, vmID, podID, portName)
	if len(iaasID) < 8 || len(paasID) < 8 || len(vmID) < 8 {
		return portName
	}
	rs := uuid.GetUUID8Byte(iaasID)
	rs += uuid.GetUUID8Byte(paasID)
	rs += uuid.GetUUID8Byte(vmID)
	if len(podID) > 8 {
		rs += uuid.GetUUID8Byte(podID)
	}
	rs += "_"
	rs += portName
	klog.Infof("GetExternalPortName OutPUT portName[%v]", rs)
	return rs
}

func GetOriginalPortName(longName string) string {
	klog.Infof("GetOriginalPortName Input:%s", longName)
	var oriName string
	const FixIPPortPrefixLen int = 17
	//IaaS-Tenants-UUID-8bit + PaaS-UUID-8bit + VM-UUID-8bit + "_"
	const SouthPortPrefixLen int = 25
	//IaaS-Tenants-UUID-8bit + PaaS-UUID-8bit + VM-UUID-8bit + POD-UUID-8bit + "_"
	const NorthPortPrefixLen int = 33

	if len(longName) < FixIPPortPrefixLen {
		oriName = longName
		klog.Infof("GetOriginalPortName Output:%s", oriName)
		return oriName
	}
	if len(longName) < SouthPortPrefixLen {
		oriName = longName[FixIPPortPrefixLen:]
		klog.Infof("GetOriginalPortName Output:%s", oriName)
		return oriName
	}
	if len(longName) < NorthPortPrefixLen {
		oriName = longName[SouthPortPrefixLen:]
		klog.Infof("GetOriginalPortName Output:%s", oriName)
		return oriName
	}
	oriName = longName[NorthPortPrefixLen:]
	klog.Infof("GetOriginalPortName Output:%s", oriName)
	return oriName
}

var CheckNetService = func(openstack iaasaccessor.IaaS) error {
	_, err := openstack.GetNetwork("this-is-a-error-uuid-for-auth")
	if err != nil && strings.Contains(err.Error(), "GetNetwork: socket-error") {
		return err
	}
	return nil
}

var InitIaaS = func() error {
	var err error

	cfg := common.GetOpenstackCfg()
	klog.Info("OpenStack Configration:", cfg)
	openstack := openstack.NewOpenstack()
	err = openstack.SetOpenstackConfig(cfg)
	if err != nil {
		klog.Warning("IaaS-SetOpenstackConfig-ERROR")
		return err
	}
	err = openstack.Auth()
	if err != nil {
		klog.Warning("IaaS-auth-ERROR")
		return err
	}
	err = CheckNetService(openstack)
	if err != nil {
		klog.Error("NetService-check-ERROR")
		return err
	}
	klog.Info("IaaS-auth-OK")
	SetIaaS(openstack.GetTenantID(), openstack)
	//id, _ := openstack.GetTenantUUID(cfg)
	//SetIaaSTenants(id)
	return nil
}

var CheckOpenstackConfig = func(conf *openstack.OpenStackConf) (*gophercloud.ProviderClient, error) {
	var config gophercloud.AuthOptions
	config.Username = conf.Username
	config.Password = conf.Password
	config.TenantName = conf.TenantName
	config.TenantID = conf.Tenantid
	config.IdentityEndpoint = conf.Url + "/tokens"
	c, err := openstack.AuthenticatedClientV2(config)
	if err != nil {
		klog.Error("CheckOpenstackConfig call AuthenticatedClient Error:", err)
		return nil, err
	}

	if c.TenantID == "" || c.TenantName == "" {
		klog.Error("openstack user info input error TenantID[", c.TenantID, "]TenantName[", c.TenantName, "]")
		return nil, errors.New("openstack user info input error")
	}

	return c, nil
}

var SaveDefaultPhysnet = func(defaultPhysnet string) error {
	key := dbaccessor.GetKeyOfDefaultPhysnet()
	klog.Infof("DefaultPhysnet: %v", defaultPhysnet)
	errE := common.GetDataBase().SaveLeaf(key, defaultPhysnet)
	if errE != nil {
		klog.Errorf("SaveLeaf:key[%v],value[%v]: %v", key, defaultPhysnet, errE)
		return errE
	}
	return nil
}

var GetDefaultPhysnet = func() (string, error) {
	key := dbaccessor.GetKeyOfDefaultPhysnet()
	physnet, errR := common.GetDataBase().ReadLeaf(key)
	if errR != nil {
		klog.Errorf("ReadLeaf:key[%v] Err: %v", key, errR)
		return "", errR
	}
	return physnet, nil
}

type IaaSObj struct {
	TimeLastGetIaas int64
	IaaSInterface   iaasaccessor.IaaS
}

type IaasObjMgr struct {
	Mng  map[string]IaaSObj
	Lock sync.RWMutex
}

var isMultipleIaasTenants bool = false
var iaasObjMgr IaasObjMgr = IaasObjMgr{Mng: make(map[string]IaaSObj)}

func SetMultipleIaasTenantsFlag(isMultiple bool) {
	isMultipleIaasTenants = isMultiple
}

func GetMultipleIaasTenantsFlag() bool {
	return isMultipleIaasTenants
}

func GetIaasObjMgrSingleton() *IaasObjMgr {
	return &iaasObjMgr
}

func (p *IaasObjMgr) Add(tenantID string, i iaasaccessor.IaaS) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Mng[tenantID] = IaaSObj{
		TimeLastGetIaas: time.Now().Unix(),
		IaaSInterface:   i}
	return nil
}

func (p *IaasObjMgr) Del(tenantID string) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if _, exist := p.Mng[tenantID]; exist {
		delete(p.Mng, tenantID)
		return nil
	}
	klog.Warningf("IaaS[%v] is not exist", tenantID)
	return nil
}

func (p *IaasObjMgr) Get(tenantID string) (*IaaSObj, error) {
	p.Lock.RLock()
	defer p.Lock.RUnlock()
	if _, exist := p.Mng[tenantID]; exist {
		if p.Mng[tenantID].IaaSInterface == nil {
			klog.Errorf("IaaS[%v] Interface is nil", tenantID)
			return nil, errors.New(constvalue.ErrOfIaasInterfaceNil)
		}
		ias := p.Mng[tenantID]
		return &ias, nil
	}

	klog.Errorf("IaaS[%v] is not exist", tenantID)
	return nil, errors.New(constvalue.ErrOfIaasTenantNotExist)
}

func SaveIaasTenantInfoToDB(conf *openstack.OpenStackConf) error {
	confBytes, err := json.Marshal(conf)
	if err != nil {
		klog.Errorf("SaveIaasTenantInfoToDB: json.Marshal(%v) FAILED, error: %v", conf, err)
		return errobj.ErrMarshalFailed
	}
	tenantKey := dbaccessor.GetKeyOfIaasTenantInfo(conf.Tenantid)
	return common.GetDataBase().SaveLeaf(tenantKey, string(confBytes))
}

func DelIaasTenantInfoFromDB(tenantid string) error {
	tenantKey := dbaccessor.GetKeyOfIaasTenantInfo(tenantid)
	return common.GetDataBase().DeleteLeaf(tenantKey)
}

func GetIaasTenantInfoFromDB(tenantid string) (string, error) {
	tenantKey := dbaccessor.GetKeyOfIaasTenantInfo(tenantid)
	tenantStr, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil {
		return "", err
	}
	return tenantStr, nil
}

func InitIaasByKnitterJSON(tenantID string) error {
	Scene, err := GetSceneByKnitterJSON()
	if err != nil {
		return err
	}
	switch Scene {
	case constvalue.TECS:
		return InitIaasForTECS(tenantID)
	case constvalue.EMBEDDED:
		return SetIaaS(constvalue.DefaultIaasTenantID,
			networkserver.GetEmbeddedNetwrokManager())
	case constvalue.VNM:
		cfg, err := version.GetConfObject(constvalue.KnitterJSONPath, "manager")
		if err != nil {
			klog.Errorf("GetConfObject Error: %v", err)
			return err
		}
		return InitNoAuth(cfg)
	default:
		klog.Errorf("Scene[%v] is err", Scene)
		return errors.New("Scene err")
	}
}

func InitIaasForTECS(tenantID string) error {
	opstkStr, err := GetIaasTenantInfoFromDB(tenantID)
	if err != nil {
		klog.Errorf("InitIaasForTecs[%v] err: %v", tenantID, err)
		return err
	}
	opstk := openstack.NewOpenstack()
	err = opstk.SetOpenstackConfig(opstkStr)
	if err != nil {
		klog.Errorf("IaaS-SetOpenstackConfig-ERROR")
		return err
	}
	err = opstk.Auth()
	if err != nil {
		klog.Errorf("IaaS-auth-ERROR")
		return err
	}
	err = CheckNetService(opstk)
	if err != nil {
		klog.Error("NetService-check-ERROR")
		return err
	}
	klog.Info("IaaS-auth-OK")
	SetIaaS(opstk.GetTenantID(), opstk)
	return nil
}

type PaasTenantWithIaasID struct {
	TenantName     string `json:"tenant_name"`
	TenantUUID     string `json:"tenant_uuid"`
	Networks       string `json:"networks"`
	Interfaces     string `json:"interfaces"`
	Quota          int    `json:"net_quota"`
	NetNum         int    `json:"net_number"`
	IsCancelling   bool   `json:"is_cancelling"`
	CreateTime     string `json:"create_time"`
	IaasTenantID   string `json:"iaas_tenant_id"`
	IaasTenantName string `json:"iaas_tenant_name"`
}

func GetIaasTenantIDByPaasTenantID(paasTenantID string) (string, error) {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(paasTenantID)
	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil {
		klog.Errorf("GetIaasTenantIDAndNameByPaasTenantID: ReadLeaf(key: %v) FAILED, error: %v", tenantKey, err)
		return "", err
	}

	var tenantInfo PaasTenantWithIaasID
	err = json.Unmarshal([]byte(tenantValue), &tenantInfo)
	if err != nil {
		klog.Errorf("GetIaasTenantIDAndNameByPaasTenantID: json.Unmarshal(%v) FAILED, error: %v", string(tenantValue), err)
		return "", err
	}
	return tenantInfo.IaasTenantID, nil
}

func SaveTenantInfoWithIaasIDToDB(paasTenantID, iaasTenantID, iaasTenantName string) error {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(paasTenantID)
	var tenantInfo PaasTenantWithIaasID
	if paasTenantID == constvalue.PaaSTenantAdminDefaultUUID {
		tenantInfo = PaasTenantWithIaasID{
			TenantUUID:     constvalue.PaaSTenantAdminDefaultUUID,
			TenantName:     constvalue.PaaSTenantAdminDefaultUUID,
			Quota:          100,
			CreateTime:     time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05Z"),
			IaasTenantID:   iaasTenantID,
			IaasTenantName: iaasTenantName,
		}
	} else {
		tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
		if err != nil {
			klog.Errorf("SaveTenantInfoWithIaasIDToDB: ReadLeaf(key: %v) FAILED, error: %v", tenantKey, err)
			return err
		}
		json.Unmarshal([]byte(tenantValue), &tenantInfo)
		tenantInfo.IaasTenantID = iaasTenantID
		tenantInfo.IaasTenantName = iaasTenantName
	}

	value, _ := json.Marshal(tenantInfo)
	err := common.GetDataBase().SaveLeaf(tenantKey, string(value))
	return err
}

func InitNoAuth(cfg *jason.Object) error {
	var err error
	klog.Info("Now-Use-NoAuth-Openstack")
	noauthCfg, err := cfg.GetObject("iaas", "noauth_openstack", "config")
	if err != nil {
		klog.Error("InitNoAuth: get noauth_openstack config error")
		return fmt.Errorf("%v:InitNoAuth: get noauth_openstack config error", err)
	}
	ip, err := noauthCfg.GetString("ip")
	if err != nil {
		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack ip does not exist")
		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack ip does not exist", err)
	} else if ip == "" {
		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack ip is blank string")
		return errors.New("initnoauth: in noauth_openstack init, noauth_openstack ip is blank string")
	}
	url, err := noauthCfg.GetString("url")
	if err != nil {
		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack url does not exist")
		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack url does not exist", err)
	}
	tenantID, err := noauthCfg.GetString("tenant_id")
	if err != nil {
		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack tenant_id does not exist")
		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack tenant_id does not exist", err)
	}

	defaultProviderNetwokType, err := noauthCfg.GetString("default_network_type")
	if err != nil {
		defaultProviderNetwokType = constvalue.DefaultProviderNetworkType
		klog.Infof("InitNoAuth: in noauth_openstack init, noauth_openstack default_network_type does not exist,"+
			" use default: %s", defaultProviderNetwokType)
	}
	klog.Infof("InitNoAuth: in noauth_openstack init, noauth_openstack default_network_type: %s",
		defaultProviderNetwokType)

	defalyPhysnet, err := GetDefaultPhysnet()
	if err != nil {
		klog.Infof("Frist init defailt physnet")
		defalyPhysnet = ""
	}

	openstackConf := noauth_openstack.NoauthOpenStackConf{IP: ip, Tenantid: tenantID, URL: url}
	neutronConf := noauth_openstack.NoauthNeutronConf{
		Port:   "9696",
		ApiVer: "v2.0",
		ProviderConf: noauth_openstack.DefaultProviderConf{
			PhyscialNetwork: defalyPhysnet,
			NetworkType:     defaultProviderNetwokType,
		},
	}
	InitNoauthOpenStack(openstackConf, neutronConf)

	return nil
}

func InitNoauthOpenStack(openstackConf noauth_openstack.NoauthOpenStackConf, neutronConf noauth_openstack.NoauthNeutronConf) {
	if openstackConf.Tenantid == "" {
		openstackConf.Tenantid = noauth_openstack.DefaultNoauthOpenStackTenantId
	}
	opn := noauth_openstack.NewNoauthOpenStack(openstackConf, neutronConf)
	for {
		err := CheckNetService(opn)
		if err != nil {
			klog.Warning("NetService-check-ERROR")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	klog.Infof("Iaas-auth-OK, opn: %v", &opn)
	SetIaaS(constvalue.DefaultIaasTenantID, opn)
	SaveDefaultPhysnet(neutronConf.ProviderConf.PhyscialNetwork)
}

var GetSceneByKnitterJSON = func() (string, error) {
	return constvalue.EMBEDDED, nil
}

func InitIaasForDefault(paasTenantID string) (*IaaSObj, error) {
	var Scene string = ""
	var err error = nil
	var iaasTenantID string = constvalue.DefaultIaasTenantID
	var iaasTenantName string = constvalue.DefaultIaasTenantName
	Scene, err = GetSceneByKnitterJSON()
	if err != nil {
		klog.Errorf("GetSceneByKnitterJSON err: %v", err)
		return nil, err
	}
	switch Scene {

	case constvalue.TECS:
		iaasTenantID, iaasTenantName, err = InitDefaultIaasForTECS(paasTenantID)
	case constvalue.EMBEDDED:
		SetIaaS(constvalue.DefaultIaasTenantID,
			networkserver.GetEmbeddedNetwrokManager())
	case constvalue.VNM:
		cfg, err := version.GetConfObject(constvalue.KnitterJSONPath, "manager")
		if err != nil {
			klog.Errorf("GetConfObject Error: %v", err)
			return nil, err
		}
		err = InitNoAuth(cfg)
	default:
		klog.Errorf("Scene[%v] is err", Scene)
		return nil, errors.New("Scene err")
	}
	if err != nil {
		return nil, err
	}
	SaveTenantInfoWithIaasIDToDB(paasTenantID,
		iaasTenantID, iaasTenantName)
	return GetIaasObjMgrSingleton().Get(iaasTenantID)
}

func InitDefaultIaasForTECS(paasTenantID string) (string, string, error) {
	cfgStr := common.GetOpenstackCfg()
	opstk := openstack.NewOpenstack()
	err := opstk.SetOpenstackConfig(cfgStr)
	if err != nil {
		klog.Errorf("IaaS-SetOpenstackConfig-ERROR")
		return "", "", err
	}
	err = opstk.Auth()
	if err != nil {
		klog.Errorf("IaaS-auth-ERROR")
		return "", "", err
	}
	err = CheckNetService(opstk)
	if err != nil {
		klog.Error("NetService-check-ERROR")
		return "", "", err
	}
	klog.Infof("IaaS-auth-OK")
	iaasTenantID := opstk.GetTenantID()
	iaasTenantName := opstk.GetTenantName()
	SetIaaS(iaasTenantID, opstk)
	tenantKey := dbaccessor.GetKeyOfIaasTenantInfo(iaasTenantID)
	iaasInfoStr, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil || iaasInfoStr == "" {
		errS := common.GetDataBase().SaveLeaf(tenantKey, cfgStr)
		if errS != nil {
			klog.Errorf("Save iaas info to DB err: %v", errS)
			return iaasTenantID, iaasTenantName, errors.New("fail to save iaas info to DB")
		}
	}
	return iaasTenantID, iaasTenantName, nil
}

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
	"encoding/json"
	"fmt"
	NETHTTP "net/http"
	"strconv"
	"strings"
	"time"

	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/embedded"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/noauth_openstack"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/antonholmquist/jason"
)

type EncapOpenStack struct {
	Config *openstack.OpenStackConf `json:"openstack"`
}

type VnpmConf struct {
	URL string `json:"url"`
}

type EncapNoauthOpenStackConf struct {
	NoauthOpenStack noauth_openstack.NoauthOpenStackConf `json:"noauth_openstack"`
}

func SaveOpenStackConfg(value []byte) error {
	key := dbaccessor.GetKeyOfOpenstack()
	err1 := common.GetDataBase().SaveLeaf(key, string(value))
	if err1 != nil {
		klog.Error(err1)
		return err1
	}
	WaitForEtcdClusterSync(2)
	return nil
}

var DelOpenStackConfg = func() error {
	key := dbaccessor.GetKeyOfOpenstack()
	err1 := common.GetDataBase().DeleteLeaf(key)
	if err1 != nil {
		klog.Error(err1)
		return err1
	}
	return nil
}

var SaveVnfmConfg = func(value []byte) error {
	key := dbaccessor.GetKeyOfVnfm()
	err1 := common.GetDataBase().SaveLeaf(key, string(value))
	if err1 != nil {
		klog.Error(err1)
		return err1
	}
	WaitForEtcdClusterSync(2)
	return nil
}

func HandleRegularCheckConfg(timeInterval string) error {
	key := dbaccessor.GetKeyOfRecycleResourceByTimerUrl()
	err := common.GetDataBase().SaveLeaf(key, string(timeInterval))
	if err != nil {
		klog.Error("HandleRegularCheckConfg call GetDataBase().SaveLeaf ERROR", err)
		return BuildErrWithCode(NETHTTP.StatusInternalServerError, err)
	}
	klog.Info("Configration RegularCheck interval Time:[", timeInterval, "] OK!")
	return nil
}

func WaitForEtcdClusterSync(second int) {
	klog.Infof("Wait %d Second for etcd cluster sync openstack info", second)
	time.Sleep(time.Duration(second) * time.Second)
}

func HandleOpenStackConfg(conf *openstack.OpenStackConf) (*openstack.OpenStackConf, error) {
	var err error

	client, err := iaas.CheckOpenstackConfig(conf)
	if err != nil {
		klog.Error("CheckOpenstackConfig Error:", err)
		return nil, BuildErrWithCode(NETHTTP.StatusBadRequest, err)
	}
	cfg := *conf
	cfg.Tenantid = client.TenantID
	cfg.TenantName = client.TenantName
	value, _ := json.Marshal(cfg)
	err = SaveOpenStackConfg(value)
	if err != nil {
		klog.Warning("SaveOpenStackConfg Error:", err)
		return nil, BuildErrWithCode(NETHTTP.StatusInternalServerError, err)
	}
	err = iaas.SaveIaasTenantInfoToDB(&cfg)
	if err != nil {
		klog.Errorf("SaveIaasTenantInfoToDB err: %v", err)
		return nil, BuildErrWithCode(NETHTTP.StatusInternalServerError, err)
	}
	err = SaveAdminTenantInfoToDB(cfg.Tenantid, cfg.TenantName)
	if err != nil {
		klog.Errorf("SaveAdminTenantInfoToDB err: %v", err)
		return nil, BuildErrWithCode(NETHTTP.StatusInternalServerError, err)
	}

	var req int = 0
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) != nil {
		req = iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetAttachReq()
	}
	iaas.InitIaaS()
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) != nil {
		iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).SetAttachReq(req)
	}
	return &cfg, nil
}

//var CheckVnfmConfig = func(conf vnfm.VnfmConf) error {
//	delURL := dbaccessor.GetVnfmDeletePortUrl(conf.URL, conf.NfInstanceID, "testnwport")
//	statusCode, _ := http.GetHTTPClientObj().Delete(delURL)
//	if statusCode != 404 {
//		return errors.New("checkvnfmconfig err")
//	}
//	return nil
//}

func saveAdminTenantInfoWithIaasTenantID() {
	iaasTenantID, _ := iaas.GetIaasTenantIDByPaasTenantID(constvalue.PaaSTenantAdminDefaultUUID)
	if iaasTenantID == "" {
		iaas.SaveTenantInfoWithIaasIDToDB(constvalue.PaaSTenantAdminDefaultUUID,
			constvalue.DefaultIaasTenantID, constvalue.DefaultIaasTenantName)
	}
}

func initIaas(cfg *jason.Object) error {
	//common.SetIaaSTenants(uuid.NIL.String())
	embedded, err := cfg.GetBoolean("iaas", "embedded")
	if err == nil && embedded == true {
		klog.Info("Now-Use-embedded-network-server")
		iaas.SetIaaS(constvalue.DefaultIaasTenantID,
			networkserver.GetEmbeddedNetwrokManager())
		saveAdminTenantInfoWithIaasTenantID()
		return nil
	}
	return nil
}

func InitEnv4Manger(confObj *jason.Object) error {
	defer klog.Flush()
	etcdAPIVer, err := confObj.GetInt64("etcd", "api_version")
	if err != nil {
		etcdAPIVer = int64(etcd.DefaultEtcdAPIVersion)
		klog.Warningf("InitEnv4Manger: get etcd api verison error: %v, use default: %d", err, etcdAPIVer)
	} else {
		klog.Infof("InitEnv4Manger: get etcd api verison: %d", etcdAPIVer)
	}

	etcdURL, _ := confObj.GetString("etcd", "urls")
	if etcdURL == "" {
		klog.Errorf("InitEnv4Manger: etcd urls is null")
		return errors.New("InitEnv4Manger: etcd urls is null")
	}
	klog.Info("InitEnv4Manger: etcd urls:", etcdURL)

	interval, _ := confObj.GetString("interval", "seconds")
	serviceURL, err3 := confObj.GetString("self_service", "url")
	isMultiple, _ := confObj.GetBoolean("multiple_iaas_tenants")

	klog.Info("ETCD URL:", etcdURL)
	klog.Info("serviceUrl :", serviceURL)
	if err3 != nil {
		klog.Error("get Input param from config file ERROR")
		return fmt.Errorf("%v:Configration ERROR not find self_service-platform", err3)
	}
	common.SetDataBase(etcd.NewEtcdWithRetry(int(etcdAPIVer), etcdURL))
	common.CheckDB()
	iaas.SetMultipleIaasTenantsFlag(isMultiple)

	err = common.RegisterSelfToDb(serviceURL)
	if err != nil {
		klog.Errorf("register knitter_master self to etcd failed, error: %v", err)
		return fmt.Errorf("%v:register knitter_master self to etcd error", err)
	}

	err = initIaas(confObj)
	if err != nil {
		klog.Warningf("initIaas err: %v", err)
	}

	GetSyncMgt().SetInterval(interval)
	SetNetQuota(confObj)
	UpdateEtcd4NetQuota()

	LoadAllResourcesToCache()
	CancelResidualTenants()
	return nil
}

func LoadAllResourcesToCache() {
	for _, lroFunc := range LoadResourceObjectFuncs {
		LoadResouceObjectsLoop(lroFunc)
	}
	klog.Infof("LoadAllResourcesToCache: load all type resource SUCC")
}

var LoadResourceObjectFuncs = []LoadResouceObjectFunc{
	LoadAllPortObjs,
	LoadPhysicalPortObjects,
	LoadAllNetworkObjects,
	LoadAllSubnetObjects,
	LoadAllIPGroupObjects}

func LoadResouceObjectsLoop(lroFunc LoadResouceObjectFunc) {
	for {
		var err error
		err = lroFunc()
		if err == nil || IsKeyNotFoundError(err) {
			klog.Infof("LoadResouceObjectsLoop: lroFunc[%v] SUCC", lroFunc)
			return
		}
		klog.Infof("LoadResouceObjectsLoop: lroFunc[%v] error: %v, just wait retry", lroFunc, err)
		time.Sleep(constvalue.GetLoadReourceRetryIntervalInSec * time.Second)
	}
}

func ConvertAttachReqMax(maxReqAttach string) int {
	if maxReqAttach == "" {
		return openstack.MaxReqForAttach
	}
	attachReq, err := strconv.Atoi(maxReqAttach)
	if err != nil {
		return openstack.MaxReqForAttach
	}
	if attachReq <= 0 || attachReq > 30 {
		return openstack.MaxReqForAttach
	}
	return attachReq
}

//func InitNoAuth(cfg *jason.Object) error {
//	var err error
//	defer func() {
//		var evt event.Event
//		target := event.EventObject{Uuid: "keystone", Name: "keystone", Kind: "keystone"}
//		source := event.EventObject{Kind: "nwmaster", Name: "nwmaster-" + common.GetManagerUUID(), Uuid: "nwmaster-" + common.GetManagerUUID()}
//		if err != nil {
//			evt = event.NewPlatformEvent("OpenstackAuthFailed.", event.BuildError(err.Error()), target, source)
//
//		} else {
//			evt = event.NewPlatformEvent("OpenstackAuthSuccessfully.", "Openstack auth successfully.", target, source)
//		}
//		evt.Report()
//	}()
//	klog.Info("Now-Use-NoAuth-Openstack")
//	noauthCfg, err := cfg.GetObject("iaas", "noauth_openstack", "config")
//	if err != nil {
//		klog.Error("InitNoAuth: get noauth_openstack config error")
//		return fmt.Errorf("%v:InitNoAuth: get noauth_openstack config error", err)
//	}
//	ip, err := noauthCfg.GetString("ip")
//	if err != nil {
//		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack ip does not exist")
//		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack ip does not exist", err)
//	} else if ip == "" {
//		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack ip is blank string")
//		return errors.New("initnoauth: in noauth_openstack init, noauth_openstack ip is blank string")
//	}
//	url, err := noauthCfg.GetString("url")
//	if err != nil {
//		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack url does not exist")
//		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack url does not exist", err)
//	}
//	tenantID, err := noauthCfg.GetString("tenant_id")
//	if err != nil {
//		klog.Error("InitNoAuth: in noauth_openstack init, noauth_openstack tenant_id does not exist")
//		return fmt.Errorf("%v:InitNoAuth: in noauth_openstack init, noauth_openstack tenant_id does not exist", err)
//	}
//
//	defaultProviderNetwokType, err := noauthCfg.GetString("default_network_type")
//	if err != nil {
//		defaultProviderNetwokType = constvalue.DefaultProviderNetworkType
//		klog.Infof("InitNoAuth: in noauth_openstack init, noauth_openstack default_network_type does not exist,"+
//			" use default: %s", defaultProviderNetwokType)
//	}
//	klog.Infof("InitNoAuth: in noauth_openstack init, noauth_openstack default_network_type: %s",
//		defaultProviderNetwokType)
//
//	defalyPhysnet, err := GetDefaultPhysnet()
//	if err != nil {
//		klog.Infof("Frist init defailt physnet")
//		defalyPhysnet = ""
//	}
//
//	openstackConf := noauth_openstack.NoauthOpenStackConf{IP: ip, Tenantid: tenantID, URL: url}
//	neutronConf := noauth_openstack.NoauthNeutronConf{
//		Port:   "9696",
//		ApiVer: "v2.0",
//		ProviderConf: noauth_openstack.DefaultProviderConf{
//			PhyscialNetwork: defalyPhysnet,
//			NetworkType:     defaultProviderNetwokType,
//		},
//	}
//	common.InitNoauthOpenStack(openstackConf, neutronConf)
//
//	return nil
//}

//func InitAuth(cfg *jason.Object) {
//	klog.Info("Now-Use-Auth-Openstack")
//	common.InitIaaS()
//	if common.GetIaaS() != nil {
//		maxReqAttach, _ := cfg.GetString("max_req_attach")
//		common.GetIaaS().SetAttachReq(ConvertAttachReqMax(maxReqAttach))
//	}
//	return
//}

func SetNetQuota(cfg *jason.Object) {
	adminQuota, _ := cfg.GetString("net_quota", "admin")
	noAdminQuota, _ := cfg.GetString("net_quota", "no_admin")
	QuotaAdmin, _ = ConvertQuota(adminQuota, DefaultQuotaAdmin)
	QuotaNoAdmin, _ = ConvertQuota(noAdminQuota, DefaultQuotaNoAdmin)
	klog.Infof("QuotaAdmin:%v, QuotaNoAdmin:%v", QuotaAdmin, QuotaNoAdmin)
}

func ConvertQuota(quota string, defaultValue int) (int, bool) {
	if quota == "" {
		return defaultValue, false
	}
	quotaInt, err := strconv.Atoi(quota)
	if err != nil {
		return defaultValue, false
	}
	if quotaInt <= 0 || quotaInt > MaxQuota {
		return defaultValue, false
	}
	return quotaInt, true
}

func UpdateEtcd4NetQuota() error {
	tenantsURL := dbaccessor.GetKeyOfTenants()
	nodes, err := common.GetDataBase().ReadDir(tenantsURL)
	if err != nil {
		klog.Warning("UpdateEtcd4NetQuota:Read Tenants dir[",
			tenantsURL, "] from ETCD Error:", err)
		return err
	}
	for _, node := range nodes {
		tenantID := strings.TrimPrefix(node.Key, tenantsURL+"/")
		if tenantID == "admin" {
			continue
		}
		tenant := &Tenant{Quota: 0}
		tenantKey := dbaccessor.GetKeyOfTenantSelf(tenantID)
		tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
		if err == nil {
			err = json.Unmarshal([]byte(tenantValue), tenant)
			if err == nil && tenant.Quota == 0 {
				tenant.Quota = QuotaNoAdmin
				tenant.NetNum = GetNetNumOfTenant(tenantID)
				value, _ := json.Marshal(tenant)
				err = common.GetDataBase().SaveLeaf(tenantKey, string(value))
				if err != nil {
					klog.Error("UpdateEtcd4NetQuota:Save Tenants key[",
						tenantKey, "] to ETCD Error:", err)
					continue
				}
				klog.Infof("UpdateEtcd4NetQuota Successful tenantID:%v,Quota:%v,NetNum:%v",
					tenantID, tenant.Quota, tenant.NetNum)
			}
		} else {
			klog.Error("UpdateEtcd4NetQuota:Read Tenants Leaf[",
				tenantKey, "] to ETCD Error:", err)
			continue
		}
	}
	return nil
}

var GetNetNumOfTenant = func(tenantID string) int {
	netObjs, err := GetNetObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		klog.Warningf("GetNetNumOfTenant: ListByTenantID[tenantID: %s] FAIL, error:%v", tenantID, err)
		return 0
	}
	return len(netObjs)
}

func BuildErrWithCode(code int, err error) error {
	status := NETHTTP.StatusText(code)
	if status == "" {
		return fmt.Errorf("%v::%v", NETHTTP.StatusInternalServerError, err)
	}
	return fmt.Errorf("%v::%v", code, err)
}

func IsKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, "Key not found") {
		klog.Error(err)
		return true
	}
	return false
}

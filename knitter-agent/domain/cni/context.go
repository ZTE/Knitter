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
	"errors"
	"strings"

	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"

	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/etcd"

	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/knitter-agent/infra/k8s"
	"github.com/ZTE/Knitter/pkg/leveldb"
	"time"
)

const DefaultSyncVnisIntvalInMin = 60

type Range struct {
	Start int
	End   int
}

type AgentContext struct {
	Mtu                string
	VMID               string
	ClusterID          string
	oseToken           string
	HostType           string
	RunMode            string
	SyncSwitch         bool
	PaasNwConfPath     string
	PhysnetPfMap       *map[string]string
	DefaultMaxVfs      int
	VfNumMap           *map[string]int
	VfRanges           *map[string]Range
	VfRangeConfigured  bool
	ExternalIP         string
	SendVdp            bool
	ClusterType        string
	ClusterUUID        string
	HostIP             string
	SyncVniFlag        bool
	SyncVniIntvalInMin int64
	AdminTenantUUID    string
	DB                 dbaccessor.DbAccessor
	RemoteDB           dbaccessor.DbAccessor
	Mc                 manager.ManagerClient
	MtrC               monitor.MonitorClient
	K8s                k8s.K8sClient
}

type BondInfo struct {
	PhyNw      string
	BondType   string
	BondPair   []string
	BondMaster string
}

var BondList map[string]BondInfo

var (
	ctx AgentContext
)

var GetGlobalContext = func() *AgentContext {
	return &ctx
}

func GetBondInfo(phyNw string) (BondInfo, error) {
	info, ok := BondList[phyNw]
	if ok {
		return info, nil
	}
	return BondInfo{}, errors.New("bond info is not config")
}

func InitConfigration4Agent(cfg *jason.Object) error {
	klog.Infof("InitEnv4Agent: Init etcd client!")
	etcdAPIVer, err := cfg.GetInt64("etcd", "api_version")
	if err != nil {
		etcdAPIVer = int64(etcd.DefaultEtcdAPIVersion)
		klog.Warningf("InitConfigration4Agent: get etcd api verison error: %v, use default: %d", err, etcdAPIVer)
	} else {
		klog.Infof("InitConfigration4Agent: get etcd api verison: %d", etcdAPIVer)
	}

	urls, _ := ctx.GetEtcdServerURLs(cfg)
	ctx.RemoteDB = etcd.NewEtcdWithRetry(int(etcdAPIVer), urls)

	db, err := leveldb.NewLevelDBClient(constvalue.LocalDBDataDir)
	if err != nil {
		klog.Errorf("InitConfigration4Agent: NewLevelDBClient(%s) error: %v", constvalue.LocalDBDataDir, err)
		return err
	}
	ctx.DB = db

	ctx.MtrC = monitor.MonitorClient{}
	err = ctx.MtrC.InitClient(cfg)
	if err != nil {
		klog.Errorf("InitEnv4Agent:Init monitor error! Error:-%v", err)
		return fmt.Errorf("%v:InitEnv4Agent:Init monitor error", err)
	}
	klog.Infof("InitEnv4Agent: Init knitter_manager client!")
	ctx.Mc = manager.ManagerClient{}
	manageriniterr := ctx.Mc.InitClient(cfg)
	if manageriniterr != nil {
		klog.Errorf("InitEnv4Agent:Init cni manager error! Error:-%v", manageriniterr)
		return fmt.Errorf("%v:InitEnv4Agent:Init manager error", manageriniterr)
	}
	for {
		checkErr := ctx.Mc.CheckKnitterManager()
		if checkErr != nil {
			klog.Errorf("InitEnv4Agent:CheckKnitterManager error! -%v",
				checkErr)
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	klog.Infof("InitEnv4Agent: Init cluster type!")
	clusterType, _ := ctx.GetClusterType(cfg)
	ctx.ClusterType = clusterType
	klog.Infof("InitEnv4Agent:Init k8s client!")
	k8sURL, _ := ctx.GetK8sServerURL(cfg)
	ctx.K8s = k8s.K8sClient{ServerURL: k8sURL}

	ctx.ClusterUUID, _ = cfg.GetString("cluster_uuid")
	klog.Infof("InitEnv4Agent:get ClusterUUID = :", ctx.ClusterUUID)

	ctx.SetRunMode(cfg)
	ctx.SetSyncSwitch(cfg)

	err = ctx.SetVMID(cfg)
	if err != nil {
		klog.Errorf("InitEnv4Agent: setVmId error! -%v", err)
		return fmt.Errorf("%v:InitEnv4Agent: setVmId error", err)
	}

	clusterErr := ctx.SetClusterID(cfg)
	if clusterErr != nil {
		klog.Errorf("InitEnv4Agent:setTenantId error! -%v", clusterErr)
		return fmt.Errorf("%v:InitEnv4Agent:setTenantId error", clusterErr)
	}

	err = ctx.SetHostType(cfg)
	if err != nil {
		klog.Errorf("InitEnv4Agent:setHostType error!-%v", err)
		return fmt.Errorf("%v:InitEnv4Agent:setHostType  error", err)
	}

	mtu, err := cfg.GetString("host", "mtu")
	if err != nil {
		klog.Warningf("InitEnv4Agent: get mtu FAILED, error: %v, use default 1500 byte", err)
		mtu = constvalue.DefaultMtu
	}
	ctx.Mtu = mtu
	klog.Infof("InitEnv4Agent:get MTU = :", ctx.Mtu)

	ctx.SendVdp, err = cfg.GetBoolean("sdn_proxy", "send_vdp")
	if err != nil {
		klog.Errorf("InitEnv4Agent:get send_vdp error: %v, set to default true", err)
		ctx.SendVdp = true
	}
	klog.Infof("InitEnv4Agent:get SendVdp = :", ctx.SendVdp)

	ctx.HostIP, _ = cfg.GetString("host", "ip")
	klog.Infof("InitEnv4Agent:get HostIp = :", ctx.HostIP)

	syncVniFlag, err := cfg.GetBoolean("sdn_proxy", "sync_vni")
	if err != nil {
		klog.Errorf("InitEnv4Agent:get sync_vni error: %v, set to default false", err)
		syncVniFlag = false
	}
	klog.Errorf("InitEnv4Agent:get sync_vni %v", syncVniFlag)
	ctx.SyncVniFlag = syncVniFlag

	syncIntval, err := cfg.GetInt64("sdn_proxy", "sync_intval_mins")
	if err != nil {
		klog.Errorf("InitEnv4Agent:get sync_intval error: %v, set to default time: %d",
			err, DefaultSyncVnisIntvalInMin)
		syncIntval = DefaultSyncVnisIntvalInMin
	}
	klog.Errorf("InitEnv4Agent:get sync_vni %v", syncIntval)
	ctx.SyncVniIntvalInMin = syncIntval

	err = ctx.SetExternalIP(cfg)
	if err != nil {
		klog.Errorf("InitEnv4Agent:SetExternalIp error!-%v", err)
		return fmt.Errorf("%v:InitEnv4Agent:SetExternalIp  error", err)
	}
	//todo to delete
	ctx.AdminTenantUUID = constvalue.PaaSTenantAdminDefaultUUID

	return nil
}

func (self *AgentContext) GetOseServerURL(cfg *jason.Object) (string, error) {
	OseURL, _ := cfg.GetString("k8s", "url")
	if OseURL == "" {
		klog.Errorf("getOseServerUrl:ose url is null")
		return "", errors.New("getOseServerUrl:ose url is null")
	}
	url := strings.Replace(OseURL, "http:", "https:", -1)
	klog.Info("getOseServerUrl:ose url:", url)
	return url, nil
}

func (self *AgentContext) GetK8sServerURL(cfg *jason.Object) (string, error) {
	K8sURL, _ := cfg.GetString("k8s", "url")
	if K8sURL == "" {
		klog.Errorf("GetK8sServerUrl:ose url is null")
		return "", errors.New("getK8sServerUrl:ose url is null")
	}
	klog.Info("GetK8sServerUrl:ose url:", K8sURL)
	return K8sURL, nil
}

func (self *AgentContext) SetClusterID(cfg *jason.Object) error {
	k8sURL, _ := cfg.GetString("k8s", "url")
	klog.Infof("serverconfjson.GetString k8s url: %v", k8sURL)
	if k8sURL == "" {
		klog.Errorf("GetK8SUrlFromConfigure:k8s url is null")
		return errors.New("k8s url is null")
	}
	urlStr := strings.Replace(k8sURL, "http://", "", -1)
	urlArray := strings.Split(urlStr, ":")
	self.ClusterID = urlArray[0]
	klog.Infof("K8s Cluster ID[%v]", self.ClusterID)
	return nil
}

func (self *AgentContext) SetMTU(cfg *jason.Object) error {
	//Get mtu param from Netconf and add to portinfo
	self.Mtu = "1400"
	//Get mtu param from Netconf and add to portinfo
	mtu, err := cfg.GetString("mtu")

	if err != nil {
		klog.Errorf("setMTU:GetString error! -%v", err)
		return fmt.Errorf("%v:setMTU:GetString error", err)
	}
	klog.Info("setMTU:New Pod k8s MTU:", mtu)
	self.Mtu = mtu
	return nil
}

func (self *AgentContext) CheckHostType(hostType string) error {
	switch hostType {
	case "virtual_machine":
		return nil
	case "bare_metal":
		return nil
	}
	return errors.New("CheckHostType ERROR:" + hostType)
}

func (self *AgentContext) SetRunMode(cfg *jason.Object) error {
	self.RunMode = "overlay"
	klog.Info("setRanMode: RunMode :", self.RunMode)
	return nil
}

func (self *AgentContext) SetSyncSwitch(cfg *jason.Object) error {
	syncSwitch, err := cfg.GetBoolean("run_mode", "sync")
	if err == nil && syncSwitch == true {
		self.SyncSwitch = true
	} else {
		self.SyncSwitch = false
	}
	klog.Info("setRanMode: SyncSwitch :", self.SyncSwitch)
	return nil
}

func (self *AgentContext) SetHostType(cfg *jason.Object) error {
	hostType, errHostType := cfg.GetString("host", "type")
	vmID, errVMID := cfg.GetString("host", "vm_id")
	if (errHostType != nil) && (errVMID == nil) && (vmID != "") {
		hostType = "virtual_machine"
	} else if errHostType != nil {
		klog.Error("setHostType:GetString ERROR", errHostType)
		return fmt.Errorf("%v:setHostType:GetString ERROR", errHostType)
	}
	klog.Info("setHostType: hostType :", hostType)
	err := self.CheckHostType(hostType)
	if err != nil {
		klog.Error("setHostType:CheckHostType Error:", err)
		return fmt.Errorf("%v:setHostType:CheckHostType Error", err)
	}
	self.HostType = hostType
	return nil
}

func (self *AgentContext) SetVMID(cfg *jason.Object) error {
	vmID, err := cfg.GetString("host", "vm_id")
	if err != nil {
		klog.Errorf("setVmId:vmId err")
		return fmt.Errorf("%v:vmId err", err)
	}
	if vmID == "" && self.HostType == "virtual_machine" {
		klog.Errorf("setVmId:vmId is null")
		return errors.New("vmId is null")
	}
	klog.Info("setVmId: vmId :", vmID)
	self.VMID = vmID
	return nil
}

func (self *AgentContext) GetOseToken(oseID string) (string, error) {
	oseURL := dbaccessor.GetKeyOfOseToken(oseID)
	oseResp, oseErr := self.RemoteDB.ReadLeaf(oseURL)
	if oseErr != nil {
		klog.Errorf("setOseToken:ReadLeaf error! -%v", oseErr)
		return "", oseErr
	}
	oseJSON, _ := jason.NewObjectFromBytes([]byte(oseResp))
	oseToken, err := oseJSON.GetString("token")
	if err != nil {
		klog.Errorf("setOseToken: GetString token error!")
		return "", fmt.Errorf("%v:setOseToken: GetString error", err)
	}
	klog.Info("setTenantId: tenantid :", oseToken)
	return oseToken, nil
}

func (self *AgentContext) SetPathOfPaasNwConf(cfg *jason.Object) error {
	paaNwConfPath, err := cfg.GetString("phy", "net_cfg")
	if err != nil {
		klog.Errorf("setPathOfPaasNwConf:GetString error! -%v", err)
		return fmt.Errorf("%v:setPathOfPaasNwConf:GetString error", err)
	}
	self.PaasNwConfPath = paaNwConfPath
	return nil
}

func (self *AgentContext) GetEtcdServerURLs(cfg *jason.Object) (string, error) {
	urls, _ := cfg.GetString("etcd", "urls")
	if urls == "" {
		klog.Errorf("getEtcdServerUrl:urls is null")
		return "", errors.New("getEtcdServerUrl:urls is null")
	}
	klog.Info("getEtcdServerUrl:etcd urls:", urls)
	return urls, nil
}

func (self *AgentContext) SetExternalIP(cfg *jason.Object) error {
	if infra.IsEnhancedMode() {
		return nil
	}
	externalIP, err := cfg.GetString("external", "ip")
	if err != nil {
		klog.Errorf("SetExternalIp:GetString error! -%v", err)
		return fmt.Errorf("%v:SetExternalIp:GetString error", err)
	}
	self.ExternalIP = externalIP
	klog.Info("SetExternalIp: external ip :", externalIP)
	return nil
}

func (self *AgentContext) GetClusterType(cfg *jason.Object) (string, error) {
	clusterType, _ := cfg.GetString("cluster_type")
	if clusterType == "" {
		klog.Errorf("GetClusterType:cluster_type is null")
		return "", errors.New("getClusterType:cluster_type is null")
	}
	klog.Infof("GetClusterType:cluster_type is: %v", clusterType)
	return clusterType, nil
}

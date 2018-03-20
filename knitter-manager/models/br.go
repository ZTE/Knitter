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
	"io"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"net/http"
	"strings"

	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/antonholmquist/jason"
	"io/ioutil"
	"time"
)

const (
	adminTenantID string = "admin"
)

type StorDestConf struct {
	StorType string `json:"storageType"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	UserNm   string `json:"user"`
	Passwd   string `json:"password"`
	Path     string `json:"path"`
}

type RestoreReq struct {
	JobID         string       `json:"jobDn"` // backup request ID
	Type          int          `json:"type"`  // 1: fullsize backup, 2: minisize backup
	StorConf      StorDestConf `json:"storage"`
	DirectoryName string       `json:"directoryName"`
}

type RestoreJobState struct {
	JobDn string
	State int
}

var NwRestore = func(RestoreReq RestoreReq) error {
	jobDn := RestoreReq.JobID
	klog.Infof("NwRestore: restore jobDn: %s START", jobDn)
	klog.Infof("NwRestore: jobDn: %s start to judge pod not exist", jobDn)
	err := ShouldNotPodExist()
	if err != nil {
		klog.Errorf("NwRestore: ShouldNotPodExist FAILED, error: %v", err)
		return err
	}

	klog.Infof("NwRestore: jobDn: %s start to clear pod and network", jobDn)
	err = ClearAllPortAndNetwork()
	if err != nil {
		klog.Errorf("NwRestore: ClearPodAndNetwork FAILED, error: %v", err)
	}

	klog.Infof("NwRestore: jobDn: %s start to restore from etcd-admtool", jobDn)
	err = RestoreFromEtcd(RestoreReq)
	if err != nil {
		klog.Errorf("NwRestore: RestoreFromETcd FAILED, error: %v", err)
		return err
	}
	err = RestoreTenants()
	if err != nil {
		klog.Errorf("NwRestore: RestoreTenants FAILED, error: %v", err)
		return err
	}

	klog.Infof("NwRestore: jobDn: %s start to RestorePublic", jobDn)
	err = RestorePublic()
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("NwRestore: RestorePublic FAILED, error: %v", err)
		return err
	}

	klog.Infof("NwRestore: jobDn: %s start to ClearDcs", jobDn)
	err = ClearDcs()
	if err != nil {
		klog.Errorf("NwRestore: ClearDcs FAILED, error: %v", err)
		return err
	}
	err = RecoverInitNetwork()
	if err != nil {
		klog.Errorf("NwRestore: RecoverInitNetwork() FAILED, error : %v", err)
		return err
	}
	klog.Infof("NwRestore: restore jobDn: %s SUCC", jobDn)
	return nil

}

var ShouldNotPodExist = func() error {
	klog.Infof("shouldNotPodExist: START")
	tenantIds, err := getAllTenantIds()
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("ShouldPodNotExist: getAllTenantIds FAILED, error: %v", err)
		return err
	}
	for _, tid := range tenantIds {
		pod := &Pod{}
		pod.SetTenantID(tid)
		pods := PodListAll(pod)
		if len(pods) > 0 {
			return errobj.ErrTenantHasPodsInUse
		}
	}

	klog.Infof("shouldNotPodExist SUCC")
	return nil
}
var ClearAllPortAndNetwork = func() error {
	klog.Info("ClearAllPodAndNetwork START")
	tenantIds, err := getAllTenantIds()
	if err != nil {
		if IsKeyNotFoundError(err) {
			klog.Infof("ClearAllPortAndNetwork: getAllTenantIds is nil, clear SUCC")
			return nil
		}
		klog.Errorf("ClearAllPortAndNetwork: gerAllTenantIds FAILED, ERROR: %v", err)
		return err
	}
	for _, tid := range tenantIds {
		err = ClearIPGroups(tid)
		if err != nil {
			klog.Errorf("ClearAllPortAndNetwork: ClearIPGroups(tenantId: %s) FAILED, error: %v",
				tid, err)
			return err
		}

		err = ClearLogicalPorts(tid)
		if err != nil {
			klog.Errorf("ClearAllPortAndNetwork: ClearLogicalPorts(tenantId: %s) FAILED, error: %v",
				tid, err)
			return err
		}
		err = ClearNetworks(tid)
		if err != nil {
			klog.Errorf("ClearAllPortAndNetwork: ClearPrivateNetworks(tenantId: %s) FAILED, error: %v", tid, err)
			return err
		}

	}
	klog.Info("ClearAllPortAndNetwork SUCC")
	return nil

}

var RestoreFromEtcd = func(etcdRestorejob RestoreReq) error {
	klog.Infof("RestoreFromEtcd: RestoreFromEtcd START")

	klog.Infof("Preparing to send restore Post ")
	//todo should get msb from config file
	msbURL, err := GetMsbURL()
	if err != nil {
		klog.Errorf("RestoreFromEtcd: getMsbUrl() FAILED, ERROR: %v", err)
		return err
	}
	etcdAdmToolRestoreURL := GetEtcdAdmToolRestoreURL(msbURL)
	restoreJobStateMap := MakeRestoreStateMap(etcdRestorejob)
	klog.Info("Send restore jobs to etcdAdmTool")
	SendRestoreJobs(etcdAdmToolRestoreURL, etcdRestorejob, restoreJobStateMap)

	time.Sleep(1 * time.Second)
	klog.Info("Get restore jobs states")
	GetRestoreJobStates(etcdAdmToolRestoreURL, restoreJobStateMap)

	err = CheckRestoreJobState(restoreJobStateMap)
	if err != nil {
		klog.Errorf("RestoreFromEtcd : RestoreFromEtcd FAILED,error: %v", err)
		return err
	}
	klog.Infof("RestoreFromEtcd: RestoreFromEtcd SUCC")
	return nil

}

var GetMsbURL = func() (string, error) {
	managerKey := dbaccessor.GetKeyOfKnitterManagerUrl()
	managerStr, err := common.GetDataBase().ReadLeaf(managerKey)
	if err != nil {
		klog.Errorf("GetMsbUrl: read key : [%s] FAILED, ERROR: %v", managerKey, err)
		return "", err
	}
	return strings.Split(managerStr, "/nw")[0], err
}

var GetEtcdAdmToolRestoreURL = func(msbUrl string) string {
	return msbUrl + "/etcdat/v1/tenants/admin/restore"
}

var MakeRestoreStateMap = func(etcdRestorejob RestoreReq) map[string]*RestoreJobState {
	var restoreJobStateMap = make(map[string]*RestoreJobState)
	restoreJobStateMap["tenants"] = &RestoreJobState{JobDn: etcdRestorejob.JobID + "_tenants", State: 0}
	restoreJobStateMap["public"] = &RestoreJobState{JobDn: etcdRestorejob.JobID + "_public", State: 0}
	restoreJobStateMap["dcs"] = &RestoreJobState{JobDn: etcdRestorejob.JobID + "_dcs", State: 0}
	return restoreJobStateMap
}

var SendRestoreJobs = func(etcdAdmToolRestoreUrl string, etcdRestorejob RestoreReq, restoreJobStateMap map[string]*RestoreJobState) {
	klog.Infof("SendRestorePosts START")
	for k := range restoreJobStateMap {
		var postBody = etcdRestorejob
		postBody.DirectoryName = "/paasnet/" + k
		postBody.JobID = postBody.JobID + "_" + k
		klog.Infof("SendRestorePosts :etcdAdmToolRestoreUrl[%v],postBody[%v]", etcdAdmToolRestoreUrl, postBody)
		err := SendRestoreJob(etcdAdmToolRestoreUrl, postBody)
		if err != nil {
			restoreJobStateMap[k].State = constvalue.RestoreProgressStatProcessFail
			klog.Errorf("SendRestoreJobs: SendRestoreJob fail %v", err)
		}

	}
	klog.Infof("SendRestorePosts END")
}

var SendRestoreJob = func(etcdAdmToolRestoreUrl string, etcdRestoreJob RestoreReq) error {
	klog.Info("SendRestorePost START")
	body, err := json.Marshal(etcdRestoreJob)
	if err != nil {
		klog.Errorf("SendRestorePost: json.Marshal(createNetParam: %v) SUCC", err)
		return err
	}
	err = HTTPPost(etcdAdmToolRestoreUrl, body)
	if err != nil {
		klog.Errorf("SendRestorePost: HttpPost(url: %s, body: %v) SUCC", etcdAdmToolRestoreUrl, string(body))
		return err
	}
	klog.Info("SendRestorePost SUCC")

	return nil
}

var GetRestoreJobStates = func(etcdAdmToolRestoreUrl string, restoreJobStateMap map[string]*RestoreJobState) {
	klog.Infof("GetRestoreJobStates START")
	for i := 0; i < constvalue.GetRestoreStateRetryTimes; i++ {
		klog.Infof("GetRestoreJobStates: [%s] time ", i)
		for k, v := range restoreJobStateMap {
			url := etcdAdmToolRestoreUrl + "?jobDn=" + v.JobDn
			klog.Infof("GetRestoreJobStates:HTTPGet(%s)", url)
			resp, err := HTTPGet(url)
			if err != nil {
				klog.Errorf("GetRestoreJobState:HTTPGet(%s) FAILED, ERROR: %v", url, err)
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				klog.Errorf("GetRestoreJobState:ioutil.ReadAll(%s) FAILED, ERROR: %v", string(body), err)
				continue
			}
			resp.Body.Close()
			klog.Infof("GetRestoreJobState: jason.NewObjectFromBytes(%s)", string(body))
			jobStateBody, err := jason.NewObjectFromBytes(body)
			if err != nil {
				klog.Errorf("GetRestoreJobState:jason.NewObjectFromBytes(%s) FAILED, ERROR: %v", string(body), err)
				continue
			}
			jobStateTmp, err := jobStateBody.GetInt64("state")
			klog.Info("GetRestoreJobState:[jobStateTmp: %v]", jobStateTmp)
			if err != nil {
				klog.Errorf("GetRestoreJobState:jobStateBody.GetString(%s) ERROR:%v", err, jobStateTmp)
			}
			jobStateDetail, err := jobStateBody.GetString("detail")
			klog.Info("GetRestoreJobState:[jobStateDetail:%v]", jobStateDetail)
			if err != nil {
				klog.Errorf("Get:RestoreJobState:jobStateBody.GetString(detail) ERROR:%v", err)

			}
			jobState := int(jobStateTmp)
			klog.Infof("GetRestoreJobState:strings.Contains:[%v]", strings.Contains(jobStateDetail, "directory not exist"), jobState)
			if jobState == constvalue.RestoreProgressStatProcessFail && strings.Contains(jobStateDetail, "directory not exist") {
				jobState = constvalue.RestoreProgressStatProcessSucc
			}
			restoreJobStateMap[k].State = jobState
		}
		finishedJobTotalNum := 0
		for _, v := range restoreJobStateMap {
			if v.State == constvalue.RestoreProgressStatProcessSucc {
				finishedJobTotalNum++
			}
			if v.State == constvalue.RestoreProgressStatProcessFail {
				klog.Infof("GetRestoreJobState:[v.state = %v] return ", v.State)
				return
			}
		}
		if finishedJobTotalNum == len(restoreJobStateMap) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	klog.Infof("GetRestoreJobStates END")

}

var CheckRestoreJobState = func(restoreJobStateMap map[string]*RestoreJobState) error {
	klog.Infof("CheckRestoreJobState: START")
	for _, v := range restoreJobStateMap {
		if v.State == -1 || v.State == 0 {
			RollBackEtcd(restoreJobStateMap)
			klog.Errorf("CheckRestoreJobState: restore job [%v]FAILED, job state is [%v]", v.JobDn, v.State)
			return errobj.ErrEtcdRestoreFromEtcdAdmToolFailed
		}
	}
	klog.Infof("CheckRestoreJobState: restore job SUCC")
	return nil
}

var RollBackEtcd = func(restoreJobStateMap map[string]*RestoreJobState) {
	for k, v := range restoreJobStateMap {
		if v.State == constvalue.RestoreProgressStatProcessSucc {
			key := dbaccessor.GetKeyOfRoot() + "/" + k
			err := common.GetDataBase().DeleteDir(key)
			if err != nil && !IsKeyNotFoundError(err) {
				klog.Errorf("RollBackEtcd: DeleteDir(key: %s)FAILED, error: %v", key, err)
			}
		}
	}

}

var RestoreTenants = func() error {
	klog.Info("RestoreTenants: START")
	tenantIds, err := getAllTenantIds()
	if err != nil {
		klog.Errorf("RestoreTenants: getAllTenantIds FAILED, error: %v", err)
		return err
	}
	klog.Info("RestoreTenants: start to RestoreTenantResource")
	for _, tid := range tenantIds {
		err := RestoreTenantResource(tid)
		if err != nil {
			klog.Errorf("RestoreTenants: RestoreTenantResource(tenantId: %s) FAILED, error: %v",
				tid, err)
			return err
		}
		klog.Infof("RestoreTenants: end to RestoreTenantResource for tenantId: %s", tid)
	}
	klog.Info("RestoreTenants: SUCC")
	return nil
}

var RestoreTenantResource = func(tid string) error {
	klog.Infof("RestoreTenantResource: tenantID: %s START", tid)
	klog.Infof("RestoreTenants: start to clearPods for tenantId: %s", tid)
	err := clearPods(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: clearPods(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}
	err = clearRouters(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: clearRouters(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}

	klog.Infof("RestoreTenants: start to RestoreNetworks for tenantId: %s", tid)
	err = RestoreNetworks(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: RestoreNetworks(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}

	klog.Infof("RestoreTenants: start to ClearIPGroups for tenantId: %s", tid)
	err = ClearIPGroups(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: ClearIPGroups(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}

	klog.Infof("RestoreTenants: start to ClearLogicalPorts for tenantId: %s", tid)
	err = ClearLogicalPorts(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: ClearLogicalPorts(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}

	err = ClearPhysicalPorts(tid)
	if err != nil {
		klog.Errorf("RestoreTenantResource: clearPhysicalPorts(tenantId: %s) FAILED, error: %v",
			tid, err)
		return err
	}

	klog.Infof("RestoreTenantResource: tenantID: %s SUCC", tid)
	return nil
}

func clearRouters(tid string) error {
	key := dbaccessor.GetKeyOfRouterGroup(tid)
	err := common.GetDataBase().DeleteDir(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("clearRouters: DeleteDir(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("clearRouters: DeleteDir(key: %s) SUCC", key)
	return nil
}

func clearPods(tid string) error {
	key := dbaccessor.GetKeyOfPodNsGroup(tid)
	err := common.GetDataBase().DeleteDir(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("clearPods: DeleteDir(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("clearPods: DeleteDir(key: %s) SUCC", key)
	return nil
}

var ClearLogicalPorts = func(tid string) error {
	ports, err := GetPortObjRepoSingleton().ListByTenantID(tid)
	if err != nil {
		klog.Infof("ClearLogicalPorts:ListByTenantID(tenantID: %s) FAIL, error: %v", tid, err)
		return err
	}
	if len(ports) == 0 {
		klog.Info("ClearLogicalPorts: no ports left, omit it")
		return nil
	}

	for _, port := range ports {
		portID := port.ID
		err = clearLogicalPort(tid, portID)
		if err != nil {
			klog.Errorf("ClearLogicalPorts: clearLogicalPort(portID: %s) FAILED, error: %v", portID, err)
			return err
		}
	}
	return nil
}

var ClearIPGroups = func(tid string) error {
	igObjects, err := GetIPGroupObjRepoSingleton().ListByTenantID(tid)
	if err != nil {
		klog.Infof("ClearIPGroups:ListByTenantID(tenantID: %s) FAIL, error: %v", tid, err)
		return err
	}

	if len(igObjects) == 0 {
		klog.Info("ClearIPGroups: no ip groups left, omit it")
		return nil
	}

	for _, igObject := range igObjects {
		err = clearIPGroup(tid, igObject.ID)
		if err != nil {
			klog.Errorf("ClearIPGroups: clearIPGroup(ID: %s) FAILED, error: %v", igObject.ID, err)
			return err
		}
	}
	return nil
}

func clearIPGroup(tid string, igID string) error {
	klog.Infof("clearIPGroup: ip_group_id: %s START", igID)
	igSelfKey := createIPGroupKey(igID)
	igStr, err := common.GetDataBase().ReadLeaf(igSelfKey)
	if err != nil {
		klog.Errorf("clearIPGroup: ReadLeaf(%s) FAILED, error: %v", igSelfKey, err)
		return err
	}

	igInDb := IPGroupInDB{}
	json.Unmarshal([]byte(igStr), &igInDb)
	for _, ip := range igInDb.IPs {
		err = iaas.GetIaaS(tid).DeletePort(ip.PortID)
		if err != nil {
			klog.Errorf("clearIPGroup: DeletePort(%s) FAILED, error: %v", ip.PortID, err)
		}
	}

	err = deleteIGFromDBAndCache(igID)
	if err != nil {
		klog.Errorf("clearIPGroup: deleteIGFromDBAndCache(%s) FAILED, error: %v", igID, err)
		return err
	}

	klog.Infof("clearIPGroup: clear ip_group_id: %s SUCC", igID)
	return nil
}

var ClearPhysicalPorts = func(tenantID string) error {
	ports, err := GetPhysPortObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		klog.Infof("ClearPhysicalPorts:ListByTenantID(tenantID: %s) FAIL, error: %v", tenantID, err)
		return err
	}
	if len(ports) == 0 {
		klog.Info("ClearPhysicalPorts: no ports left, omit it")
		return nil
	}

	for _, port := range ports {
		err = clearPhysicalPort(tenantID, port)
		if err != nil {
			klog.Errorf("ClearPhysicalPorts: clearPhysicalPort(%v) FAILED, error: %v", port, err)
			return err
		}
	}
	return nil
}

func clearLogicalPortFromIaas(tenantID, portID string) error {
	err := iaas.GetIaaS(tenantID).DeletePort(portID)
	if err != nil {
		klog.Errorf("clearPortFromIaas: DeletePort(portID: %s) FAILED, error: %v", portID, err)
		return err
	}
	klog.Tracef("clearPortFromIaas: clear portID: %s SUCC")
	return nil
}

func clearLogicalPort(tenantID, portID string) error {
	klog.Infof("clearLogicalPort: clear tid: %s, interface_id: %s START", portID)
	err := clearLogicalPortFromIaas(tenantID, portID)
	if err != nil {
		klog.Errorf("clearLogicalPort: clearLogicalPortFromIaas(portID: %s) FAIL, error: %v", portID, err)
		return err
	}

	err = DeleteLogicalPort(portID)
	if err != nil {
		klog.Errorf("clearLogicalPort: DeleteLogicalPort(portID: %s) FAIL, error: %v", portID, err)
		return err
	}

	err = GetPortObjRepoSingleton().Del(portID)
	if err != nil {
		klog.Errorf("clearLogicalPort: GetPortObjRepoSingleton().Del(portID: %s) FAIL, error: %v", portID, err)
		return err
	}
	klog.Tracef("clearLogicalPort: delete port portID: %s SUCC", portID)
	return nil
}

func clearPortFromIaas(tenantID, portID, nodeID string) error {
	err := iaas.GetIaaS(tenantID).DetachPortFromVM(nodeID, portID)
	if err != nil {
		klog.Warningf("clearPortFromIaas: DetachPortFromVM(portID:%s, nodeID: %s) FAILED, error: %v",
			nodeID, portID, err)
	}

	err = iaas.GetIaaS(tenantID).DeletePort(portID)
	if err != nil {
		klog.Errorf("clearPortFromIaas: DeletePort(portID: %s) FAILED, error: %v", portID, err)
		return err
	}
	klog.Tracef("clearPortFromIaas: clear portID: %s SUCC")
	return nil
}

func clearPhysicalPort(tenantID string, port *PhysPortObj) error {
	klog.Infof("clearPhysicalPort: clear tid: %s, interface_id: %s START", port.ID)
	err := clearPortFromIaas(tenantID, port.ID, port.NodeID)
	if err != nil {
		klog.Errorf("clearPhysicalPort: clearPortFromIaas(portID: %s, nodeID: %s) FAIL, error: %v",
			port.ID, port.NodeID, err)
		return err
	}

	err = DeletePhysicalPort(port.ID)
	if err != nil {
		klog.Errorf("clearPhysicalPort: DeletePhysicalPort(portID: %s) FAIL, error: %v", port.ID, err)
		return err
	}

	err = GetPhysPortObjRepoSingleton().Del(port.ID)
	if err != nil {
		klog.Errorf("clearPhysicalPort: GetPhysPortObjRepoSingleton().Del(portID: %s) FAIL, error: %v", port.ID, err)
		return err
	}
	klog.Tracef("clearPhysicalPort: delete PhysPort portID: %s SUCC", port.ID)
	return nil
}

var RestoreNetworks = func(tenantId string) error {
	klog.Infof("RestoreNetworks: clear tenantID: %s all networks START", tenantId)
	nets, err := GetNetObjRepoSingleton().ListByTenantID(tenantId)
	if err != nil {
		klog.Errorf("RestoreNetworks: ListByTenantID[tenantID: %s} FAILED, error: %v", tenantId, err)
		return err
	}

	for _, net := range nets {
		err := restorePrivNetwork(tenantId, net.ID)
		if err != nil {
			klog.Errorf("RestoreNetworks: restorePrivNetwork: %s: %s FAILED, error: %v", tenantId, net.ID, err)
			return err
		}
	}
	klog.Infof("RestoreNetworks: clear tenantID: %s all networks SUCC", tenantId)
	return nil
}

type createNetwork struct {
	Name        string `json:"name"`
	Cidr        string `json:"cidr"`
	Gateway     string `json:"gateway"`
	Public      bool   `json:"public"`
	Description string `json:"	description"`
}

type encapCreateNetwork struct {
	Network createNetwork `json:"network"`
}

type createProviderNetwork struct {
	Name                string `json:"name"`
	Cidr                string `json:"cidr"`
	Gateway             string `json:"gateway"`
	Public              bool   `json:"public"`
	ProviderNetworkType string `json:"provider:network_type"`
	ProviderPhysNet     string `json:"provider:physical_network"`
	ProviderSegID       string `json:"provider:segmentation_id"`
}

type encapCreateProviderNetwork struct {
	ProviderNetwork createProviderNetwork `json:"provider_network"`
}

func restorePrivNetwork(tenantID string, netID string) error {
	klog.Infof("restorePrivNetwork: for tenantID: %s, netID: %s START", tenantID, netID)
	_, err := iaas.GetIaaS(tenantID).GetNetwork(netID)
	if err == nil {
		klog.Infof("restorePrivNetwork: GetIaaS().GetNetwork(netId: %s) SUCC, just retain it", netID)
		return nil
	}
	net, err := GetNetObjRepoSingleton().Get(netID)
	if err != nil {
		klog.Errorf("restorePrivNetwork: GetNetObjRepoSingleton().Get(networID: %s) FAIL, error: %v", netID, err)
		return err
	}

	subnet, err := GetSubnetObjRepoSingleton().Get(net.SubnetID)
	if err != nil {
		klog.Errorf("restorePrivNetwork: GetSubnetObjRepoSingleton().Get(SubnetID: %s) FAIL, error: %v",
			net.SubnetID, err)
		return err
	}

	// clear network from knitter-manager side
	err = DeleteNetwork(netID)
	if err != nil {
		klog.Errorf("restorePrivNetwork: DeleteNetwork(networID: %s) FAIL, error: %v", netID, err)
		return err
	}

	if !net.IsExternal {
		err = CreateNet(net, subnet)
		if err != nil {
			klog.Errorf("restorePrivNetwork: CreateNet(net: %v) FAIL, error: %v", net, err)
			return err
		}
	}

	klog.Infof("restorePrivNetwork: for tenantID: %s, netID: %s SUCC", tenantID, netID)
	return nil
}

func isProviderNetowrk(net *Net) bool {
	if net.Provider.NetworkType != "" &&
		net.Provider.PhysicalNetwork != "" &&
		net.Provider.SegmentationID != "" {
		return true
	}
	return false
}

func createTenantNet(net *Net) error {
	createNetParam := encapCreateNetwork{
		Network: createNetwork{
			Name:        net.Network.Name,
			Cidr:        net.Subnet.Cidr,
			Gateway:     net.Subnet.GatewayIp,
			Public:      net.Public,
			Description: net.Description,
		},
	}

	url := "http://127.0.0.1:9527" + "/nw/v1/tenants/" + net.TenantUUID + "/" + "networks"
	body, err := json.Marshal(&createNetParam)
	if err != nil {
		klog.Errorf("createTenantNet: json.Marshal(createNetParam: %v) SUCC", createNetParam)
		return err
	}
	err = HTTPPost(url, body)
	if err != nil {
		klog.Errorf("createTenantNet: HttpPost(url: %s, body: %v) SUCC", url, string(body))
		return err
	}

	return nil
}

func createTenantProviderNet(net *Net) error {
	createProviderNetParam := encapCreateProviderNetwork{
		ProviderNetwork: createProviderNetwork{
			Name:                net.Network.Name,
			Cidr:                net.Subnet.Cidr,
			Gateway:             net.Subnet.GatewayIp,
			Public:              net.Public,
			ProviderNetworkType: net.Provider.NetworkType,
			ProviderPhysNet:     net.Provider.PhysicalNetwork,
			ProviderSegID:       net.Provider.SegmentationID,
		},
	}

	url := "http://127.0.0.1:9527" + "/nw/v1/tenants/" + net.TenantUUID + "/" + "networks"
	body, err := json.Marshal(&createProviderNetParam)
	if err != nil {
		klog.Errorf("createTenantProviderNet: json.Marshal(createProviderNetParam: %v) SUCC", createProviderNetParam)
		return err
	}
	err = HTTPPost(url, body)
	if err != nil {
		klog.Errorf("createTenantProviderNet: HttpPost(url: %s, body: %v) SUCC", url, string(body))
		return err
	}

	return nil
}

var CreateNet = func(netObj *NetworkObject, subnet *SubnetObject) error {
	net := &Net{
		Network: iaasaccessor.Network{
			Name: netObj.Name,
			Id:   netObj.ID,
		},
		Subnet: iaasaccessor.Subnet{
			Id:        subnet.ID,
			NetworkId: subnet.NetworkID,
			Name:      subnet.Name,
			Cidr:      subnet.CIDR,
			GatewayIp: subnet.GatewayIP,
			TenantId:  subnet.TenantID,
		},
		Provider: iaasaccessor.NetworkExtenAttrs{
			Id:              netObj.ID,
			Name:            netObj.Name,
			NetworkType:     netObj.ExtAttrs.NetworkType,
			PhysicalNetwork: netObj.ExtAttrs.PhysicalNetwork,
			SegmentationID:  netObj.ExtAttrs.SegmentationID,
		},
		TenantUUID:  netObj.TenantID,
		Public:      netObj.IsPublic,
		ExternalNet: netObj.IsExternal,
		CreateTime:  netObj.CreateTime,
		Status:      netObj.Status,
		Description: netObj.Description,
	}

	if isProviderNetowrk(net) {
		err := createTenantProviderNet(net)
		klog.Errorf("CreateNet: createTenantProviderNet(net: %v) FAILED, error: %v", *net, err)
		return err
	}

	err := createTenantNet(net)
	if err != nil {
		klog.Errorf("CreateNet: createTenantNet(net: %v) FAILED, error: %v", *net, err)
		return err
	}
	return nil
}

var RestorePublic = func() error {
	err := clearPublic()
	if err != nil {
		klog.Errorf("RestorePublic: clearPublic FAILED, error: %v", err)
		return err
	}

	netIds, err := GetAdminPublicNetIds()
	if err != nil {
		klog.Errorf("RestorePublic: GetAdminPublicNetIds() FAILED, error: %v", err)
		return err
	}

	klog.Infof("RestorePublic: find all public networks[%v]", netIds)
	err = rebuildPublicNetworks(adminTenantID, netIds)
	if err != nil {
		klog.Errorf("RestorePublic: rebuild PublicNetworks[%v] FAILED, error: %v", netIds, err)
		return err
	}

	klog.Infof("RestorePublic: restore all public networks[%v] SUCC", netIds)
	return nil
}

func isNetworkPublic(value string) bool {
	netObj, err := jason.NewObjectFromBytes([]byte(value))
	if err != nil {
		klog.Infof("isNetworkPublic: jason.NewObjectFromBytes(%v) FAILED, error: %v",
			value, err)
		return false
	}
	isPub, err := netObj.GetBoolean("Public")
	if err != nil {
		klog.Infof("isNetworkPublic: netObj.GetBoolean(Public) FAILED, error: %v", err)
		return false
	}
	return isPub
}

var GetAdminPublicNetIds = func() ([]string, error) {
	pubNetsKey := dbaccessor.GetKeyOfNetworkGroup(adminTenantID)
	nodes, err := common.GetDataBase().ReadDir(pubNetsKey)
	if err != nil {
		klog.Errorf("GetAdminPublicNetIds: ReadDir(key: %s) FAILED, error: %v", pubNetsKey, err)
		return nil, err
	}

	netIds := []string{}
	for _, node := range nodes {
		netID := strings.TrimPrefix(node.Key, pubNetsKey+"/")
		netKey := dbaccessor.GetKeyOfNetworkSelf(adminTenantID, netID)
		value, err := common.GetDataBase().ReadLeaf(netKey)
		if err != nil {
			klog.Errorf("GetAdminPublicNetIds: ReadLeaf[key: %s] FAILED, error: %v", netKey, err)
			return nil, err
		}
		if !isNetworkPublic(value) {
			continue
		}
		netIds = append(netIds, netID)
	}
	klog.Infof("GetAdminPublicNetIds: get all public network ids: %v SUCC", netIds)
	return netIds, nil
}

func rebuildPublicNetworks(tenantID string, netIds []string) error {
	for _, netID := range netIds {
		err := rebuildPublicNet(tenantID, netID)
		if err != nil {
			klog.Errorf("rebuildPublicNetworks: rebuildPublicNet(tenantId: %s, netId: %s) FAILED, error: %v",
				tenantID, netID, err)
			return err
		}
		klog.Infof("rebuildPublicNet: rebuild public network(tenantId: %s, netId: %s) SUCC",
			tenantID, netID)
	}
	klog.Infof("rebuildPublicNet: rebuild all public network(tenantId: %s, netIds: %v) SUCC",
		tenantID, netIds)
	return nil
}

func rebuildPublicNet(tenantID, netID string) error {
	key := dbaccessor.GetKeyOfPublicNetwork(netID)
	value := dbaccessor.GetKeyOfNetworkSelf(tenantID, netID)
	err := common.GetDataBase().SaveLeaf(key, value)
	if err != nil {
		klog.Errorf("rebuildPublicNet: SaveLeaf(key: %s, value: %s) FAILED, error: %v",
			key, value, err)
		return err
	}
	klog.Infof("rebuildPublicNet: restore public network(tenantId: %s, netId: %s) SUCC", tenantID, netID)
	return nil
}

func clearPublic() error {
	pubKey := dbaccessor.GetKeyOfPublic()
	err := common.GetDataBase().DeleteDir(pubKey)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("clearPublic: DeleteDir(key: %s) FAILED, error: %v", pubKey, err)
		return err
	}
	return nil
}

var ClearDcs = func() error {
	dcsKey := dbaccessor.GetKeyOfDcs()
	err := common.GetDataBase().DeleteDir(dcsKey)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("ClearDcs: DeleteDir(key: %s)FAILED, error: %v", dcsKey, err)
		return err
	}
	return nil
}

func ClearRuntime() error {
	rtKey := dbaccessor.GetKeyOfRuntime()
	err := common.GetDataBase().DeleteDir(rtKey)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Infof("ClearRuntime: DeleteDir(key: %s)FAILED, error: %v", rtKey, err)
		return err
	}
	return nil
}

var HTTPPostFunc = func(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return http.Post(url, contentType, body)
}
var HTTPGetFunc = func(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func HTTPPost(url string, bodyContent []byte) (err error) {
	klog.Infof("HttpPost: request url=%v", url)
	contentType := "application/json"

	body := bytes.NewReader([]byte(bodyContent))
	resp, err := HTTPPostFunc(url, contentType, body)
	if err != nil {
		klog.Errorf("HttpPost:bytes.NewReader(%v) err: %v", body, err)
		return err
	}

	if resp.StatusCode != 200 {
		klog.Errorf("HttpPost:status code: %d", resp.StatusCode)
		return errobj.ErrHTTPPostStatusCode
	}

	return nil
}

var HTTPGet = func(url string) (*http.Response, error) {
	klog.Infof("HttpGet: request url=%v", url)
	//contentType := "application/json"
	resp, err := HTTPGetFunc(url)
	//resp, err := HTTPGetFunc(url, contentType)
	if err != nil {
		klog.Errorf("HttpGet err: %v", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		klog.Errorf("HttpGet:status code: %d", resp.StatusCode)
		return nil, errobj.ErrHTTPPostStatusCode
	}
	return resp, nil
}

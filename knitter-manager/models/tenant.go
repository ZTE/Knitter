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
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/astaxie/beego/context"
	"github.com/rackspace/gophercloud"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const CancelIntervalInSec = 20
const dhcpPortIDPrefix = "dhcp"

var DefaultCreateTenantTime string = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05Z")

type Tenant struct {
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

type PaasTenant struct {
	TenantName string     `json:"name"`
	TenantUUID string     `json:"id"`
	NetNum     int        `json:"net_number"`
	CreateTime string     `json:"created_at"`
	Quotas     PaasQuotas `json:"quotas"`
	Status     string     `json:"status"`
}

type PaasQuotas struct {
	Network int `json:"network"`
}

type ExclusiveTenant struct {
	PaasName       string `json:"paas_name"`
	PassID         string `json:"pass_id"`
	IaasEndpoint   string `json:"iaas_endpoint"`
	IaasTenantName string `json:"iaas_tenant_name"`
	IaasUserName   string `json:"iaas_user_name"`
	IaasPaasword   string `json:"iaas_paasword"`
	IaasTenantID   string `json:"iaas_tenant_id"`
	Networks       string `json:"networks"`
	Interfaces     string `json:"interfaces"`
	Quota          int    `json:"net_quota"`
	NetNum         int    `json:"net_number"`
	IsCancelling   bool   `json:"is_cancelling"`
	CreateTime     string `json:"create_time"`
}

func (self *Tenant) SaveTenantToEtcd() error {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(self.TenantUUID)
	tenantValue, _ := json.Marshal(*self)
	err := common.GetDataBase().SaveLeaf(tenantKey, string(tenantValue))
	if err != nil {
		klog.Errorf("SaveTenantToEtcd: SaveLeaf: key[%s], value[%v]:Error[%v]", tenantKey, tenantValue, err)
		return err
	}
	return nil
}

func (self *Tenant) UpdateQuota(quota int) error {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(self.TenantUUID)
	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil {
		return BuildErrWithCode(http.StatusNotFound, err)
	}
	err = json.Unmarshal([]byte(tenantValue), self)
	if err != nil {
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}
	self.Quota = quota
	value, err := json.Marshal(self)
	if err != nil {
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}
	err = common.GetDataBase().SaveLeaf(tenantKey, string(value))
	if err != nil {
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}
	klog.Infof("tenant:%v, quota:%v, UpdateQuota Successful", self.TenantUUID, self.Quota)
	return nil
}

func (self *Tenant) GetTenantInfo() (*Tenant, error) {
	//if self.TenantUUID == constvalue.PaaSTenantAdminDefaultUUID {
	//	return makePaasTenant(getAdminInfo()), nil
	//}
	tenant, err := getTenantInfo(self.TenantUUID)
	if err != nil {
		return nil, BuildErrWithCode(http.StatusNotFound, err)
	}
	return tenant, nil
}

func forceDeleteNetwork(tenantID, id string) error {
	klog.Infof("forceDeleteNetwork: delete network[id: %s] START", id)
	err := deleteResidualPorts(tenantID, id)
	if err != nil {
		klog.Errorf("forceDeleteNetwork: deleteResidualPorts of network[id: %s] FAIL, error: %v", id, err)
		return err
	}

	err = iaas.GetIaaS(tenantID).DeleteNetwork(id)
	if err != nil {
		klog.Errorf("DelNetwork: delete network(id: %s) from iaas FAIL, error: %v", id, err)
		return err
	}

	klog.Infof("forceDeleteNetwork: delete network[id: %s] SUCC", id)
	return nil
}

func isDHCPPort(deviceID string) bool {
	return strings.HasPrefix(deviceID, dhcpPortIDPrefix)
}

// judge if port is attach to vm, actually router device also match, just ignore at present
func isPortAttachedToVM(deviceID string) bool {
	return deviceID != ""
}

var clearResidualPort = func(tenantID string, port *iaasaccessor.Interface) error {
	if isDHCPPort(port.DeviceId) {
		klog.Debugf("clearResidualPort: skip dhcp port: %v", port)
		return nil
	}

	if isPortAttachedToVM(port.DeviceId) {
		err := iaas.GetIaaS(tenantID).DetachPortFromVM(port.DeviceId, port.Id)
		if err != nil {
			klog.Warningf("deleteResidualPorts: DetachPortFromVM(DeviceId: %s, portID: %s) error: %v",
				port.DeviceId, port.Id, err)
			// pass through, vm pci will leak
		}
		klog.Tracef("clearResidualPort: detach port: %v SUCC", port)
	}

	err := iaas.GetIaaS(tenantID).DeletePort(port.Id)
	if err != nil {
		klog.Warningf("deleteResidualPorts: DeletePort(id: %s) error: %v", port.Id, err)
		return err
	}
	klog.Tracef("clearResidualPort: clear port: %v SUCC", port)
	return nil
}

var deleteResidualPorts = func(tenantID, networkID string) error {
	// list all ports
	ports, err := iaas.GetIaaS(tenantID).ListPorts(networkID)
	if err != nil {
		klog.Errorf("deleteResidualPorts: get all residuel ports for networkID :%s error: %v", networkID, err)
		return err
	}

	var residualPortCnt int
	// loop detach/delete all ports
	for _, port := range ports {
		err = clearResidualPort(tenantID, port)
		if err != nil {
			klog.Warningf("deleteResidualPorts: clearResidualPort(portID: %v) error: %v", port.Id, err)
			residualPortCnt++
		}
	}

	if residualPortCnt > 0 {
		klog.Errorf("deleteResidualPorts: still has %d ports not proceed", residualPortCnt)
		return errobj.ErrNetworkHasPortsInUse
	}
	klog.Infof("deleteResidualPorts: all ports[ %v ] deleted SUCC", ports)
	return nil

}

func deleteTenantDir(tenantID string) error {
	key := dbaccessor.GetKeyOfTenant(tenantID)
	err := common.GetDataBase().DeleteDir(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("deleteTenantDir: Delete tenant [ID: %s from etcd error", tenantID)
		return err
	}
	return nil
}

var ClearNetworks = func(tenantID string) error {
	nets, err := GetNetObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		klog.Errorf("ClearNetworks: GetNetObjRepoSingleton().ListByTenantID(tenantID: %s) FAILED, error: %v",
			tenantID, err)
		return err
	}

	klog.Infof("ClearNetworks: start to clear tenantID: %s all networks: %v", tenantID, nets)
	for _, net := range nets {
		err := DeleteNetwork(net.ID)
		if err != nil {
			klog.Errorf("ClearNetworks: restorePrivNetwork: %s: %s FAILED, error: %v", tenantID, net.ID, err)
			return err
		}
	}
	klog.Infof("ClearNetworks: clear tenantID: %s all networks SUCC", tenantID)
	return nil
}

func (self *Tenant) Cancel() error {
	self.IsCancelling = true
	err := self.SaveTenantToEtcd()
	if err != nil {
		klog.Errorf("SaveTenantToEtcd(tenant: %v) FAILED, error: %v", self, err)
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	go CancelTenantLoop(self.TenantUUID)
	klog.Infof("start cancel tenantID: %s goroutine SUCC", self.TenantUUID)
	return nil
}

func makePaasTenant(tenant *Tenant) *PaasTenant {
	var status string
	if tenant.IsCancelling {
		status = "DELETING"
	} else {
		status = "ACTIVE"
	}
	return &PaasTenant{
		TenantUUID: tenant.TenantUUID,
		TenantName: tenant.TenantName,
		NetNum:     tenant.NetNum,
		CreateTime: tenant.CreateTime,
		Quotas:     PaasQuotas{Network: tenant.Quota},
		Status:     status,
	}
}

func getAdminInfo() *Tenant {
	adminInfo := Tenant{
		TenantUUID: "admin",
		TenantName: "admin",
		NetNum:     GetNetNumOfTenant("admin"),
		Quota:      QuotaAdmin,
		CreateTime: DefaultCreateTenantTime,
	}

	return &adminInfo
}

func getTenantInfo(tenantID string) (*Tenant, error) {
	//if tenantID == constvalue.PaaSTenantAdminDefaultUUID {
	//	return getAdminInfo(), nil
	//}

	tenantKey := dbaccessor.GetKeyOfTenantSelf(tenantID)
	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err != nil {
		klog.Errorf("getTenantInfo: ReadLeaf(key: %s) FAILED, error: %v", tenantKey, err)
		return nil, err
	}

	var tenantInfo Tenant
	err = json.Unmarshal([]byte(tenantValue), &tenantInfo)
	if err != nil {
		klog.Errorf("getTenantInfo: json.Unmarshal(%v) FAILED, error: %v", string(tenantValue), err)
		return nil, err
	}
	klog.Infof("getTenantInfo: get tenant info tenantID: %s SUCC, info detail: %v", tenantID, tenantInfo)
	return &tenantInfo, nil
}

func GetAllTenants() []*Tenant {
	klog.Infof("GetAllTenants: START")
	tids, err := getAllTenantIds()
	if err != nil {
		klog.Errorf("GetAllTenants: getAllTenantIds FAILED, error: %v", err)
		return nil
	}

	tenants := []*Tenant{}
	for _, tid := range tids {
		//if tid == constvalue.PaaSTenantAdminDefaultUUID {
		//	tenants = append(tenants, makePaasTenant(getAdminInfo()))
		//	continue
		//}

		tenantInfo, err := getTenantInfo(tid)
		if err != nil {
			continue
		}
		tenants = append(tenants, tenantInfo)
	}

	klog.Infof("GetAllTenants: get all tenants: %v SUCC", tenants)
	if len(tenants) == 0 {
		return nil
	}
	return tenants
}

func GetAllNormalTenantIds() ([]string, error) {
	klog.Infof("GetAllNormalTenantIds: START")
	tids, err := getAllTenantIds()
	if err != nil {
		klog.Errorf("GetAllNormalTenantIds: getAllTenantIds FAILED, error: %v", err)
		return nil, err
	}

	tenantIds := []string{}
	for _, tid := range tids {
		if tid == constvalue.PaaSTenantAdminDefaultUUID {
			tenantIds = append(tenantIds, tid)
			continue
		}
		tenantInfo, err := getTenantInfo(tid)
		if err != nil {
			klog.Errorf("GetAllNormalTenantIds: getTenantInfo[%v] FAILED, error: %v", tid, err)
			return nil, err
		}
		if tenantInfo != nil && tenantInfo.IsCancelling {
			continue
		}
		tenantIds = append(tenantIds, tid)
	}

	klog.Infof("GetAllNormalTenantIds: get all tenantIds: %v SUCC", tenantIds)
	return tenantIds, nil
}

func isTenantCancellable(tenantID string) bool {
	if tenantID == constvalue.PaaSTenantAdminDefaultUUID {
		return false
	}
	return true
}

var GetCancellingTenantsInfo = func() []string {
	klog.Infof("GetCancellingTenantsInfo: START")
	tids, err := getAllTenantIds()
	if err != nil {
		klog.Errorf("GetCancellingTenantsInfo: getAllTenantIds FAILED, error: %v", err)
		return nil
	}

	var cancellingTenantIDs []string
	for _, tid := range tids {
		if !isTenantCancellable(tid) {
			klog.Infof("GetCancellingTenantsInfo: isTenantCancellable(id: %s) FAILED, error: %v", tid, err)
			continue
		}

		tenantInfo, err := getTenantInfo(tid)
		if err != nil {
			klog.Errorf("GetCancellingTenantsInfo: getTenantInfo FAILED, error: %v", err)
			continue
		}
		if tenantInfo.IsCancelling {
			cancellingTenantIDs = append(cancellingTenantIDs, tenantInfo.TenantUUID)
		}
	}

	klog.Infof("GetCancellingTenantsInfo: SUCC, get all cancelling"+
		" tenantIDs : %v", cancellingTenantIDs)
	return cancellingTenantIDs
}

// todo: will be replaced by upper GetTenantPods()
var GetPodsOfTenant = func(tid string) []*Pod {
	pod := &Pod{}
	pod.SetTenantID(tid)
	return PodListAll(pod)
}

var BeforeExecTenantCheck = func(ctx *context.Context) {
	userID := ctx.Input.Param(":user")
	if userID != constvalue.PaaSTenantAdminDefaultUUID {
		validUser := false
		tenantIds, err := GetAllNormalTenantIds()
		if err != nil {
			ctx.Redirect(http.StatusInternalServerError, ctx.Request.URL.RequestURI())
			ctx.Output.JSON(map[string]string{"ERROR": http.StatusText(http.StatusInternalServerError),
				"message": err.Error()}, false, false)
			return
		}
		for _, tenantID := range tenantIds {
			if userID == tenantID {
				validUser = true
				break
			}
		}

		if !validUser {
			ctx.Redirect(http.StatusNotFound, ctx.Request.URL.RequestURI())
			ctx.Output.JSON(map[string]string{"ERROR": "Bad Tenant Id",
				"message": "Bad Tenant Id[" + userID + "]"}, false, false)
		}
	}
}

const (
	CancelWaitPodsDeletedTimeout = 15 * 60
	CancelWaitPodsDeletedIntval  = 15
)

var getWaitTenantPodsDeletedTimout = func() int {
	timeoutInSec := CancelWaitPodsDeletedTimeout
	key := dbaccessor.GetKeyOfCancelWaitPodsDeletedTimeout()
	timeoutStr, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Warningf("waitTenantsPodsDeleted: Get Wait Timeout from etcd path: %s failed, use default %d seconds",
			key, timeoutInSec)
	} else {
		timeoutConv, err := strconv.Atoi(timeoutStr)
		if err != nil {
			klog.Warningf("strconv.Atoi(%s) failed, error: %v, using default %d seconds",
				timeoutStr, err, timeoutInSec)
		} else {
			timeoutInSec = timeoutConv
		}
	}
	return timeoutInSec
}

var waitTenantPodsDeleted = func(tenantID string) []*LogicPod {
	timeoutInSec := getWaitTenantPodsDeletedTimout()
	startTime := time.Now().Unix()
	endTime := startTime + int64(timeoutInSec)
	var pods []*LogicPod
	for curTime := startTime; curTime < endTime; curTime = time.Now().Unix() {
		//pods := GetTenantPods(tenantID)
		pods := GetPodsOfTenant(tenantID)
		if pods == nil || len(pods) == 0 {
			klog.Infof("waitTenantPodsDeleted: wait tenant [%s] all pods deleted done", tenantID)
			return nil
		}
		klog.Warningf("waitTenantPodsDeleted: Tenant [%v] still has [%v] pods in use, retry if not timeout",
			tenantID, len(pods))
		time.Sleep(time.Second * CancelWaitPodsDeletedIntval)
	}

	klog.Warningf("waitTenantPodsDeleted: wait tenant [%v] all pods deleted timeout", tenantID)
	return pods
}

var ClearTenantPods = func(tenantID string) error {
	pods := waitTenantPodsDeleted(tenantID)
	if pods == nil || len(pods) == 0 {
		return nil
	}

	for _, pod := range pods {
		err := pod.Delete()
		if err != nil {
			klog.Errorf("ClearTenantPods: pod.Delete(podName: %s) FAIL, error: %v", pod.PodName, err)
			return err
		}
	}
	klog.Debugf("ClearTenantPods: delete all tenant: %s pods SUCC", tenantID)
	return nil
}

var CancelTenant = func(tenantID string) error {
	klog.Infof("CancelTenant: delTenantAllNetworks for tenantID: %s START", tenantID)
	err := ClearTenantPods(tenantID)
	if err != nil {
		klog.Infof("CancelTenant: ClearTenantPods[tenantID: %s] error: %v", tenantID, err)
		return err
	}

	err = ClearIPGroups(tenantID)
	if err != nil {
		klog.Errorf("CancelTenant: ClearIPGroups for TenantID[%s] FAILED, error: %v", tenantID, err)
		return err
	}

	err = ClearLogicalPorts(tenantID)
	if err != nil {
		klog.Errorf("CancelTenant: ClearLogicalPorts for TenantID[%s] FAILED, error: %v", tenantID, err)
		return err
	}

	err = ClearPhysicalPorts(tenantID)
	if err != nil {
		klog.Errorf("CancelTenant: ClearPhysicalPorts for TenantID[%s] FAILED, error: %v", tenantID, err)
		return err
	}

	err = ClearNetworks(tenantID)
	if err != nil {
		klog.Errorf("CancelTenant: ClearNetworks TenantID[%s] FAILED, error: %v", tenantID, err)
		return err
	}

	err = deleteTenantDir(tenantID)
	if err != nil {
		klog.Errorf("CancelTenant: deleteTenantDir TenantID[%s] FAILED, error: %v", tenantID, err)
		return err
	}

	klog.Infof("CancelTenant: delTenantAllNetworks for tenantID: %s SUCC", tenantID)
	return nil
}

var CancelTenantLoop = func(tenantID string) {
	klog.Infof("CancelTenantLoop: goroutine cancel tenantID: %s START", tenantID)
	for {
		err := CancelTenant(tenantID)
		if err == nil {
			break
		}
		klog.Warningf("CancelTenantLoop: CancelTenant tenantID: %s error: %v, just wait and retry", err, tenantID)
		time.Sleep(time.Duration(CancelIntervalInSec) * time.Second)
	}
	klog.Infof("CancelTenantLoop: goroutine cancel tenantID: %s  SUCC", tenantID)
}

func CancelResidualTenants() {
	klog.Info("CancelResidualTenants: START")
	tids := GetCancellingTenantsInfo()
	if tids == nil {
		klog.Info("CancelResidualTenants: GetCancellingTenantsInfo() result nil, just return")
		return
	}

	klog.Infof("CancelResidualTenants: GetCancellingTenantsInfo() result: %v", tids)
	for _, tid := range tids {
		go CancelTenantLoop(tid)
		klog.Debugf("CancelResidualTenants: start CancelTenantLoop goroutine for tenantID: %v", tid)
	}
	klog.Info("CancelResidualTenants: trigger FINI")
}

func (self *ExclusiveTenant) Create() error {
	if IsTenantExistInDB(self.PassID) {
		return BuildErrWithCode(http.StatusConflict, errors.New("tenant is already exist"))
	}
	opstk := openstack.NewOpenstack()
	conf := gophercloud.AuthOptions{
		IdentityEndpoint: self.IaasEndpoint + "/tokens",
		Username:         self.IaasUserName,
		Password:         self.IaasPaasword,
		TenantName:       self.IaasTenantName,
		AllowReauth:      true,
	}
	opstk.SetConfig(conf)
	err := opstk.Auth()
	if err != nil {
		klog.Errorf("IaaS-auth-ERROR")
		return BuildErrWithCode(http.StatusNotAcceptable, errors.New("auth error"))
	}
	err = iaas.CheckNetService(opstk)
	if err != nil {
		klog.Error("NetService-check-ERROR")
		return BuildErrWithCode(http.StatusNotAcceptable, errors.New("check network error"))
	}
	klog.Info("IaaS-auth-OK")
	self.IaasTenantID = opstk.GetTenantID()
	iaas.SetIaaS(self.IaasTenantID, opstk)
	self.NetNum = 1
	self.IsCancelling = false
	self.Networks = dbaccessor.GetKeyOfNetworkGroup(self.PassID)
	self.Interfaces = dbaccessor.GetKeyOfInterfaceGroup(self.PassID)
	self.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	self.Quota = QuotaNoAdmin
	err = self.SaveIaasAndPaasInfoToDB()
	if err != nil {
		iaas.GetIaasObjMgrSingleton().Del(self.IaasTenantID)
		return BuildErrWithCode(http.StatusInternalServerError, errors.New("save tenant infomation error"))
	}
	_, err = CreateNetworkLan(self.PassID)
	if err != nil {
		deleteTenantDir(self.PassID)
		iaas.GetIaasObjMgrSingleton().Del(self.IaasTenantID)
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}
	return nil
}

func (self *ExclusiveTenant) SaveIaasAndPaasInfoToDB() error {
	opnCfg := &openstack.OpenStackConf{
		Username:   self.IaasUserName,
		Password:   self.IaasPaasword,
		Url:        self.IaasEndpoint,
		Tenantid:   self.IaasTenantID,
		TenantName: self.IaasTenantName,
	}
	err := iaas.SaveIaasTenantInfoToDB(opnCfg)
	if err != nil {
		klog.Errorf("SaveIaasTenantInfoToDB err: %v", err)
		return err
	}
	tenant := Tenant{
		TenantName:     self.PaasName,
		TenantUUID:     self.PassID,
		Networks:       self.Networks,
		Interfaces:     self.Interfaces,
		Quota:          self.Quota,
		NetNum:         self.NetNum,
		IsCancelling:   self.IsCancelling,
		CreateTime:     self.CreateTime,
		IaasTenantID:   self.IaasTenantID,
		IaasTenantName: self.IaasTenantName,
	}
	err = tenant.SaveTenantToEtcd()
	if err != nil {
		klog.Errorf("SaveTenantToEtcd err: %v", err)
		iaas.DelIaasTenantInfoFromDB(self.IaasTenantID)
		return err
	}
	return nil
}

func IsTenantExistInDB(tenantID string) bool {
	tenantKey := dbaccessor.GetKeyOfTenantSelf(tenantID)
	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err == nil {
		tenant := &Tenant{}
		json.Unmarshal([]byte(tenantValue), tenant)
		klog.Errorf("Tenant [%v] is already in etcd, status [is being deleted? = %v]!",
			tenant.TenantUUID, tenant.IsCancelling)
		return true
	}
	return false
}

func CreateNetworkLan(tenantID string) (string, error) {
	net := Net{}
	net.Network.Name = "lan"
	net.Subnet.Cidr = "100.100.0.0/16"
	net.Subnet.GatewayIp = "100.100.0.1"
	net.Public = false
	net.TenantUUID = tenantID
	net.Description = "Default tenant network"
	klog.Infof("Create Tenant uuid [ %v ]!", tenantID)
	err := net.Create()
	if err != nil {
		klog.Errorf("Create tenant [%v]'s lan network error!", net.TenantUUID)
		return "", err
	}
	return net.Network.Id, nil
}

func GetPaasTenantInfoFromDB(tenentID string) (*Tenant, error) {
	tenant, err := getTenantInfo(tenentID)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

var getAllTenantIds = func() ([]string, error) {
	key := dbaccessor.GetKeyOfTenants()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("getAllTenantIds: get all tenantIds FAILED, error: %v", err)
		return nil, err
	}

	tenantIds := []string{}
	for _, node := range nodes {
		tid := strings.TrimPrefix(node.Key, key+"/")
		tenantIds = append(tenantIds, tid)
	}

	klog.Infof("getAllTenantIds: get all tenantIds: %v SUCC", tenantIds)
	return tenantIds, nil
}

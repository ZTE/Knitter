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
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type IPMode int

const (
	MaxIpsInGroup = 32
	IpsPrefix     = "["
	IpsSuffix     = "]"
	EmptyIP       = ""

	_ IPMode = iota
	AddressMode
	CountMode
)

var igMutex sync.Mutex

func lockIG() {
	igMutex.Lock()
}
func unlockIG() {
	igMutex.Unlock()
}

type IPGroup struct {
	TenantID    string
	NetworkID   string
	SubnetID    string
	SubnetCidr  string
	SubnetPools []AllocationPool
	Name        string
	ID          string
	IPs         string
	IPsSlice    []string
	AddIPs      []string
	DelIPs      map[string]string
	SizeStr     string
	Size        int
	AddSize     int
	Mode        IPMode
}

type IPGroupInDB struct {
	ID        string   `json:"id"`
	TenantID  string   `json:"tenant_id"`
	Name      string   `json:"name"`
	NetworkID string   `json:"network_id"`
	IPs       []IPInDB `json:"ips"`
}

type IPInDB struct {
	IPAddr  string `json:"ip_addr"`
	Used    bool   `json:"used"`
	PortID  string `json:"port_id"`
	MacAddr string `json:"mac_addr"`
}

func getIPGroupsKey() string {
	return KnitterManagerKeyRoot + "/ipgroups"
}

func createIPGroupKey(igID string) string {
	return getIPGroupsKey() + "/" + igID
}

func (self *IPGroup) Create() (*IPGroupObject, error) {
	klog.Infof("Now in Create IpGroup Function")
	lockIG()
	defer unlockIG()
	if self.Name == "" {
		klog.Errorf("IpGroup Create error: name is empty")
		return nil, BuildErrWithCode(http.StatusBadRequest, errors.New("name is empty"))
	}

	err := self.CheckNet()
	if err != nil {
		klog.Errorf("IpGroup Create CheckNet error: [%v]", err.Error())
		return nil, err
	}

	err = self.Operate(nil)
	if err != nil {
		klog.Errorf("IpGroup Create operate error: [%v], id: [%v]", err.Error(), self.ID)
		return nil, err
	}

	return getIGFromCache(self.TenantID, self.ID)
}

func (self *IPGroup) analyzeUpdateIPs(igInDb *IPGroupInDB) error {
	//empty string means do not change ips
	if self.Mode == AddressMode {
		return self.analyzeUpdateIPStr(igInDb)
	}

	return self.analyzeUpdateIPCount(igInDb)
}

func (self *IPGroup) analyzeUpdateIPStr(igInDb *IPGroupInDB) error {
	//empty string means do not change ips
	if self.IPs == "" {
		klog.Debugf("IpGroup analyzeUpdateIps: ip is empty string, do not update ips")
		return nil
	}

	ipStrInDb := make([]string, 0)
	for _, ip := range igInDb.IPs {
		ipStrInDb = append(ipStrInDb, ip.IPAddr)
	}

	//calculate AddIps
	if self.IPs != "" {
		self.AddIPs = StringSliceDiff(self.IPsSlice, ipStrInDb)
	}

	//calculate DelIps
	DelIPAddrs := StringSliceDiff(ipStrInDb, self.IPsSlice)
	if self.DelIPs == nil {
		self.DelIPs = make(map[string]string, 0)
	}
	for _, ip := range igInDb.IPs {
		if InSliceString(ip.IPAddr, DelIPAddrs) {
			if ip.Used {
				klog.Errorf("IpGroup analyzeUpdateIps error: ip[%v] is in use, id: [%v]", ip.IPAddr, self.ID)
				return BuildErrWithCode(http.StatusConflict, errors.New("ip["+ip.IPAddr+"] is in use"))
			}
			self.DelIPs[ip.IPAddr] = ip.PortID
		}
	}

	return nil
}

func (self *IPGroup) AnalyzeIPs(igInDb *IPGroupInDB) error {
	if igInDb == nil {
		//create ip group scenario
		self.analyzeCreateIPs()
		return nil
	}

	//update ip group scenario
	err := self.analyzeUpdateIPs(igInDb)
	if err != nil {
		klog.Errorf("IpGroup analyzeUpdateIps error: [%v]", err.Error())
		return err
	}
	return nil
}

func (self *IPGroup) CheckNet() error {
	if self.NetworkID == "" {
		return BuildErrWithCode(http.StatusBadRequest, errors.New("network id is empty"))
	}

	// create ip group scenario
	netDB, err := GetNetObjRepoSingleton().Get(self.NetworkID)
	if err != nil {
		klog.Errorf("IpGroup operate get network error: [%v]", err.Error())
		return BuildErrWithCode(http.StatusNotFound, errors.New("network not exists"))
	}

	if netDB.IsPublic {
		if self.TenantID != constvalue.PaaSTenantAdminDefaultUUID {
			klog.Errorf("IpGroup operate error, can not operate public network")
			return BuildErrWithCode(http.StatusForbidden, errors.New("can not operate public network"))
		}
	} else if netDB.TenantID != self.TenantID {
		klog.Errorf("IpGroup operate failed, network tenant[%v] not same with input tenant[%v]", netDB.TenantID, self.TenantID)
		return BuildErrWithCode(http.StatusNotFound, errors.New("network not exists"))
	}

	self.SubnetID = netDB.SubnetID
	subnetDB, err := GetSubnetObjRepoSingleton().Get(self.SubnetID)
	if err != nil {
		klog.Errorf("IpGroup operate get subnet error: [%v]", err.Error())
		return err
	}
	self.SubnetCidr = subnetDB.CIDR
	self.SubnetPools = subnetDB.AllocPools
	return nil
}

func (self *IPGroup) Operate(igInDb *IPGroupInDB) error {
	if self.IsDuplicateName(igInDb) {
		klog.Errorf("IpGroup operate error, ip group name exists, name: [%v]", self.Name)
		return BuildErrWithCode(http.StatusConflict, errors.New("ip group name exists"))
	}

	if !self.IsValidIPs() {
		klog.Errorf("IpGroup operate error, invalid ips, name: [%v], ips: [%v]", self.Name, self.IPs)
		return BuildErrWithCode(http.StatusBadRequest, errors.New("invalid ips"))
	}

	err := self.AnalyzeIPs(igInDb)
	if err != nil {
		klog.Errorf("IpGroup analyzeIps error: [%v]", err.Error())
		return err
	}
	klog.Infof("IpGroup addIps: [%v], delIps: [%v], AddIPsCount: [%v]", self.AddIPs, self.DelIPs, self.AddSize)
	if igInDb == nil {
		igInDb = &IPGroupInDB{Name: self.Name, NetworkID: self.NetworkID, ID: self.ID, TenantID: self.TenantID}
	} else if self.Name != "" {
		igInDb.Name = self.Name
	}

	//todo check addips if available

	//create ips
	err = self.createIps(&(igInDb.IPs))
	if err != nil {
		klog.Errorf("IpGroup operate createIps error: [%v], id: [%v]", err.Error(), self.ID)
		return err
	}

	//delete ips
	failedips := make([]string, 0)
	for ipAddr, portID := range self.DelIPs {
		klog.Infof("IpGroup delete ip[%v] start", ipAddr)
		err = iaas.GetIaaS(self.TenantID).DeletePort(portID)
		if err != nil {
			klog.Errorf("IpGroup Delete DeletePort error: [%v], id: [%v], port_id: [%v]", err.Error(), self.ID, portID)
			failedips = append(failedips, ipAddr)
			continue
		}

		RemoveIPFromSlice(&(igInDb.IPs), ipAddr)
	}

	err = saveIGToDBAndCache(igInDb)
	if err != nil {
		klog.Errorf("IpGroup Create saveIpGroupToEtcd error: [%v], ipgroup: [%v]", err.Error(), self.ID)
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	if len(failedips) > 0 {
		klog.Errorf("IpGroup Delete error, failedips[%v]", failedips)
		return BuildErrWithCode(http.StatusInternalServerError, errors.New("ip["+strings.Join(failedips, ",")+"] delete failed"))
	}

	return nil
}

func (self *IPGroup) deleteIps(ips *[]IPInDB) error {
	failedips := make([]string, 0)
	klog.Infof("IpGroup Delete start: ips[%v]", *ips)
	for _, ip := range *ips {
		err := iaas.GetIaaS(self.TenantID).DeletePort(ip.PortID)
		if err != nil {
			klog.Errorf("IpGroup Delete deleteIp error: [%v], id: [%v], port_id: [%v]", err.Error(), self.ID, ip.PortID)
			failedips = append(failedips, ip.IPAddr)
		}
	}

	if len(failedips) > 0 {
		klog.Errorf("IpGroup Delete error, failedips[%v]", failedips)
		return BuildErrWithCode(http.StatusInternalServerError, errors.New("ip["+strings.Join(failedips, ",")+"] delete failed"))
	}

	return nil
}

func (self *IPGroup) createIps(ips *[]IPInDB) error {
	var err error
	var ports []*iaasaccessor.Interface = nil
	iaas := iaas.GetIaaS(self.TenantID)
	defer func() {
		if err != nil {
			for _, port := range ports {
				portErr := iaas.DeletePort(port.Id)
				if portErr != nil {
					klog.Warningf("createIps rollback iaas port failed. error: [%v], id: [%v]", portErr.Error(), port.Id)
				}
			}

			klog.Errorf("createIps failed. error: [%v]", err.Error())
		}
	}()

	bulkPortsReq := self.makeBulkPortsReq()
	klog.Infof("IpGroup bulkPortsReq: [%v]", *bulkPortsReq)

	if len(bulkPortsReq.Ports) > 0 {
		ports, err = iaas.CreateBulkPorts(bulkPortsReq)
		if err != nil {
			klog.Errorf("IpGroup Create CreateBulkPorts error: [%v], ipgroup: [%v]", err.Error(), self.ID)
			return BuildErrWithCode(http.StatusInternalServerError, err)
		}

		for _, port := range ports {
			*ips = append(*ips, IPInDB{
				IPAddr:  port.Ip,
				MacAddr: port.MacAddress,
				PortID:  port.Id,
				Used:    false,
			})
		}
	}

	return nil
}

func (self *IPGroup) IsDuplicateName(igInDb *IPGroupInDB) bool {
	if self.Name == "" {
		return false
	}

	igs := getTenantIGsFromCache(self.TenantID)
	netID := self.NetworkID
	if igInDb != nil {
		netID = igInDb.NetworkID
	}
	for _, ig := range igs {
		if ig.Name == self.Name && ig.NetworkID == netID && ig.ID != self.ID {
			return true
		}
	}

	return false
}

func (self *IPGroup) IsValidIPs() bool {
	if self.SizeStr == "" && self.IPs == "" {
		self.Mode = AddressMode
		return true
	} else if self.SizeStr != "" && self.IPs != "" {
		return false
	} else if self.SizeStr == "" && self.IPs != "" {
		self.Mode = AddressMode
		return self.isValidSpecificIPs()
	}

	self.Mode = CountMode
	return self.isValidIPCount()
}

//valid ips starts with "[" ends with "]", different ips seperate by ",", max ip numer is 32
//valid ips: 1. ""
//           2. "[1.1.1.1,1.1.1.2]"
//invalid ips: 1. "[]"
//             2. "dddd"
//             3. "[ddd]"
//             4. "[1.1.1]"
//             5. "[1.1.1.1/2.2.2.2]"
func (self *IPGroup) isValidSpecificIPs() bool {
	if !strings.HasPrefix(self.IPs, IpsPrefix) || !strings.HasSuffix(self.IPs, IpsSuffix) {
		return false
	}

	ipsStr := strings.TrimSuffix(strings.TrimPrefix(self.IPs, IpsPrefix), IpsSuffix)
	ips := strings.Split(ipsStr, ",")
	ips = StringSliceUnique(ips)

	//check ip number
	if len(ips) > MaxIpsInGroup {
		return false
	}
	for _, ip := range ips {
		if !isFixIPInCidr(ip, self.SubnetCidr) {
			return false
		}
	}

	self.IPsSlice = ips
	return true
}

func isFixIPInCidr(ip, cidr string) bool {
	ipObject := net.ParseIP(ip)
	if ipObject == nil {
		return false
	}

	//check ip is in cidr pool or not
	_, ipNet, _ := net.ParseCIDR(cidr)
	if ipNet != nil && !ipNet.Contains(ipObject) {
		return false
	}

	return true
}

func isFixIPInIPRanges(ip string, pools []AllocationPool) bool {
	for _, pool := range pools {
		subPool := subnets.AllocationPool{Start: pool.Start, End: pool.End}
		if IsFixIPInIPRange(ip, subPool) {
			return true
		}
	}

	return false
}

func (self *IPGroup) makeMgrPortReq(ip string) *mgriaas.MgrPortReq {
	agentPortReq := agtmgr.AgtPortReq{
		TenantID: self.TenantID,
		FixIP:    ip,
		VnicType: "normal",
	}
	return &mgriaas.MgrPortReq{
		TenantId:   self.TenantID,
		NetworkId:  self.NetworkID,
		SubnetId:   self.SubnetID,
		AgtPortReq: agentPortReq,
	}
}

func (self *IPGroup) makeBulkPortsReq() *mgriaas.MgrBulkPortsReq {
	if self.Mode == AddressMode {
		return self.makeBulkPortsReqForAddrMode()
	}

	return self.makeBulkPortsReqForCountMode()
}

func saveIGToDBAndCache(ig *IPGroupInDB) error {
	var errDB error
	var errCache error
	defer func() {
		if errCache != nil {
			deleteIGFromDB(ig.ID)
		}
	}()

	errDB = saveIGToDB(ig)
	if errDB != nil {
		klog.Errorf("saveIPGroupToDBAndCache saveIPGroupToEtcd error: [%v], id: [%v]", errDB.Error(), ig.ID)
		return errDB
	}

	igObject := TransIGInDBToIGObject(ig)
	_, errCache = GetIPGroupObjRepoSingleton().Get(ig.ID)
	if errCache != nil {
		errCache = GetIPGroupObjRepoSingleton().Add(igObject)
	} else {
		errCache = GetIPGroupObjRepoSingleton().Update(igObject)
	}

	if errCache != nil {
		klog.Errorf("saveIPGroupToDBAndCache save to cache error: [%v], id: [%v]", errCache.Error(), ig.ID)
		return errCache
	}

	return nil
}

func saveIGToDB(ig *IPGroupInDB) error {
	key := createIPGroupKey(ig.ID)
	value, _ := json.Marshal(*ig)
	klog.Infof("saveIGToDB IpGroupInfo: [%v]", string(value))
	err := common.GetDataBase().SaveLeaf(key, string(value))
	if err != nil {
		klog.Errorf("saveIGToDB error: [%v], key: [%v]", err.Error(), key)
		return err
	}

	return nil
}

func (self *IPGroup) GetIGsByNet() ([]*IPGroupObject, error) {
	return GetIPGroupObjRepoSingleton().ListByNetworkID(self.NetworkID)
}

func (self *IPGroup) GetIGs() ([]*IPGroupObject, error) {
	klog.Infof("Now in GetIGs Function")
	finalIGs := make([]*IPGroupObject, 0)
	privateIGs := getTenantIGsFromCache(self.TenantID)
	for _, privateIG := range privateIGs {
		if self.NetworkID == "" || self.NetworkID == privateIG.NetworkID {
			finalIGs = append(finalIGs, privateIG)
		}
	}

	if !IsAdminTenant(self.TenantID) {
		adminIGs := getTenantIGsFromCache(constvalue.PaaSTenantAdminDefaultUUID)
		for _, adminIG := range adminIGs {
			if self.NetworkID == "" || self.NetworkID == adminIG.NetworkID {
				finalIGs = append(finalIGs, adminIG)
			}
		}
	}

	return finalIGs, nil
}

func (self *IPGroup) GetIPGroupByName() (*IPGroupObject, error) {
	igs, err := self.GetIGs()
	if err != nil {
		klog.Errorf("GetIPGroupByName GetIGsByNetAndTenant error: [%v], name: [%v]", err.Error(), self.Name)
		return nil, err
	}

	for _, igObject := range igs {
		if igObject.Name == self.Name {
			return igObject, nil
		}
	}

	klog.Warningf("ip group not exist. name: [%v]", self.Name)
	return nil, errors.New("ip group not exist. group name:[" + self.Name + "]")
}

func deleteIGFromDBAndCache(igID string) error {
	err := deleteIGFromDB(igID)
	if err != nil && !etcd.IsNotFindError(err) {
		klog.Errorf("deleteIGFromDBAndCache deleteIGFromDB error: [%v], igID: [%v]", err.Error(), igID)
		return err
	}

	err = GetIPGroupObjRepoSingleton().Del(igID)
	if err != nil {
		klog.Errorf("deleteIGFromDBAndCache Del error: [%v], igID: [%v]", err.Error(), igID)
		return err
	}

	return nil
}

func deleteIGFromDB(igID string) error {
	key := createIPGroupKey(igID)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil {
		klog.Errorf("deleteIpGroupFromEtcd DeleteDir error: [%v], key: [%v]", err.Error(), key)
		return err
	}

	return nil
}

func getTenantIGsFromCache(tenantID string) []*IPGroupObject {
	igObjects, err := GetIPGroupObjRepoSingleton().ListByTenantID(tenantID)
	if err != nil {
		klog.Warningf("IpGroup ListByTenantID error: [%v], tenantID: [%v]", err.Error(), tenantID)
		return nil
	}

	return igObjects
}

func (self *IPGroup) Get() (*IPGroupObject, error) {
	return getIGFromCache(self.TenantID, self.ID)
}

func getIGFromCache(tenantID, igID string) (*IPGroupObject, error) {
	igObject, err := GetIPGroupObjRepoSingleton().Get(igID)
	if err != nil {
		klog.Errorf("getIGFromCache error: [%v], id: [%v]", err.Error(), igID)
		if strings.Contains(err.Error(), errobj.ErrRecordNotExist.Error()) {
			return nil, BuildErrWithCode(http.StatusNotFound, errors.New("ip group not exists"))
		}
		return nil, err
	}

	if !IsAdminTenant(igObject.TenantID) && tenantID != igObject.TenantID {
		klog.Errorf("getIGFromCache input tenant[%v] not same with real tenant[%v], id: [%v]", tenantID, igObject.TenantID, igID)
		return nil, BuildErrWithCode(http.StatusNotFound, errors.New("ip group not exists"))
	}

	return igObject, nil
}

func getIGWithTenantCheck(id, tenantID string) (*IPGroupObject, error) {
	igObject, err := GetIPGroupObjRepoSingleton().Get(id)
	if err != nil {
		klog.Errorf("getIGWithTenantCheck get ip group from cache error: [%v], id: [%v]", err.Error(), id)
		if strings.Contains(err.Error(), errobj.ErrRecordNotExist.Error()) {
			return nil, BuildErrWithCode(http.StatusNotFound, errors.New("ip group not exists"))
		}

		return nil, err
	}

	if tenantID != igObject.TenantID {
		if IsAdminTenant(igObject.TenantID) {
			klog.Errorf("getIGWithTenantCheck error, can not operate public network, tenant: [%v]", tenantID)
			return nil, BuildErrWithCode(http.StatusForbidden, errors.New("can not operate public network"))
		}

		return nil, BuildErrWithCode(http.StatusNotFound, errors.New("ip group not exists"))
	}

	return igObject, nil
}

func IsAdminTenant(tenantID string) bool {
	return tenantID == constvalue.PaaSTenantAdminDefaultUUID
}

func (self *IPGroup) GetIGWithCheck() (*IPGroupInDB, error) {
	igObject, err := getIGWithTenantCheck(self.ID, self.TenantID)
	if err != nil {
		klog.Errorf("CheckIGAndNetExist CheckIG error: [%v], id: [%v]", err.Error(), self.ID)
		return nil, err
	}

	self.NetworkID = igObject.NetworkID
	paasNet, err := GetNetObjRepoSingleton().Get(self.NetworkID)
	if err != nil {
		klog.Errorf("IpGroup Update get network error: [%v], net: [%v]", err.Error(), self.NetworkID)
		return nil, err
	}

	self.SubnetID = paasNet.SubnetID
	subnetDB, err := GetSubnetObjRepoSingleton().Get(self.SubnetID)
	if err != nil {
		klog.Errorf("IpGroup Update get subnet error: [%v]", err.Error())
		return nil, err
	}

	self.SubnetCidr = subnetDB.CIDR
	self.SubnetPools = subnetDB.AllocPools
	igInDb := TransIGObjectToIGInDB(igObject)
	return igInDb, nil
}

func (self *IPGroup) Update() (*IPGroupObject, error) {
	klog.Infof("Now in Update IpGroup Function")
	lockIG()
	defer unlockIG()

	igInDb, err := self.GetIGWithCheck()
	if err != nil {
		klog.Errorf("IpGroup Update CheckIGAndNetExist error: [%v], id: [%v]", err.Error(), self.ID)
		return nil, err
	}

	err = self.Operate(igInDb)
	if err != nil {
		klog.Errorf("IpGroup Update operate error: [%v], id: [%v]", err.Error(), self.ID)
		return nil, err
	}

	klog.Infof("leave IpGroup Update")
	return getIGFromCache(self.TenantID, self.ID)
}

func (self *IPGroup) ObtainIP() (*iaasaccessor.Interface, error) {
	klog.Infof("Now in ObtainIP IpGroup Function")
	lockIG()
	defer unlockIG()

	igObject, err := getIGFromCache(self.TenantID, self.ID)
	if err != nil {
		klog.Errorf("ObtainIP getIGFromCache error: [%v], id: [%v]", err.Error(), self.ID)
		return nil, err
	}
	klog.Infof("ObtainIP get ig: [%+v]", igObject)

	igInDb := TransIGObjectToIGInDB(igObject)
	for _, ip := range igInDb.IPs {
		if !ip.Used {
			port, err := iaas.GetIaaS(self.TenantID).GetPort(ip.PortID)
			if err != nil {
				klog.Errorf("ObtainIP GetPort error: [%v], portID: [%v]", err.Error(), ip.PortID)
				continue
			}
			port.IPGroupID = self.ID

			newIGInDB := makeNewIPGroupObject(igInDb, ip.PortID, true)
			err = saveIGToDBAndCache(newIGInDB)
			if err != nil {
				klog.Errorf("ObtainIP saveIGToDBAndCache error: [%v], portID: [%v]", err.Error(), ip.PortID)
				continue
			}
			return port, nil
		}
	}
	return nil, errors.New("no available IP. group name:[" + igInDb.Name + "]")
}

func makeNewIPGroupObject(oldIG *IPGroupInDB, portID string, used bool) *IPGroupInDB {
	newIG := &IPGroupInDB{ID: oldIG.ID, TenantID: oldIG.TenantID, Name: oldIG.Name, NetworkID: oldIG.NetworkID}
	for _, newIP := range oldIG.IPs {
		if portID == newIP.PortID {
			newIP.Used = used
		}
		newIG.IPs = append(newIG.IPs, newIP)
	}

	return newIG
}

func (self *IPGroup) ReleaseIP(portID string) error {
	klog.Infof("Now in ReleaseIP IpGroup Function, portId: [%v]", portID)
	lockIG()
	defer unlockIG()

	igObject, err := getIGFromCache(self.TenantID, self.ID)
	if err != nil {
		klog.Errorf("ReleaseIP getIGFromCache error: [%v], id: [%v]", err.Error(), self.ID)
		return err
	}
	klog.Infof("ReleaseIP get ig: [%+v]", igObject)

	igInDb := TransIGObjectToIGInDB(igObject)
	for _, ip := range igInDb.IPs {
		if ip.PortID == portID && ip.Used {
			newIGInDB := makeNewIPGroupObject(igInDb, ip.PortID, false)
			err = saveIGToDBAndCache(newIGInDB)
			if err != nil {
				klog.Errorf("ReleaseIP saveIGToDBAndCache error: [%v], portID: [%v]", err.Error(), ip.PortID)
				return BuildErrWithCode(http.StatusInternalServerError, err)
			}

			return nil
		}
	}

	return nil
}

func (self *IPGroup) Delete() error {
	klog.Infof("Now in Delete IpGroup Function")
	lockIG()
	defer unlockIG()

	igObject, err := getIGWithTenantCheck(self.ID, self.TenantID)
	if err != nil {
		klog.Errorf("IpGroup Delete CheckIG error: [%v], id: [%v]", err.Error(), self.ID)
		return err
	}

	for _, ip := range igObject.IPs {
		if ip.Used {
			klog.Errorf("IpGroup Delete error: ip[%v] is in use, id: [%v]", ip.IPAddr, self.ID)
			return BuildErrWithCode(http.StatusConflict, errors.New("ip["+ip.IPAddr+"] is in use"))
		}
	}

	//delete ips
	err = self.deleteIps(&(igObject.IPs))
	if err != nil {
		return err
	}

	err = deleteIGFromDBAndCache(self.ID)
	if err != nil {
		klog.Errorf("IpGroup Delete deleteIGFromDBAndCache error: [%v], id: [%v]", err.Error(), self.ID)
		return BuildErrWithCode(http.StatusInternalServerError, err)
	}

	return nil
}

// SliceDiff returns diff slice of slice1 - slice2.
func StringSliceDiff(slice1, slice2 []string) (diffslice []string) {
	for _, v := range slice1 {
		if !InSliceString(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}

// InSliceString checks given string in string slice.
func InSliceString(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// SliceUnique cleans repeated values in slice.
func StringSliceUnique(slice []string) (uniqueslice []string) {
	for _, v := range slice {
		if !InSliceString(v, uniqueslice) {
			uniqueslice = append(uniqueslice, v)
		}
	}
	return
}

func RemoveIPFromSlice(s *[]IPInDB, ipAddr string) {
	for k, ip := range *s {
		if ip.IPAddr == ipAddr {
			*s = append((*s)[:k], (*s)[k+1:]...)
			return
		}
	}
}

func TransIGObjectToIGInDB(igObject *IPGroupObject) *IPGroupInDB {
	return &IPGroupInDB{
		ID:        igObject.ID,
		Name:      igObject.Name,
		TenantID:  igObject.TenantID,
		NetworkID: igObject.NetworkID,
		IPs:       igObject.IPs,
	}
}

func TransIGInDBToIGObject(igInDB *IPGroupInDB) *IPGroupObject {
	return &IPGroupObject{
		ID:        igInDB.ID,
		Name:      igInDB.Name,
		TenantID:  igInDB.TenantID,
		NetworkID: igInDB.NetworkID,
		IPs:       igInDB.IPs,
	}
}

func (self *IPGroup) analyzeUpdateIPCount(igInDb *IPGroupInDB) error {
	count, _ := strconv.Atoi(self.SizeStr)
	self.Size = count

	changed := self.Size - len(igInDb.IPs)
	if changed >= 0 {
		self.AddSize = changed
		return nil
	}

	if self.DelIPs == nil {
		self.DelIPs = make(map[string]string, 0)
	}
	for _, ip := range igInDb.IPs {
		if !ip.Used {
			self.DelIPs[ip.IPAddr] = ip.PortID
			changed++
		}

		if changed == 0 {
			break
		}
	}
	if changed < 0 {
		return BuildErrWithCode(http.StatusConflict, errors.New("ip_count invalid"))
	}

	return nil
}

func (self *IPGroup) analyzeCreateIPs() {
	if self.Mode == AddressMode {
		if len(self.IPsSlice) > 0 {
			self.AddIPs = self.IPsSlice
		}
	} else {
		count, _ := strconv.Atoi(self.SizeStr)
		self.Size = count
		self.AddSize = self.Size
	}
}

func (self *IPGroup) isValidIPCount() bool {
	count, err := strconv.Atoi(self.SizeStr)
	if err != nil || count > MaxIpsInGroup || count < 0 {
		return false
	}

	return true
}

func (self *IPGroup) makeBulkPortsReqForAddrMode() *mgriaas.MgrBulkPortsReq {
	portsReq := &mgriaas.MgrBulkPortsReq{}
	for _, ip := range self.AddIPs {
		if ip == "" {
			continue
		}

		portsReq.Ports = append(portsReq.Ports, self.makeMgrPortReq(ip))
	}

	return portsReq
}

func (self *IPGroup) makeBulkPortsReqForCountMode() *mgriaas.MgrBulkPortsReq {
	portsReq := &mgriaas.MgrBulkPortsReq{}
	for i := 0; i < self.AddSize; i++ {
		portsReq.Ports = append(portsReq.Ports, self.makeMgrPortReq(EmptyIP))
	}

	return portsReq
}

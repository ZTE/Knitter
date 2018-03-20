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

package brintsubrole

import (
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/knitter-agent/infra/alg"
	"github.com/ZTE/Knitter/pkg/klog"
	"os"
	"sync"
)

type TenantNetworkValue struct {
	Vni    int             `json:"vni"`
	VlanID string          `json:"vlan_id"`
	PodIds alg.StringSlice `json:"pod_ids"`
}

type TenantNetworkTableRole struct {
	// key: network_id
	tenantNetworkMap map[string]TenantNetworkValue
	lock             sync.RWMutex
	dp               infra.DataPersister
}

func (this *TenantNetworkTableRole) Get(networkID string) (*TenantNetworkValue, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	value, ok := this.tenantNetworkMap[networkID]
	if ok {
		return &value, nil
	}
	klog.Errorf("netId[%s]'s value not found", networkID)
	return nil, errobj.ErrRecordNtExist
}

func (this *TenantNetworkTableRole) GetAll() map[string]TenantNetworkValue {
	return this.tenantNetworkMap
}

func (this *TenantNetworkTableRole) Load() error {
	klog.Info("attempt to load TenantNetworkTable.")
	if this.tenantNetworkMap == nil {
		this.tenantNetworkMap = make(map[string]TenantNetworkValue)
	}

	_, err := os.Stat("/dev/shm/" + this.dp.DirName + "/" + this.dp.FileName)
	if err != nil {
		klog.Errorf("TenantNetworkTable: get, TenantNetworkTable table doesn't exist!, err: %v", err)
		return infra.ErrGetRAMDiskFailed
	}

	err = this.dp.LoadFromMemFile(&this.tenantNetworkMap)
	if err != nil {
		klog.Errorf("TenantNetworkTable: get, restore TenantNetworkTable table failed!, err: %v", err)
		return err
	}

	return nil
}

func (this *TenantNetworkTableRole) Insert(networkID string, vni int, vlanID string) error {
	klog.Info("In-TenantNetworkTable-Insert")
	this.lock.Lock()
	defer this.lock.Unlock()
	this.tenantNetworkMap[networkID] = TenantNetworkValue{VlanID: vlanID, Vni: vni, PodIds: alg.NewStringSlice()}

	klog.Info("SaveToMemFile", this.dp.DirName, "-----", this.dp.FileName)
	err := this.dp.SaveToMemFile(this.tenantNetworkMap)
	if err != nil {
		klog.Errorf("tenantNetworkMap: insert, store to ram failed!")
		return err
	}
	klog.Info("SaveToMemFile")
	return nil
}

func (this *TenantNetworkTableRole) Delete(netID string) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.tenantNetworkMap, netID)

	err := this.dp.SaveToMemFile(this.tenantNetworkMap)
	if err != nil {
		klog.Infof("tenantNetworkMap: delete, store to ram failed!")
		return err
	}
	return nil
}

func (this *TenantNetworkTableRole) IncRefCount(networkID, podNs, podName string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	value, ok := this.tenantNetworkMap[networkID]
	if !ok {
		klog.Infof("IncRefCount: networkId: %s not found", networkID)
		return errobj.ErrMapNtFound
	}

	podID := podNs + ":" + podName
	err := value.PodIds.Add(podID)
	if err == alg.ErrElemExist {
		klog.Infof("TenantNetworkTableRole:IncRefCount podId has existed!")
		return nil
	}
	this.tenantNetworkMap[networkID] = value

	err = this.dp.SaveToMemFile(this.tenantNetworkMap)
	if err != nil {
		klog.Infof("IncRefCount: increase store to ram failed!")
		return err
	}
	return nil
}

func (this *TenantNetworkTableRole) DecRefCount(networkID, podNs, podName string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	value, ok := this.tenantNetworkMap[networkID]
	if !ok {
		klog.Infof("DecRefCount: networkId: %s not found", networkID)
		return errobj.ErrMapNtFound
	}

	podID := podNs + ":" + podName
	err := value.PodIds.Remove(podID)
	if err == alg.ErrElemNtExist {
		klog.Infof("DecRefCount:decCount podId: %s not exist", podID)
		return nil
	}
	this.tenantNetworkMap[networkID] = value

	err = this.dp.SaveToMemFile(this.tenantNetworkMap)
	if err != nil {
		klog.Infof("DecRefCount: decrease store to ram failed!")
		return err
	}
	return nil
}

func (this *TenantNetworkTableRole) NeedDelete(networkID string) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	value, ok := this.tenantNetworkMap[networkID]
	if !ok {
		klog.Infof("DecRefCount: networkId: %s not found", networkID)
		return false
	}

	if len(value.PodIds) == 0 {
		return true
	}

	return false
}

var tenantNetworkTableSingleton *TenantNetworkTableRole
var tenantNetworkTableSingletonLock sync.Mutex

func GetTenantNetworkTableSingleton() *TenantNetworkTableRole {
	tenantNetworkTableSingletonLock.Lock()
	defer tenantNetworkTableSingletonLock.Unlock()

	if tenantNetworkTableSingleton == nil {
		tenantNetworkTableSingleton = &TenantNetworkTableRole{
			make(map[string]TenantNetworkValue),
			sync.RWMutex{},
			infra.DataPersister{
				DirName:  "brintBridge",
				FileName: "brintTenantNetworkTable.dat",
			},
		}
	}

	return tenantNetworkTableSingleton
}

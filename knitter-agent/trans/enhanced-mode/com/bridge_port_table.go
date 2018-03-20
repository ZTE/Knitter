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

package com

import (
	"fmt"
	"sync"

	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/knitter-agent/infra/memtbl"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/pkg/klog"
)

type TenantNetworkID struct {
	TenantID     string `json:"tenant_id"`
	NetworkPlane string `json:"network_plane"`
	NetworkID    string `json:"network_id"`
}

type Br0PortRepo interface {
	Load() error
	memtbl.MemTblOp
}

type Br0PortValue struct {
	TenantNetID TenantNetworkID `json:"tenant_network_id"`
	PortID      string          `json:"port_id"`
}

type Br0PortTable struct {
	// key: br0VethName
	// value: br0PortValue
	Br0PortMap map[string]interface{}
	rwLock     sync.RWMutex
	dp         infra.DataPersister
}

// key: br0VethName
func (this *Br0PortTable) Get(key string) (interface{}, error) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	value, ok := this.Br0PortMap[key]
	if ok {
		return value.(Br0PortValue), nil
	}
	return nil, errobj.ErrRecordNtExist
}

func (this *Br0PortTable) GetAll() (map[string]interface{}, error) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()

	retMap := make(map[string]interface{})
	for key, value := range this.Br0PortMap {
		retMap[key] = value
	}

	return retMap, nil
}

func (this *Br0PortTable) Load() error {
	klog.Infof("attempt to load br0VethTable.")
	_, err := osencap.OsStat(this.dp.GetFilePath())
	if err != nil {
		klog.Errorf("Br0PortTable:get, br0 table doesn't exist!, err: %v", err)
		return fmt.Errorf("%v:br0 veth table doesn't exist", err)
	}
	portMap := make(map[string]Br0PortValue)
	err = adapter.DataPersisterLoadFromMemFile(&this.dp, &portMap)
	if err != nil {
		klog.Errorf("Br0PortTable:get, restore br0 veth table failed!, err: %v", err)
		return fmt.Errorf("%v:load br0 veth table failed", err)
	}
	for k, v := range portMap {
		this.Br0PortMap[k] = v
	}
	klog.Errorf("Br0PortTable:Load br0 port table: %v", this.Br0PortMap)

	return nil
}

// key: br0VethName, value: br0PortValue{TenantNetworkId{tanentId, networkPlane, networkId}, portId}
//func (this *br0PortTable) Insert(br0VethName, tanentId, networkPlane, networkId, portId string) error {
func (this *Br0PortTable) Insert(key string, value interface{}) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.Br0PortMap[key] = value

	err := adapter.DataPersisterSaveToMemFile(&this.dp, this.Br0PortMap)
	if err != nil {
		delete(this.Br0PortMap, key)
		klog.Errorf("Br0PortTable:insert, store to ram failed!")
		return fmt.Errorf("%v:store to ram failed", err)
	}
	return nil
}

// key: br0VethName
func (this *Br0PortTable) Delete(key string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	delete(this.Br0PortMap, key)

	err := adapter.DataPersisterSaveToMemFile(&this.dp, this.Br0PortMap)
	if err != nil {
		klog.Errorf("Br0PortTable:delete, store to ram failed!")
		return fmt.Errorf("%v:store to ram failed", err)
	}
	return nil
}

var br0PortTableSingleton Br0PortRepo
var br0PortTableSingletonLock sync.Mutex

var GetBr0PortTableSingleton = func() Br0PortRepo {
	if br0PortTableSingleton != nil {
		return br0PortTableSingleton
	}

	br0PortTableSingletonLock.Lock()
	defer br0PortTableSingletonLock.Unlock()
	if br0PortTableSingleton == nil {
		br0PortTableSingleton = &Br0PortTable{
			make(map[string]interface{}),
			sync.RWMutex{},
			infra.DataPersister{DirName: "br0Bridge", FileName: "br0PortTable.dat"}}
	}
	return br0PortTableSingleton
}

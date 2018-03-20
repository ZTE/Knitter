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

package portrole

import (
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
	"os"
	"sync"
)

type PortTableRole struct {
	// key: portId
	// value: portName
	portMap map[string]string
	rwLock  sync.RWMutex
	dp      infra.DataPersister
}

func (this *PortTableRole) Get(portID string) (string, error) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	portName, ok := this.portMap[portID]
	if ok {
		return portName, nil
	}
	return "", errobj.ErrRecordNtExist
}

func (this *PortTableRole) Load() error {
	klog.Infof("attempt to load portTable.")
	_, err := os.Stat(this.dp.GetFilePath())
	if err != nil {
		klog.Errorf("PortTableRole:Load, port table doesn't exist!, err: %v", err)
		return errobj.ErrTblNtExist
	}
	err = this.dp.LoadFromMemFile(&this.portMap)
	if err != nil {
		klog.Errorf("PortTableRole:Load, restore port table failed!, err: %v", err)
		return err
	}
	return nil
}

func (this *PortTableRole) Insert(portID, portName string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.portMap[portID] = portName

	err := this.dp.SaveToMemFile(this.portMap)
	if err != nil {
		klog.Errorf("PortTableRole:Insert, store to ram failed!")
		return err
	}
	return nil
}

func (this *PortTableRole) Delete(portID string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	delete(this.portMap, portID)

	err := this.dp.SaveToMemFile(this.portMap)
	if err != nil {
		klog.Errorf("PortTableRole:delete, store to ram failed!")
		return err
	}
	return nil
}

func (this *PortTableRole) IsExist(vethName string) bool {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()

	for _, value := range this.portMap {
		if vethName == value {
			klog.Infof("PortTableRole:IsExist: veth:%s exist in bridge br-int", vethName)
			return true
		}
	}
	return false
}

var portTableSingleton *PortTableRole
var br0PortTableSingletonLock sync.Mutex

func GetPortTableSingleton() *PortTableRole {
	if portTableSingleton != nil {
		return portTableSingleton
	}

	br0PortTableSingletonLock.Lock()
	defer br0PortTableSingletonLock.Unlock()
	if portTableSingleton == nil {
		portTableSingleton = &PortTableRole{
			make(map[string]string),
			sync.RWMutex{},
			infra.DataPersister{
				DirName:  "bridge",
				FileName: "portTable.dat",
			},
		}
		err := portTableSingleton.Load()
		if err != nil && err.Error() != errobj.ErrTblNtExist.Error() {
			klog.Errorf("GetPortTableSingleton: portTableSingleton.Load() error: %v", err)
			return nil
		}

	}
	return portTableSingleton
}

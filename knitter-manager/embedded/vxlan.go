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

package networkserver

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	LOG "github.com/ZTE/Knitter/pkg/klog"
	"strconv"
	"sync"
)

const (
	StartVxlanID int = 5000
	EndVxlanID   int = 15000
	ErrVxlanID   int = 88888
)

type VxlanIDManager struct {
	lock   sync.Mutex
	IDList map[string]*bool
}

var vxlanManager *VxlanIDManager

func GetVxlanManager() *VxlanIDManager {
	if vxlanManager == nil {
		vxlan := VxlanIDManager{}
		vxlan.load()
		vxlanManager = &vxlan
	}
	return vxlanManager
}

func (_ *VxlanIDManager) GetErrVxlanID() string {
	return strconv.Itoa(ErrVxlanID)
}

func (self *VxlanIDManager) Alloc() (string, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.IDList == nil {
		self.IDList = make(map[string]*bool)
	}
	for i := StartVxlanID; i < EndVxlanID; i++ {
		id := strconv.Itoa(i)
		if self.IDList[id] == nil {
			used := true
			self.IDList[id] = &used
			err := self.save()
			if err != nil {
				delete(self.IDList, id)
				continue
			}
			return id, nil
		}
	}

	return strconv.Itoa(ErrVxlanID), errors.New("can-not-alloc-vxlan-id")
}

func (self *VxlanIDManager) Free(id string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.IDList[id] == nil {
		return errors.New("free-vxlan-id-error:not-in-use")
	}
	idItem := self.IDList[id]
	delete(self.IDList, id)
	err := self.save()
	if err != nil {
		self.IDList[id] = idItem
		return err
	}
	return nil
}

func (self *VxlanIDManager) load() (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	key := dbaccessor.GetKeyOfEmbeddedServerVnis()
	value, err := ReadData(key)
	if err != nil {
		LOG.Error("Read-embedded-server-data-from-etcd-error")
		return err
	}
	return json.Unmarshal([]byte(value), &self)
}

func (self *VxlanIDManager) save() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerVnis()
	value, err := json.Marshal(self)
	if err != nil {
		LOG.Error("Marshal-ERROR:", err.Error())
		return err
	}

	err = SaveData(key, string(value))
	if err != nil {
		LOG.Error("Save-embedded-server-VxlanIDManager-data-to-etcd-error")
		return err
	}
	return nil
}

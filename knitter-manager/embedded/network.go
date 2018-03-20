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
	iaas "github.com/ZTE/Knitter/pkg/iaas-accessor"
	LOG "github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
	"strings"
	"sync"
)

/*************************************************************************/
type NetworkExtenAttrs struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	NetworkType     string `json:"provider:network_type"`
	PhysicalNetwork string `json:"provider:physical_network"`
	SegmentationID  string `json:"provider:segmentation_id"`
}

func (self *NetworkExtenAttrs) load(id string) (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerNetworkID(id)
	value, err := ReadData(key)
	if err != nil {
		LOG.Error("Read-embedded-server-data-from-etcd-error")
		return err
	}
	return json.Unmarshal([]byte(value), &self)
}

func (self *NetworkExtenAttrs) save() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerNetworkID(self.ID)
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

func (self *NetworkExtenAttrs) delete() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerNetworkID(self.ID)
	err = DeleteData(key)
	if err != nil {
		LOG.Error("Save-embedded-server-VxlanIDManager-data-to-etcd-error")
		return err
	}
	return nil
}

type Networks struct {
	lock sync.RWMutex
	list map[string]*NetworkExtenAttrs
}

var netManager *Networks

func GetNetManager() *Networks {
	if netManager == nil {
		new := Networks{}
		new.list = make(map[string]*NetworkExtenAttrs)
		new.load()
		netManager = &new
	}
	return netManager
}

func (self *Networks) load() error {
	self.lock.Lock()
	defer self.lock.Unlock()
	key := dbaccessor.GetKeyOfEmbeddedServerNetworks()
	nodes, err := ReadDataDir(key)
	if err != nil {
		LOG.Warning("Read Network dir[", key,
			"] from ETCD Error:", err)
		return nil
	}

	for _, node := range nodes {
		id := strings.TrimPrefix(node.Key, key+"/")
		item := NetworkExtenAttrs{}
		err = item.load(id)
		if err != nil {
			continue
		}
		self.list[item.ID] = &item
	}
	return nil
}

func (self *Networks) IsExistNetwork(id string) bool {
	return (self.list[id] != nil)
}

func (self *Networks) GetNetworkSegmentationID(netID string) string {
	if self.IsExistNetwork(netID) {
		return self.list[netID].SegmentationID
	}
	return GetVxlanManager().GetErrVxlanID()
}

/*************************************************************************/
func (self *Networks) CreateNetwork(name string) (*iaas.Network, error) {
	LOG.Info("EMBEDDED-CreateNetwork:", name)

	vni, err := GetVxlanManager().Alloc()
	if err != nil {
		LOG.Error("EMBEDDED-CreateNetwork-ERROR:", err)
		return nil, err
	}

	LOG.Info("EMBEDDED-Alloc-Vxlan:", vni)
	newNetwork := NetworkExtenAttrs{Name: name,
		ID: uuid.NewUUID()}
	if self.IsExistNetwork(newNetwork.ID) {
		GetVxlanManager().Free(vni)
		LOG.Error("EMBEDDED-CreateNetwork-ERROR:", "network-exist-now")
		return nil, errors.New("network-exist-now")
	}

	newNetwork.NetworkType = "vxlan"
	newNetwork.PhysicalNetwork = "paas-network-default"
	newNetwork.SegmentationID = vni

	err = newNetwork.save()
	if err != nil {
		return nil, err
	}

	self.lock.Lock()
	defer self.lock.Unlock()
	self.list[newNetwork.ID] = &newNetwork
	LOG.Infof("EMBEDDED-creat-network: %+v", newNetwork)
	return &iaas.Network{Name: newNetwork.Name, Id: newNetwork.ID}, nil
}

func (self *Networks) DeleteNetwork(id string) error {
	LOG.Infof("EMBEDDED-DeleteNetwork:", id)

	if self.IsExistNetwork(id) == false {
		LOG.Error("EMBEDDED-DeleteNetwork-Error:", id)
		return errors.New("can-not-find-network")
	}

	delNetwork := self.list[id]

	//delete-subnet-if-exist
	sid, err := GetSubnetManager().GetSubnetID(id)
	if err == nil && sid != "" {
		err := GetSubnetManager().DeleteSubnet(sid)
		if err != nil {
			LOG.Error("Delete-subnet[", sid, "]in-network[",
				id, "]-ERROR:", err.Error())
			return err
		}
	}

	err = GetVxlanManager().Free(self.list[id].SegmentationID)
	if err != nil {
		return err
	}
	err = delNetwork.delete()
	if err != nil {
		return err
	}

	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.list, id)
	return nil
}

func (self *Networks) GetNetworkID(networkName string) (string, error) {
	LOG.Info("EMBEDDED-GetNetworkID:", networkName)
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, v := range self.list {
		if v.Name == networkName {
			LOG.Info("EMBEDDED-GetNetworkID:name[", networkName,
				"]id[", v.ID, "]")
			return v.ID, nil
		}
	}
	LOG.Error("EMBEDDED-GetNetworkID-Error:", networkName)
	return "", errors.New("can-not-find-network")
}

func (self *Networks) GetNetwork(id string) (*NetworkExtenAttrs, error) {
	LOG.Info("EMBEDDED-GetNetwork:", id)
	self.lock.RLock()
	defer self.lock.RUnlock()

	if self.IsExistNetwork(id) == false {
		LOG.Error("EMBEDDED-GetNetwork-Error:", id)
		return nil, errors.New("can-not-find-network")
	}
	LOG.Infof("EMBEDDED-get-network: %+v", self.list[id])
	return self.list[id], nil
}

func (self *Networks) GetNetworkExtenAttrs(
	id string) (*NetworkExtenAttrs, error) {
	LOG.Info("EMBEDDED-GetNetworkExtenAttrs:", id)
	self.lock.RLock()
	defer self.lock.RUnlock()

	if self.IsExistNetwork(id) == false {
		LOG.Error("EMBEDDED-GetNetworkExtenAttrs-Error:", id)
		return nil, errors.New("can-not-find-network")
	}
	LOG.Infof("EMBEDDED-creat-network: %+v", self.list[id])
	return self.list[id], nil
}

func (self *Networks) GetAttachReq() int {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return 0
}

func (self *Networks) SetAttachReq(req int) {
	self.lock.Lock()
	defer self.lock.Unlock()
}

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
	"fmt"
	"net"

	"github.com/ZTE/Knitter/pkg/db-accessor"
	iaas "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	LOG "github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
	"strings"
	"sync"
)

type Interface struct {
	Name       string `json:"name"`
	ID         string `json:"port_id"`
	IP         string `json:"ip"`
	MacAddress string `json:"mac_address"`
	NetworkID  string `json:"network_id"`
	SubnetID   string `json:"subnet_id"`
}

func (self *Interface) load(id string) (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerPortID(id)
	value, err := ReadData(key)
	if err != nil {
		LOG.Error("Read-embedded-server-data-from-etcd-error")
		return err
	}
	return json.Unmarshal([]byte(value), &self)
}

func (self *Interface) save() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerPortID(self.ID)
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

func (self *Interface) delete() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerPortID(self.ID)
	err = DeleteData(key)
	if err != nil {
		LOG.Error("Save-embedded-server-VxlanIDManager-data-to-etcd-error")
		return err
	}
	return nil
}

type Interfaces struct {
	lock sync.RWMutex
	list map[string]*Interface
}

var portManager *Interfaces

func GetPortManager() *Interfaces {
	if portManager == nil {
		new := Interfaces{}
		new.list = make(map[string]*Interface)
		new.load()
		portManager = &new
	}
	return portManager
}

func (self *Interfaces) load() error {
	self.lock.Lock()
	defer self.lock.Unlock()
	key := dbaccessor.GetKeyOfEmbeddedServerPorts()
	nodes, err := ReadDataDir(key)
	if err != nil {
		LOG.Warning("Read Network dir[", key,
			"] from ETCD Error:", err)
		return nil
	}

	for _, node := range nodes {
		id := strings.TrimPrefix(node.Key, key+"/")
		item := Interface{}
		err = item.load(id)
		if err != nil {
			continue
		}
		self.list[item.ID] = &item
	}
	return nil
}

func (self *Interfaces) IsExistPort(id string) bool {
	return (self.list[id] != nil)
}

func (self *Interfaces) isExistPortOnSubNet(sid string) bool {
	for _, p := range self.list {
		if p.SubnetID == sid {
			return true
		}
	}
	return false
}

func getMacAddr(ip []byte) string {
	//fa:16:3e:1a:5d:91
	newMac := fmt.Sprintf("fa:16:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])
	return newMac
}

func paasPort2IaasPort(port *Interface) *iaas.Interface {
	tmpPort := iaas.Interface{}
	tmpPort.Ip = port.IP
	tmpPort.Id = port.ID
	tmpPort.Name = port.Name
	tmpPort.MacAddress = port.MacAddress
	tmpPort.NetworkId = port.NetworkID
	tmpPort.SubnetId = port.SubnetID
	return &tmpPort
}

func (self *Interfaces) CreatePort(networkID, subnetID, networkPlane,
	ip, mac, vnicType string) (*iaas.Interface, error) {
	LOG.Info("EMBEDDED-CreatePort:[", networkID, "][", subnetID, "]")

	if GetNetManager().IsExistNetwork(networkID) == false {
		LOG.Error("EMBEDDED-CreatePort-error:[", networkID, "][is-not-exist]")
		return nil, errors.New("create-port-error:network-isnot-exist")
	}

	if GetSubnetManager().IsExistSubnet(subnetID) == false {
		LOG.Error("EMBEDDED-CreatePort-error:[", subnetID, "][is-not-exist]")
		return nil, errors.New("create-port-error:subnet-isnot-exist")
	}

	ip, err := GetSubnetManager().allocIP(subnetID, ip)
	if err != nil {
		LOG.Error("EMBEDDED-CreatePort-error:[", subnetID, "][alloc-ip-error]")
		return nil, err
	}
	newMac := getMacAddr(net.ParseIP(ip).To4())
	newPort := Interface{}
	newPort.IP = ip
	newPort.ID = uuid.NewUUID()
	newPort.NetworkID = networkID
	newPort.SubnetID = subnetID
	newPort.Name = networkPlane
	newPort.MacAddress = newMac

	err = newPort.save()
	if err != nil {
		return nil, err
	}
	self.lock.Lock()
	defer self.lock.Unlock()
	self.list[newPort.ID] = &newPort

	LOG.Infof("EMBEDDED-create-subnet: %+v", newPort)
	return paasPort2IaasPort(&newPort), nil
}

func (self *Interfaces) CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*iaas.Interface, error) {
	LOG.Info("EMBEDDED-CreateBulkPorts-enter")
	if req == nil {
		LOG.Warning("EMBEDDED-CreateBulkPorts-req-is-nil")
		return nil, nil
	}

	ports := make([]*iaas.Interface, 0)
	for _, reqPort := range req.Ports {
		port, err := self.CreatePort(reqPort.NetworkId, reqPort.SubnetId, reqPort.PortName,
			reqPort.FixIP, "", "normal")
		if err != nil {
			//rollback
			LOG.Infof("EMBEDDED-CreateBulkPorts-start-rollback")
			for _, succPort := range ports {
				self.DeletePort(succPort.Id)
			}
			return nil, errors.New("Create-Bulk-Ports-failed: " + err.Error())
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func (self *Interfaces) GetPort(id string) (*iaas.Interface, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.getPort(id)
}

func (self *Interfaces) getPort(id string) (*iaas.Interface, error) {
	if self.IsExistPort(id) {
		return paasPort2IaasPort(self.list[id]), nil
	}
	return nil, errors.New("get-port-error:not-exist")
}

func (self *Interfaces) DeletePort(id string) error {
	if self.IsExistPort(id) == false {
		return errors.New("delete-port-error:not-exist")
	}

	delPort := self.list[id]
	err := GetSubnetManager().freeIP(delPort.SubnetID, delPort.IP)
	if err != nil {
		return err
	}
	err = delPort.delete()
	if err != nil {
		return err
	}
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.list, id)
	return nil
}

func (self *Interfaces) ListPorts(networkID string) ([]*iaas.Interface, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	ports := make([]*iaas.Interface, 0)
	for _, port := range self.list {
		if port.NetworkID == networkID {
			iaasPort := paasPort2IaasPort(port)
			ports = append(ports, iaasPort)
			LOG.Tracef("ListPorts: add port[%v] to result array", iaasPort)
		}
	}

	LOG.Tracef("ListPorts: list SUCC, result: %+v", ports)
	return ports, nil
}

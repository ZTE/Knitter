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

package bridgerole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brcom-sub-role"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brint-sub-role"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/port-role"
	"github.com/ZTE/Knitter/pkg/klog"
)

type BrintRole struct {
	brintsubrole.PortRole

	bridgeRole brcomsubrole.BridgeRole
	portRole   brcomsubrole.PortRole
}

func (this *BrintRole) GetVlanID(networkID string) (string, error) {
	value, err := brintsubrole.GetTenantNetworkTableSingleton().Get(networkID)
	if err != nil {
		klog.Errorf("BrintRole.GetVlanId err: %v", err)
		return "", err
	}
	return value.VlanID, nil
}

func (this *BrintRole) InsertTenantNetworkTable(networkID string, vni int, vlanID string) error {
	return brintsubrole.GetTenantNetworkTableSingleton().Insert(networkID, vni, vlanID)
}

func (this *BrintRole) GetTenantNetworkTable(networkID string) (*brintsubrole.TenantNetworkValue, error) {
	return brintsubrole.GetTenantNetworkTableSingleton().Get(networkID)
}

func (this *BrintRole) LoadTenantNetworkTable() error {
	return brintsubrole.GetTenantNetworkTableSingleton().Load()
}

func (this *BrintRole) DelTenantNetworkTable(networkID string) error {
	return brintsubrole.GetTenantNetworkTableSingleton().Delete(networkID)
}

func (this *BrintRole) NeedDelTenantNetworkTable(networkID string) bool {
	return brintsubrole.GetTenantNetworkTableSingleton().NeedDelete(networkID)
}

func (this *BrintRole) IncRefCount(networkID, podNs, podName string) error {
	return brintsubrole.GetTenantNetworkTableSingleton().IncRefCount(networkID, podNs, podName)
}

func (this *BrintRole) DecRefCount(networkID, podNs, podName string) error {
	return brintsubrole.GetTenantNetworkTableSingleton().DecRefCount(networkID, podNs, podName)
}

func (this *BrintRole) AddBridge(properties ...string) {
	this.bridgeRole.AddBridge(constvalue.OvsBrint, properties...)
}

func (this *BrintRole) AddPort(port string, ofPort uint, properties ...string) (string, error) {
	return this.portRole.AddPort(constvalue.OvsBrint, port, ofPort, properties...)
}

func (this *BrintRole) GetPortTable(portID string) (string, error) {
	return portrole.GetPortTableSingleton().Get(portID)
}

func (this *BrintRole) InsertPortTable(portID, portName string) error {
	return portrole.GetPortTableSingleton().Insert(portID, portName)
}

func (this *BrintRole) DelPortTable(portID string) error {
	return portrole.GetPortTableSingleton().Delete(portID)
}

func (this *BrintRole) GetDefaultGwNetworkID() string {
	return brintsubrole.GetDefaultGwSingleton().NetworkID

}

func (this *BrintRole) GetDefaultGwIP() string {
	return brintsubrole.GetDefaultGwSingleton().IP
}

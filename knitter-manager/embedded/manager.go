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
	"errors"
	iaas "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
)

/*************************************************************************/
type NetworkManager struct {
	net  *Networks
	sub  *Subnets
	port *Interfaces
	vx   *VxlanIDManager
}

var manager *NetworkManager

func GetEmbeddedNetwrokManager() *NetworkManager {
	if manager == nil {
		manager = &NetworkManager{}
		manager.net = GetNetManager()
		manager.sub = GetSubnetManager()
		manager.port = GetPortManager()
		manager.vx = GetVxlanManager()
	}
	return manager
}

/*************************************************************************/
func (_ *NetworkManager) CreateNetwork(name string) (*iaas.Network, error) {
	return GetNetManager().CreateNetwork(name)
}

func (_ *NetworkManager) DeleteNetwork(id string) error {
	return GetNetManager().DeleteNetwork(id)
}

func (_ *NetworkManager) GetNetworkID(networkName string) (string, error) {
	return GetNetManager().GetNetworkID(networkName)
}

func (_ *NetworkManager) GetNetwork(id string) (*iaas.Network, error) {
	net, err := GetNetManager().GetNetwork(id)
	if err != nil {
		return nil, err
	}
	network := iaas.Network{Id: net.ID, Name: net.Name}
	return &network, nil
}

func (_ *NetworkManager) GetNetworkExtenAttrs(id string) (*iaas.NetworkExtenAttrs, error) {
	value, err := GetNetManager().GetNetworkExtenAttrs(id)
	if err != nil {
		return nil, err
	}
	attrs := iaas.NetworkExtenAttrs{
		Id: value.ID, Name: value.Name,
		NetworkType:     value.NetworkType,
		PhysicalNetwork: value.PhysicalNetwork,
		SegmentationID:  value.SegmentationID,
	}
	return &attrs, nil
}

func (_ *NetworkManager) GetAttachReq() int {
	return GetNetManager().GetAttachReq()
}

func (_ *NetworkManager) SetAttachReq(req int) {
	defer GetNetManager().SetAttachReq(req)
}

func (_ *NetworkManager) CreateSubnet(id, cidr, gw string, allocationPools []subnets.AllocationPool) (*iaas.Subnet, error) {
	return GetSubnetManager().CreateSubnet(id, cidr, gw, allocationPools)
}

func (_ *NetworkManager) DeleteSubnet(id string) error {
	return GetSubnetManager().DeleteSubnet(id)
}

func (_ *NetworkManager) GetSubnetID(networkID string) (string, error) {
	return GetSubnetManager().GetSubnetID(networkID)
}

func (_ *NetworkManager) GetSubnet(id string) (*iaas.Subnet, error) {
	return GetSubnetManager().GetSubnet(id)
}

func (_ *NetworkManager) CreatePort(networkID, subnetID,
	networkPlane, ip, mac, vnicType string) (*iaas.Interface, error) {
	return GetPortManager().CreatePort(networkID, subnetID,
		networkPlane, ip, mac, vnicType)
}

func (_ *NetworkManager) CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*iaas.Interface, error) {
	return GetPortManager().CreateBulkPorts(req)
}

func (_ *NetworkManager) GetPort(id string) (*iaas.Interface, error) {
	return GetPortManager().GetPort(id)
}

func (_ *NetworkManager) DeletePort(id string) error {
	return GetPortManager().DeletePort(id)
}

func (_ *NetworkManager) ListPorts(networkID string) ([]*iaas.Interface, error) {
	return GetPortManager().ListPorts(networkID)
}

/***************************************************************
*
****************************************************************/

func (self *NetworkManager) GetTenantUUID(cfg string) (string, error) {
	return "e8b764da-5fe5-51ed-8af8-a5a5a5a5a5a5", nil
}

func (self *NetworkManager) Auth() error {
	return nil
}

func (self *NetworkManager) GetType() string {
	return "EMBEDDED"
}

func (self *NetworkManager) CreateRouter(name, extNetID string) (string, error) {
	return "", errors.New("can-not-support-CreateRouter")
}

func (self *NetworkManager) UpdateRouter(id, name, extNetID string) error {
	return errors.New("can-not-support-UpdateRouter")
}

func (self *NetworkManager) GetRouter(id string) (*iaas.Router, error) {
	return nil, errors.New("can-not-support-GetRouter")
}

func (self *NetworkManager) DeleteRouter(id string) error {
	return errors.New("can-not-support-DeleteRouter")
}

func (self *NetworkManager) AttachPortToVM(vmID,
	portID string) (*iaas.Interface, error) {
	return nil, errors.New("can-not-support-AttachPortToVM")
}

func (self *NetworkManager) DetachPortFromVM(vmID,
	portID string) error {
	return errors.New("can-not-support-DetachPortFromVM")
}

func (self *NetworkManager) AttachNetToRouter(routerID,
	subNetID string) (string, error) {
	return "", errors.New("can-not-support-AttachNetToRouter")
}

func (self *NetworkManager) DetachNetFromRouter(routerID,
	netID string) (string, error) {
	return "", errors.New("can-not-support-DetachNetFromRouter")
}

func (self *NetworkManager) CreateProviderNetwork(
	name, nwType, phyNet, sID string, vlanTransparent bool) (*iaas.Network, error) {
	return nil, errors.New("not-support-CreateProviderNetwork")
}

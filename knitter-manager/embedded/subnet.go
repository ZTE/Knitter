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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	iaas "github.com/ZTE/Knitter/pkg/iaas-accessor"
	LOG "github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"net"
	"strconv"
	"strings"
	"sync"
)

type PaasSubnet struct {
	Sub  *iaas.Subnet       `json:"subnets"`
	Pool map[string]*net.IP `json:"ip_used"`
	lock sync.Mutex
}

func (self *PaasSubnet) load(id string) (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerSubnetID(id)
	value, err := ReadData(key)
	if err != nil {
		LOG.Error("Read-embedded-server-data-from-etcd-error")
		return err
	}
	return json.Unmarshal([]byte(value), &self)
}

func (self *PaasSubnet) save() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerSubnetID(self.Sub.Id)
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

func (self *PaasSubnet) delete() (err error) {
	key := dbaccessor.GetKeyOfEmbeddedServerSubnetID(self.Sub.Id)
	err = DeleteData(key)
	if err != nil {
		LOG.Error("Save-embedded-server-VxlanIDManager-data-to-etcd-error")
		return err
	}
	return nil
}

type Subnets struct {
	lock sync.RWMutex
	list map[string]*PaasSubnet
}

var subManager *Subnets

func GetSubnetManager() *Subnets {
	if subManager == nil {
		new := Subnets{}
		new.list = make(map[string]*PaasSubnet)
		new.load()
		subManager = &new
	}
	return subManager
}

func (self *Subnets) load() error {
	self.lock.Lock()
	defer self.lock.Unlock()
	key := dbaccessor.GetKeyOfEmbeddedServerSubnets()
	nodes, err := ReadDataDir(key)
	if err != nil {
		LOG.Warning("Read Network dir[", key,
			"] from ETCD Error:", err)
		return nil
	}

	for _, node := range nodes {
		id := strings.TrimPrefix(node.Key, key+"/")
		item := PaasSubnet{}
		err = item.load(id)
		if err != nil {
			continue
		}
		self.list[id] = &item
	}
	return nil
}

func (self *Subnets) IsIPUsed(subnetid, ipAddr string) bool {
	if !self.IsExistSubnet(subnetid) || self.list[subnetid].Pool == nil {
		return false
	}

	for _, v := range self.list[subnetid].Pool {
		if v != nil && v.String() == ipAddr {
			return true
		}
	}

	return false
}

func (self *Subnets) IsExistSubnet(id string) bool {
	return (self.list[id] != nil)
}

func UnitToBytes(i uint32) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToUint(buf []byte) uint32 {
	return uint32(binary.BigEndian.Uint32(buf))
}

func IPAddrPlus(ipNet *net.IPNet, offSet uint32) net.IP {
	ipStart := BytesToUint(ipNet.IP.To4())
	ipArr := UnitToBytes(ipStart + offSet)
	ip := net.IPv4(ipArr[0], ipArr[1], ipArr[2], ipArr[3])
	return ip
}

func getSpecIPAddrFromNet(ipPool *net.IPNet, pool map[string]*net.IP,
	specIP string, ones int, minIP uint32) (string, error) {
	ipBytes := net.ParseIP(specIP)
	if !ipPool.Contains(ipBytes) {
		LOG.Error("ipaddr-not-in-CIDR:", ipPool.String())
		return "", errors.New("ipaddr-not-in-CIDR")
	}

	offset := BytesToUint(ipBytes.To4()) & (0xffffffff >> uint(ones))
	if uint32(offset) < minIP {
		LOG.Error("ipaddr-offset-invalid:", int(offset))
		return "", errors.New("ipaddr-offset-invalid")
	}
	idx := strconv.Itoa(int(offset))
	pool[idx] = &ipBytes
	LOG.Info("ipaddr---offset[", idx, "]IP[", specIP, "]")
	return specIP, nil
}

func getIPAddrFromNet(ipPool *net.IPNet,
	pool map[string]*net.IP, specIP string) (string, error) {
	const ReserveIPOffset uint32 = 2
	ones, all := ipPool.Mask.Size()
	minIP := ReserveIPOffset
	maxIP := (1 << uint(all-ones)) - ReserveIPOffset
	LOG.Info("Alloc-ipaddr-in-range[", strconv.Itoa(int(minIP)),
		"---", strconv.Itoa(int(maxIP)), "]")

	if specIP != "" {
		return getSpecIPAddrFromNet(ipPool, pool, specIP, ones, minIP)
	}

	for i := minIP; i < maxIP; i++ {
		idx := strconv.Itoa(int(i))
		ip := IPAddrPlus(ipPool, i)
		LOG.Info("ipaddr---offset[", idx, "]IP[", ip.String(), "]")
		if pool[idx] == nil && ipPool.Contains(ip) {
			pool[idx] = &ip
			LOG.Info("ipaddr---offset[", idx, "]IP[", ip.String(), "]")
			return ip.String(), nil
		}
	}
	LOG.Error("No-unsued-ipaddr-in-CIDR:", ipPool.String())
	return "", errors.New("no-unused-ipaddrress")
}

func (self *Subnets) allocIP(id, specIP string) (string, error) {
	if self.IsExistSubnet(id) == false {
		return "", errors.New("Subnet-isnot-exist:" + id)
	}

	subnet := self.list[id].Sub
	_, ipPool, _ := net.ParseCIDR(subnet.Cidr)
	if specIP != "" && (net.ParseIP(specIP) == nil || self.IsIPUsed(id, specIP)) {
		return "", errors.New("ip-is-invalid-or-in-use:" + specIP)
	}
	subNet := self.list[id]
	subNet.lock.Lock()
	defer subNet.lock.Unlock()
	if subNet.Pool == nil {
		subNet.Pool = make(map[string]*net.IP)
	}
	pool := subNet.Pool
	newIP, err := getIPAddrFromNet(ipPool, pool, specIP)
	if err != nil {
		return "", err
	}
	err = subNet.save()
	if err != nil {
		return "", err
	}
	return newIP, nil
}

func (self *Subnets) freeIP(id, ip string) error {
	if self.IsExistSubnet(id) == false {
		return errors.New("Subnet-isnot-exist:" + id)
	}
	subNet := self.list[id]
	subNet.lock.Lock()
	defer subNet.lock.Unlock()
	_, ipPool, _ := net.ParseCIDR(subNet.Sub.Cidr)
	ipAddr := net.ParseIP(ip)
	ipStart := BytesToUint(ipPool.IP.To4())
	ipDel := BytesToUint(ipAddr.To4())
	idxInt := ipDel - ipStart
	idx := strconv.Itoa(int(idxInt))
	pool := subNet.Pool
	if ipPool.Contains(ipAddr) && pool[idx] != nil {
		tmp := subNet.Pool[idx]
		subNet.Pool[idx] = nil
		err := subNet.save()
		if err != nil {
			subNet.Pool[idx] = tmp
			return err
		}
		return nil
	}

	return errors.New("ipaddress-not-alloc-by-subnet:" + id)
}

func (self *Subnets) CreateSubnet(id, cidr, gw string, allocationPool []subnets.AllocationPool) (*iaas.Subnet, error) {
	LOG.Info("EMBEDDED-CreateSubnet:[", id, "][", cidr, "]")
	if self.list == nil {
		self.list = make(map[string]*PaasSubnet)
	}
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		LOG.Error("EMBEDDED-CreateSubnet-ERROR:[cidr][", cidr, "]")
		return nil, fmt.Errorf("%v:Create-subnet-error:[CIDR error]", err)
	}
	LOG.Info("EMBEDDED-CreateSubnet-ParseCIDR:[", cidr, "]-OK")

	network, err := GetNetManager().GetNetwork(id)
	if err != nil {
		LOG.Error("EMBEDDED-CreateSubnet-ERROR:",
			"[netowork-not-exist-now][", id, "]")
		return nil, err
	}

	newSub := iaas.Subnet{}
	newSub.Name = "sub_" + network.Name
	newSub.Id = uuid.NewUUID()
	newSub.NetworkId = network.ID
	newSub.GatewayIp = gw
	newSub.Cidr = cidr
	newSub.TenantId = "paas-network-tenant-uuid"
	paasSub := PaasSubnet{Sub: &newSub}
	err = paasSub.save()
	if err != nil {
		return nil, err
	}
	self.lock.Lock()
	defer self.lock.Unlock()
	self.list[newSub.Id] = &paasSub
	LOG.Infof("EMBEDDED-create-subnet: %+v", newSub)
	return &newSub, nil
}

func (self *Subnets) DeleteSubnet(id string) error {
	LOG.Info("EMBEDDED-DeleteSubnet:[", id, "]")

	if self.IsExistSubnet(id) == false {
		LOG.Error("EMBEDDED-DeleteSubnet-ERROR:[",
			id, "]can-not-find")
		return errors.New("can-not-find-subnet-by-id:" + id)
	}

	if GetPortManager().isExistPortOnSubNet(id) {
		LOG.Error("EMBEDDED-DeleteSubnet-ERROR:[", id, "]-have-ports")
		return errors.New("Exist-port-on-subnet:" + id)
	}

	delSub := self.list[id]
	err := delSub.delete()
	if err != nil {
		return err
	}

	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.list, id)
	return nil
}

func (self *Subnets) GetSubnetID(networkID string) (string, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	LOG.Info("EMBEDDED-GetSubnetID:[", networkID, "]")
	for _, v := range self.list {
		if v.Sub.NetworkId == networkID {
			LOG.Info("EMBEDDED-GetSubnetID:net[", networkID,
				"]sub[", v.Sub.Id, "]")
			return v.Sub.Id, nil
		}
	}
	LOG.Error("EMBEDDED-GetSubnetID-ERROR:[", networkID, "]")
	return "", errors.New("can-not-find-subnet-id-by-network-id:" + networkID)
}

func (self *Subnets) GetSubnet(id string) (*iaas.Subnet, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	LOG.Info("EMBEDDED-GetSubnet:[", id, "]")
	if self.IsExistSubnet(id) {
		LOG.Infof("EMBEDDED-create-subnet: %+v", self.list[id].Sub)
		return self.list[id].Sub, nil
	}
	LOG.Error("EMBEDDED-GetSubnet-ERROR:[", id, "]")
	return nil, errors.New("can-not-find-subnet-by-id:" + id)
}

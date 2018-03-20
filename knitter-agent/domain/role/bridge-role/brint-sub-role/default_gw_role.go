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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brtun-sub-role"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/port-role"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	VethGwBrint   = "vethGWbrint"
	VethGwHost    = "vethGWhost"
	DefaultTenant = "admin"
)

type Gateway struct {
	NetworkID string `json:"network_id"`
	Network   string `json:"network"`
	Vni       int    `json:"vni"`
	VlanID    string `json:"vlan_id"`
	IP        string `json:"ip"`
	Mac       string `json:"mac"`
	Mask      string `json:"mask"`
}

var defaultGw Gateway

func GetDefaultGwSingleton() *Gateway {
	return &defaultGw
}

func GetNetworkID(userName, networkName string) (string, error) {
	ctx := cni.GetGlobalContext()
	url := ctx.Mc.GetNetworkURL(userName, networkName)
	statCode, networkInfoByte, err := ctx.Mc.Get(url)

	if err != nil {
		klog.Errorf("Get network url  error! -%v ", err)
		return "", fmt.Errorf("%v:Get-network-url-error", err)
	}
	if statCode != 200 {
		klog.Errorf("Get network id error! -%v ", err)
		return "", errors.New("get-network-id-error")
	}
	networkInfoJSON, _ := jason.NewObjectFromBytes(networkInfoByte)
	networkID, err := networkInfoJSON.GetString("network_id")
	if err != nil {
		klog.Errorf("Get network id error! -%v ", err)
		return "", err
	}
	return networkID, nil
}

func GetDefaultNetworkID(networkName string) (string, error) {
	return GetNetworkID(DefaultTenant, networkName)
}
func GetVniByNetworkID(networkID string) (int, error) {
	ctx := cni.GetGlobalContext()
	url := ctx.Mc.GetSegmentIDURLByID(DefaultTenant, networkID)
	statCode, vxlanByte, err := ctx.Mc.Get(url)
	if err != nil || statCode != 200 {
		klog.Errorf("Get vxlan info error! -%v ", err)
		return -1, errors.New("get-vxlan-info-error")
	}
	vxlanJSON, _ := jason.NewObjectFromBytes(vxlanByte)
	vniStr, err := vxlanJSON.GetString("vni")
	if err != nil {
		klog.Errorf("Get vxlan info error! -%v ", err)
		return -1, err
	}
	vni, _ := strconv.Atoi(vniStr)
	return vni, nil
}

func isIptablesRuleNotFoundError(output string) bool {
	if strings.Contains(output, "iptables: No chain/target/match by that name") {
		return true
	}
	return false
}

func delSnatRule(cidr, srcIP string) error {
	iptables, _ := exec.LookPath("iptables")
	iptablesArgsMasq := []string{"-t", "nat", "-D", "POSTROUTING", "-s", cidr, "-j", "MASQUERADE"}
	klog.Infof("delSnatRule: delete ops: %s + %s", iptables, iptablesArgsMasq)
	iptablesOutputMasq, err := exec.Command(iptables, iptablesArgsMasq...).CombinedOutput()
	if err != nil && !isIptablesRuleNotFoundError(string(iptablesOutputMasq)) {
		klog.Errorf("delSnatRule: Unable to exec iptables , err: %v, output: %s", err, string(iptablesOutputMasq))
		return err
	}
	klog.Infof("delSnatRule: delete ops: iptablesOutput is: %s", iptablesOutputMasq)

	iptablesArgs := []string{"-t", "nat", "-D", "POSTROUTING", "-s", cidr,
		"!", "-d", cidr, "-j", "SNAT", "--to-source", srcIP}
	klog.Infof("delSnatRule: delete ops: %s + %s", iptables, iptablesArgs)
	iptablesOutput, err := exec.Command(iptables, iptablesArgs...).CombinedOutput()
	if err != nil && !isIptablesRuleNotFoundError(string(iptablesOutput)) {
		klog.Errorf("delSnatRule: Unable to exec iptables , err: %v, output: %s", err, string(iptablesOutput))
		return err
	}
	klog.Infof("delSnatRule: delete ops: iptablesOutput is: %s", iptablesOutput)

	iptablesArgsOld := []string{"-t", "nat", "-D", "POSTROUTING", "-s", cidr,
		"-j", "SNAT", "--to-source", srcIP}
	klog.Infof("delSnatRule: delete ops: %s + %s", iptables, iptablesArgsOld)
	iptablesOutputOld, err := exec.Command(iptables, iptablesArgsOld...).CombinedOutput()
	if err != nil && !isIptablesRuleNotFoundError(string(iptablesOutputOld)) {
		klog.Errorf("delSnatRule: Unable to exec iptables , err: %v, output: %s", err, string(iptablesOutputOld))
		return err
	}
	klog.Infof("delSnatRule: delete ops: iptablesOutputOld is: %s", iptablesOutputOld)

	return nil
}
func SetSnatRule(cidr string, srcIP string) error {
	var err error
	for {
		err = delSnatRule(cidr, srcIP)
		if err == nil {
			klog.Info("SetSnatRule: delSnatRule SUCC", err)
			break
		}
		time.Sleep(5 * time.Second)
	}
	iptables, _ := exec.LookPath("iptables")
	iptablesArgs := []string{"-t", "nat", "-A", "POSTROUTING", "-s", cidr, "-j", "MASQUERADE"}
	klog.Infof("SetSnatRule: %s + %s", iptables, iptablesArgs)
	iptablesOutput, err := exec.Command(iptables, iptablesArgs...).CombinedOutput()
	if err != nil {
		klog.Errorf("Unable to exec iptables , err: %v, output: %s", err, string(iptablesOutput))
		return err
	}
	klog.Infof("SetSnatRule: iptablesOutput is: %s", iptablesOutput)

	/*iptables -D FORWARD -j REJECT --reject-with icmp-host-prohibited*/
	delRejectArgs := []string{"-D", "FORWARD", "-j", "REJECT",
		"--reject-with", "icmp-host-prohibited"}
	klog.Infof("Delete-reject-rule: %s + %s", iptables, delRejectArgs)
	exec.Command(iptables, delRejectArgs...).CombinedOutput()
	return nil
}

func ActivateVethPair(veth ovs.VethPair) error {
	err := bind.ActiveVethPort(veth.VethNameOfPod)
	if err != nil {
		klog.Infof("Activate VethNameOfPod: %s error! -%v", veth.VethNameOfPod, err)
		return err
	}
	bind.ActiveVethPort(veth.VethNameOfBridge)
	if err != nil {
		klog.Infof("Activate VethNameOfBridge:%s error! -%v", veth.VethNameOfBridge, err)
		return err
	}
	return nil
}

func SetPort4Gw(vethName string, port *manager.Port) error {
	link, err := netlink.LinkByName(vethName)
	if err != nil {
		klog.Errorf("%s not exist error !: -%v", vethName, err)
		return err
	}
	addr := port.MakeAddr()
	addr.Label = vethName
	klog.Info("bind-Pod-addLinkToContainer:addr info :", addr.String())
	err = netlink.AddrAdd(link, &addr)
	if err != nil {
		klog.Errorf("Set ip for gw failed !: -%v", err)
		return err
	}
	mtu, _ := strconv.Atoi(port.MTU)
	netlink.LinkSetMTU(link, mtu)
	if err != nil {
		klog.Errorf("Set mtu for gw failed !: -%v", err)
		return err
	}
	mac, _ := net.ParseMAC(port.MACAddress)
	netlink.LinkSetHardwareAddr(link, mac)
	if err != nil {
		klog.Errorf("Set mac for gw failed !: -%v", err)
		return err
	}
	return nil
}

func CreatePort4Gw(defaultNetName, vethName string) (*manager.Port, error) {
	ctx := cni.GetGlobalContext()
	req := manager.CreatePortReq{
		AgtPortReq: agtmgr.AgtPortReq{
			TenantID:    constvalue.PaaSTenantAdminDefaultUUID,
			NetworkName: defaultNetName,
			PortName:    "eth_gw",
			NodeID:      ctx.VMID,
			PodNs:       constvalue.PaaSTenantAdminDefaultUUID,
			PodName:     "",
			FixIP:       "",
			ClusterID:   cni.GetGlobalContext().ClusterUUID}}
	portByte, err := ctx.Mc.CreateNeutronPort(vethName, req, DefaultTenant)
	if err != nil {
		klog.Errorf("createPort: CreateNeutronPort for CreatePortReq[%v] failed, error! -%v", req, err)
		return nil, err
	}
	port := &manager.Port{}
	err = port.Extract(portByte, "", constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		klog.Errorf("manager port extract error! -%v", err)
		return nil, err
	}
	port.MTU = ctx.Mtu
	return port, nil
}

func StoreGwIaasPort2Etcd(mport *manager.Port) error {
	ctx := cni.GetGlobalContext()
	bytePort, _ := json.Marshal(mport)
	key := dbaccessor.GetKeyOfGwForNode(ctx.ClusterID, ctx.HostIP)
	err := ctx.DB.SaveLeaf(key, string(bytePort))
	if err != nil {
		klog.Errorf("Save gw iaas port to etcd error! -%v", err)
		return err
	}
	return nil
}

func DeleteGwIaasPort4Etcd() error {
	ctx := cni.GetGlobalContext()
	key := dbaccessor.GetKeyOfGwForNode(ctx.ClusterID, ctx.HostIP)
	err := ctx.DB.DeleteLeaf(key)
	if err != nil {
		klog.Errorf("Detele gw iaas port to etcd error! -%v", err)
		return err
	}
	return nil
}

func OvsVsctl(args ...string) error {
	ovsctl, _ := exec.LookPath("ovs-vsctl")
	ovsctlOutput, err := exec.Command(ovsctl, args...).CombinedOutput()
	if err != nil {
		klog.Errorf("ovs-vsctl exec error ! -%v output: %s", err,
			string(ovsctlOutput))
		return err
	}
	return nil
}

func AddPort2Ovs(bridge, port string, properties ...string) error {
	args := []string{"--if-exists", "del-port", port, "--", "add-port", bridge, port}
	if len(properties) > 0 {
		args = append(args, properties...)
	}
	return OvsVsctl(args...)

}

func printDefaultGw(port *manager.Port) {
	klog.Infof("----------default gateway --------------> ")
	klog.Infof("Network--> %v ", defaultGw.Network)
	klog.Infof("NetworkID--> %v ", defaultGw.NetworkID)
	klog.Infof("Vni--> %v ", defaultGw.Vni)
	klog.Infof("Vlan--> %v ", defaultGw.VlanID)
	klog.Infof("\n")
	klog.Infof("Port Id --> %v ", port.ID)
	klog.Infof("Port Name --> %v ", port.Name)
	klog.Infof("Port Network Id --> %v ", port.NetworkID)
	klog.Infof("Port Network Name --> %v ", port.NetworkName)
	klog.Infof("Port Network Type --> %v ", port.NetworkType)
	klog.Infof("Port Tenant Id --> %v ", port.TenantID)
	klog.Infof("Port Ip --> %v ", port.FixedIPs[0].Address)
	klog.Infof("Port CIDR --> %v ", port.CIDR)
	klog.Infof("Port Gateway Ip --> %v ", port.GatewayIP)
	klog.Infof("Port MAC --> %v ", port.MACAddress)
	klog.Infof("Port Is Default Gateway --> %v ", port.IsDefaultGateway)
	klog.Infof("Port MTU --> %v ", port.MTU)
	klog.Infof("----------------------------------------> ")
}

const (
	DefaultGWFakePodNs   = "DefaultGWFakePodNs"
	DefaultGWFakePodName = "DefaultGWFakePodName"
)

func CreateDefaultGw(port *manager.Port) error {
	ctx := cni.GetGlobalContext()
	defaultGw.IP = port.FixedIPs[0].Address
	defaultGw.Mac = port.MACAddress
	ipNet := manager.MakeIPNetByCidr(port.CIDR)
	defaultGw.Mask = string(ipNet.Mask)
	printDefaultGw(port)
	klog.Infof("CreateDefaultGw:Insert default Gw: %v to tenantNetworkTable", defaultGw)
	table := GetTenantNetworkTableSingleton()
	_, err := table.Get(defaultGw.NetworkID)
	if err != nil {
		klog.Infof("CreateDefaultGw:default network:%s not found, insert tenantNetworkTable", defaultGw.NetworkID)
		err = table.Insert(defaultGw.NetworkID, defaultGw.Vni, defaultGw.VlanID)
		if err != nil {
			klog.Errorf("CreateDefaultGw:table.Insert(vlan-vni) err: %v", err)
			return err
		}
		err = table.IncRefCount(port.NetworkID, DefaultGWFakePodNs, DefaultGWFakePodName)
		if err != nil {
			klog.Errorf("CreateDefaultGw:table.IncRefCount err: %v", err)
			return err
		}
	}

	klog.Info("Add Network[", port.NetworkID, "] to FlowMgr")
	err = brtunsubrole.GetFlowMgrSingleton().AddNetwork(port.NetworkID, defaultGw.Vni, defaultGw.VlanID)
	if err != nil && err.Error() != errobj.ErrNetExist.Error() {
		klog.Errorf("Store flow table error! -%v", err)
		return err
	}
	var vethPair ovs.VethPair
	_, err1 := netlink.LinkByName(VethGwHost)
	_, err2 := netlink.LinkByName(VethGwBrint)
	if err1 == nil || err2 == nil {
		vethPair = ovs.VethPair{VethNameOfPod: VethGwHost,
			VethNameOfBridge: VethGwBrint}
		err := ovs.DeleteVethPair(vethPair)
		if err != nil {
			klog.Error("Delete-default-GW-veth-pair-ERROR")
			return err
		}
	}
	klog.Info("Create veth pair.")
	vethPair, err = ovs.CreateVethPair(VethGwHost, VethGwBrint)
	if err != nil {
		klog.Errorf("Create veth pair error! -%v", err)
	}
	klog.Infof("CreateVethPair[%v] finish", vethPair)
	klog.Info("Activate veth pair.")
	err = ActivateVethPair(vethPair)
	if err != nil {
		klog.Errorf("Activate veth pair error! -%v", err)
		return err
	}
	klog.Info("Set port for gw.")
	err = SetPort4Gw(VethGwHost, port)
	if err != nil {
		klog.Errorf("Set port for gw error! -%v", err)
		return err
	}

	klog.Info("Add port to ovs with vlan id.")
	err = AddPort2Ovs(constvalue.OvsBrint, VethGwBrint, fmt.Sprintf("tag=%s", defaultGw.VlanID))
	if err != nil {
		klog.Errorf("Add port to ovs with vlan id error! -%v", err)
	}

	klog.Infof("Insert to PortMgr: vethPair[%v], mport[%v]", vethPair, port)
	err = portrole.GetPortTableSingleton().Insert(port.ID, vethPair.VethNameOfBridge)
	if err != nil {
		klog.Errorf("Save veth pair error! -%v", err)
	}
	klog.Infof("Set SNAT rule: cidr:%s, externalIp:%s", port.CIDR, ctx.ExternalIP)
	err = SetSnatRule(port.CIDR, ctx.ExternalIP)
	if err != nil {
		klog.Errorf("Set SNAT rule error! -%v", err)
		return err
	}
	klog.Infof("Create default gw success! ")
	return nil
}

func GetGwPortFromEtcd(defaultNetName string) (*manager.Port, error) {
	ctx := cni.GetGlobalContext()
	key := dbaccessor.GetKeyOfGwForNode(ctx.ClusterID, ctx.HostIP)
	portStr, err := ctx.DB.ReadLeaf(key)
	if err != nil {
		klog.Infof("Get gw iaas port from etcd failed ! -%v", err)
		return nil, err
	}
	klog.Infof("Gw iaas port is %v", portStr)
	port := &manager.Port{}
	err = json.Unmarshal([]byte(portStr), port)
	if err != nil {
		klog.Errorf("Json unmarshal error! -%v", err)
		return nil, err
	}
	networkID, _ := GetDefaultNetworkID(defaultNetName)
	if port.NetworkName != defaultNetName || port.NetworkID != networkID {
		ctx.Mc.DeleteNeutronPort(port.ID, port.TenantID)
		DeleteGwIaasPort4Etcd()
		return nil, errors.New("port in ETCD is timeout")
	}
	return port, nil
}

func GetDefaultNetName(cfg *jason.Object) string {
	netName, err := cfg.GetString("external", "net_name")
	if err == nil || netName != "" {
		return netName
	}
	return "net_api"
}

func CreateDefaultGwPort(defaultNetName string) (*manager.Port, error) {
	klog.Info("Create iaas port for default gw!")
	ctx := cni.GetGlobalContext()
	port, err := CreatePort4Gw(defaultNetName, VethGwHost)
	if err != nil {
		klog.Errorf("Create port for default gw error! -%v", err)
		return nil, err
	}
	klog.Infof("Store gw iaas port to etcd. mport[%v]", port)
	for i := 0; i < 5; i++ {
		err = StoreGwIaasPort2Etcd(port)
		if err != nil {
			klog.Errorf("Save gw port to etcd error! -%v", err)
			time.Sleep(5 * time.Second)
		} else {
			klog.Infof("Save gw port to etcd success! -%v", err)
			break
		}
	}
	if err != nil {
		klog.Errorf("Save gw port to etcd error! -%v", err)
		ctx.Mc.DeleteNeutronPort(port.ID, port.TenantID)
		if err != nil {
			klog.Errorf("Delete neutron port error! -%v", err)
		}
		return nil, err
	}
	return port, nil
}

func ConfigDefaultGwObj(defaultNetName string) error {
	var vlanID string
	networkID, err := GetDefaultNetworkID(defaultNetName)
	if err != nil {
		klog.Errorf("Get network id error! -%v", err)
		return err
	}

	klog.Info("Get-Default-network-ID:", networkID)
	vni, err := GetVniByNetworkID(networkID)
	if err != nil {
		klog.Errorf("Get vni by network id error! -%v", err)
		return err
	}

	tenantNetworkValue, err := GetTenantNetworkTableSingleton().Get(networkID)
	if err != nil {
		klog.Errorf("Get vni/vlan from ram file by network error! -%v", err)
		knitterAgentObj := knitteragtobj.GetKnitterAgtObjSingleton()
		vlanID = knitterAgentObj.VlanIDAllocatorRole.Alloc()
	} else {
		klog.Errorf("Get vni/vlan from ram file by network success! -%v", err)
		vlanID = tenantNetworkValue.VlanID
	}

	klog.Info("Get-VNI[", vni, "]VLAN-ID[", vlanID, "]-of-Default-network:", networkID)
	defaultGw.Network = defaultNetName
	defaultGw.Vni = vni
	defaultGw.VlanID = vlanID
	defaultGw.NetworkID = networkID
	return nil
}

func InitDefaultGw(defaultNetName string) error {
	port, err := GetGwPortFromEtcd(defaultNetName)
	if err != nil {
		klog.Errorf("Get gw iaas port from etcd error !: -%v", err)
		port, err = CreateDefaultGwPort(defaultNetName)
		if err != nil {
			klog.Errorf("CreateDefaultGwPort error! -%v", err)
			return err
		}
	}

	err = ConfigDefaultGwObj(defaultNetName)
	if err != nil {
		klog.Errorf("ConfigDefaultGwObj error ! -%v", err)
		return err
	}

	err = CreateDefaultGw(port)
	if err != nil {
		klog.Errorf("Create deafault gw error ! -%v", err)
		return err
	}
	return nil
}

func RetryInitDefaultGw(defaultNetName string) error {
	for {
		err := InitDefaultGw(defaultNetName)
		if err != nil {
			klog.Warningf("ERROR:Init-default-gw-error[%v]", err)
			time.Sleep(30 * time.Second)
			continue
		}
		return nil
	}
}

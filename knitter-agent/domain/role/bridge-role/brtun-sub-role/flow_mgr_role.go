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

package brtunsubrole

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brcom-sub-role"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"strconv"
	"sync"
)

type TunPort struct {
	Agent *dbaccessor.Agent `json:"agent"`
	ID    int               `json:"port_id"`
	Name  string            `json:"port_name"`
}

type TunNet struct {
	ID     string `json:"net_id"`
	Vni    int    `json:"vxlan_id"`
	VlanID string `json:"vlan_id"`
}

type FlowMgrRole struct {
	FlowTableRole       FlowTableRole
	PortIDAllocatorRole PortIDAllocatorRole
	PortRole            brcomsubrole.PortRole
	PortList            []*TunPort
	NetList             []*TunNet
}

var flowMgr *FlowMgrRole
var flowMgrLock sync.Mutex

func GetFlowMgrSingleton() *FlowMgrRole {
	flowMgrLock.Lock()
	defer flowMgrLock.Unlock()
	if flowMgr != nil {
		return flowMgr
	}

	flowMgr = &FlowMgrRole{}
	flowMgr.FlowTableRole.Init()
	return flowMgr
}

func (this *FlowMgrRole) getPort(agent *dbaccessor.Agent) *TunPort {
	for _, port := range this.PortList {
		if port.Agent.Id == agent.Id {
			return port
		}
	}
	return nil
}

func (this *FlowMgrRole) isLinked(agent *dbaccessor.Agent) bool {
	port := this.getPort(agent)
	if port == nil {
		return false
	}

	if port.ID == InvalidTunnulPortID &&
		port.Name == strconv.Itoa(InvalidTunnulPortID) {
		return false
	}

	return true
}

func (this *FlowMgrRole) AddNetwork(networkID string, vni int, vlanID string) error {
	newNet := TunNet{ID: networkID, Vni: vni, VlanID: vlanID}
	for _, net := range this.NetList {
		if (net.ID == newNet.ID) || (net.Vni == newNet.Vni) ||
			(net.VlanID == newNet.VlanID) {
			klog.Error("Add-Net[", net.ID, "] error, net exist.")
			return errobj.ErrNetExist
		}
	}

	this.NetList = append(this.NetList, &newNet)
	klog.Info("Add-Net[", newNet.ID, "][", newNet.Vni,
		"vs", newNet.VlanID, "]-OK")
	this.FlowTableRole.Update(this.NetList, this.PortList)
	return nil
}

func (this *FlowMgrRole) RemoveNetwork(netID string) error {
	var indexOfNet int = constvalue.InvalidNetID
	for index, net := range this.NetList {
		if net.ID == netID {
			indexOfNet = index
		}
	}
	if indexOfNet != constvalue.InvalidNetID {
		tmpNetList := this.NetList
		this.NetList = append(tmpNetList[:indexOfNet], tmpNetList[indexOfNet+1:]...)
	} else {
		klog.Error("Delete-Net-error:Cannot-find-net:", netID)
		return errors.New("delete-Net-error:Cannot-find")
	}
	klog.Error("Del-Net[", netID, "]-OK")
	this.FlowTableRole.Update(this.NetList, this.PortList)
	return nil
}

func (this *FlowMgrRole) createVxlan(remote, local *dbaccessor.Agent) (int, error) {
	portID := this.PortIDAllocatorRole.Alloc()
	arg0 := "type=vxlan"
	arg1 := "options:df_default=false"
	arg2 := "options:in_key=flow"
	arg3 := "options:out_key=flow"
	arg4 := "options:dst_port=6789"
	arg5 := fmt.Sprintf("options:remote_ip=%s", remote.Ip)
	arg6 := fmt.Sprintf("options:local_ip=%s", local.Ip)
	_, err := this.PortRole.AddPort(constvalue.OvsBrtun, remote.Id, uint(portID),
		arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	if err != nil {
		this.PortIDAllocatorRole.Free(portID)
		klog.Warning("Add-PORT-for-agent[", remote.Ip,
			"] error:", err.Error())
		return InvalidTunnulPortID, err
	}
	return portID, nil
}

func (this *FlowMgrRole) createLink(remote, local *dbaccessor.Agent) error {
	portID, err := this.createVxlan(remote, local)
	if err != nil {
		klog.Warning("Create-Vxlan-tunnel-with[", remote.Ip, "] error.")
		return err
	}

	port := this.getPort(remote)
	if port == nil {
		klog.Info("Craate-port-for-new-agent")
		port = &TunPort{ID: portID,
			Name: remote.Id, Agent: remote}
		this.PortList = append(this.PortList, port)
	} else {
		klog.Info("Create-port-for-down-agent")
		port.ID = portID
		port.Name = remote.Id
		port.Agent = remote
	}

	portData, _ := json.Marshal(port)
	klog.Info("Create-PORT-for-agent[", remote.Ip,
		"] PORT:", string(portData))

	this.FlowTableRole.Update(this.NetList, this.PortList)
	return nil
}

func (this *FlowMgrRole) deleteLink(agent *dbaccessor.Agent) error {
	portTun := this.getPort(agent)
	if portTun == nil {
		return errors.New("cannot-find-tunnel-port")
	}

	_, err := this.PortRole.DeletePort(agent.Id)
	if err != nil {
		klog.Warning("Del-PORT-of-agent[", agent.Ip,
			"] error:", err.Error())
		return err
	}

	this.PortIDAllocatorRole.Free(portTun.ID)
	portTun.ID = InvalidTunnulPortID
	portTun.Name = strconv.Itoa(InvalidTunnulPortID)
	portTun.Agent = agent

	portData, _ := json.Marshal(portTun)
	klog.Info("Delete-PORT-of-agent[", agent.Ip,
		"] PORT:", string(portData))

	this.FlowTableRole.Update(this.NetList, this.PortList)
	return nil
}

func (this *FlowMgrRole) Sync(topo *dbaccessor.Sync) error {
	var local *dbaccessor.Agent = topo.Client
	for _, remote := range topo.Agents {
		if remote.Id == local.Id {
			klog.Info("Skip-self[", remote.Ip, "]")
			continue
		}
		if this.isLinked(remote) == false &&
			remote.Status == dbaccessor.AgentStatusReady {
			klog.Info("Add-port-for-agent[", remote.Ip, "]")
			this.createLink(remote, local)
			continue
		}
		if this.isLinked(remote) == true &&
			remote.Status == dbaccessor.AgentStatusDown {
			klog.Info("Del-port-for-agent[", remote.Ip, "]")
			this.deleteLink(remote)
			continue
		}
		klog.Info("Link-with-agent[", remote.Ip, "] OK")
	}

	return nil
}

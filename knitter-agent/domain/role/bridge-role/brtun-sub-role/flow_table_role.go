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
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brcom-sub-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"strconv"
	"time"
)

type FlowTableRole struct {
	bridgeRole brcomsubrole.BridgeRole
	portRole   brcomsubrole.PortRole
}

func (this FlowTableRole) Init() {
	for {
		_, err := this.bridgeRole.ForceAddBridge(constvalue.OvsBrtun)
		if err == nil {
			break
		} else {
			klog.Warning("waiting-for-ovs-poweron ", "error:", err.Error())
			time.Sleep(time.Second * 5)
		}
	}

	this.bridgeRole.AddBridge(constvalue.OvsBrint)
	this.portRole.AddPort(constvalue.OvsBrint, "patch-tun-paas", constvalue.DefaultTunIntPort,
		"type=patch", "options:peer=patch-int-paas")
	this.portRole.AddPort(constvalue.OvsBrtun, "patch-int-paas", constvalue.DefaultTunIntPort,
		"type=patch", "options:peer=patch-tun-paas")
	this.delAllFlow()
	this.addDefaultFlow()
}

func (this FlowTableRole) addDefaultFlow() {
	this.addFlow("table=0,priority=2,in_port=1,actions=resubmit(,1)")
	this.addFlow("table=0,priority=1,actions=resubmit(,4)")
	this.addFlow("table=1,priority=1,dl_dst=00:00:00:00:00:00/01:00:00:00:00:00,actions=resubmit(,20)")
	this.addFlow("table=1,priority=1,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00,actions=resubmit(,21)")
	this.addFlow("table=1,priority=0,actions=drop")
	this.addFlow("table=4,priority=0,actions=drop")
	this.addFlow("table=10,priority=1,actions=learn(table=20,priority=1,hard_timeout=300,NXM_OF_VLAN_TCI[0..11],NXM_OF_ETH_DST[]=NXM_OF_ETH_SRC[],load:0->NXM_OF_VLAN_TCI[],load:NXM_NX_TUN_ID[]->NXM_NX_TUN_ID[],output:NXM_OF_IN_PORT[]),output:1")
	this.addFlow("table=20,priority=0,actions=resubmit(,21)")
	this.addFlow("table=21,priority=0,actions=drop")
}

func (this FlowTableRole) delAllFlow() {
	this.deleteFlows("")
}

// deleteFlows deletes all matching flows from the bridge. The arguments are
// passed to fmt.Sprintf().
func (this FlowTableRole) deleteFlows(flow string, args ...interface{}) (string, error) {
	if len(args) > 0 {
		flow = fmt.Sprintf(flow, args...)
	}
	return this.bridgeRole.OfctlExec("del-flows", constvalue.OvsBrtun, flow)
}

// addFlow adds a flow to the bridge. The arguments are passed to fmt.Sprintf().
func (this FlowTableRole) addFlow(flow string, args ...interface{}) (string, error) {
	if len(args) > 0 {
		flow = fmt.Sprintf(flow, args...)
	}
	return this.bridgeRole.OfctlExec("add-flow", constvalue.OvsBrtun, flow)
}

func (this FlowTableRole) Update(NetList []*TunNet, PortList []*TunPort) {
	this.deleteFlows("table=4")
	this.deleteFlows("table=21")
	this.addDefaultFlow()

	outputStr := this.getOutput(PortList)
	for _, net := range NetList {
		flow1 := fmt.Sprintf("table=4, priority=1, tun_id=%d, actions=mod_vlan_vid:%s,resubmit(,10)",
			net.Vni, net.VlanID)
		this.addFlow(flow1)
		flow2 := fmt.Sprintf("table=21, priority=1, dl_vlan=%s, actions=strip_vlan,set_tunnel:%d%s",
			net.VlanID, net.Vni, outputStr)
		this.addFlow(flow2)
	}
}

func (self FlowTableRole) getOutput(PortList []*TunPort) string {
	var outputList string
	for _, port := range PortList {
		if port.ID == InvalidTunnulPortID &&
			port.Name == strconv.Itoa(InvalidTunnulPortID) {
			continue
		}
		outputList += fmt.Sprintf(",output:%d", port.ID)
	}
	klog.Info("Update-FLOW-TABLE-with-ports[", outputList, "]")
	return outputList
}

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
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brtun-sub-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
)

type BrtunRole struct {
	BridgeRole     brcomsubrole.BridgeRole
	PortRole       brcomsubrole.PortRole
	SyncClientRole brtunsubrole.SyncClientRole
}

func (this *BrtunRole) SyncSwitch(cfg *jason.Object) bool {
	runMode, err0 := cfg.GetString("run_mode", "type")
	syncSwitch, err1 := cfg.GetBoolean("run_mode", "sync")
	if err0 != nil || err1 != nil {
		klog.Warning("Sync-shouldbe-not-run-in-config")
		return false
	}
	if syncSwitch == false || runMode != "overlay" {
		klog.Warning("Sync-shouldbe-not-run-in-config")
		return false
	}

	url, err2 := cfg.GetString("manager", "url")
	if err2 != nil {
		klog.Error("Can-not-get-manager-url")
		return false
	}
	ipaddr, err3 := cfg.GetString("internal", "ip")
	if err3 != nil {
		klog.Error("Can-not-get-internal-ip")
		return false
	}
	brtunsubrole.SetManager(url, ipaddr)

	klog.Info("SYNC: Switch is on, Start sync with manager")
	err4 := this.SyncClientRole.Init()
	if err4 != nil {
		klog.Error("Can-not-InitSync")
		return false
	}
	return true
}

func (this *BrtunRole) StartSync() {
	for {
		this.SyncClientRole.HeartBeat()
	}
}

func (this *BrtunRole) AddPort(port string, ofPort uint, properties ...string) (string, error) {
	return this.PortRole.AddPort(constvalue.OvsBrtun, port, ofPort, properties...)
}

func (this *BrtunRole) AddNetwork(networkID string, vni int, vlanID string) error {
	return brtunsubrole.GetFlowMgrSingleton().AddNetwork(networkID, vni, vlanID)
}

func (this *BrtunRole) RemoveNetwork(networkID string) error {
	return brtunsubrole.GetFlowMgrSingleton().RemoveNetwork(networkID)
}

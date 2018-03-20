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

package brcomsubrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
)

type BridgeRole struct {
}

func (this *BridgeRole) AddBridge(bridge string, properties ...string) {
	args := []string{"--may-exist", "add-br", bridge}
	if len(properties) > 0 {
		args = append(args, "--", "set", "Bridge", bridge)
		args = append(args, properties...)
	}
	this.VsctlExec(args...)
}

func (this *BridgeRole) ForceAddBridge(bridge string, properties ...string) (string, error) {
	args := []string{"--if-exists", "del-br", bridge, "--", "add-br",
		bridge}
	if len(properties) > 0 {
		args = append(args, "--", "set", "Bridge", bridge)
		args = append(args, properties...)
	}
	return this.VsctlExec(args...)
}

func (this *BridgeRole) VsctlExec(args ...string) (string, error) {

	return osencap.Exec(constvalue.OvsVsctl, args...)
}

func (this *BridgeRole) OfctlExec(args ...string) (string, error) {
	args = append([]string{"-O", "OpenFlow10"}, args...)
	return osencap.Exec(constvalue.OvsOfctl, args...)
}

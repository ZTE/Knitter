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
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
)

type PortRole struct {
	BridgeRole BridgeRole
}

func (this *PortRole) AddPort(bridge string, port string, ofPort uint, properties ...string) (string, error) {
	args := []string{"--if-exists", "del-port", port, "--", "add-port", bridge, port, "--",
		"set", "Interface", port, fmt.Sprintf("ofport_request=%d", ofPort)}
	if len(properties) > 0 {
		args = append(args, properties...)
	}
	return osencap.Exec(constvalue.OvsVsctl, args...)
}

// DeletePort removes an interface from the bridge. (It is an error if the
// interface is not currently a bridge port.)
func (this *PortRole) DeletePort(port string) (string, error) {
	return this.BridgeRole.VsctlExec("del-port", port)
}

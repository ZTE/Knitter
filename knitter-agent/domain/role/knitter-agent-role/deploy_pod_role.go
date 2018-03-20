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

package knitteragentrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/openshift"
	"github.com/ZTE/Knitter/pkg/klog"
)

type DeployPodRole struct {
}

func (this *DeployPodRole) ConnectToBrdep(containerID string) error {
	err := openshift.ConnectToDocker0(containerID)
	if err != nil {
		klog.Infof("cmdAdd: Connect deployer pod to Docker0 failed!")
	} else {
		klog.Infof("cmdAdd: Connect deployer pod to Docker0 success!")
	}
	return nil
}

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

package models

import (
//"github.com/ZTE/Knitter/pkg/klog"
)
import (
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
)

type HealthObj struct {
	Level int64  `json:"health_level"`
	State string `json:"state"`
}

func (self *HealthObj) GetHealthLevel() int64 {
	//klog.Info("Health Level return 0!")
	return 0
}

func (self *HealthObj) GetHealthState() string {
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) != nil && common.GetDataBase() != nil {
		return "good"
	}
	return "bad"
}

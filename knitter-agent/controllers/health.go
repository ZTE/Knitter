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

package controllers

import (
	"github.com/ZTE/Knitter/knitter-agent/scheduler"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"runtime/debug"
)

// Operations about network
type HealthController struct {
	beego.Controller
}

// @Title Get
// @Description get health level
// @Success 200 {object} models.HealthObj
// @router / [get]
func (self *HealthController) Get() {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("HealthController: Get enter recover, error: %v", err)
			klog.Error("Stack: ", string(debug.Stack()))
			klog.Errorf("HealthController: Get exit recover")
			klog.Flush()
		}
	}()

	klog.Info("Request health check!")
	var obj = scheduler.HealthObj{}
	level := obj.GetHealthLevel()
	obj.Level = level
	self.Data["json"] = obj
	self.ServeJSON()

	return
}

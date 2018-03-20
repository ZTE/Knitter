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
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"io/ioutil"
)

// Operations about Configration
type CfgController struct {
	beego.Controller
}

func (self *CfgController) CfgOpenStack(body []byte) {
	var encapConf models.EncapOpenStack
	json.Unmarshal(body, &encapConf)
	cfg, err := models.HandleOpenStackConfg(encapConf.Config)
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}

	self.Data["json"] = models.EncapOpenStack{Config: cfg}
	self.ServeJSON()
}

func (self *CfgController) CfgRegularCheck(timeInterval string) {
	err := models.HandleRegularCheckConfg(timeInterval)
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}

	self.Data["json"] = map[string]string{"interval_time": timeInterval}
	self.ServeJSON()
}

// @Title configration openstack
// @Description configration openstack
// @Success 200 {string} success
// @Failure 403 invalid request body
// @Failure 406 config openstack error
// @router / [post]
func (self *CfgController) Post() {
	defer RecoverRsp500(&self.Controller)
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info("Req Body:", string(body))
	cfg, err := jason.NewObjectFromBytes(body)
	if err != nil {
		Err400(&self.Controller, err)
		return
	}

	openstack, err := cfg.GetObject("openstack")
	if err == nil && openstack != nil {
		klog.Info("openstack:", *openstack)
		self.CfgOpenStack(body)
		return
	}

	time, err := cfg.GetString("regular_check")
	if err == nil && time != "" {
		klog.Info("Regular Check Interval Time is --->", time)
		self.CfgRegularCheck(time)
		return
	}

	ErrorRequstRsp400(&self.Controller, string(body))
	return
}

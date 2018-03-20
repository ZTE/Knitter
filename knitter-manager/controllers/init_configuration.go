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
	"context"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"time"
)

type InitCfgController struct {
	beego.Controller
}

const (
	CtxTimeout = 4 * time.Minute // etcd timeout
)

var mutexInitConf int = 0

// @Title configration openstack
// @Description configration openstack
// @Success 200 {string} success
// @Failure 403 invalid request body
// @Failure 406 config openstack error
// @router / [post]
func (self *InitCfgController) Post() {
	var err error

	if mutexInitConf == 1 {
		err := models.BuildErrWithCode(http.StatusLocked, errobj.ErrDoing)
		HandleErr(&self.Controller, err)
		return
	}
	mutexInitConf = 1
	ctx, cancel := context.WithTimeout(context.Background(), CtxTimeout)
	defer cancel()
	defer RecoverRsp500(&self.Controller)
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Infof("Req Body: %v", string(body))
	cfg, errNewObjectFromBytes := jason.NewObjectFromBytes(body)
	if errNewObjectFromBytes != nil {
		err = models.BuildErrWithCode(http.StatusUnsupportedMediaType, errobj.ErrJSON)
		HandleErr(&self.Controller, err)
		mutexInitConf = 0
		return
	}

	err = models.CfgInit(cfg, ctx)
	if err != nil {
		HandleErr(&self.Controller, err)
		mutexInitConf = 0
		return
	}

	self.Data["json"] = "Init configuration success"
	self.ServeJSON()
	mutexInitConf = 0
	return
}

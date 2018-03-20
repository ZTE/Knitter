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
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
)

// Operations about Master
type SyncController struct {
	beego.Controller
}

// @Title get VLan ID of network for sr_iov in ITRAN2.0 site side
// @Description create port and return it to knitter
// @Param	body		body 	models.Master	true		"The master get VLan ID"
// @Success 200 {object} models.GetVLanInfoResp
// @Failure 403 body is empty
// @router /:internal_ip [get]
func (o *SyncController) Get() {
	defer RecoverRsp500(&o.Controller)
	sourceIP := o.Ctx.Input.IP()
	if sourceIP == "127.0.0.1" {
		err := errors.New("get-sync-client-ip-error")
		klog.Error("Sync-error:", err.Error())
		HandleErr406(&o.Controller, err)
		return
	}

	syncMgt := models.GetSyncMgt()
	requestIP := o.GetString(":internal_ip")
	syncRsp := syncMgt.Sync(requestIP)
	klog.Tracef("Sync-OK-from[%s]-Data: %v", syncRsp)

	o.Data["json"] = syncRsp
	o.ServeJSON()
}

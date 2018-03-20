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
	"io/ioutil"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/scheduler"
	"github.com/ZTE/Knitter/pkg/klog"
	"runtime/debug"
	"time"
)

// Operations about object
type PodController struct {
	beego.Controller
	operation string
	reqBody   []byte
}

func Init(o *jason.Object) error {
	return scheduler.Init(o)
}

func DoesIllegalOperation(op string) bool {
	if strings.ToUpper(op) == "ATTACH" {
		return false
	}
	if strings.ToUpper(op) == "DETACH" {
		return false
	}
	return true
}

func (self *PodController) getOperation() error {
	operation := self.GetString("operation")
	if DoesIllegalOperation(operation) {
		err := errors.New("operation is Illegal")
		klog.Error("getOperation ERROR:", err)
		self.Data["json"] = map[string]string{"ERROR": err.Error()}
		self.Redirect(self.Ctx.Request.URL.RequestURI(), 403)
		self.ServeJSON()
		return err
	}
	self.operation = operation
	return nil
}

func (self *PodController) getReqBody() error {
	reqBody, err := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	if err != nil {
		klog.Error("getReqBody ERROR:", err)
		self.Data["json"] = map[string]string{"ERROR": err.Error()}
		self.Redirect(self.Ctx.Request.URL.RequestURI(), 406)
		self.ServeJSON()
		return err
	}
	self.reqBody = reqBody
	return nil
}

func (self *PodController) attach() error {
	err := scheduler.Attach(self.reqBody)
	if err != nil {
		klog.Error("attach ERROR:", err)
		self.Data["json"] = map[string]string{"ERROR": err.Error(), "STATUS": "409"}
		self.Redirect(self.Ctx.Request.URL.RequestURI(), 409)
		time.Sleep(60 * time.Second)
		return err
	}
	self.Data["json"] = map[string]string{"Success": "Attach ports to POD", "STATUS": "200"}
	return nil
}

func (self *PodController) detach() error {
	err := scheduler.Detach(self.reqBody)
	if err != nil {
		if strings.Contains(err.Error(), constvalue.SKIP) {
			self.Data["json"] = map[string]string{"Success": constvalue.SKIP,
				"STATUS": "200"}
			return nil
		}
		klog.Error("detach ERROR:", err)
		self.Data["json"] = map[string]string{"ERROR": err.Error(), "STATUS": "409"}
		self.Redirect(self.Ctx.Request.URL.RequestURI(), 409)
		return err
	}
	self.Data["json"] = map[string]string{"Success": "Detach ports from POD", "STATUS": "200"}
	return nil
}

// @Title create
// @Description attach ports to pod or detach ports from pod
// @Param	body		body 	skel.CmdArgs	true		"cni args for pod"
// @Success 200 {string}
// @Failure 403 invalid request operation
// @Failure 406 invalid request json body
// @Failure 409 operation return error
// @router / [post]
func (self *PodController) Post() {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("PodController: Post enter recover, error: ", err)
			klog.Error("Stack: ", string(debug.Stack()))
			klog.Errorf("PodController: Post exit recover")
			klog.Flush()
		}
	}()

	if err := self.getOperation(); err != nil {
		return
	}
	klog.Info("Request operation: ", self.operation)

	if err := self.getReqBody(); err != nil {
		return
	}
	klog.Info(string(self.reqBody))

	if strings.ToUpper(self.operation) == "ATTACH" {
		self.attach()
	}
	if strings.ToUpper(self.operation) == "DETACH" {
		self.detach()
	}

	self.ServeJSON()
	return
}

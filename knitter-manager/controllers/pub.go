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
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

func IsUUID(str string) bool {
	const UUID string = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	rxUUID := regexp.MustCompile(UUID)
	return rxUUID.MatchString(str)
}

func RecoverRsp500(o *beego.Controller) {
	klog.Flush()
	if err := recover(); err != nil {
		klog.Info("RecoverRsp500:", err)
		klog.Infof("stack[%v]", debug.Stack())
		o.Data["json"] = map[string]string{"ERROR": "Panic",
			"Stack": string(debug.Stack())}
		o.Redirect(o.Ctx.Request.URL.RequestURI(), 500)
		o.ServeJSON()
	}
}

func RecoverRsp401(o *beego.Controller) {
	err := errors.New("never register openstack user count")
	klog.Info("RecoverRsp401:", err)
	o.Data["json"] = map[string]string{"ERROR": "Bad IaaS account",
		"message": err.Error()}
	o.Redirect(o.Ctx.Request.URL.RequestURI(), 401)
	o.ServeJSON()
}

func ErrorRequstRsp400(o *beego.Controller, body string) {
	err := errors.New("Request is error " + body)
	Err400(o, err)
}

func HandleErr406(o *beego.Controller, err error) {
	klog.Info("HandleErr406:", err)
	o.Data["json"] = map[string]string{"ERROR": "Proccess error",
		"message": err.Error()}
	o.Redirect(o.Ctx.Request.URL.RequestURI(), 406)
	o.ServeJSON()
}

func UnmarshalErr403(o *beego.Controller, err error) {
	klog.Info("UnmarshalErr403:", err)
	o.Data["json"] = map[string]string{"ERROR": "Bad Request Json body",
		"message": err.Error()}
	o.Redirect(o.Ctx.Request.URL.RequestURI(), 403)
	o.ServeJSON()
}

func HandleErr(o *beego.Controller, err error) {
	klog.Info("HandleErr:", err)

	parts := strings.Split(err.Error(), "::")
	var i int
	var msg string

	if len(parts) < 2 {
		i = http.StatusInternalServerError
		msg = http.StatusText(i)
	} else {
		i, _ = strconv.Atoi(parts[0])
		if i == 0 {
			i = http.StatusInternalServerError
		}

		msg = http.StatusText(i)
		if msg == "" {
			i = http.StatusInternalServerError
			msg = http.StatusText(i)
		}
	}

	o.Data["json"] = map[string]string{"ERROR": msg,
		"message": parts[len(parts)-1]}
	o.Redirect(o.Ctx.Request.URL.RequestURI(), i)
	o.ServeJSON()
}

func Err400(o *beego.Controller, err error) {
	HandleErr(o, models.BuildErrWithCode(http.StatusBadRequest, err))
}

func Err500(o *beego.Controller, err error) {
	HandleErr(o, models.BuildErrWithCode(http.StatusInternalServerError, err))
}

func NotfoundErr404(o *beego.Controller, err error) {
	HandleErr(o, models.BuildErrWithCode(http.StatusNotFound, err))
}

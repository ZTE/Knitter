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
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"strconv"
)

type LogController struct {
	beego.Controller
}

// @Title modify log level
// @Description modify log level for klog
// @Failure 403 params is invalid
// @router /:log_level [put]
func (l *LogController) Put() {
	logLevel := l.GetString(":log_level")
	level, err := strconv.Atoi(logLevel)
	if err != nil {
		klog.Errorf("input log level is: %s, can not transform to integer", logLevel)
		Err400(&l.Controller, err)
		return
	}

	levelNum := klog.Level(level)

	if levelNum < klog.TraceLevel || levelNum >= klog.NumLevel {
		klog.Errorf("input new log level is: %s, invalid level", logLevel)
		Err400(&l.Controller, errors.New("invalid log level number, should between 0 to 5"))
		return
	}

	klog.SetLogLevel(levelNum)
	klog.Errorf("error level")
	klog.Warningf("warning level")
	klog.Infof("info level")
	klog.Debugf("debug level")
	klog.Tracef("trace level")
	l.Data["json"] = fmt.Sprintf("modify log level to %d success", levelNum)
	l.ServeJSON()
}

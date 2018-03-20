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

package version

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
)

func init() {
	verFlag = flag.Bool("version", false, "print version infomation")
}

var verFlag *bool

/* judge whether command line params has --version option, return bool type */
func HasVerFlag() bool {
	return *verFlag == true
}

var moduleName string
var verType string
var versionInfo string
var buildTime string
var gitHash string

// print version info
func PrintVersion() {
	printStr := fmt.Sprintf("**********Version**********\n%s\nvertype: %s\nversion: %s\ngit-hash: %s\nbuilt at: %s\n",
		moduleName, verType, versionInfo, gitHash, buildTime)
	fmt.Println(printStr)
}

const (
	DefaultValue          = ""
	DefaultConfigFilePath = "."
	ConfigFileName        = "knitter.json"
)

func getJasonObj(bytes []byte, module string) (*jason.Object, error) {
	obj, err := jason.NewObjectFromBytes(bytes)
	if err != nil {
		klog.Error("getJsonObj: jason.NewObjectFromBytes(", string(bytes),
			") failed, error: ", err)
		return nil, fmt.Errorf("%v:jason.NewObjectFromBytes error", err)
	}
	objCfg, err := obj.GetObject("conf", module)
	return objCfg, err
}

var GetConfObject = func(confPath string, moduleName string) (*jason.Object, error) {
	if confPath == "" {
		confPath = DefaultConfigFilePath + "/" + ConfigFileName
	}
	confData, err := ioutil.ReadFile(confPath)
	if err != nil {
		klog.Error("readFileContent: ioutil.ReadFile(", confData,
			") failed, error: ", err)
		return nil, fmt.Errorf("%v:ioutil.ReadFile error", err)
	}

	return getJasonObj(confData, moduleName)
}

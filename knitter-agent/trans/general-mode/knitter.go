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

package modelscom

import (
	"github.com/ZTE/Knitter/knitter-agent/domain"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
)

func InitContext4Agent(cfg *jason.Object) error {
	defer klog.Flush()
	klog.Infof("InitEnv4Agent: Init configration for General MODE!")
	err := cni.InitConfigration4Agent(cfg)
	if err != nil {
		klog.Errorf("InitConfigration4Agent return error:%v", err)
		return err
	}

	infra.Init()
	if err := domain.Init(cfg); err != nil {
		klog.Errorf("InitEnv4Agent: domain.Init() error: %v", err)
		return err
	}
	return nil
}

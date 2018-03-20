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

package networkserver

import (
	"github.com/ZTE/Knitter/knitter-manager/public"
	LOG "github.com/ZTE/Knitter/pkg/klog"
	"github.com/coreos/etcd/client"
)

var SaveData = func(k, v string) error {
	LOG.Tracef("ETCD-W:[%v][%v]", k, v)
	return common.GetDataBase().SaveLeaf(k, string(v))
}

var ReadDataDir = func(k string) ([]*client.Node, error) {
	node, err := common.GetDataBase().ReadDir(k)
	LOG.Tracef("ETCD-RD:[%v][%v]", k, node)
	return node, err
}

var ReadData = func(k string) (string, error) {
	v, err := common.GetDataBase().ReadLeaf(k)
	LOG.Tracef("ETCD-R:[%v][%v]", k, v)
	return v, err
}

var DeleteData = func(k string) error {
	LOG.Tracef("ETCD-D:[%v]", k)
	return common.GetDataBase().DeleteLeaf(k)
}

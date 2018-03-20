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

package accessor

import (
	"errors"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type CloudTInfo struct {
	PortInfo       string
	VethInfo       string
	OvsbrInfo      string
	MechDriverType string
}

var Set4CloudT = func(db dbaccessor.DbAccessor, url, portid string, t CloudTInfo) error {

	key4cloudt := []string{url + "/" + portid + "/self",
		url + "/" + portid + "/veths",
		url + "/" + portid + "/ovsbrs",
		url + "/" + portid + "/driver"}

	var value4cloudt [4]string
	value4cloudt[0] = t.PortInfo
	value4cloudt[1] = t.VethInfo
	value4cloudt[2] = t.OvsbrInfo
	value4cloudt[3] = t.MechDriverType

	if db == nil {
		klog.Errorf("db is null!!")
		return errors.New("db is null")
	}

	for index := range key4cloudt {
		err := db.SaveLeaf(key4cloudt[index], value4cloudt[index])
		if err != nil {
			klog.Errorf("Set4CloudT key -%v,error!-%v", key4cloudt[index], err)
			return err
		}
	}
	return nil
}

var Get4CloudT = func(db dbaccessor.DbAccessor, url, portid string) (*CloudTInfo, error) {

	key4cloudt := []string{url + "/" + portid + "/self",
		url + "/" + portid + "/veths",
		url + "/" + portid + "/ovsbrs",
		url + "/" + portid + "/driver"}
	var value4cloudt [4]string

	if db == nil {
		klog.Errorf("db is null!!")
		return nil, errors.New("db is null")
	}
	for index := range key4cloudt {
		str, err := db.ReadLeaf(key4cloudt[index])
		if err != nil {
			klog.Errorf("Get4CloudT key -%v,error!-%v", key4cloudt[index], err)
			return nil, err
		}
		value4cloudt[index] = str
	}
	var cloudt CloudTInfo
	cloudt.PortInfo = value4cloudt[0]
	cloudt.VethInfo = value4cloudt[1]
	cloudt.OvsbrInfo = value4cloudt[2]
	cloudt.MechDriverType = value4cloudt[3]

	return &cloudt, nil
}

var Del4CloudT = func(db dbaccessor.DbAccessor, url, portid string) error {

	key4cloudt := url + "/" + portid

	if db == nil {
		klog.Errorf("db is null!!")
		return errors.New("db is null")
	}
	err := db.DeleteDir(key4cloudt)
	if err != nil {
		klog.Errorf("Del4CloudT key -%v,error!-%v", key4cloudt, err)
		return err
	}
	return nil
}

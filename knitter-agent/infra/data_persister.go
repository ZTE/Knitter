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

package infra

import (
	"encoding/json"
	"github.com/ZTE/Knitter/pkg/klog"
	"io/ioutil"
	"os"
)

const DirPrefix string = "/dev/shm/"

type DataPersister struct {
	DirName  string
	FileName string
}

//To serialize data from struct
func (this *DataPersister) serialize(i interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		klog.Errorf("Serialize:json.Marshal error! -%v", err)
		return nil, ErrJSONMarshalFailed
	}
	klog.Infof("serialize:map is : %v", i)
	return data, nil
}

//To deserialize data to struct
func (this *DataPersister) deserialize(data []byte, i interface{}) error {
	err := json.Unmarshal(data, i)
	if err != nil {
		klog.Errorf("DeSerialize: json.Unmarshal error! -%v", err)
		return ErrJSONUnmarshalFailed
	}
	return nil
}

//To store data to file
func (this *DataPersister) SaveToMemFile(i interface{}) error {
	data, err := this.serialize(i)
	if err != nil {
		klog.Errorf("Store : Serialize error!  %v", err)
		return err
	}
	rdPath, err := this.getRamdiskPath()
	if err != nil {
		klog.Errorf("Store:GetRamdiskPath error! -%v", err)
		return err
	}
	filePath := rdPath + "/" + this.FileName
	err = ioutil.WriteFile(filePath, data, 0755)
	if err != nil {
		klog.Errorf("Store:Failed to write data %s to file %q: err: %v", data, filePath, err)
		return ErrWriteFileFailed
	}
	return nil
}

//To restore data from file
func (this *DataPersister) LoadFromMemFile(i interface{}) error {
	rdPath, err := this.getRamdiskPath()
	if err != nil {
		klog.Errorf("ReStore:GetRamdiskPath error! -%v", err)
		return err
	}
	filePath := rdPath + "/" + this.FileName
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		klog.Errorf("ReStore:ioutil.ReadFile error! -%v", err)
		return ErrReadFileFailed
	}
	err = this.deserialize(data, i)
	if err != nil {
		klog.Errorf("ReStore:DeSerialize error! -%v", err)
		return err
	}
	return nil
}

func (this *DataPersister) IsExist() bool {
	rdPath := DirPrefix + this.DirName + "/" + this.FileName
	_, err := os.Stat(rdPath)
	return err == nil || os.IsExist(err)
}

func (this *DataPersister) GetFilePath() string {
	return DirPrefix + this.DirName + "/" + this.FileName
}

//To get ramdisk path  /dev/shm/{path}
func (this *DataPersister) getRamdiskPath() (string, error) {
	rdPath := DirPrefix + this.DirName
	_, err := os.Stat(rdPath)
	if err != nil {
		err = os.Mkdir(rdPath, 0644)
		if err != nil {
			klog.Errorf("getRamdiskPath:Unable to exec mkdir %s, err: %v", rdPath, err)
			return "", ErrGetRAMDiskFailed
		}
	}
	return rdPath, nil
}

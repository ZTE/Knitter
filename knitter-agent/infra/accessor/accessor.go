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
	"context"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
	"time"
)

type Accessor interface {
	Init()
	Set(key string, value string) error
	Get(url string) (*jason.Object, error)
	DeleteDir(url string) error
	DeleteLeaf(url string) error
}

type Etcd struct {
	URL       string
	ServerAPI client.KeysAPI
}

func (e *Etcd) GetEtcdServerURLFromConfigure(serverconf []byte) (string, error) {
	klog.Info("GetEtcdServerURLFromConfigure: serverconf", serverconf)
	serverconfjson, jsonerr := jason.NewObjectFromBytes([]byte(serverconf))
	if jsonerr != nil {
		klog.Errorf("GetEtcdServerUrlFromConfigure:jason.NewObjectFromBytes serverconf error! -%v", jsonerr)
		return "", fmt.Errorf("%v:GetEtcdServerUrlFromConfigure:jason.NewObjectFromBytes serverconf error", jsonerr)
	}
	klog.Info("NewObjectFromBytes etcd url", serverconfjson)
	etcdurl, _ := serverconfjson.GetString("etcd", "url")
	klog.Info("serverconfjson.GetString etcd url", etcdurl)
	if etcdurl == "" {
		klog.Errorf("getEtcdServerUrlFromConfigureL:etcd url is null")
		return "", errors.New("etcd url is null")
	}
	return etcdurl, nil
}

func (e *Etcd) CreateAPI(cfg client.Config) (client.KeysAPI, error) {
	c, err := client.New(cfg)
	if err != nil {
		klog.Errorf("createClient:create etcdclient error!-%v", err)
		return nil, fmt.Errorf("%v:createClient:create etcdclient error", err)
	}
	api := client.NewKeysAPI(c)
	if api != nil {
		return api, nil
	}
	return nil, errors.New("createApi: NewKeysAPI error")
}

func (e *Etcd) Init(cniargs string, serverinfo []byte) error {
	//etcd server url
	etcdurl, err := e.GetEtcdServerURLFromConfigure(serverinfo)
	if err != nil {
		klog.Errorf("init:getEtcdServerUrlFromConfigure  error!-%v", err)
		return fmt.Errorf("%v:init:getEtcdServerUrlFromConfigure  error", err)
	}
	e.URL = etcdurl
	//etcd client
	cfg := client.Config{
		Endpoints: []string{e.URL},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	//etcd api
	api, _ := e.CreateAPI(cfg)
	e.ServerAPI = api
	return nil
}

func (e *Etcd) Set(key string, value string) error {
	_, err := e.ServerAPI.Set(context.Background(), key, value, nil)
	if err != nil {
		klog.Errorf("set error!-%v", err)
		return err
	}
	return nil
}

func (e *Etcd) Get(key string) (*jason.Object, error) {
	obj, err := e.ServerAPI.Get(context.Background(), key, nil)
	if err != nil {
		klog.Errorf("get error!-%v", err)
		return nil, err
	}
	str := obj.Node.Value
	json, _ := jason.NewObjectFromBytes([]byte(str))
	return json, nil
}

func (e *Etcd) GetDir(key string) (*client.Response, error) {
	obj, err := e.ServerAPI.Get(context.Background(), key, nil)
	if err != nil {
		klog.Errorf("get error!-%v", err)
		return nil, err
	}
	return obj, nil
}

func (e *Etcd) GetPodID(podinfo *jason.Object) (string, error) {
	podid, err := podinfo.GetString("metadata", "uid")
	if err != nil {
		klog.Errorf("storePodAndPort2Etcd:get podid error!-%v", err)
		return "", fmt.Errorf("%v:storePodAndPort2Etcd:get podid error", err)
	}
	return podid, nil
}

func (e *Etcd) GetPodName(podinfo *jason.Object) (string, error) {
	podname, err := podinfo.GetString("metadata", "name")
	if err != nil {
		klog.Errorf("storePodAndPort2Etcd:get podname error!-%v", err)
		return "", fmt.Errorf("%v:storePodAndPort2Etcd:get podname error", err)
	}
	return podname, nil
}

func (e *Etcd) GetPodNS(podinfo *jason.Object) (string, error) {
	podns, err := podinfo.GetString("metadata", "namespace")
	if err != nil {
		klog.Errorf("storePodAndPort2Etcd:get podns error!-%v", err)
		return "", fmt.Errorf("%v:storePodAndPort2Etcd:get podns error", err)
	}
	return podns, nil
}

func (e *Etcd) DeleteDir(url string) error {
	ops := &client.DeleteOptions{
		Recursive: true,
		Dir:       true}
	_, err := e.ServerAPI.Delete(context.Background(), url, ops)
	if err != nil {
		klog.Errorf("deleteDirFromEtcd:Delete error!-%v", err)
		return fmt.Errorf("%v:deletePodFromEtcd:Delete error", err)
	}
	return nil
}

func (e *Etcd) DeleteLeaf(url string) error {
	_, err := e.ServerAPI.Delete(context.Background(), url, nil)
	if err != nil {
		klog.Errorf("deleteLeafFromEtcd:Delete error!-%v", err)
		return fmt.Errorf("%v:deleteLeafFromEtcd:Delete error", err)
	}
	return nil
}

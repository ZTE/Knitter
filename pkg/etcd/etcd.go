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

package etcd

import (
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"strings"
	"time"
)

const (
	MAXTIME               = 5
	DefaultEtcdAPIVersion = 2

	EtcdInitRetryIntervalInSec = 10
)

type Etcd struct {
	client client.KeysAPI
	config client.Config
}

func IsNotFindError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, "100: Key not found") {
		klog.Trace(err)
		return true
	}
	return false
}

func NewEtcdWithRetry(ver int, urls string) dbaccessor.DbAccessor {
	var etcdObj dbaccessor.DbAccessor
	for {
		etcdObj = NewEtcd(ver, urls)
		if etcdObj != nil {
			klog.Infof("NewEtcdWithRetry: NewEtcd(%d, %s) SUCC", ver, urls)
			return etcdObj
		}

		klog.Warningf("NewEtcdWithRetry: NewEtcd(%d, %s) FAIL, wait retry", ver, urls)
		time.Sleep(EtcdInitRetryIntervalInSec * time.Second)
	}
}

func NewEtcd(ver int, urls string) dbaccessor.DbAccessor {
	klog.Infof("-------------------New ETCD, ver: %d, urls: %s", ver, urls)
	endpoints := strings.Split(urls, ",")
	if ver == 3 {
		klog.Infof("New ETCD v3 urls: %s", endpoints)
		return NewEtcdV3(endpoints)
	}

	klog.Infof("New ETCD v2 urls: %s", endpoints)
	return NewEtcdV2(endpoints)
}

func NewEtcdV2(urls []string) *Etcd {
	klog.Info("-------------------New ETCD config: [", urls, "]")
	var etcdCfg client.Config = client.Config{
		Endpoints:               urls,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: 10 * time.Second,
	}
	etcd := Etcd{config: etcdCfg, client: nil}
	for i := 0; i < MAXTIME; i++ {
		err := etcd.auth()
		if err != nil {
			klog.Error("NewEtcd:", err.Error())
			time.Sleep(3 * time.Second)
			continue
		}
		return &etcd
	}
	return nil
}

func (self *Etcd) auth() error {
	c, err := client.New(self.config)
	if err != nil {
		klog.Error(err)
		return err
	}

	kapi := client.NewKeysAPI(c)
	self.client = kapi
	return nil
}

//func (self *Etcd) SaveLeaf(k, v string) error {
//	klog.Info("add ETCD: key[", k, "]value[", v, "]")
//	_, err := self.client.Set(context.Background(), k, v, nil)
//	if err != nil {
//		klog.Error("SaveLeaf",err)
//		return err
//	}
//	return nil
//}
func (self *Etcd) SaveLeaf(k, v string) error {
	klog.Trace("add ETCD: key[", k, "]value[", v, "]")
	var err error
	for i := 0; i < MAXTIME; i++ {
		_, err1 := self.client.Set(context.Background(), k, v, nil)
		if err1 != nil {
			klog.Error("SaveLeaf", err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return nil
	}
	return err
}

//func (self *Etcd) ReadDir(k string) ([]*client.Node, error) {
//	klog.Info("get ETCD: key[", k, "]")
//	rsp, err := self.client.Get(context.Background(), k, nil)
//	if err != nil {
//		klog.Error(err)
//		return nil, err
//	}
//	return rsp.Node.Nodes, nil
//}

func (self *Etcd) ReadDir(k string) ([]*client.Node, error) {
	klog.Trace("get ETCD: key[", k, "]")
	var err error
	for i := 0; i < MAXTIME; i++ {
		rsp, err1 := self.client.Get(context.Background(), k, nil)
		if IsNotFindError(err1) {
			return nil, err1
		}
		if err1 != nil {
			klog.Error(err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return rsp.Node.Nodes, nil
	}
	return nil, err
}

//func (self *Etcd) ReadLeaf(k string) (string, error) {
//	klog.Info("get ETCD: key[", k, "]")
//	rsp, err := self.client.Get(context.Background(), k, nil)
//	if err != nil {
//		klog.Error(err)
//		return "", err
//	}
//	return rsp.Node.Value, nil
//}
func (self *Etcd) ReadLeaf(k string) (string, error) {
	klog.Trace("get ETCD: key[", k, "]")
	var err error
	for i := 0; i < MAXTIME; i++ {
		time.Sleep(100 * time.Millisecond)
		rsp, err1 := self.client.Get(context.Background(), k, nil)
		if IsNotFindError(err1) {
			return "", err1
		}
		if err1 != nil {
			klog.Error(err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return rsp.Node.Value, nil
	}
	return "", err
}

//func (self *Etcd) DeleteLeaf(k string) error {
//	klog.Info("del ETCD: key[", k, "]")
//	_, err := self.client.Delete(context.Background(), k, nil)
//	if err != nil {
//		klog.Error("deleteLeaf: Delete error!", err)
//		return err
//	}
//	return nil
//}
func (self *Etcd) DeleteLeaf(k string) error {
	klog.Trace("del ETCD: key[", k, "]")
	var err error
	for i := 0; i < MAXTIME; i++ {
		_, err1 := self.client.Delete(context.Background(), k, nil)
		if IsNotFindError(err1) {
			return err1
		}
		if err1 != nil {
			klog.Error("deleteLeaf: Delete error!", err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return nil
	}
	return err
}

//func (self *Etcd) DeleteDir(url string) error {
//	opt := client.DeleteOptions{}
//	opt.Dir = true
//	opt.Recursive = true
//	klog.Info("del ETCD: key[", url, "]")
//	_, err := self.client.Delete(context.Background(), url, &opt)
//	if err != nil {
//		klog.Error("deleteDir:Delete error!", err)
//		return err
//	}
//	return nil
//}
func (self *Etcd) DeleteDir(url string) error {
	opt := client.DeleteOptions{}
	opt.Dir = true
	opt.Recursive = true
	klog.Trace("del ETCD: key[", url, "]")
	var err error
	for i := 0; i < MAXTIME; i++ {
		_, err1 := self.client.Delete(context.Background(), url, &opt)
		if IsNotFindError(err1) {
			return err1
		}
		if err1 != nil {
			klog.Error("deleteDir:Delete error!", err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return nil
	}
	return err
}

//func (self *Etcd) WatcherDir(url string) (*client.Response, error) {
//	opts := client.WatcherOptions{
//		Recursive:  true,
//		AfterIndex: 0,
//	}
//
//	//klog.Info("watch ETCD dir [", url, "]")
//	got := self.client.Watcher(url, &opts)
//	resp, err := got.Next(context.Background())
//	if err != nil {
//		klog.Error(err)
//		return nil,err
//	}
//
//	return resp,nil
//}
func (self *Etcd) WatcherDir(url string) (*client.Response, error) {
	opts := client.WatcherOptions{
		Recursive:  true,
		AfterIndex: 0,
	}

	//klog.Info("watch ETCD dir [", url, "]")
	got := self.client.Watcher(url, &opts)
	var err error
	for i := 0; i < MAXTIME; i++ {
		resp, err1 := got.Next(context.Background())
		if IsNotFindError(err1) {
			return nil, err1
		}
		if err1 != nil {
			klog.Error(err1)
			err = err1
			time.Sleep(3 * time.Second)
			continue
		}
		return resp, nil
	}
	return nil, err
}

func (self *Etcd) Lock(k string) bool {
	opts := client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       time.Minute,
	}
	_, err := self.client.Set(context.Background(), k, "true", &opts)
	if err == nil {
		return true
	}
	return false
}

func (self *Etcd) Unlock(k string) bool {
	for i := 0; i < MAXTIME; i++ {
		_, err1 := self.client.Delete(context.Background(), k, nil)
		if err1 == nil {
			klog.Infof("unlock true %v", k)
			return true
		}
		if IsNotFindError(err1) {
			return false
		}
		if err1 != nil {
			klog.Errorf("Unlock: Delete error! %v", err1)
			time.Sleep(3 * time.Second)
			continue
		}
	}
	return false
}

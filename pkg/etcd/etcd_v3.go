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
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"

	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	DefaultDialTimeout    = 5 * time.Second
	DefaultRetryTimeout   = 3 * time.Second
	DefaultRequestTimeout = 10 * time.Second
)

var (
	ErrKeyNotFound         = errors.New("100: Key not found")
	ErrTooManyRecordsFound = errors.New("too many records found")
)

type KvPair struct {
	Key   string
	Value string
}

type DbRecords struct {
	Records []KvPair
}

type EtcdV3 struct {
	client *clientv3.Client
	config clientv3.Config
}

func NewEtcdV3(urls []string) *EtcdV3 {
	klog.Infof("-------------------New ETCD v3 endpoints: [%v]", urls)
	etcdCfg := clientv3.Config{
		Endpoints:   urls,
		DialTimeout: DefaultDialTimeout * time.Second,
	}

	etcd := EtcdV3{config: etcdCfg}
	for i := 0; i < MAXTIME; i++ {
		err := etcd.auth()
		if err != nil {
			klog.Errorf("NewEtcdV3 error: %v, retry auth after 3 sec", err)
			time.Sleep(DefaultRetryTimeout * time.Second)
			continue
		}
		return &etcd
	}
	return nil
}

func (self *EtcdV3) auth() error {
	cli, err := clientv3.New(self.config)
	if err != nil {
		klog.Errorf("Etcd.auth: clientv3.New(%v) error: %v", self.config, err)
		return err
	}

	self.client = cli
	return nil
}

func (self *EtcdV3) SaveLeaf(k, v string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	_, err := self.client.Put(ctx, k, v)
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.SaveLeaf: self.client.Put(ctx.TODO, key: %s, value: %s) error: %v", k, v, err)
		return err
	}
	return nil
}

func (self *EtcdV3) ReadLeaf(k string) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	rsp, err := self.client.Get(ctx, k)
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.ReadLeaf: self.client.Get(ctx.TODO, key: %s) error: %v", k, err)
		return "", err
	}

	if length := len(rsp.Kvs); length == 0 {
		klog.Errorf("Etcd.ReadLeaf: key: %s's value not found", k)
		return "", ErrKeyNotFound
	} else if length > 1 {
		klog.Errorf("Etcd.ReadLeaf: key: %s's too many values found", k)
		return "", ErrTooManyRecordsFound
	}

	return string(rsp.Kvs[0].Value), nil
}

func (self *EtcdV3) DeleteLeaf(k string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	_, err := self.client.Delete(ctx, k)
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.DeleteLeaf: delete key: %s's value error: %v", k, err)
		return err
	}
	return nil
}

func (self *EtcdV3) ReadDirV3(k string) (*DbRecords, error) {
	if !strings.HasSuffix(k, "/") {
		k += "/"
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	rsp, err := self.client.Get(ctx, k, clientv3.WithPrefix())
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.ReadDirV3: self.client.Get(ctx.TODO, key: %s, WithPrefix) error: %v", k, err)
		return nil, err
	}

	records := DbRecords{Records: make([]KvPair, 0)}
	for _, kv := range rsp.Kvs {
		basename := strings.TrimPrefix(string(kv.Key), k)
		if !strings.Contains(basename, "/") {
			records.Records = append(records.Records, KvPair{Key: string(kv.Key), Value: string(kv.Value)})
		}
	}

	return &records, nil
}

func (self *EtcdV3) ReadDir(dir string) ([]*client.Node, error) {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	rsp, err := self.client.Get(ctx, dir, clientv3.WithPrefix())
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.ReadDirV3: self.client.Get(ctx.TODO, key: %s, WithPrefix) error: %v", dir, err)
		return nil, err
	}

	cliNodes := make([]*client.Node, 0)
	for _, kv := range rsp.Kvs {
		baseName := strings.TrimPrefix(string(kv.Key), dir)
		if baseName == "" {
			klog.Infof("ReadDir: skip dir: %s self", string(kv.Key))
			continue
		}

		node := &client.Node{}
		segs := strings.Split(baseName, "/")
		if len(segs) > 1 {
			// directory
			subDir := segs[0]
			node.Dir = true
			node.Key = dir + subDir
		} else {
			// leaf
			node.Key = string(kv.Key)
			node.Value = string(kv.Value)
		}

		var findFlag bool = false
		for _, cliNode := range cliNodes {
			if cliNode.Key == node.Key {
				findFlag = true
			}
		}
		if !findFlag {
			cliNodes = append(cliNodes, node)
		}
	}

	klog.Debugf("Etcd.ReadDir: ReadDir(dir: %s) SUCC, result: %v", dir, cliNodes)
	return cliNodes, nil
}

func (self *EtcdV3) DeleteDir(k string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	_, err := self.client.Delete(ctx, k, clientv3.WithPrefix())
	cancelFunc()
	if err != nil {
		klog.Errorf("Etcd.DeleteDir: self.client.Delete(key: %s) error: %v", k, err)
		return err
	}
	klog.Debugf("Etcd.DeleteDir: self.client.Delete(key: %s) SUCC", k)
	return nil
}

func (self *EtcdV3) WatcherDir(k string) (*client.Response, error) {
	// todo, will not implement until discussed whether compatible with v2 dbaccessor api
	//panic("method not implemented")
	return nil, nil
}

func (self *EtcdV3) Lock(k string) bool {
	// todo, will not implement until discussed whether compatible with v2 dbaccessor api
	//panic("method not implemented")
	return false
}

func (self *EtcdV3) Unlock(k string) bool {
	// todo, will not implement until discussed whether compatible with v2 dbaccessor api
	//panic("method not implemented")
	return false
}

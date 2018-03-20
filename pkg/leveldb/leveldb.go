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

package leveldb

import (
	"errors"
	"strings"

	"github.com/coreos/etcd/client"
	lvldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

var (
	KeyNotFoundErr = errors.New("100: Key not found")
)

type LevelDB struct {
	db *lvldb.DB
}

func (l *LevelDB) SaveLeaf(k, v string) error {
	// write through to disk
	err := l.db.Put([]byte(k), []byte(v), &opt.WriteOptions{Sync: true})
	if err != nil {
		klog.Errorf("Set key: %s to value: %s error: %v", k, v, err)
		return err
	}

	klog.Debugf("Set key: %s to value: %s SUCC", k, v)
	return nil
}

func (l *LevelDB) ReadLeaf(k string) (string, error) {
	v, err := l.db.Get([]byte(k), nil)
	if err != nil {
		klog.Errorf("Get key: %s 's value error: %v", k, err)
		return "", err
	}

	klog.Debugf("Get key: %s of value: %s SUCC", k, string(v))
	return string(v), nil
}

func (l *LevelDB) DeleteLeaf(k string) error {
	err := l.db.Delete([]byte(k), &opt.WriteOptions{Sync: true})
	if err != nil {
		klog.Errorf("Delete key: %s error: %v", k, err)
		return err
	}

	klog.Debugf("Delete key: %s SUCC", k)
	return nil
}

func (l *LevelDB) ReadDir(dir string) ([]*client.Node, error) {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	cliNodes := make([]*client.Node, 0)
	iter := l.db.NewIterator(util.BytesPrefix([]byte(dir)), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		baseName := strings.TrimPrefix(string(key), dir)
		if baseName == "" {
			klog.Debugf("ReadDir: skip dir: %s self", dir)
			continue
		}

		klog.Tracef("ReadDir: append new key/value[%s : %s] to result", string(key), string(value))
		node := &client.Node{}

		segs := strings.Split(baseName, "/")
		if len(segs) > 1 {
			// directory
			subDir := segs[0]
			node.Dir = true
			node.Key = dir + subDir
		} else {
			// leaf
			node.Key = string(key)
			node.Value = string(value)
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

	klog.Debugf("ReadDir: result: %v", cliNodes)
	return cliNodes, nil
}

func (l *LevelDB) DeleteDir(dir string) error {
	iter := l.db.NewIterator(util.BytesPrefix([]byte(dir)), nil)
	for iter.Next() {
		key := iter.Key()
		err := l.DeleteLeaf(string(key))
		if err != nil {
			klog.Errorf("DeleteDir: delete key: %s error: %v", key, err)
			return err
		}
	}

	klog.Debugf("DeleteDir: %s SUCC", dir)
	return nil
}

// todo
func (l *LevelDB) WatcherDir(url string) (*client.Response, error) {
	return nil, nil
}

// todo
func (l *LevelDB) Lock(url string) bool {
	return true
}

// todo
func (l *LevelDB) Unlock(url string) bool {
	return true
}

// @param: path -- directory to store data, not need pre-create
func NewLevelDBClient(path string) (dbaccessor.DbAccessor, error) {
	db, err := lvldb.OpenFile(path, nil)
	if err != nil {
		klog.Errorf("lvldb.Open(path: %s, nil) error: %v", path, err)
		return nil, err
	}

	klog.Infof("open leveldb path:%s SUCC", path)
	return &LevelDB{db: db}, nil
}

package services

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/antonholmquist/jason"

	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/ZTE/Knitter/pkg/klog"
)

func InitConfigurations4Monitor(confObj *jason.Object, kubeconfig string) error {
	klogdir, err := confObj.GetString("log_dir")
	if err != nil {
		klog.Errorf("confObj.GetString(log_dir) err, error is [%v]", err)
		return err
	}
	klog.ConfigLog(klogdir)
	if kubeconfig == "" {
		klog.Infof("-kubeconfig not specified")
		flag.Set("kubeconfig", "/etc/kubernetes/kubectl.kubeconfig")
	}

	etcdAPIVer, err := confObj.GetInt64("etcd", "api_version")
	if err != nil {

		etcdAPIVer = int64(etcd.DefaultEtcdAPIVersion)
		klog.Warningf("InitEnv4Manger: get etcd api verison error: %v, use default: %d", err, etcdAPIVer)
	} else {
		klog.Infof("InitEnv4Manger: get etcd api verison: %d", etcdAPIVer)
	}

	etcdURL, _ := confObj.GetString("etcd", "urls")
	if etcdURL == "" {
		klog.Errorf("InitEnv4Monitor: etcd service query url is null")
		return errors.New("InitEnv4Monitor: etcd service query url is null")
	}
	infra.SetDataBase(etcd.NewEtcdWithRetry(int(etcdAPIVer), etcdURL))
	infra.CheckDB()
	klog.Info("InitEnv4Manger: etcd service query url:", etcdURL)

	manageriniterr := infra.InitManagerClient(confObj)

	if manageriniterr != nil {
		klog.Errorf("InitEnv4Monitor:Init cni manager error! Error:-%v", manageriniterr)
		return fmt.Errorf("%v:InitEnv4Monitor:Init manager error", manageriniterr)
	}

	waitManagerClient()
	return nil
}

func waitManagerClient() {
	for {
		managerClient := infra.GetManagerClient()
		checkErr := managerClient.CheckKnitterManager()
		if checkErr != nil {
			klog.Errorf("InitEnv4Monitor:CheckKnitterManager error! -%v",
				checkErr)
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
}

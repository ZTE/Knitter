package infra

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
)

var clientSet *kubernetes.Clientset

func InitKubernetesClientset(kubeCnfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeCnfig)
	if err != nil {
		klog.Errorf("clientcmd.BuildConfigFromFlags( ) err, error is [%v]", err)
		return err
	}

	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("kubernetes.NewForConfig(config) err, error is [%v]", err)
		return err
	}
	return nil
}

func GetClientset() *kubernetes.Clientset {
	klog.Debugf("")
	return clientSet
}

var uuidCluster string

func SetClusterID() {
	uuidCluster = GetClusterUUID()
}
func GetClusterID() string {
	return uuidCluster
}
func GetClusterUUID() string {
	key := dbaccessor.GetKeyOfClusterUUID()
	id, err := GetDataBase().ReadLeaf(key)
	if err == nil {
		return id
	}
	if !errobj.IsKeyNotFoundError(err) {
		return uuid.NIL.String()
	}
	id = uuid.NewUUID()
	err = GetDataBase().SaveLeaf(key, id)
	if err == nil {
		return id
	}
	return uuid.NIL.String()
}

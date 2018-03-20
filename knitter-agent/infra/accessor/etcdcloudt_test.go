package accessor

import (
	"errors"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/coreos/etcd/client"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

type TestEtcdCloudT struct {
	code string
}

func GetTestEtcd() *TestEtcdCloudT {
	api := TestEtcdCloudT{code: ""}
	return &api
}

func (self *TestEtcdCloudT) SaveLeaf(k, v string) error {

	var err = errors.New("saveleaf rtn err")

	if strings.Contains(k, "testportiderror") {
		klog.Info("SaveLeaf Rtn Err")
		return err
	}
	return nil
}
func (self *TestEtcdCloudT) ReadDir(k string) ([]*client.Node, error) {
	return nil, nil
}
func (self *TestEtcdCloudT) ReadLeaf(k string) (string, error) {

	var err = errors.New("readleaf rtn err")

	if strings.Contains(k, "testportidnormal/self") {
		return "portinfo", nil
	}
	if strings.Contains(k, "testportidnormal/veths") {
		return "vethinfo", nil
	}
	if strings.Contains(k, "testportidnormal/ovsbrs") {
		return "ovsbrinfo", nil
	}
	if strings.Contains(k, "testportidnormal/driver") {
		return "driverinfo", nil
	}
	if strings.Contains(k, "testportiderror") {
		klog.Info("ReadLeaf Rtn Err")
		return "", err
	}
	return "", nil
}
func (self *TestEtcdCloudT) DeleteLeaf(k string) error {
	return nil
}
func (self *TestEtcdCloudT) DeleteDir(url string) error {
	var err = errors.New("deletedir rtn err")

	if strings.Contains(url, "testportiderror") {
		klog.Info("DeleteDir Rtn Err")
		return err
	}
	return nil
}
func (self *TestEtcdCloudT) WatcherDir(url string) (*client.Response, error) {
	return nil, nil
}
func (self *TestEtcdCloudT) Lock(k string) bool {
	return true
}
func (self *TestEtcdCloudT) Unlock(k string) bool {
	return true
}

func Test_Set4CloudT_Normal(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"
	testcloudt := CloudTInfo{PortInfo: "portinfo",
		VethInfo:       "vethinfo",
		OvsbrInfo:      "ovsbrinfo",
		MechDriverType: "driverinfo"}

	err := Set4CloudT(GetTestEtcd(), testurl, testportidnor, testcloudt)

	Convey("Subject:Test_Set4CloudT_Normal", t, func() {
		Convey("return data err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
	})
}
func Test_Set4CloudT_Except1(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"
	testcloudt := CloudTInfo{PortInfo: "portinfo",
		VethInfo:       "vethinfo",
		OvsbrInfo:      "ovsbrinfo",
		MechDriverType: "driverinfo"}
	expErrStr := "db is null"

	err := Set4CloudT(nil, testurl, testportidnor, testcloudt)

	Convey("Subject:Test_Set4CloudT_Except1", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}
func Test_Set4CloudT_Except2(t *testing.T) {

	testurl := "testurl"
	testportiderr := "testportiderror"
	testcloudt := CloudTInfo{PortInfo: "portinfo",
		VethInfo:       "vethinfo",
		OvsbrInfo:      "ovsbrinfo",
		MechDriverType: "driverinfo"}
	expErrStr := "saveleaf rtn err"

	err := Set4CloudT(GetTestEtcd(), testurl, testportiderr, testcloudt)

	Convey("Subject:Test_Set4CloudT_Except2", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}

func Test_Get4CloudT_Normal(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"
	expPort := "portinfo"
	expVeth := "vethinfo"
	expOvsbr := "ovsbrinfo"
	expDriver := "driverinfo"

	cloudt, err := Get4CloudT(GetTestEtcd(), testurl, testportidnor)
	Convey("Subject:Test_Get4CloudT_Normal", t, func() {
		Convey("return data err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
		Convey("return data ShouldEqual"+expPort, func() {
			So(cloudt.PortInfo, ShouldEqual, expPort)
		})
		Convey("return data ShouldEqual"+expVeth, func() {
			So(cloudt.VethInfo, ShouldEqual, expVeth)
		})
		Convey("return data ShouldEqual"+expOvsbr, func() {
			So(cloudt.OvsbrInfo, ShouldEqual, expOvsbr)
		})
		Convey("return data ShouldEqual"+expDriver, func() {
			So(cloudt.MechDriverType, ShouldEqual, expDriver)
		})
	})
}
func Test_Get4CloudT_Except1(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"
	expErrStr := "db is null"

	_, err := Get4CloudT(nil, testurl, testportidnor)

	Convey("Subject:Test_Get4CloudT_Except1", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}
func Test_Get4CloudT_Except2(t *testing.T) {

	testurl := "testurl"
	testportiderr := "testportiderror"

	expErrStr := "readleaf rtn err"

	_, err := Get4CloudT(GetTestEtcd(), testurl, testportiderr)

	Convey("Subject:Test_Get4CloudT_Except2", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}

func Test_Del4CloudT_Normal(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"

	err := Del4CloudT(GetTestEtcd(), testurl, testportidnor)

	Convey("Subject:Test_Del4CloudT_Normal", t, func() {
		Convey("return data err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
	})
}
func Test_Del4CloudT_Except1(t *testing.T) {

	testurl := "testurl"
	testportidnor := "testportidnormal"
	expErrStr := "db is null"

	err := Del4CloudT(nil, testurl, testportidnor)
	Convey("Subject:Test_Del4CloudT_Except1", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}
func Test_Del4CloudT_Except2(t *testing.T) {

	testurl := "testurl"
	testportiderr := "testportiderror"
	expErrStr := "deletedir rtn err"

	err := Del4CloudT(GetTestEtcd(), testurl, testportiderr)
	Convey("Subject:Test_Del4CloudT_Except2", t, func() {
		Convey("return data err ShouldEqual"+expErrStr, func() {
			So(err.Error(), ShouldEqual, expErrStr)
		})
	})
}

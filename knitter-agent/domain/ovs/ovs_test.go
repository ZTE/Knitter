package ovs

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	. "github.com/golang/gostub"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetOvsBrg(t *testing.T) {

	Convey("TestGetOvsBrg \n", t, func() {
		Convey("TestGetOvsBrg"+"---want---"+"Error(No file) "+"\n", func() {
			confPath := "ovsTestErr4.conf"
			phyNw := "physnet1"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"Error(From GetSection(ovs)) "+"\n", func() {
			confPath := "ovsTestErr.conf"
			phyNw := "physnet1"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"Error(From GetKey(bridge_mappings)) "+"\n", func() {
			confPath := "ovsTestErr1.conf"
			phyNw := "physnet1"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"Error(From GetPhyNwOvsPair len<2) "+"\n", func() {
			confPath := "ovsTestErr2.conf"
			phyNw := "physnet1"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"Error(From phyNwOvsPair[phyNw]) "+"\n", func() {
			confPath := "ovsTestErr3.conf"
			phyNw := "physnet2"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"All OK "+"\n", func() {
			confPath := "ovsTestOK.conf"
			phyNw := "physnet1"
			_, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldBeNil)
		})
		Convey("TestGetOvsBrg"+"---want---"+"OK if phyNw is nil and one configure"+"\n", func() {
			confPath := "ovsTestOK1.conf"
			ovsExp := "br-phy1"
			phyNw := ""
			ovs, err := GetOvsBrg(phyNw, confPath)
			So(err, ShouldBeNil)
			So(ovs, ShouldEqual, ovsExp)
		})
	})
}

func TestGetNwMechDriver(t *testing.T) {
	Convey("TestGetNwMechDriver \n", t, func() {
		Convey("TestGetNwMechDriver"+"---want---"+"Error (No file) "+"\n", func() {
			confPath := "ovsTestErr4.conf"
			phyNw := "physnet1"
			portType := "normal"
			_, err := GetNwMechDriver(phyNw, portType, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetNwMechDriver"+"---want---"+"Error (From SearchDriver) "+"\n", func() {
			confPath := "ovsTestErr5.conf"
			phyNw := "physnet1"
			portType := "normal"
			_, err := GetNwMechDriver(phyNw, portType, confPath)
			So(err, ShouldNotBeNil)
		})
		Convey("TestGetNwMechDriver"+"---want---"+"All OK "+"\n", func() {
			confPath := "ovsTestOK.conf"
			phyNw := "physnet1"
			portType := "normal"
			_, err := GetNwMechDriver(phyNw, portType, confPath)
			So(err, ShouldBeNil)
		})
	})
}

//func TestFtDeleteVethPair(t *testing.T){
//	Convey("TestFtDeleteVethPair \n", t, func() {
//		Convey("TestFtDeleteVethPair"+"---want---"+"All OK"+"\n", func() {
//			vethPair,err := CreateVethPair()
//			err = DeleteVethPair(vethPair)
//			So(err, ShouldBeNil)
//		})
//	})
//}
//
//func TestFtCreateVethPair(t *testing.T){
//
//	Convey("TestFtCreateVethPair \n", t, func() {
//		Convey("TestFtCreateVethPair"+"---want---"+"All OK"+"\n", func() {
//			_,err := CreateVethPair()
//			So(err, ShouldBeNil)
//		})
//	})
//}

func TestWaitOvsUsable(t *testing.T) {
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 4},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable:\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestWaitOvsUsable_addOvsBr_RetrySucc(t *testing.T) {
	errStr := "ovs-vsctl error"
	outPuts := []Output{
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 39},
		{StubVals: Values{"", nil}, Times: 4},
	}
	stubs := StubFuncSeq(&osencap.Exec, outPuts)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_addOvsBr_RetrySucc:\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestWaitOvsUsable_addOvsBr_RetryFail(t *testing.T) {
	errStr := "ovs-vsctl error"
	outPuts := []Output{
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 40},
	}
	stubs := StubFuncSeq(&osencap.Exec, outPuts)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_addOvsBr_RetryFail:\n", t, func() {
		So(err, ShouldEqual, errobj.ErrWaitOvsUsableFailed)
	})
}

func TestWaitOvsUsable_checkOvsPort_AddPort_RetrySucc(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 1},
		{StubVals: Values{"", nil}, Times: 3},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_checkOvsPort_AddPort_RetrySucc:\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestWaitOvsUsable_checkOvsPort_AddPort_RetryFailed(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 40},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_checkOvsPort_AddPort_RetryFailed:\n", t, func() {
		So(err, ShouldEqual, errobj.ErrWaitOvsUsableFailed)
	})
}

func TestWaitOvsUsable_checkOvsPort_DelPort_RetrySucc(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 1},
		{StubVals: Values{"", nil}, Times: 3},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_checkOvsPort_AddPort_RetrySucc:\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestWaitOvsUsable_checkOvsPort_DelPort_RetryFail(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"", nil}, Times: 1},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 40},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_checkOvsPort_DelPort_RetryFail:\n", t, func() {
		So(err, ShouldEqual, errobj.ErrWaitOvsUsableFailed)
	})
}

func TestWaitOvsUsable_delOvsBr_RetrySucc(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 3},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 39},
		{StubVals: Values{"", nil}, Times: 3},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_delOvsBr_RetrySucc:\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestWaitOvsUsable_delOvsBr_RetryFail(t *testing.T) {
	errStr := "database connection failed"
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 3},
		{StubVals: Values{"excute failed", errors.New(errStr)}, Times: 40},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	stubs.StubFunc(&TimeSleepFunc)
	defer stubs.Reset()

	err := WaitOvsUsable()
	Convey("TestWaitOvsUsable_delOvsBr_RetryFail:\n", t, func() {
		So(err, ShouldEqual, errobj.ErrWaitOvsUsableFailed)
	})
}

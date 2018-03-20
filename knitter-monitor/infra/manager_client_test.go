package infra

import (
	"errors"
	"github.com/antonholmquist/jason"
	"github.com/golang/gostub"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"strings"
	"testing"
)

func TestInitClient(t *testing.T) {
	jsonConf := []byte(`{
  "conf":{
  "agent": {
      "manager": {
        "url": "http://192.168.1.204:9527"
      },
      "host": {
        "vm_id": "XXXXXXXXXXXXXXXXXXXX",
        "remote_net_type": "sdn"
      }
      }}}`)
	obj, _ := jason.NewObjectFromBytes(jsonConf)
	objCfg, _ := obj.GetObject("conf", "agent")

	Convey("TestInitClient\n", t, func() {
		Convey("All-OK", func() {
			err := InitManagerClient(objCfg)
			So(err, ShouldBeNil)
		})
	})
}

func TestInitClientErrManager(t *testing.T) {
	jsonConf := []byte(`{
  "conf":{
  "agent": {
      "manager": {
        "url": ""
      },
      "host": {
        "vm_id": "XXXXXXXXXXXXXXXXXXXX",
        "remote_net_type": "sdn"
      }
      }}}`)
	obj, _ := jason.NewObjectFromBytes(jsonConf)
	objCfg, _ := obj.GetObject("conf", "agent")

	Convey("TestInitClient\n", t, func() {
		Convey("ErrManager", func() {
			err := InitManagerClient(objCfg)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestInitClientNoManager(t *testing.T) {
	jsonConf := []byte(`{
  "conf":{
  "agent": {
      "host": {
        "vm_id": "XXXXXXXXXXXXXXXXXXXX",
        "remote_net_type": "sdn"
      }
      }}}`)
	obj, _ := jason.NewObjectFromBytes(jsonConf)
	objCfg, _ := obj.GetObject("conf", "agent")

	Convey("TestInitClient\n", t, func() {
		Convey("NoManager", func() {
			err := InitManagerClient(objCfg)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetManagerClient(t *testing.T) {
	mcExpect := &ManagerClient{}
	managerClient = mcExpect
	Convey("TestGetManagerClient", t, func() {
		mc := GetManagerClient()
		So(mc, ShouldEqual, mcExpect)
	})
}

func TestCreateNeutronPortOK(t *testing.T) {
	mc := &ManagerClient{}
	port := &ManagerCreateBulkPortsReq{}
	HTTPPost = testPost2Master
	HTTPClose = testClose
	HTTPReadAll = testReadAll
	Convey("TestInitClient\n", t, func() {
		Convey("NoVMID", func() {
			rsp, err := mc.CreateNeutronBulkPorts("testUrl200",
				port, "tenant-uuid-for-req")
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
	})
}

func testClose(resp *http.Response) error {
	return nil
}

func testReadAll(resp *http.Response) ([]byte, error) {
	testByte := []byte("respod message")
	return testByte, nil
}

func testPost2Master(url string, bodyType string, postData []byte) (*http.Response, error) {
	var err = errors.New("url error")
	resp := &http.Response{}
	resp.StatusCode = 123
	if strings.Contains(url, "testUrlError") {
		return resp, err
	} else if strings.Contains(url, "testUrlOK") {
		return resp, nil
	} else if strings.Contains(url, "testUrl200") {
		resp.StatusCode = 200
		return resp, nil
	}
	return resp, nil
}

func TestCreateNeutronPort0Err(t *testing.T) {
	mc := ManagerClient{}
	port := &ManagerCreateBulkPortsReq{}
	HTTPPost = testPost2Master
	HTTPClose = testClose
	HTTPReadAll = testReadAll
	Convey("TestInitClient\n", t, func() {
		Convey("NoVMID", func() {
			rsp, err := mc.CreateNeutronBulkPorts("testUrlOK",
				port, "tenant-uuid-for-req")
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
		})
	})
}

func TestCreateNeutronPort1Err(t *testing.T) {
	mc := ManagerClient{}
	port := &ManagerCreateBulkPortsReq{}
	HTTPPost = testPost2Master
	HTTPClose = testClose
	HTTPReadAll = testReadAll
	Convey("TestInitClient\n", t, func() {
		Convey("NoVMID", func() {
			rsp, err := mc.CreateNeutronBulkPorts("testUrlError",
				port, "tenant-uuid-for-req")
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
		})
	})
}

func TestCreateNeutronPort2Err(t *testing.T) {
	mc := ManagerClient{}
	port := &ManagerCreateBulkPortsReq{}
	HTTPPost = testPost2Master
	HTTPClose = testClose
	Convey("TestInitClient\n", t, func() {
		Convey("NoVMID", func() {
			rsp, err := mc.CreateNeutronBulkPorts("testUrlError",
				port, "tenant-uuid-for-req")
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
		})
	})
}

func TestHttpGet(t *testing.T) {
	mc := ManagerClient{}
	HTTPGet = testGet2Master
	HTTPClose = testClose
	HTTPReadAll = testReadAll

	Convey("TestHttpGet \n", t, func() {
		Convey("Error", func() {
			url := "testUrlError"
			_, _, err := mc.Get(url)
			So(err, ShouldNotBeNil)
		})
		Convey("All OK", func() {
			url := "testUrlOK"
			_, _, err := mc.Get(url)
			So(err, ShouldBeNil)
		})
	})
}

func testGet2Master(url string) (*http.Response, error) {
	var err = errors.New("url error")
	resp := &http.Response{}
	resp.StatusCode = 123
	if strings.Contains(url, "testUrlError") {
		return resp, err
	} else if strings.Contains(url, "testUrlOK") {
		return resp, nil
	} else if strings.Contains(url, "testUrl200") {
		resp.StatusCode = 200
		return resp, nil
	}
	return resp, nil
}

func TestDelete(t *testing.T) {
	mc := ManagerClient{}
	HTTPDelete = testDelete2Master
	HTTPClose = testClose
	HTTPReadAll = testReadAll

	Convey("TestDelete \n", t, func() {
		Convey("TestDelete"+"---want---"+"Error(From DeleteFunc)", func() {
			url := "testUrlError"
			_, _, err := mc.Delete(url)
			So(err, ShouldNotBeNil)
		})
		Convey("TestDelete"+"---want---"+"All OK", func() {
			url := "testUrlOK"
			_, _, err := mc.Delete(url)
			So(err, ShouldBeNil)
		})
	})
}

func testDelete2Master(url string) (*http.Response, error) {
	var err = errors.New("url error")
	resp := &http.Response{}
	resp.StatusCode = 123
	if strings.Contains(url, "testUrlError") {
		return resp, err
	} else if strings.Contains(url, "testUrlOK") {
		return resp, nil
	}
	return resp, nil
}

func TestDeleteNeutronPortOK(t *testing.T) {
	mc := ManagerClient{}
	HTTPDelete = testDelete2Master
	HTTPReadAll = testReadAll
	HTTPClose = testClose
	Convey("TestDeleteNeutronPortOK\n", t, func() {
		err := mc.DeleteNeutronPort("port-uuid-for-delete-req",
			"tenant-uuid-for-req-testUrlOK")
		So(err, ShouldNotBeNil)
	})
}

func TestDeleteNeutronPort2Err(t *testing.T) {
	mc := ManagerClient{}
	HTTPDelete = testDelete2Master
	HTTPReadAll = testReadAll
	HTTPClose = testClose
	Convey("TestDeleteNeutronPort2Err\n", t, func() {
		err := mc.DeleteNeutronPort("port-uuid-for-delete-req",
			"tenant-uuid-for-req-testUrlError")
		So(err, ShouldNotBeNil)
	})
}

func TestCheckKnitterManager(t *testing.T) {
	mc := ManagerClient{}
	resp := &http.Response{StatusCode: 200}
	gostub.StubFunc(&HTTPGet, resp, nil)
	netinfo := string(`{"state": "good"}`)
	gostub.StubFunc(&HTTPReadAll, []byte(netinfo), nil)
	HTTPClose = testClose
	Convey("TestCheckKnitterManager\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldBeNil)
	})
}

func TestCheckKnitterManagerErr(t *testing.T) {
	mc := ManagerClient{}
	gostub.StubFunc(&HTTPGet, nil, errors.New("get-err"))
	Convey("TestCheckKnitterManagerErr\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldNotBeNil)
	})
}

func TestCheckKnitterManagerErr1(t *testing.T) {
	mc := ManagerClient{}
	resp := &http.Response{StatusCode: 123}
	gostub.StubFunc(&HTTPGet, resp, nil)
	netinfo := string(`{"state": "good"}`)
	gostub.StubFunc(&HTTPReadAll, []byte(netinfo), nil)
	HTTPClose = testClose
	Convey("TestCheckKnitterManagerErr1\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldNotBeNil)
	})
}

func TestCheckKnitterManagerErr2(t *testing.T) {
	mc := ManagerClient{}
	resp := &http.Response{StatusCode: 200}
	gostub.StubFunc(&HTTPGet, resp, nil)
	netinfo := string("{'state': 'good'}")
	gostub.StubFunc(&HTTPReadAll, []byte(netinfo), nil)
	HTTPClose = testClose
	Convey("TestCheckKnitterManagerErr2\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldNotBeNil)
	})
}

func TestCheckKnitterManagerErr3(t *testing.T) {
	mc := ManagerClient{}
	resp := &http.Response{StatusCode: 200}
	gostub.StubFunc(&HTTPGet, resp, nil)
	netinfo := string(`{"sta": "good"}`)
	gostub.StubFunc(&HTTPReadAll, []byte(netinfo), nil)
	HTTPClose = testClose
	Convey("TestCheckKnitterManagerErr3\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldNotBeNil)
	})
}

func TestCheckKnitterManagerErr4(t *testing.T) {
	mc := ManagerClient{}
	resp := &http.Response{StatusCode: 200}
	gostub.StubFunc(&HTTPGet, resp, nil)
	netinfo := string(`{"state": "bad"}`)
	gostub.StubFunc(&HTTPReadAll, []byte(netinfo), nil)
	HTTPClose = testClose
	Convey("TestCheckKnitterManagerErr4\n", t, func() {
		err := mc.CheckKnitterManager()
		So(err, ShouldNotBeNil)
	})
}

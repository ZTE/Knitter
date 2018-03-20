package cni

import (
	"testing"

	"github.com/containernetworking/cni/pkg/skel"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	cniPram := &CniParam{}
	args := &skel.CmdArgs{ContainerID: "11111",
		Netns:     "/proc/123/ns/net",
		IfName:    "eth0",
		Args:      "IgnoreUnknown=1;K8S_POD_NAMESPACE=itran;K8S_POD_NAME=713f7c06-6b23-49e4-b76b-8eba9e2b8445-1-zdfb7;K8S_POD_INFRA_CONTAINER_ID=c018d18f0de26371f02bef8e3802c33c002e33889ddb6f5409e3df5f52e1803d",
		Path:      "/opt/cni/bin:/opt/knitter/bin",
		StdinData: []byte("{\"stdin_data\":\"\"}")}
	Convey("TestInit\n", t, func() {
		Convey("TestInit"+" ok \n", func() {
			err := cniPram.init(args)
			So(err, ShouldBeNil)
		})
		Convey("TestInit"+" mashal error \n", func() {
			args.StdinData = []byte("{\"stdin_data\":\"\",}")
			err := cniPram.init(args)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestAnalyzeCniParam(t *testing.T) {
	cniPram := &CniParam{}
	args := &skel.CmdArgs{ContainerID: "11111",
		Netns:     "/proc/123/ns/net",
		IfName:    "eth0",
		Args:      "IgnoreUnknown=1;K8S_POD_NAMESPACE=itran;K8S_POD_NAME=713f7c06-6b23-49e4-b76b-8eba9e2b8445-1-zdfb7;K8S_POD_INFRA_CONTAINER_ID=c018d18f0de26371f02bef8e3802c33c002e33889ddb6f5409e3df5f52e1803d",
		Path:      "/opt/cni/bin:/opt/knitter/bin",
		StdinData: []byte("{\"stdin_data\":\"\"}")}
	Convey("TestAnalyzeCniParam\n", t, func() {
		Convey("TestAnalyzeCniParam"+" ok \n", func() {
			err := cniPram.AnalyzeCniParam(args)
			So(err, ShouldBeNil)
		})
		Convey("TestAnalyzeCniParam"+" init error \n", func() {
			args.StdinData = []byte("{\"stdin_data\":\"\",}")
			err := cniPram.AnalyzeCniParam(args)
			So(err, ShouldNotBeNil)
		})
	})
}

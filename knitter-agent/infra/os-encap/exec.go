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

package osencap

import (
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
	"os/exec"
)

var Exec = func(cmd string, args ...string) (string, error) {
	cmdpath, err := exec.LookPath(cmd)
	if err != nil {
		klog.Errorf("exec.LookPath err: %v, cmd: %s", err, cmd)
		return "", infra.ErrExecLookPathFailed
	}

	var output []byte
	output, err = exec.Command(cmdpath, args...).CombinedOutput()
	if err != nil {
		klog.Errorf("exec.Command.CombinedOutput err: %v, cmd: %s, output: %s", err, cmd, string(output))
		return "", infra.ErrExecCombinedOutputFailed
	}
	//klog.Info("CMD[", cmdpath, "]ARGS[", args, "]OUT[", string(output), "]")
	return string(output), nil
}

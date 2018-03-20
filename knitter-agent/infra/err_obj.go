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

package infra

import (
	"errors"
)

var ErrJSONMarshalFailed = errors.New("json marshal failded")
var ErrJSONUnmarshalFailed = errors.New("json unmarshal failded")
var ErrGetRAMDiskFailed = errors.New("get ram disk path failded")
var ErrWriteFileFailed = errors.New("write file failded")
var ErrReadFileFailed = errors.New("read file failded")
var ErrOsStatFailed = errors.New("os stat failded")
var ErrExecLookPathFailed = errors.New("exec lookpath failded")
var ErrExecCombinedOutputFailed = errors.New("exec combined output failded")

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

package adapter

import (
	"github.com/ZTE/Knitter/pkg/openstack/driver/http"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
)

var (
	CreateBulkPorts = func(c *gophercloud.ServiceClient, opts ports.CreateOptsBuilder) ([]*ports.Port, error) {
		return ports.CreateBulk(c, opts).ExtractBulk()
	}
	DoHttpPost = func(url string, body map[string]interface{}, headers map[string]string) (int, []byte, error) {
		return http.GetHTTPClientObj().Post(url, body, headers)
	}
	DoHttpGet = func(url string, headers map[string]string) (int, []byte, error) {
		return http.GetHTTPClientObj().Get(url, headers)
	}
	DoHttpDelete = func(url string, headers map[string]string) (int, error) {
		return http.GetHTTPClientObj().Delete(url, headers)
	}
)

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

package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"io/ioutil"
	"net/http"
)

type HTTPMethods interface {
	Get(url string, headers map[string]string) (int, []byte, error)
	Post(url string, body map[string]interface{}, headers map[string]string) (int, []byte, error)
	Delete(url string, headers map[string]string) (int, error)
}

type httpClient struct{}

var httpClientObject HTTPMethods = &httpClient{}

var GetHTTPClientObj = func() HTTPMethods {
	return httpClientObject
}

func (self *httpClient) Get(url string, headers map[string]string) (int, []byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Error("Get: http NewRequest error: ", err.Error())
		return http.StatusInternalServerError, nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	if headers != nil {
		for k, v := range headers {
			if v != "" {
				request.Header.Set(k, v)
			} else {
				request.Header.Del(k)
			}
		}
	}

	response, err := client.Do(request)
	var status int
	if response != nil {
		status = response.StatusCode
	} else {
		status = http.StatusInternalServerError
	}
	if err != nil {
		klog.Error("Get: client.Do error: ", err.Error())
		return status, nil, fmt.Errorf("%v:http client.Do error", err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Error("Get: ioutil.ReadAll: [", url, "] error: ", err.Error())
		return status, nil, fmt.Errorf("%v:ioutil.ReadAll response error", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		klog.Error("Get: http client.Do[", url, "] response error, status code: ", response.StatusCode, "response body: ", string(body))
		return status, body, errors.New("http client.Do response error")
	}
	return status, body, nil
}

func (self *httpClient) Post(url string, body map[string]interface{}, headers map[string]string) (int, []byte, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		klog.Error("Post: json.Marshal [", body, "]error: ", err.Error())
		return http.StatusInternalServerError, nil, fmt.Errorf("%v:Post: json.Marshal body error", err)
	}

	bodyReader := bytes.NewReader(bodyBytes)
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		klog.Error("Post: http NewRequest error: ", err.Error())
		return http.StatusInternalServerError, nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	if headers != nil {
		for k, v := range headers {
			if v != "" {
				request.Header.Set(k, v)
			} else {
				request.Header.Del(k)
			}
		}
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	var status int
	if response != nil {
		status = response.StatusCode
	} else {
		status = http.StatusInternalServerError
	}
	if err != nil {
		klog.Error("Post: client.Do error: ", err.Error())
		return status, nil, errors.New("http client.Do error")
	}

	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Error("Post: ioutil.ReadAll: [", url, "] error: ", err.Error())
		return status, nil, fmt.Errorf("%v:ioutil.ReadAll response error", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		klog.Error("Post: http client.Do[", url, "] response error, status code: ", response.StatusCode, ", response body: ", string(respBody))
		return status, respBody, errors.New("http client.Do response error")
	}
	return status, respBody, nil
}

func (self *httpClient) Delete(url string, headers map[string]string) (int, error) {
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		klog.Error("Delete: http NewRequest error: ", err.Error())
		return http.StatusInternalServerError, fmt.Errorf("%v:http NewRequest error", err)
	}

	if headers != nil {
		for k, v := range headers {
			if v != "" {
				request.Header.Set(k, v)
			} else {
				request.Header.Del(k)
			}
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return response.StatusCode, err
}

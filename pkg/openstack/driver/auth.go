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

package driver

import (
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/openstack/driver/http"
	"github.com/antonholmquist/jason"
	"strings"
	"sync"
)

const (
	NetworkEndpointType = "network"
	ComputeEndpointType = "compute"
)

type AuthClient struct {
	AuthConfig
}

type AuthConfig struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	IdentityEndpoint string

	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username. In Identity V3, either
	// UserID or a combination of Username and DomainID or DomainName are needed.
	Username, UserID string

	// Exactly one of Password or APIKey is required for the Identity V2 and V3
	// APIs. Consult with your provider's control panel to discover your account's
	// preferred method of authentication.
	Password, APIKey string

	// At most one of DomainID and DomainName must be provided if using Username
	// with Identity V3. Otherwise, either are optional.
	DomainID, DomainName string

	// The TenantID and TenantName fields are optional for the Identity V2 API.
	// Some providers allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID, TenantName string

	// AllowReauth should be set to true if you grant permission for Gophercloud to
	// cache your credentials in memory, and to allow Gophercloud to attempt to
	// re-authenticate automatically if/when your token expires.  If you set it to
	// false, it will not cache these settings, but re-authentication will not be
	// possible.  This setting defaults to false.
	AllowReauth bool

	// ReauthFunc is the function used to re-authenticate the user if the request
	// fails with a 401 HTTP response code. This a needed because there may be multiple
	// authentication functions for different Identity service versions.
	ReauthFunc func() error

	// TokenID allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	TokenID string

	NetworkEndpoint string

	ComputeEndpoint string

	mutex sync.Mutex
}

var instance *AuthConfig
var authOnce sync.Once

func getAuthSingleton() *AuthConfig {
	authOnce.Do(func() {
		instance = &AuthConfig{}
	})

	return instance
}

func (auth *AuthConfig) setConf(config OpenStackConf) {
	auth.AllowReauth = true
	auth.Username = config.Username
	auth.Password = config.Password
	auth.IdentityEndpoint = NormalizeURL(config.Url) + "/tokens"
	auth.TenantID = config.Tenantid
}

func AuthCheck(config *OpenStackConf) (*AuthClient, error) {
	auth := AuthClient{}
	auth.setConf(*config)
	err := auth.auth()
	return &auth, err
}

func (auth *AuthConfig) auth() error {
	auth.mutex.Lock()
	defer auth.mutex.Unlock()
	url := auth.IdentityEndpoint
	body := auth.makeTokenv2Body()
	status, rspBytes, err := http.GetHTTPClientObj().Post(url, body, nil)
	if err != nil {
		klog.Error("auth: Post url[", url, "], status[", status, "] error: ", err.Error())
		return fmt.Errorf("%v,%v,%v:auth: Post request error", url, status, err)
	}

	rspJasObj, err := jason.NewObjectFromBytes(rspBytes)
	if err != nil {
		klog.Error("auth: NewObjectFromBytes error: ", err.Error())
		return fmt.Errorf("%v:auth: NewObjectFromBytes parse response body error", err)
	}

	err = auth.setTokenAndTenantInfo(rspJasObj)
	if err != nil {
		klog.Error("auth: setTokenAndTenantInfo error: ", err.Error())
		return fmt.Errorf("%v:auth: setTokenAndTenantInfo error", err)
	}

	err = auth.setEndpoints(rspJasObj)
	if err != nil {
		klog.Error("auth: setEndpoints error: ", err.Error())
		return fmt.Errorf("%v:auth: setEndpoints error", err)
	}

	return nil
}

func (auth *AuthConfig) setTokenAndTenantInfo(authJSON *jason.Object) error {
	tokenId, err := authJSON.GetString("access", "token", "id")
	if err != nil {
		klog.Error("auth: GetString access->token->id error: ", err.Error())
		return fmt.Errorf("%v:auth: GetString access->token->id error", err)
	}
	//set token id
	auth.TokenID = tokenId

	tenantId, err := authJSON.GetString("access", "token", "tenant", "id")
	if err != nil {
		klog.Error("auth: GetString access->token->tenant->id error: ", err.Error())
		return fmt.Errorf("%v:auth: GetString access->token->tenant->id error", err)
	}
	auth.TenantID = tenantId

	tenantName, err := authJSON.GetString("access", "token", "tenant", "name")
	if err != nil {
		klog.Error("auth: GetString access->token->tenant->name error: ", err.Error())
		return fmt.Errorf("%v:auth: GetString access->token->tenant->name error", err)
	}
	auth.TenantName = tenantName
	return nil
}

func (auth *AuthConfig) setEndpoints(authJSON *jason.Object) error {
	srvCatalogList, err := authJSON.GetValueArray("access", "serviceCatalog")
	if err != nil {
		klog.Error("auth: GetValueArray access->serviceCatalog error: ", err.Error())
		return fmt.Errorf("%v:auth: GetValueArray access->serviceCatalog error", err)
	}

	//set endpoint info
	for _, srvCatalog := range srvCatalogList {
		srvCatalogEle, _ := srvCatalog.Object()
		typename, _ := srvCatalogEle.GetString("type")
		if err != nil {
			klog.Error("auth: GetString access->serviceCatalog->type error: ", err.Error())
			return fmt.Errorf("%v:auth: GetString access->serviceCatalog->type error", err)
		}
		if typename == NetworkEndpointType {
			endpoints, err := srvCatalogEle.GetValueArray("endpoints")
			if err != nil {
				klog.Error("auth: GetValueArray endpoints error: ", err.Error())
				return fmt.Errorf("%v:auth: GetValueArray endpoints error", err)
			}

			endpointsEle, _ := endpoints[0].Object()
			auth.NetworkEndpoint, _ = endpointsEle.GetString("publicURL")
			auth.NetworkEndpoint = NormalizeURL(auth.NetworkEndpoint) + "v2.0/"
		} else if typename == ComputeEndpointType {
			endpoints, err := srvCatalogEle.GetValueArray("endpoints")
			if err != nil {
				klog.Error("auth: GetValueArray endpoints error: ", err.Error())
				return fmt.Errorf("%v:auth: GetValueArray endpoints error", err)
			}

			endpointsEle, _ := endpoints[0].Object()
			auth.ComputeEndpoint, _ = endpointsEle.GetString("publicURL")
			auth.ComputeEndpoint = NormalizeURL(auth.ComputeEndpoint)
		}
	}
	return nil
}

func (auth *AuthConfig) makeTokenv2Body() map[string]interface{} {
	if auth.UserID != "" {
		return nil
	}
	if auth.APIKey != "" {
		return nil
	}
	if auth.DomainID != "" {
		return nil
	}
	if auth.DomainName != "" {
		return nil
	}

	// Populate the request map.
	authMap := make(map[string]interface{})

	if auth.Username != "" {
		if auth.Password != "" {
			authMap["passwordCredentials"] = map[string]interface{}{
				"username": auth.Username,
				"password": auth.Password,
			}
		} else {
			return nil
		}
	} else if auth.TokenID != "" {
		authMap["token"] = map[string]interface{}{
			"id": auth.TokenID,
		}
	} else {
		return nil
	}

	if auth.TenantID != "" {
		authMap["tenantId"] = auth.TenantID
	}
	if auth.TenantName != "" {
		authMap["tenantName"] = auth.TenantName
	}

	return map[string]interface{}{"auth": authMap}
}

func NormalizeURL(url string) string {
	if !strings.HasSuffix(url, "/") {
		return url + "/"
	}
	return url
}

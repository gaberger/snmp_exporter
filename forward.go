// Copyright 2020 Forward Networks Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"crypto/tls"
	"strings"
	// "strings"
	// "encoding/json"
	"fmt"
	// "net/url"
	// "strings"

	// "net/url"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-resty/resty/v2"
	"gopkg.in/thedevsaddam/gojsonq.v2"
	// "forwardnetworks.com/snmp_exporter/config"
	// "github.com/go-resty/resty/v2"
	// "github.com/soniah/gosnmp"
)

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func updateList(index int, value string, data []string) []string {
	for k := range data {
		if k == index {
			data[k] = value
		}
	}
	return data
}

// type ResponseStruct struct {
// 	Aliases    []string    `json:"aliases"`
// 	Topo       interface{} `json:"coveredTopo"`
// 	FileLines  interface{} `json:"fileLines"`
// 	Duplex     bool        `json:"fullDuplex"`
// 	IPAddr     []string    `json:"ipAddresses"`
// 	isDown     bool        `json:"isDown"`
// 	macAddress string      `json:"macAddress"`
// 	mtu        string      `json:"mtu"`
// 	name       string      `json:"name"`
// 	speed      string      `json:"speedMbps"`
// 	intfType   string      `json:"type"`
// }

type ResponseStruct []struct {
	Aliases []string `json:"aliases"`
}

func getFwdInterfaceAlias(device string, logger log.Logger) string {
	// var intfEsc = strings.ToLower(url.QueryEscape(intf))
	var server = "dev-vm-9:8443"
	var snapshot = 298
	var user = os.Getenv("FWD_ADMINUSER")
	var password = os.Getenv("FWD_ADMINPASSWORD")
	var urlString = fmt.Sprintf("https://%s/api/snapshots/%d/devices/%s/interfaces", server, snapshot, device)
	var result string

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetHeader("Accept", "application/json")
	client.SetBasicAuth(user, password)
	level.Debug(logger).Log("msg", "Calling Forward API server", server)

	// _, err := client.R().SetResult(&r).Get(urlString)
	resp, err := client.R().Get(urlString)

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		result = ""
	} else {
		result = resp.String()
	}
	return result
}

func getAlias(json string, intf string) []string {
	var r ResponseStruct
	var lowerIntf = strings.ToLower(intf)
	var result []string
	gojsonq.New().FromString(json).From("interfaces").Out(&r)

	for _, v := range r {
		for _, i := range v.Aliases {
			if (lowerIntf == i){
				result = v.Aliases
				break
			}
		}
	}
	// intfResult.Out(&r)
	return result
}

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

// ResponseStruct to capture interface aliases
type ResponseStruct []struct {
	Aliases []string `json:"aliases"`
}

func callForwardAPI(url string, logger log.Logger) (string, error) {
	var server = os.Getenv("FWD_SERVER")
	var user = os.Getenv("FWD_ADMINUSER")
	var password = os.Getenv("FWD_ADMINPASSWORD")

	if len(server) == 0 || len(user) == 0 || len(password) == 0 {
		level.Debug(logger).Log("msg", "FWD Environment not set\n Set: FWD_SERVER, FWD_ADMINUSER, FWD_ADMINPASSWORD")
		return "", fmt.Errorf("FWD ENV not set")
	}

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetHeader("Accept", "application/json")
	client.SetBasicAuth(user, password)

	var fwdURL = fmt.Sprintf("https://%s/%s", server, url)
	level.Debug(logger).Log("msg", "Calling Forward API ", fwdURL)

	resp, err := client.R().Get(fwdURL)
	if err != nil {
		if resp.StatusCode() != 200 {
			return "", fmt.Errorf("response status code: %d", resp.StatusCode())
		}
	}
	return resp.String(), nil
}

func getLatestSnapshot(networkID string, logger log.Logger) (string, error) {
	var url = fmt.Sprintf("api/networks/%s/snapshots/latestProcessed", networkID)

	resp, err := callForwardAPI(url, logger)
	if err != nil {
		level.Error(logger).Log("Somthing went wrong ", err)
		return "", nil
	}
	level.Debug(logger).Log("msg", "Latest Snapshot ", resp)

	result, err := gojsonq.New().FromString(resp).FindR("id")

	level.Debug(logger).Log("RESULT", result)

	if err != nil {
		level.Error(logger).Log(err)
		return "", fmt.Errorf("Could not retrieve snapshot")
	}
	snapshotID, _ := result.String()
	level.Debug(logger).Log("SNAPSHOTID", snapshotID)

	if err != nil {
		return "", fmt.Errorf("Error in HTTP Status: %s", err)
	}
	return snapshotID, nil
}

func getFwdInterfaceAlias(device string, logger log.Logger) (string, error) {
	var network = os.Getenv("FWD_NETWORK")
	var snapshotID string
	level.Debug(logger).Log("msg", "Calling getFwdInterfaceAlias")

	if len(network) == 0 {
		level.Debug(logger).Log("msg", "FWD_NETWORK not set")
		return "", fmt.Errorf("FWD_NETWORK not set")
	}

	respSnapshot, respSnapshotErr := getLatestSnapshot(network, logger)
	if respSnapshotErr != nil {
		return "", fmt.Errorf("Error in HTTP Status: %s", respSnapshotErr)
	}
	snapshotID = respSnapshot

	var url = fmt.Sprintf("api/snapshots/%s/devices/%s/interfaces", snapshotID, device)

	respInterface, respInterfaceErr := callForwardAPI(url, logger)

	if respInterfaceErr != nil {
		return "", fmt.Errorf("Error in HTTP Status: %s", respInterfaceErr)
	}
	return respInterface, nil
}

func getAlias(json string, intf string) []string {
	var r ResponseStruct
	var lowerIntf = strings.ToLower(intf)
	var result []string
	gojsonq.New().FromString(json).From("interfaces").Out(&r)

	for _, v := range r {
		for _, i := range v.Aliases {
			if lowerIntf == i {
				result = v.Aliases
				break
			}
		}
	}
	// intfResult.Out(&r)
	return result
}

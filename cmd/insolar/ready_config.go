/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type ReadyConfigParams struct {
	PrivateKey        string   `json:"private_key"`
	PublicKey         string   `json:"public_key"`
	MajorityRule      uint     `json:"majority_rule"`
	Roles             []string `json:"roles"`
	NumBootstrapNodes uint     `json:"num_bootstrap_nodes"`
	Host              string   `json:"host"`
}

func readConfig(keysPath string) ReadyConfigParams {
	data, err := ioutil.ReadFile(filepath.Clean(keysPath))
	check("[ readConfig ] couldn't read keys from: "+keysPath, err)
	params := ReadyConfigParams{}
	json.Unmarshal(data, &params)

	return params
}

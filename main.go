/*
Copyright 2014 The Kubernetes Authors.

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

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/xyproto/simpleredis/v2"
)

var (
	masterPool  *simpleredis.ConnectionPool
	replicaPool *simpleredis.ConnectionPool
)

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	list := simpleredis.NewList(replicaPool, key)
	members, err := list.GetAll()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	membersJSON, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(membersJSON)
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(masterPool, key)
	if err := list.Add(value); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info, err := masterPool.Get(0).Do("INFO")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	infoBytes, ok := info.([]byte)
	if !ok {
		http.Error(rw, "Unexpected response from Redis", http.StatusInternalServerError)
		return
	}

	rw.Write(infoBytes)
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.SplitN(item, "=", 2)
		if len(splits) == 2 {
			key := splits[0]
			val := splits[1]
			environment[key] = val
		}
	}

	envJSON, err := json.MarshalIndent(environment, "", "  ")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(envJSON)
}

func main() {
	masterPool = simpleredis.NewConnectionPoolHost("redis-master:6379")
	defer masterPool.Close()
	replicaPool = simpleredis.NewConnectionPoolHost("redis-replica:6379")
	defer replicaPool.Close()

	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
	//n.Run(":4000")
}

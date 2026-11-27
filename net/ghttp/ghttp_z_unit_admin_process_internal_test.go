// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Internal tests for passing listener file descriptors to a reloaded process.

package ghttp

import (
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/test/gtest"
)

// Test_bufferToServerFdMap_DottedName checks that server names containing "."
// survive the encode/decode round trip of the reload fd map.
func Test_bufferToServerFdMap_DottedName(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		sfm := map[string]listenerFdMap{
			"api.v1":  {schemeHttp: ":8080#3", schemeHttps: ""},
			"default": {schemeHttp: ":8081#4", schemeHttps: ":8443#5"},
		}
		buffer, err := gjson.Encode(sfm)
		t.AssertNil(err)
		t.Assert(bufferToServerFdMap(buffer), sfm)
	})
}

// Test_Server_Start_ReloadFdByInstanceName checks that a reloaded server picks up
// the inherited listener stored under its instance name, even when its config
// name differs from the instance name.
func Test_Server_Start_ReloadFdByInstanceName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not support socket file descriptor passing")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := ln.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	}()
	file, err := ln.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	}()
	var (
		instanceName = "reload-4166.api"
		inheritPort  = ln.Addr().(*net.TCPAddr).Port
	)
	buffer, err := gjson.Encode(map[string]listenerFdMap{
		instanceName: {
			schemeHttp:  fmt.Sprintf("127.0.0.1:%d#%d", inheritPort, file.Fd()),
			schemeHttps: "",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(adminActionReloadEnvKey, string(buffer))

	s := GetServer(instanceName)
	s.BindHandler("/", func(r *Request) {
		r.Response.Write("ok")
	})
	s.SetDumpRouterMap(false)
	gtest.C(t, func(t *gtest.T) {
		// The config name stays "default" while the instance name is "reload-4166.api".
		t.AssertNE(s.GetName(), instanceName)
		t.AssertNil(s.Start())
		defer func() {
			t.AssertNil(s.Shutdown())
		}()
		time.Sleep(100 * time.Millisecond)
		t.Assert(s.GetListenedPort(), inheritPort)
	})
}

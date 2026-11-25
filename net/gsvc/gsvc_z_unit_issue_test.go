// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gsvc_test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gsvc"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/gutil"
)

// Test_Issue4180 tests that NewServiceWithKV returns an error instead of panicking
// for malformed service keys, such as a service name containing the separator.
// See https://github.com/gogf/gf/issues/4180
func Test_Issue4180(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// A service name containing "/" shifts the version into the endpoints segment.
		s, err := gsvc.NewServiceWithKV(
			"/service/default/default/gogf-grpcx/demo2/latest/192.168.100.1:56586", "",
		)
		t.AssertNil(s)
		t.AssertNE(err, nil)
		t.Assert(gerror.Code(err), gcode.CodeInvalidParameter)
	})
	gtest.C(t, func(t *gtest.T) {
		// Malformed endpoint in a key with the right number of segments.
		s, err := gsvc.NewServiceWithKV(
			"/service/default/default/demo2/latest/192.168.100.1", "",
		)
		t.AssertNil(s)
		t.AssertNE(err, nil)
		t.Assert(gerror.Code(err), gcode.CodeInvalidParameter)
	})
	gtest.C(t, func(t *gtest.T) {
		// A valid key still parses.
		s, err := gsvc.NewServiceWithKV(
			"/service/default/default/demo2/latest/192.168.100.1:56586,192.168.2.15:56586",
			`{"protocol":"grpc"}`,
		)
		t.AssertNil(err)
		t.Assert(s.GetName(), "demo2")
		t.Assert(s.GetVersion(), "latest")
		t.Assert(s.GetEndpoints().String(), "192.168.100.1:56586,192.168.2.15:56586")
		t.Assert(s.GetMetadata().Get(gsvc.MDProtocol), "grpc")
	})
	gtest.C(t, func(t *gtest.T) {
		// NewEndpoint and NewEndpoints keep panicking on invalid addresses.
		t.Assert(gsvc.NewEndpoint("192.168.100.1:56586").String(), "192.168.100.1:56586")
		t.Assert(gsvc.NewEndpoints("192.168.100.1:56586,192.168.2.15:56586").String(), "192.168.100.1:56586,192.168.2.15:56586")
		err := gutil.Try(context.Background(), func(_ context.Context) {
			gsvc.NewEndpoint("latest")
		})
		t.Assert(gerror.Code(err), gcode.CodeInvalidParameter)
		err = gutil.Try(context.Background(), func(_ context.Context) {
			gsvc.NewEndpoints("192.168.100.1:56586,latest")
		})
		t.Assert(gerror.Code(err), gcode.CodeInvalidParameter)
	})
}

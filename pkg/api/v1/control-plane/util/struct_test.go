// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package util_test

import (
	"fmt"
	"reflect"

	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/util"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

var _ = Describe("util tests", func() {
	Context("go-control-plane tests", func() {

		var (
			st *types.Struct
			pb *v2.DiscoveryRequest
		)

		BeforeEach(func() {
			pb = &v2.DiscoveryRequest{
				VersionInfo: "test",
				Node:        &envoy_api_v2_core.Node{Id: "proxy"},
			}
			var err error
			st, err = util.MessageToStruct(pb)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can transform objects", func() {
			pbst := map[string]*types.Value{
				"version_info": &types.Value{Kind: &types.Value_StringValue{StringValue: "test"}},
				"node": &types.Value{Kind: &types.Value_StructValue{StructValue: &types.Struct{
					Fields: map[string]*types.Value{
						"id": &types.Value{Kind: &types.Value_StringValue{StringValue: "proxy"}},
					},
				}}},
			}
			if !reflect.DeepEqual(st.Fields, pbst) {
				Fail(fmt.Sprintf("MessageToStruct(%v) => got %v, want %v", pb, st.Fields, pbst))
			}
		})

		It("can error when improper types re given", func() {
			out := &v2.DiscoveryRequest{}
			err := util.StructToMessage(st, out)
			Expect(err).NotTo(HaveOccurred())
			if !reflect.DeepEqual(pb, out) {
				Fail(fmt.Sprintf("StructToMessage(%v) => got %v, want %v", st, out, pb))
			}
		})

		It("will error when given nil inputs", func() {
			_, err := util.MessageToStruct(nil)
			Expect(err).To(HaveOccurred())

			err = util.StructToMessage(nil, &v2.DiscoveryRequest{})
			Expect(err).To(HaveOccurred())
		})
	})
})

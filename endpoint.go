// Copyright 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net"
)

// endpointInfo caches endpoint information from Cilium which will be useful
// for updating the MTU of connected endpoints.
type endpointInfo struct {
	addrs map[string]struct{}
	links map[string]struct{}
}

// managedIP returns true if the specified IP is a managed IP address.
func (e *endpointInfo) managedIP(ip net.IP) bool {
	return false
}

// managedLink returns true if the link with the specified name is managed by
// Cilium.
func (e *endpointInfo) managedLink(name string) bool {
	return false
}

// endpoints fetches information about endpoints from Cilium. Returns an error
// if Cilium cannot be reached or listing the endpoints fails for any reason.
func getEndpoints() (*endpointInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

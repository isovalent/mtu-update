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

	"github.com/vishvananda/netns"
)

func scanNamespaces() (netns.NsHandle, []netns.NsHandle, error) {
	return netns.None(), nil, fmt.Errorf("not implemented")
}

// updateNamespaces searches for unique namespaces in the current namespace,
// and attempts to update the device and route MTU in those namespaces if
// their primary device IPs can be found in 'epInfo'.
func updateNamespaces(deviceMTU, tunnelMTU int, epInfo *endpointInfo) error {
	return fmt.Errorf("not implemented")
}

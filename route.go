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
	"github.com/vishvananda/netlink"
)

// isDefault returns true if this is a default route, ie its mask is 0.
func isDefault(route *netlink.Route) bool {
	return route.Dst == nil
}

// getDefaultRoutes fetches the default routes (both IPv4 and IPv6).
func getDefaultRoutes() ([]netlink.Route, error) {
	scopedLog.Debug("Listing routes")
	routes, err := netlink.RouteList(nil, 0)
	if err != nil {
		return nil, err
	}

	scopedLog.Debugf("Found %d routes", len(routes))

	result := make([]netlink.Route, 0, 2)
	for _, r := range routes {
		scopedLog.Debugf("  %+v", r)
		if isDefault(&r) {
			scopedLog.Debugf("  => Adding to defaults")
			result = append(result, r)
		}
	}

	return result, nil
}

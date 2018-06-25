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

	"github.com/vishvananda/netlink"
)

// scanLinks finds all links in the current namespace and returns them.
func scanLinks() ([]netlink.Link, error) {
	scopedLog.Debug("Fetching links")

	return nil, fmt.Errorf("not implemented")
}

func updateHostLinks(allLinks []netlink.Link, deviceMTU int, epInfo *endpointInfo) {
	log.Warn("Updating host links is not implemented")
}

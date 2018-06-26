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
	"net"

	"github.com/cilium/cilium/api/v1/models"
	clientPkg "github.com/cilium/cilium/pkg/client"
)

// endpointInfo caches endpoint information from Cilium which will be useful
// for updating the MTU of connected endpoints.
type endpointInfo struct {
	addrs map[string]struct{}
	links map[string]struct{}
}

// managedIP returns true if the specified IP is a managed IP address.
func (e *endpointInfo) managedIP(ip net.IP) bool {
	_, ok := e.addrs[ip.String()]
	return ok
}

// addIP attempts to insert the specified address into the endpoint info
// structure. If the specified address is a valid IPv4 or IPv6 address, inserts
// it and returns true. Otherwise, returns false.
func (e *endpointInfo) addIP(addr string) bool {
	ip := net.ParseIP(addr)
	if ip != nil {
		e.addrs[addr] = struct{}{}
		return true
	}
	return false
}

// managedLink returns true if the link with the specified name is managed by
// Cilium.
func (e *endpointInfo) managedLink(name string) bool {
	_, ok := e.links[name]
	return ok
}

// endpointInvalid returns true if the information we need from the provided
// Endpoint object is missing.
func endpointInvalid(ep *models.Endpoint) bool {
	return ep == nil || ep.Status == nil || ep.Status.Networking == nil ||
		len(ep.Status.Networking.Addressing) < 1 ||
		ep.Status.Networking.InterfaceName == ""
}

// newEndpointInfoFromEndpoints creates a new endpointInfo structure from the
// specified endpoint models. Logs errors if any endpoinst are invalid.
func newEndpointInfoFromEndpoints(eps []*models.Endpoint) *endpointInfo {
	result := &endpointInfo{
		addrs: make(map[string]struct{}, len(eps)),
		links: make(map[string]struct{}, len(eps)),
	}
	for _, ep := range eps {
		if endpointInvalid(ep) {
			log.Warnf("Found EP with invalid model: %+v", ep)
			continue
		}
		netConfig := ep.Status.Networking
		for _, addr := range netConfig.Addressing {
			if !result.addIP(addr.IPV4) {
				log.Warnf("Skipping invalid IP %s", addr.IPV4)
			}
			if !result.addIP(addr.IPV6) {
				log.Warnf("Skipping invalid IP %s", addr.IPV6)
			}
		}
		result.links[netConfig.InterfaceName] = struct{}{}
	}

	return result
}

// endpoints fetches information about endpoints from Cilium. Returns an error
// if Cilium cannot be reached or listing the endpoints fails for any reason.
func getEndpoints() (*endpointInfo, error) {
	client, err := clientPkg.NewClient("")
	if err != nil {
		return nil, err
	}

	eps, err := client.EndpointList()
	if err != nil {
		return nil, err
	}

	return newEndpointInfoFromEndpoints(eps), nil
}

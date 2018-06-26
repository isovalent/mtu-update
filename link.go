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
	"strings"

	"github.com/vishvananda/netlink"
)

const (
	innerLinkType = "veth"
)

// linkInfo caches relevant information from the link / address probe for a
// particular link on the system.
type linkInfo struct {
	netlink.Link
	Addrs []netlink.Addr
}

func newLinkInfo(link netlink.Link, addrs []netlink.Addr) *linkInfo {
	return &linkInfo{
		Link:  link,
		Addrs: addrs,
	}
}

// scanLinks finds all links in the current namespace and returns them.
func scanLinks() ([]netlink.Link, error) {
	scopedLog.Debug("Fetching links")

	allLinks, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	scopedLog.Debugf("Found %d links", len(allLinks))
	for _, link := range allLinks {
		attrs := link.Attrs()
		scopedLog.Debugf("  %s", attrs.Name)
		scopedLog.Debugf("    Type: %s", link.Type())
		switch link.Type() {
		case "device":
			scopedLog.Debugf("    MTU: %d", attrs.MTU)
		case "veth":
			scopedLog.Debugf("    MTU: %d", attrs.MTU)
		}
	}

	return allLinks, nil
}

// getPrimaryLink fetches the primary link in the current namespace - ie link
// with the first ifindex (after loopback).
func getPrimaryLink() (*linkInfo, error) {
	links, err := scanLinks()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if link.Type() == innerLinkType {
			addrs, err := netlink.AddrList(link, netlink.FAMILY_V6)
			if err != nil {
				scopedLog.Infof("Failed to fetch address info for link %s", link.Attrs().Name)
				continue
			}
			return newLinkInfo(link, addrs), nil
		}
	}

	return nil, fmt.Errorf("failed to find primary link in %+v", links)
}

func updateHostLinks(allLinks []netlink.Link, deviceMTU int, epInfo *endpointInfo) {
	log.Debug("Updating host namespace devices")
	var (
		skipped int
		failed  int
		updated int
	)

	// First, set all of the veths to allow reception of larger MTU.
	ciliumLinks := make([]netlink.Link, 0, 4)
	for _, link := range allLinks {
		name := link.Attrs().Name
		if link.Attrs().MTU == deviceMTU {
			log.Debugf("Device %s has desired MTU", name)
			skipped++
			continue
		}
		if strings.HasPrefix(name, "cilium") {
			// Don't count; just add to ciliumLinks
			ciliumLinks = append(ciliumLinks, link)
		} else if epInfo.managedLink(name) {
			log.Debugf("Updating MTU for device %s", name)
			if err := netlink.LinkSetMTU(link, deviceMTU); err == nil {
				updated++
			} else {
				failed++
				log.WithError(err).Warn(
					"Failed to set link MTU for %s", name)
			}
		} else {
			skipped++
		}
	}

	// Next, set all of the cilium devices to allow transmit of larger MTU.
	for _, link := range ciliumLinks {
		name := link.Attrs().Name
		if link.Attrs().MTU == deviceMTU {
			log.Debugf("Device %s has desired MTU", name)
			skipped++
			continue
		}
		log.Debugf("Updating MTU for device %s", name)
		if err := netlink.LinkSetMTU(link, deviceMTU); err == nil {
			updated++
		} else {
			failed++
			log.WithError(err).Warnf(
				"Failed to set link MTU for %s", name)
		}
	}

	log.Infof("Updated %d/%d local devices, %d skipped, %d failed",
		updated, len(allLinks), skipped, failed)
}

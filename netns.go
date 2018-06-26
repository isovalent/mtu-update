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
	"path/filepath"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// inodeFromHandle gets a unique identifier (in the form of the inode) from a
// given netns handle.
func inodeFromHandle(nsHandle netns.NsHandle) (uint64, error) {
	var statInfo syscall.Stat_t
	if err := syscall.Fstat(int(nsHandle), &statInfo); err != nil {
		return 0, fmt.Errorf("stat failed: %s", err)
	}

	return statInfo.Ino, nil
}

// scanNamespaces finds a list of all network namespaces reachable from the
// current namespace. Returns a handle to the current namespace and a slice of
// child namespace handles which does not include the current namespace.
//
// The caller must eventually call Close() on every NsHandle returned here.
func scanNamespaces() (netns.NsHandle, []netns.NsHandle, error) {
	log.Debug("Fetching list of network namespaces")

	rootNS, err := netns.Get()
	if err != nil {
		return netns.None(), nil, err
	}

	rootInode, err := inodeFromHandle(rootNS)
	if err != nil {
		return netns.None(), nil,
			fmt.Errorf("Failed to get host netns: %s", err)
	}

	paths, err := filepath.Glob("/proc/*/ns/net")
	if err != nil {
		return netns.None(), nil, err
	}

	// Use a map as a set so each netns is only found once.
	namespaces := make(map[uint64]netns.NsHandle)
	for _, path := range paths {
		nsHandle, err := netns.GetFromPath(path)
		if err != nil {
			log.WithError(err).WithField("path", path).Warn(
				"Failed to fetch netns")
			continue
		}

		inode, err := inodeFromHandle(nsHandle)
		if err != nil {
			log.WithError(err).WithField("path", path).Warn(
				"Failed to get netns inode")
			continue
		}

		if _, ok := namespaces[inode]; ok {
			// If duplicate, close this copy of the open nsHandle.
			nsHandle.Close()
		} else {
			namespaces[inode] = nsHandle
		}
	}
	delete(namespaces, rootInode)

	// Convert the map to an easily iterable slice.
	result := make([]netns.NsHandle, 0, len(namespaces))
	for _, nsHandle := range namespaces {
		result = append(result, nsHandle)
	}

	return rootNS, result, nil
}

// updateNamespaceMTU attempts to update the MTU of routes and links within
// the current namespace, and returns true if the MTU was updated.
// Returns false if the update was skipped or unsuccessful.
func updateNamespaceMTU(deviceMTU, tunnelMTU int, epInfo *endpointInfo) (bool, error) {
	managed := false
	link, err := getPrimaryLink()
	if err != nil {
		return false, fmt.Errorf("Failed to find primary link: %s", err)
	}
	scopedLog.Debugf("Determining whether the link %s is managed",
		link.Attrs().Name)
	for _, addr := range link.Addrs {
		scopedLog.Debugf("  Looking at address %s", addr)
		if epInfo.managedIP(addr.IP) {
			managed = true
			break
		}
	}

	// Skip if Cilium doesn't manage the addresses or the MTU is correct.
	if !managed {
		scopedLog.Debugf("No match for addrs in link %+v, skipping", link)
		return false, nil
	}
	if link.Attrs().MTU == deviceMTU {
		scopedLog.Debugf("Device MTU matches desired MTU, skipping")
		return false, nil
	}

	// Update routes
	routes, err := getDefaultRoutes()
	if err != nil || len(routes) < 1 {
		return false, fmt.Errorf("Failed to fetch routes: %s", err)
	}
	for _, r := range routes {
		r.MTU = tunnelMTU
		err = netlink.RouteReplace(&r)
		if err == nil {
			log.WithField("route", r).Debugf("Updated MTU")
		} else {
			return false, fmt.Errorf(
				"Failed to set route MTU for %s: %s", r, err)
		}
	}

	// Update link
	err = netlink.LinkSetMTU(link.Link, deviceMTU)
	if err == nil {
		log.WithField("link", link.Link.Attrs().Name).Debugf("Updated MTU")
	} else {
		return false, fmt.Errorf("Failed to set link MTU for %s: %s",
			link.Attrs().Name, err)
	}

	return true, nil
}

// updateNamespaces searches for unique namespaces in the current namespace,
// and attempts to update the device and route MTU in those namespaces if
// their primary device IPs can be found in 'epInfo'.
//
// Returns an error only if an error occurs while fetching namespaces.
func updateNamespaces(deviceMTU, tunnelMTU int, epInfo *endpointInfo) error {
	var (
		skipped int
		failed  int
		updated int
	)

	rootNamespace, namespaces, err := scanNamespaces()
	if err != nil {
		return err
	}

	// Set routes and device MTUs inside the network namespaces
	for _, nsHandle := range namespaces {
		log.Debugf("Moving to netns %d", nsHandle)
		scopedLog = log.WithField("netns", nsHandle)

		err = netns.Set(nsHandle)
		if err != nil {
			failed++
			log.WithError(err).Warn("Failed to set netns")
			continue
		}

		ok, err := updateNamespaceMTU(deviceMTU, tunnelMTU, epInfo)
		if err != nil {
			failed++
			scopedLog.WithError(err).Warn("Failed to update MTU")
			continue
		}

		if ok {
			updated++
		} else {
			skipped++
		}
	}

	// Return from whence we came, and clean up after ourselves.
	log.Debugf("Moving to netns %d", rootNamespace)
	netns.Set(rootNamespace)
	rootNamespace.Close()
	for _, nsHandle := range namespaces {
		nsHandle.Close()
	}

	log.Infof("Updated %d/%d namespaces, %d skipped, %d failed",
		updated, len(namespaces), skipped, failed)

	return nil
}

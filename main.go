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
	"os"

	pkgMTU "github.com/cilium/cilium/pkg/mtu"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
)

const (
	autodetectMTU = 0
)

var (
	rootCmd = &cobra.Command{
		Use:   "mtu-update",
		Short: "Update the MTU inside network namespaces.",
		Run: func(cmd *cobra.Command, args []string) {
			run(cmd)
		},
	}

	// deviceMTU governs the MTU to be configured on devices. If 0, it will
	// be autodetected based on the lowest MTU of available devices.
	deviceMTU int

	// tunnelOverhead is the overhead used for configuring routes to
	// remote nodes.
	tunnelOverhead int

	// verbose will cause debug messages to be printed if true.
	verbose bool

	log       *logrus.Logger
	scopedLog *logrus.Entry
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	flags := rootCmd.Flags()
	flags.IntVarP(&deviceMTU, "mtu", "m", pkgMTU.EthernetMTU,
		"Base MTU to configure on links (0 for autodetect)")
	flags.IntVarP(&tunnelOverhead, "tunnel-overhead", "t", pkgMTU.TunnelOverhead,
		"Expected tunnel overhead for overlay traffic")
	flags.BoolVarP(&verbose, "verbose", "v", false,
		"Print verbose debug log messages")
	viper.BindPFlags(flags)

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	log = logrus.StandardLogger()
	scopedLog = logrus.NewEntry(log)
}

// detectMTU searches through the provided list of links for real devices and
// returns the lowest MTU amongst the specified devices. If no real devices
// can be found, returns an error.
func detectMTU(links []netlink.Link) (int, error) {
	log.Debug("Autodetecting MTU")
	return 0, fmt.Errorf("not implemented")
}

// sanitizeMTU takes the specified MTU and an optional set of links, and
// validates the MTU configuration. If the MTU is not specified, autodetects
// the value to be used.
//
// Returns the desired device MTU, MTU for tunnelled routes, and optional error.
func sanitizeMTU(mtu int, links []netlink.Link) (int, int, error) {
	var err error
	if mtu == autodetectMTU {
		mtu, err = detectMTU(links)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to autodetect MTU: %s", err)
		}
	}

	// All hosts must be able to receive 576B datagrams (RFC791).
	if mtu < 576 {
		return 0, 0, fmt.Errorf("MTU %d is too short", mtu)
	}

	// Maximum Geneve tunnel overhead is 310B (draft-ietf-nvo3-geneve-06).
	if tunnelOverhead < 0 || tunnelOverhead > 310 {
		return 0, 0, fmt.Errorf("invalid tunnel overhead %d",
			tunnelOverhead)
	}
	tunnelMTU := mtu - tunnelOverhead

	return mtu, tunnelMTU, nil
}

func run(cmd *cobra.Command) {
	var tunnelMTU int

	if verbose {
		log.Level = logrus.DebugLevel
	}

	epInfo, err := getEndpoints()
	if err != nil {
		log.WithError(err).Fatalf("Failed to fetch Cilium endpoints")
	}

	allLinks, err := scanLinks()
	if err != nil {
		log.WithError(err).Fatalf("Failed to scan available links")
	}

	deviceMTU, tunnelMTU, err = sanitizeMTU(deviceMTU, allLinks)
	if err != nil {
		log.WithError(err).Fatalf("Invalid MTU specified")
	}

	log.Infof("Configuring MTU using base MTU %d, tunnel MTU %d",
		deviceMTU, tunnelMTU)

	// Perform the actual MTU update
	err = updateNamespaces(deviceMTU, tunnelMTU, epInfo)
	if err != nil {
		log.WithError(err).Fatalf("Failed to find network namespaces")
	}
	updateHostLinks(allLinks, deviceMTU, epInfo)
}

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-cli/pkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
)

// OCIDiscoveryForCentralRepo is an artifact discovery endpoint utilizing an OCI image
// which contains an SQLite database listing all available plugins.
type OCIDiscoveryForCentralRepo struct {
	// name is a name of the discovery
	name string
	// image is an OCI compliant image. Which include DNS-compatible registry name,
	// a valid URI path(MAY contain zero or more ‘/’) and a valid tag
	// E.g., harbor.my-domain.local/tanzu-cli/plugins-manifest:latest
	// Contains a single SQLite database file.
	image string
}

// NewOCIDiscoveryForCentralRepo returns a new Discovery targeting the
// format of the Central Repository.
func NewOCIDiscoveryForCentralRepo(name, image string) Discovery {
	return &OCIDiscoveryForCentralRepo{
		name:  name,
		image: image,
	}
}

// List available plugins.
func (od *OCIDiscoveryForCentralRepo) List() (plugins []Discovered, err error) {
	return od.Manifest()
}

// Describe a plugin.
func (od *OCIDiscoveryForCentralRepo) Describe(name string) (p Discovered, err error) {
	plugins, err := od.Manifest()
	if err != nil {
		return
	}

	for i := range plugins {
		if plugins[i].Name == name {
			p = plugins[i]
			return
		}
	}
	err = errors.Errorf("cannot find plugin with name '%v'", name)
	return
}

// Name of the repository.
func (od *OCIDiscoveryForCentralRepo) Name() string {
	return od.name
}

// Type of the discovery.
func (od *OCIDiscoveryForCentralRepo) Type() string {
	return common.DiscoveryTypeOCI
}

// Manifest returns the manifest for a local repository.
func (od *OCIDiscoveryForCentralRepo) Manifest() ([]Discovered, error) {
	err := carvelhelpers.DownloadDBImageToCache(od.image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database file from discovery")
	}

	return nil, nil //processDiscoveryManifestData2(outputData, od.name)
}

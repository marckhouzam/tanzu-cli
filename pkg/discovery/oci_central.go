// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"

	"github.com/pkg/errors"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"

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
	//pkgDir
	_, err := carvelhelpers.DownloadImageBundleAndSaveFilesToTempDir(od.image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database file from discovery")
	}
	// defer os.RemoveAll(pkgDir)

	return nil, nil //processDiscoveryManifestData2(outputData, od.name)
}

func processDiscoveryManifestData2(data []byte, discoveryName string) ([]Discovered, error) {
	plugins := make([]Discovered, 0)

	for _, resourceYAML := range strings.Split(string(data), "---") {
		scheme, err := cliv1alpha1.SchemeBuilder.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create scheme")
		}
		s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
			apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
		var p cliv1alpha1.CLIPlugin
		_, _, err = s.Decode([]byte(resourceYAML), nil, &p)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode discovery manifests")
		}

		dp, err := DiscoveredFromK8sV1alpha1(&p)
		if err != nil {
			return nil, err
		}
		dp.Source = discoveryName
		dp.DiscoveryType = common.DiscoveryTypeOCI
		if dp.Name != "" {
			plugins = append(plugins, dp)
		}
	}
	return plugins, nil
}

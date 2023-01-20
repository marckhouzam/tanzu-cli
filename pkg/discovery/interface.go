// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package discovery is implements discovery interface for plugin discovery
// Discovery is the interface to fetch the list of available plugins, their
// supported versions and how to download them either stand-alone or scoped to a server.
// A separate interface for discovery helps to decouple discovery (which is usually
// tied to a server or user identity) from distribution (which can be shared).
package discovery

import (
	"errors"

	"github.com/vmware-tanzu/tanzu-cli/pkg/constants"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-plugin-runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config"
)

// Discovery is the interface to fetch the list of available plugins
type Discovery interface {
	// Name of the repository.
	Name() string

	// List available plugins.
	List() ([]Discovered, error)

	// Describe a plugin.
	Describe(name string) (Discovered, error)

	// Type returns type of discovery.
	Type() string
}

type Discovery2 interface {
	Discovery
	// List all available versions of the specified plugin
	ListVersions(name string, target cliv1alpha1.Target) ([]*Discovered, error)
}

// CreateDiscoveryFromV1alpha1 creates discovery interface from v1alpha1 API
func CreateDiscoveryFromV1alpha1(pd configapi.PluginDiscovery) (Discovery, error) {
	switch {
	case pd.OCI != nil:
		if config.IsFeatureActivated(constants.FeatureCentralRepository) {
			return NewOCIDiscoveryForCentralRepo(pd.OCI.Name, pd.OCI.Image), nil
		}
		return NewOCIDiscovery(pd.OCI.Name, pd.OCI.Image), nil
	case pd.Local != nil:
		return NewLocalDiscovery(pd.Local.Name, pd.Local.Path), nil
	case pd.Kubernetes != nil:
		return NewKubernetesDiscovery(pd.Kubernetes.Name, pd.Kubernetes.Path, pd.Kubernetes.Context), nil
	case pd.REST != nil:
		return NewRESTDiscovery(pd.REST.Name, pd.REST.Endpoint, pd.REST.BasePath), nil
	}
	return nil, errors.New("unknown plugin discovery source")
}

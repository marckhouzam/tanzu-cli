// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package centralconfig implements an interface to deal with the central configuration.
package centralconfig

import (
	"path/filepath"

	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
)

// CentralConfigFileName is the name of the central config file
const CentralConfigFileName = "central_config.yaml"

// CentralConfigEntryKey represents the key of a central configuration entry.
type CentralConfigEntryKey struct {
	Key string
}

// CentralConfigEntryValue represents the value of a central configuration entry.
type CentralConfigEntryValue struct {
	Value interface{}
}

// CentralConfig is used to interact with the central configuration.
type CentralConfig interface {
	GetCentralConfigEntry(key CentralConfigEntryKey) *CentralConfigEntryValue
}

// NewCentralConfigReader returns a CentralConfig reader that can
// be used to read central configuration values.
func NewCentralConfigReader(pd *types.PluginDiscovery) CentralConfig {
	// The central config is stored in the cache
	centralConfigFile := filepath.Join(common.DefaultCacheDir, common.PluginInventoryDirName, pd.OCI.Name, CentralConfigFileName)

	return &centralConfigYamlReader{configFile: centralConfigFile}
}

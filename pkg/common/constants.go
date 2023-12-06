// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package common defines generic constants and structs
package common

// Plugin status and scope constants
const (
	PluginStatusInstalled            = "installed"
	PluginStatusNotInstalled         = "not installed"
	PluginStatusUpdateAvailable      = "update available"
	PluginStatusReadyForInstallation = "ready for installation"
	PluginScopeStandalone            = "Standalone"
	PluginScopeContext               = "Context"
)

// DiscoveryType constants
const (
	DiscoveryTypeOCI        = "oci"
	DiscoveryTypeLocal      = "local"
	DiscoveryTypeGCP        = "gcp"
	DiscoveryTypeKubernetes = "kubernetes"
	DiscoveryTypeREST       = "rest"
)

// DistributionType constants
const (
	DistributionTypeOCI   = "oci"
	DistributionTypeLocal = "local"
)

// Shared strings
const (
	TargetList = "kubernetes[k8s]/mission-control[tmc]/global"
)

// CoreName is the name of the core binary.
const CoreName = "core"

// CommandTypePlugin represents the command type is plugin
const CommandTypePlugin = "plugin"

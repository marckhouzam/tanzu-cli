// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package publisher
package publisher

type PublisherPluginAssociation struct {
	Publisher   string   `json:"publisher" yaml:"publisher"`
	Vendor      string   `json:"vendor" yaml:"vendor"`
	Description string   `json:"description" yaml:"description"`
	Plugins     []Plugin `json:"plugins" yaml:"plugins"`
}

type Plugin struct {
	Name   string `json:"name" yaml:"name"`
	Target string `json:"target" yaml:"target"`
}

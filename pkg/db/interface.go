// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package db

// PluginInventoryRow Structure of each row of the SQLite database
type PluginInventoryRow struct {
	Name               string
	Target             string
	RecommendedVersion string
	Version            string
	Hidden             string
	Description        string
	Publisher          string
	Vendor             string
	OS                 string
	Arch               string
	Digest             string
	URI                string
}

type DB interface {
	ListPluginsRows() []PluginInventoryRow

	InsertPluginRow(row PluginInventoryRow) error
}

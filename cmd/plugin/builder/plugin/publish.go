// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-cli/cmd/plugin/builder/types"
	"gopkg.in/yaml.v3"
	"os"
)

type PublisherOptions struct {
	ArtifactDir        string
	Publisher          string
	Vendor             string
	Repository         string
	PluginManifestFile string
}

type pluginArtifacts struct {
}

type PublisherImpl interface {
	PublishPlugins() error
}

func (po *PublisherOptions) PublishPlugins() error {
	log.Infof("Starting plugin publishing process...")
	log.Infof("Using plugin location: %q, Publisher: %q, Vendor: %q, Repository: %q", po.ArtifactDir, po.Publisher, po.Vendor, po.Repository)
	return nil
}

func (po *PublisherOptions) verifyArtifacts() error {
	log.Info("Verifying artifacts...")

	data, err := os.ReadFile(po.PluginManifestFile)
	if err != nil {
		return errors.Wrap(err, "fail to read the plugin manifest file")
	}

	pluginManifest := &types.PluginManifest{}
	err = yaml.Unmarshal(data, pluginManifest)
	if err != nil {
		return errors.Wrap(err, "fail to read the plugin manifest file")
	}

	//for i := range pluginManifest.Plugins {
	//}

	return nil
}

func (po *PublisherOptions) verifyPluginArtifactLocation(pm types.PluginMetadata) error {

	//// verify that plugin binary exists within provided artifacts directory
	//for _, os := range types.SupportedOS {
	//	for _, arch := range types.SupportedArch {
	//		pluginFilePath := filepath.Join(po.ArtifactDir, os, arch, "cli", pm.Name)
	//		_, error := os.Stat()
	//
	//	}
	//}

	return nil
}

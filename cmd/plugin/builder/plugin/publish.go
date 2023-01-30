// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-cli/pkg/cli"
	"github.com/vmware-tanzu/tanzu-cli/pkg/publisher"
	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
	"gopkg.in/yaml.v3"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"os"
	"path/filepath"
)

const PublisherPluginAssociationURL = "https://gist.githubusercontent.com/anujc25/894c5187b7d1da25490139fe54c1e73f/raw/b0b8f15e70110b857b3b75d0ce4f7ab04b5782c3"

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

	if po.PluginManifestFile == "" {
		po.PluginManifestFile = filepath.Join(po.ArtifactDir, cli.PluginManifestFileName)
	}

	pluginManifest, err := po.getPluginManifest()
	if err != nil {
		return err
	}

	log.Infof("Using plugin location: %q, Publisher: %q, Vendor: %q, Repository: %q, PluginManifest: %q",
		po.ArtifactDir, po.Publisher, po.Vendor, po.Repository, po.PluginManifestFile)

	log.Info("Verifying plugin artifacts...")
	if err := po.verifyPluginArtifacts(pluginManifest); err != nil {
		return errors.Wrap(err, "error while verifying artifacts")
	}
	log.Info("Successfully verified plugin artifacts")

	log.Info("Verifying plugin and publisher association...")
	if err := po.verifyPluginAndPublisherAssociation(pluginManifest); err != nil {
		return errors.Wrap(err, "error while verifying artifacts")
	}
	log.Info("Successfully verified plugin and publisher association")

	return nil
}

func (po *PublisherOptions) verifyPluginArtifacts(pluginManifest *cli.Manifest) error {
	var errList []error
	for i := range pluginManifest.Plugins {
		for _, osArch := range cli.MinOSArch {
			for _, version := range pluginManifest.Plugins[i].Versions {
				pluginFilePath := filepath.Join(po.ArtifactDir, osArch.OS(), osArch.Arch(),
					pluginManifest.Plugins[i].Target, pluginManifest.Plugins[i].Name, version,
					cli.MakeArtifactName(pluginManifest.Plugins[i].Name, osArch))

				if !utils.PathExists(pluginFilePath) {
					errList = append(errList, errors.Errorf("unable to verify artifacts for "+
						"plugin: %q, target: %q, osArch: %q, version: %q. File %q doesn't exist",
						pluginManifest.Plugins[i].Name, pluginManifest.Plugins[i].Target, osArch.String(), version, pluginFilePath))
				}
			}
		}
	}
	return kerrors.NewAggregate(errList)
}

func (po *PublisherOptions) verifyPluginAndPublisherAssociation(pluginManifest *cli.Manifest) error {
	f, err := os.CreateTemp("", "*.yaml")
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s-%s.yaml", PublisherPluginAssociationURL, po.Vendor, po.Publisher)
	log.Infof("Using url: %s", url)
	err = utils.DownloadFile(f.Name(), url)
	if err != nil {
		return errors.Wrapf(err, "error while downloading plugin publisher association file %q", url)
	}
	b, err := os.ReadFile(f.Name())
	if err != nil {
		return errors.Wrapf(err, "error while reading downloaded plugin publisher association file %q", f.Name())
	}

	registeredPluginsForPublisher := &publisher.PublisherPluginAssociation{}
	err = yaml.Unmarshal(b, registeredPluginsForPublisher)
	if err != nil {
		return errors.Wrapf(err, "error while unmarshaling downloaded plugin publisher association file %q", f.Name())
	}

	var errList []error
	for i := range pluginManifest.Plugins {
		found := false
		for j := range registeredPluginsForPublisher.Plugins {
			if pluginManifest.Plugins[i].Name == registeredPluginsForPublisher.Plugins[j].Name &&
				pluginManifest.Plugins[i].Target == registeredPluginsForPublisher.Plugins[j].Target {
				found = true
			}
		}
		if !found {
			errList = append(errList, errors.Errorf("plugin: %q with target: %q is not registered for vendor: %q, publisher: %q",
				pluginManifest.Plugins[i].Name, pluginManifest.Plugins[i].Target, po.Vendor, po.Publisher))
		}
	}
	return kerrors.NewAggregate(errList)
}

func (po *PublisherOptions) getPluginManifest() (*cli.Manifest, error) {
	data, err := os.ReadFile(po.PluginManifestFile)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read the plugin manifest file")
	}

	pluginManifest := &cli.Manifest{}
	err = yaml.Unmarshal(data, pluginManifest)
	if err != nil {
		return nil, errors.Wrap(err, "fail to read the plugin manifest file")
	}
	return pluginManifest, nil
}

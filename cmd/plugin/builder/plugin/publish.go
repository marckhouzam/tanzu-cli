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
	// Name is the name of the plugin.
	Name string

	// Target is the name of the plugin.
	Target string

	// Description is the plugin's description.
	Description string

	// Versions available for plugin.
	VersionArtifactMap map[string][]artifactMetadata
}

type artifactMetadata struct {
	// OS is the name of the os.
	OS string
	// Arch is the name of the arch.
	Arch string
	// Path is plugin binary path from where we need to publish the plugin
	Path string
	// RelativeURI is relative path within the repository for the plugins
	RelativeURI string
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

	mapPluginArtifacts, err := po.createTempArtifactsDirForPublishing(pluginManifest)
	if err != nil {
		return errors.Wrapf(err, "unable to create temp artifacts directory for publishing")
	}

	b, err := yaml.Marshal(mapPluginArtifacts)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal mapPluginArtifacts")
	}

	log.Info(string(b))

	po.publishPluginsFromPluginArtifacts(mapPluginArtifacts)

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

func (po *PublisherOptions) createTempArtifactsDirForPublishing(pluginManifest *cli.Manifest) (map[string]pluginArtifacts, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}

	mapPluginArtifacts := make(map[string]pluginArtifacts)
	for i := range pluginManifest.Plugins {
		for _, osArch := range cli.AllOSArch {
			for _, version := range pluginManifest.Plugins[i].Versions {
				locationWithinBaseDir := filepath.Join(osArch.OS(), osArch.Arch(),
					pluginManifest.Plugins[i].Target, pluginManifest.Plugins[i].Name, version,
					cli.MakeArtifactName(pluginManifest.Plugins[i].Name, osArch))

				pluginFilePath := filepath.Join(po.ArtifactDir, locationWithinBaseDir)
				tmpPluginFilePath := filepath.Join(tmpDir, locationWithinBaseDir)

				if !utils.PathExists(pluginFilePath) {
					continue
				}

				err := utils.CopyFile(pluginFilePath, tmpPluginFilePath)
				if err != nil {
					return nil, err
				}

				key := fmt.Sprintf("%s-%s", pluginManifest.Plugins[i].Target, pluginManifest.Plugins[i].Name)
				pa, exists := mapPluginArtifacts[key]
				if !exists {
					pa = pluginArtifacts{
						Name:               pluginManifest.Plugins[i].Name,
						Target:             pluginManifest.Plugins[i].Target,
						Description:        pluginManifest.Plugins[i].Description,
						VersionArtifactMap: make(map[string][]artifactMetadata),
					}
					mapPluginArtifacts[key] = pa
				}
				_, exists = pa.VersionArtifactMap[version]
				if !exists {
					pa.VersionArtifactMap[version] = make([]artifactMetadata, 0)
				}
				am := artifactMetadata{
					OS:   osArch.OS(),
					Arch: osArch.Arch(),
					Path: tmpPluginFilePath,
					RelativeURI: fmt.Sprintf("%s/%s/%s/%s:%s", osArch.OS(), osArch.Arch(),
						pluginManifest.Plugins[i].Target, pluginManifest.Plugins[i].Name, version),
				}
				pa.VersionArtifactMap[version] = append(pa.VersionArtifactMap[version], am)
			}
		}
	}
	return mapPluginArtifacts, nil
}

func (po *PublisherOptions) publishPluginsFromPluginArtifacts(mapPluginArtifacts map[string]pluginArtifacts) error {
	baseRepository := fmt.Sprintf("%s/%s/%s", po.Repository, po.Vendor, po.Publisher)
	for _, pa := range mapPluginArtifacts {
		for _, artifacts := range pa.VersionArtifactMap {
			for _, a := range artifacts {
				pluginImage := fmt.Sprintf("%s/%s", baseRepository, a.RelativeURI)
				log.Infof("imgpkg push -i %s -f %s", pluginImage, filepath.Dir(a.Path))
			}
		}
	}
	return nil
}

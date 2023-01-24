// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-cli/pkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-cli/pkg/catalog"
	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-cli/pkg/distribution"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
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

type centralRepoRow struct {
	name               string
	target             string
	recommendedVersion string
	version            string
	hidden             string
	description        string
	publisher          string
	vendor             string
	os                 string
	arch               string
	digest             string
	uri                string
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
	pluginDBDir := filepath.Join(common.DefaultCacheDir, "plugin_db")
	err := carvelhelpers.DownloadDBImage(od.image, pluginDBDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database file from discovery")
	}

	return od.getPluginsFromDB(pluginDBDir)
}

func (od *OCIDiscoveryForCentralRepo) getPluginsFromDB(dbDir string) ([]Discovered, error) {
	dbLocation := filepath.Join(dbDir, common.CentralRepoDBFileName)
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// We need to order the results properly because the logic below which converts from
	// rows to the Discovered type expects an ordering of PluginName, then Target, then Version.
	rows, err := db.Query("SELECT * FROM PluginBinaries ORDER BY PluginName,Target,Version")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// The central repository uses relative image URIs to be future-proof.
	// Determine the image prefix from the Central Repository main image.
	// If the main image is at project.registry.vmware.com/tanzu-cli/plugins/plugin_database:latest
	// then the image prefix should be project.registry.vmware.com/tanzu-cli/plugins/
	imagePrefix := path.Dir(od.image)

	currentPluginID := ""
	currentVersion := ""
	var currentPlugin *Discovered
	allPlugins := []Discovered{}
	var artifactList distribution.ArtifactList
	var artifacts distribution.Artifacts

	for rows.Next() {
		row, err := getNextRow(rows)
		if err != nil {
			return allPlugins, err
		}

		target := convertTargetFromDB(row.target)
		pluginIdFromRow := catalog.PluginNameTarget(row.name, target)
		if currentPluginID != pluginIdFromRow {
			// Found a new plugin.
			// Store the current one in the array and prepare the new one.
			if currentPlugin != nil {
				artifacts[currentVersion] = artifactList
				artifactList = distribution.ArtifactList{}
				currentPlugin.Distribution = artifacts
				allPlugins = appendPlugin(allPlugins, currentPlugin)
			}
			currentPluginID = pluginIdFromRow

			currentPlugin = &Discovered{
				Name:               row.name,
				Description:        row.description,
				RecommendedVersion: row.recommendedVersion,
				InstalledVersion:   "", // TODO(khouzam)
				SupportedVersions:  []string{},
				Optional:           false,
				Scope:              common.PluginScopeStandalone,
				Source:             "Central Repository",
				ContextName:        "", // TODO(khouzam) this is used when creating the cache.  Concept needs updating
				DiscoveryType:      common.DiscoveryTypeOCI,
				Target:             target,
				Status:             common.PluginStatusNotInstalled,
			}
			currentVersion = ""
			artifacts = distribution.Artifacts{}
		}

		// Check if we have a new version
		if currentVersion != row.version {
			// This is a new version of our current plugin.  Add it to the array of versions.
			// We can do this without verifying if the version is already there because
			// we have requested the list of plugins from the database ordered by version.
			currentPlugin.SupportedVersions = append(currentPlugin.SupportedVersions, row.version)

			// Also store the list of artifacts for the previous version then start building
			// the artifact list for the new version
			if currentVersion != "" {
				artifacts[currentVersion] = artifactList
				artifactList = distribution.ArtifactList{}
			}
			currentVersion = row.version
		}

		// The central repository uses relative URIs to be future-proof.
		// Build the full URI before creating the artifact.
		fullImagePath := fmt.Sprintf("%s/%s", imagePrefix, row.uri)
		// Create the artifact for this row.
		artifact := distribution.Artifact{
			Image:  fullImagePath,
			URI:    "",
			Digest: row.digest,
			OS:     row.os,
			Arch:   row.arch,
		}
		artifactList = append(artifactList, artifact)
	}
	// Don't forget to store the very last plugin we were building
	allPlugins = appendPlugin(allPlugins, currentPlugin)

	err = rows.Err()
	return allPlugins, err
}

func getNextRow(rows *sql.Rows) (*centralRepoRow, error) {
	var row centralRepoRow
	err := rows.Scan(
		&row.name,
		&row.target,
		&row.recommendedVersion,
		&row.version,
		&row.hidden,
		&row.description,
		&row.publisher,
		&row.vendor,
		&row.os,
		&row.arch,
		&row.digest,
		&row.uri,
	)
	return &row, err
}

func convertTargetFromDB(target string) cliv1alpha1.Target {
	target = strings.ToLower(target)
	if target == "global" {
		target = ""
	}
	return cliv1alpha1.StringToTarget(target)
}

func appendPlugin(allPlugins []Discovered, plugin *Discovered) []Discovered {
	// Now that we are done gathering the information for the plugin
	// we need to compute the recommendedVersion if it wasn't provided
	// by the database
	if err := SortVersions(plugin.SupportedVersions); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing supported versions for plugin %s: %v", plugin.Name, err)
	}
	if plugin.RecommendedVersion == "" {
		plugin.RecommendedVersion = plugin.SupportedVersions[len(plugin.SupportedVersions)-1]
	}
	allPlugins = append(allPlugins, *plugin)
	return allPlugins
}

// // DiscoveredFromK8sV1alpha1 returns discovered plugin object from k8sV1alpha1
// func DiscoveredFromSQLite(p *cliv1alpha1.CLIPlugin) (Discovered, error) {
// 	dp := Discovered{
// 		Name:               p.Name,
// 		Description:        p.Spec.Description,
// 		RecommendedVersion: p.Spec.RecommendedVersion,
// 		Optional:           p.Spec.Optional,
// 		Target:             cliv1alpha1.StringToTarget(string(p.Spec.Target)),
// 	}
// 	dp.SupportedVersions = make([]string, 0)
// 	for v := range p.Spec.Artifacts {
// 		dp.SupportedVersions = append(dp.SupportedVersions, v)
// 	}
// 	if err := SortVersions(dp.SupportedVersions); err != nil {
// 		return dp, errors.Wrapf(err, "error parsing supported versions for plugin %s", p.Name)
// 	}
// 	dp.Distribution = distribution.ArtifactsFromK8sV1alpha1(p.Spec.Artifacts)
// 	return dp, nil
// }

// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package recommendedversion is used to check for
// the currently recommended versions of the Tanzu CLI
// and inform the user if they are using an outdated version.
package recommendedversion

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-cli/pkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-cli/pkg/centralconfig"
	cliconfig "github.com/vmware-tanzu/tanzu-cli/pkg/config"
	"github.com/vmware-tanzu/tanzu-cli/pkg/constants"
	"github.com/vmware-tanzu/tanzu-cli/pkg/datastore"
	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"
)

// dataStoreLastVersionCheckKey is the data store key used to store the last
// time the version check was done
const (
	centralConfigRecommendedVersionsKey = "cli.core.cli_recommended_versions"
	dataStoreLastVersionCheckKey        = "lastVersionCheck"
	recommendedVersionCheckDelaySeconds = 24 * 60 * 60 // 24 hours
)

// CheckRecommendedCLIVersion checks the recommended versions of the Tanzu CLI
// and prints recommendations to the user if they are using an outdated version.
// Once recommendations are printed to the user, the next check is only done after 24 hours.
func CheckRecommendedCLIVersion(cmd *cobra.Command) {
	if !shouldCheckVersion() {
		return
	}

	// We will get the central configuration from the default discovery source
	discoverySource, err := config.GetCLIDiscoverySource(cliconfig.DefaultStandaloneDiscoveryName)
	if err != nil {
		return
	}

	// Get the recommended versions from the central configuration
	reader := centralconfig.NewCentralConfigReader(discoverySource)
	recommendedVersionValue, err := reader.GetCentralConfigEntry(centralConfigRecommendedVersionsKey)
	if err != nil || recommendedVersionValue == nil {
		return
	}

	value, ok := recommendedVersionValue.(string)
	if !ok {
		log.V(7).Error(err, "wrong format for recommended versions in central config")
		return
	}
	recommendedVersions, err := sortRecommendedVersionsDescending(value)
	if err != nil {
		log.V(7).Error(err, "failed to sort recommended versions")
		return
	}

	currentVersion := buildinfo.Version
	includePreReleases := utils.IsPreRelease(currentVersion)
	major := findRecommendedMajorVersion(recommendedVersions, currentVersion, includePreReleases)
	minor := findRecommendedMinorVersion(recommendedVersions, currentVersion, includePreReleases)
	patch := findRecommendedPatchVersion(recommendedVersions, currentVersion, includePreReleases)

	printVersionRecommendations(cmd.ErrOrStderr(), currentVersion, major, minor, patch)
}

// findRecommendedMajorVersion will return the recommended major version from the list of
// recommended versions. If the current version is already at the most recent major version,
// it will return an empty string.
func findRecommendedMajorVersion(recommendedVersions []string, currentVersion string, includePreReleases bool) string {
	for _, newVersion := range recommendedVersions {
		if !includePreReleases && utils.IsPreRelease(newVersion) {
			// Skip pre-release versions
			continue
		}

		// This is the most recent of all versions. If it is the same major
		// as the current version, then the current version is already the correct major version
		if utils.IsSameMajor(newVersion, currentVersion) {
			return ""
		}
		return newVersion
	}
	return ""
}

// findRecommendedMinorVersion will return the recommended minor version from the list of
// recommended versions. If the current version is already at the most recent minor version,
// it will return an empty string.
func findRecommendedMinorVersion(recommendedVersions []string, currentVersion string, includePreReleases bool) string {
	for _, newVersion := range recommendedVersions {
		if !includePreReleases && utils.IsPreRelease(newVersion) {
			// Skip pre-release versions
			continue
		}

		// Since the recommended versions are sorted in descending order,
		// the first version that is the same major version as the current version
		// will be the most recent minor to recommend.
		if utils.IsSameMajor(newVersion, currentVersion) {
			// This is the most recent of version within the same major version.
			// If it is the same minor as the current version, then the current version
			// is already the correct minor version
			if utils.IsSameMinor(newVersion, currentVersion) {
				return ""
			}
			return newVersion
		}
	}
	return ""
}

// findRecommendedPatchVersion will return the recommended patch version from the list of
// recommended versions. If the current version is already at that patch version,
// it will return an empty string.
func findRecommendedPatchVersion(recommendedVersions []string, currentVersion string, includePreReleases bool) string {
	for _, newVersion := range recommendedVersions {
		if !includePreReleases && utils.IsPreRelease(newVersion) {
			// Skip pre-release versions
			continue
		}

		// Since the recommended versions are sorted in descending order,
		// the first version that is the same minor version as the current version
		// will be the most recent patch to recommend.
		if utils.IsSameMinor(newVersion, currentVersion) {
			// This is the most recent of version within the same minor version.
			// If it is the same as the current version, then the current version
			// is already the correct patch version
			if newVersion == currentVersion {
				return ""
			}
			return newVersion
		}
	}
	return ""
}

// sortRecommendedVersionsDescending will convert the comma-separated list of recommended
// versions into an array sorted in descending order of semver
func sortRecommendedVersionsDescending(recommendedVersionStr string) ([]string, error) {
	// The value is in the form "v1.2.1,v1.1.0,v0.90.1"
	// which is a comma separated list of recommended versions for each minor version of the CLI.
	recommendedArray := strings.Split(recommendedVersionStr, ",")

	// Trim any spaces around the version strings and remove duplicates
	recommendedVersions := make([]string, 0, len(recommendedArray))
	alreadyPresent := make(map[string]bool)
	for _, newVersion := range recommendedArray {
		trimmedVersion := strings.TrimSpace(newVersion)
		if trimmedVersion != "" && !alreadyPresent[trimmedVersion] {
			recommendedVersions = append(recommendedVersions, trimmedVersion)
			alreadyPresent[trimmedVersion] = true
		}
	}

	// Now sort the versions, then reverse the order
	err := utils.SortVersions(recommendedVersions)
	if err != nil {
		return nil, err
	}

	// Reverse the order so it is descending
	for i := len(recommendedVersions)/2 - 1; i >= 0; i-- {
		opp := len(recommendedVersions) - 1 - i
		recommendedVersions[i], recommendedVersions[opp] = recommendedVersions[opp], recommendedVersions[i]
	}
	return recommendedVersions, err
}

func getRecommendationDelayInSeconds() int {
	delay := recommendedVersionCheckDelaySeconds
	delayOverride := os.Getenv(constants.ConfigVariableRecommendVersionDelayDays)
	if delayOverride != "" {
		delayOverrideValue, err := strconv.Atoi(delayOverride)
		if err == nil {
			if delayOverrideValue >= 0 {
				// Convert from days to seconds
				delay = delayOverrideValue * 24 * 60 * 60
			} else {
				// When the configured delay is negative, it means the value
				// should be in seconds.  This is used for testing purposes.
				delay = -delayOverrideValue
			}
		}
	}
	return delay
}

func shouldCheckVersion() bool {
	delay := getRecommendationDelayInSeconds()
	if delay == 0 {
		// The user has disabled the version check
		return false
	}

	// Get the last time the version check was done
	lastCheck, err := datastore.GetDataStoreValue(dataStoreLastVersionCheckKey)
	if err != nil || lastCheck == nil {
		return true
	}

	lastCheckTime, ok := lastCheck.(time.Time)
	if !ok {
		return true
	}

	return time.Since(lastCheckTime) > time.Duration(delay)*time.Second
}

func printVersionRecommendations(writer io.Writer, currentVersion, major, minor, patch string) {
	if major == "" && minor == "" && patch == "" {
		// The current version is the best recommended version
		return
	}

	// Put a delimiter before this notification so the user
	// can see it is not part of the command output
	fmt.Fprintln(writer, "\n==")

	if utils.IsNewVersion(currentVersion, major) || utils.IsNewVersion(currentVersion, minor) || utils.IsNewVersion(currentVersion, patch) {
		fmt.Fprintf(writer, "WARNING: Due to a problem it is recommended not to use the current version: %s.\n", currentVersion)
		fmt.Fprintln(writer, "Please use a recommended version:")
	} else {
		fmt.Fprintf(writer, "Note: A new version of the Tanzu CLI is available. You are at version: %s.\n", currentVersion)
		fmt.Fprintln(writer, "To benefit from the latest security and features, please update to a recommended version:")
	}

	if major != "" {
		if utils.IsNewVersion(major, currentVersion) {
			fmt.Fprintf(writer, "  - %s\n", major)
		} else {
			fmt.Fprintf(writer, "  - %s ([!] you should downgrade to a previous major version)\n", major)
		}
	}
	if minor != "" {
		if utils.IsNewVersion(minor, currentVersion) {
			fmt.Fprintf(writer, "  - %s\n", minor)
		} else {
			fmt.Fprintf(writer, "  - %s ([!] you should downgrade to a previous minor version)\n", minor)
		}
	}
	if patch != "" {
		if utils.IsNewVersion(patch, currentVersion) {
			fmt.Fprintf(writer, "  - %s\n", patch)
		} else {
			fmt.Fprintf(writer, "  - %s ([!] you should downgrade to a previous patch version)\n", patch)
		}
	}

	delay := getRecommendationDelayInSeconds()
	var delayStr string
	if delay >= 60*60 {
		delayStr = fmt.Sprintf("%d hours", delay/60/60)
	} else {
		delayStr = fmt.Sprintf("%d seconds", delay)
	}
	fmt.Fprintf(writer, "\nThis message will print at most once per %s until you update the CLI.\n"+
		"Set %s to adjust this period (0 to disable).\n",
		delayStr, constants.ConfigVariableRecommendVersionDelayDays)

	// Now that we printed the message to the use, save the time of the last check
	// so that we don't continually print the message at every command
	_ = datastore.SetDataStoreValue(dataStoreLastVersionCheckKey, time.Now())
}

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-plugin-runtime/component"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"

	"github.com/vmware-tanzu/tanzu-cli/pkg/cli"
	"github.com/vmware-tanzu/tanzu-cli/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-cli/pkg/plugininventory"
	"github.com/vmware-tanzu/tanzu-cli/pkg/pluginmanager"
	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
)

var (
	groupID          string
	showNonMandatory bool
)

func newPluginGroupCmd() *cobra.Command {
	var pluginGroupCmd = &cobra.Command{
		Use:   "group",
		Short: "Manage plugin-groups",
		Long:  "Manage plugin-groups. A plugin-group provides a list of plugins name/version combinations which can be installed in one step.",
	}
	pluginGroupCmd.SetUsageFunc(cli.SubCmdUsageFunc)

	pluginGroupCmd.AddCommand(
		newSearchCmd(),
		newGetCmd(),
	)

	return pluginGroupCmd
}

func newSearchCmd() *cobra.Command {
	var searchCmd = &cobra.Command{
		Use:               "search",
		Short:             "Search for available plugin-groups",
		Long:              "Search from the list of available plugin-groups.  A plugin-group provides a list of plugin name/version combinations which can be installed in one step.",
		Args:              cobra.MaximumNArgs(0),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			var criteria *discovery.GroupDiscoveryCriteria
			if groupID != "" {
				groupIdentifier := plugininventory.PluginGroupIdentifierFromID(groupID)
				if groupIdentifier == nil {
					return errors.Errorf("incorrect plugin-group %q specified", groupID)
				}

				criteria = &discovery.GroupDiscoveryCriteria{
					Vendor:    groupIdentifier.Vendor,
					Publisher: groupIdentifier.Publisher,
					Name:      groupIdentifier.Name,
				}
			}
			groups, err := pluginmanager.DiscoverPluginGroups(discovery.WithGroupDiscoveryCriteria(criteria))
			if err != nil {
				return err
			}

			if !showDetails {
				displayGroupsFound(groups, cmd.OutOrStdout())
			} else {
				displayGroupDetails(groups, cmd.OutOrStdout())
			}
			return nil
		},
	}

	f := searchCmd.Flags()
	f.StringVarP(&groupID, "name", "n", "", "limit the search to the plugin-group with the specified name")
	utils.PanicOnErr(searchCmd.RegisterFlagCompletionFunc("name", completeGroupNames))

	f.BoolVar(&showDetails, "show-details", false, "show the details of the specified group, including all available versions")
	f.StringVarP(&outputFormat, "output", "o", "", "output format (yaml|json|table)")
	utils.PanicOnErr(searchCmd.RegisterFlagCompletionFunc("output", completionGetOutputFormats))

	return searchCmd
}

func newGetCmd() *cobra.Command {
	var getCmd = &cobra.Command{
		Use:               "get GROUP_NAME",
		Short:             "Get the content of the specified plugin-group",
		Long:              "Get the content of the specified plugin-group.  A plugin-group provides a list of plugin name/version combinations which can be installed in one step.  This command allows to see the list of plugins included in the specified group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeGroupGet,
		RunE: func(cmd *cobra.Command, args []string) error {
			gID := args[0]

			groupIdentifier := plugininventory.PluginGroupIdentifierFromID(gID)
			if groupIdentifier == nil {
				return errors.Errorf("incorrect plugin-group %q specified", gID)
			}

			if groupIdentifier.Version == "" {
				groupIdentifier.Version = cli.VersionLatest
			}

			criteria := &discovery.GroupDiscoveryCriteria{
				Vendor:    groupIdentifier.Vendor,
				Publisher: groupIdentifier.Publisher,
				Name:      groupIdentifier.Name,
				Version:   groupIdentifier.Version,
			}
			groups, err := pluginmanager.DiscoverPluginGroups(discovery.WithGroupDiscoveryCriteria(criteria))
			if err != nil {
				return err
			}
			if len(groups) == 0 {
				return errors.Errorf("plugin-group %q cannot be found", gID)
			}

			if len(groups) > 1 {
				log.Warningf("unexpectedly found %d entries for group %q. Using the first one", len(groups), gID)
			}

			if outputFormat == "" || outputFormat == string(component.TableOutputType) {
				displayGroupContentAsTable(groups[0], cmd.OutOrStdout())
			} else {
				displayGroupContentAsList(groups[0], cmd.OutOrStdout())
			}
			return nil
		},
	}

	f := getCmd.Flags()
	f.StringVarP(&outputFormat, "output", "o", "", "output format (yaml|json|table)")
	utils.PanicOnErr(getCmd.RegisterFlagCompletionFunc("output", completionGetOutputFormats))

	f.BoolVarP(&showNonMandatory, "all", "", false, "include the non-mandatory plugins")
	_ = f.MarkHidden("all")

	return getCmd
}

func displayGroupsFound(groups []*plugininventory.PluginGroup, writer io.Writer) {
	output := component.NewOutputWriter(writer, outputFormat, "group", "description", "latest")

	for _, pg := range groups {
		id := plugininventory.PluginGroupToID(pg)
		output.AddRow(id, pg.Description, pg.RecommendedVersion)
	}
	output.Render()
}

func displayGroupDetails(groups []*plugininventory.PluginGroup, writer io.Writer) {
	// Create a specific object format so it gets printed properly in yaml or json
	type detailedObject struct {
		Name        string
		Description string
		Latest      string
		Versions    []string
	}

	// For the table format, we will use individual yaml output for each group
	if outputFormat == "" || outputFormat == string(component.TableOutputType) {
		first := true
		for _, pg := range groups {
			if !first {
				fmt.Println()
			}
			first = false
			var supportedVersions []string
			for version := range pg.Versions {
				supportedVersions = append(supportedVersions, version)
			}
			_ = utils.SortVersions(supportedVersions)
			details := detailedObject{
				Name:        plugininventory.PluginGroupToID(pg),
				Description: pg.Description,
				Latest:      pg.RecommendedVersion,
				Versions:    supportedVersions,
			}
			component.NewObjectWriter(writer, string(component.YAMLOutputType), details).Render()
		}

		return
	}

	// Non-table format.
	// Here we use an objectWriter so that the array of versions is printed as an array
	// and not a long string.
	var details []detailedObject
	for _, pg := range groups {
		var supportedVersions []string
		for version := range pg.Versions {
			supportedVersions = append(supportedVersions, version)
		}
		_ = utils.SortVersions(supportedVersions)
		details = append(details, detailedObject{
			Name:        plugininventory.PluginGroupToID(pg),
			Description: pg.Description,
			Latest:      pg.RecommendedVersion,
			Versions:    supportedVersions,
		})
	}
	component.NewObjectWriter(writer, outputFormat, details).Render()
}

func displayGroupContentAsTable(group *plugininventory.PluginGroup, writer io.Writer) {
	cyanBold := color.New(color.FgCyan).Add(color.Bold)
	cyanBoldItalic := color.New(color.FgCyan).Add(color.Bold, color.Italic)
	output := component.NewOutputWriter(writer, outputFormat, "Name", "Target", "Version")

	gID := plugininventory.PluginGroupToID(group)
	_, _ = cyanBold.Println("Plugins in Group: ", cyanBoldItalic.Sprintf("%s:%s", gID, group.RecommendedVersion))

	for _, plugin := range group.Versions[group.RecommendedVersion] {
		if showNonMandatory || plugin.Mandatory {
			output.AddRow(plugin.Name, plugin.Target, plugin.Version)
		}
	}
	output.Render()
}

func displayGroupContentAsList(group *plugininventory.PluginGroup, writer io.Writer) {
	output := component.NewOutputWriter(writer, outputFormat, "Group", "PluginName", "PluginTarget", "PluginVersion")

	gID := fmt.Sprintf("%s:%s", plugininventory.PluginGroupToID(group), group.RecommendedVersion)
	for _, plugin := range group.Versions[group.RecommendedVersion] {
		if showNonMandatory || plugin.Mandatory {
			output.AddRow(gID, plugin.Name, plugin.Target, plugin.Version)
		}
	}
	output.Render()
}

// ====================================
// Shell completion functions
// ====================================
func completeGroupGet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeGroupsAndVersion(cmd, args, toComplete)
}

func completeGroupNames(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// We need to complete a group name
	groups, err := pluginmanager.DiscoverPluginGroups(discovery.WithUseLocalCacheOnly())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(groups) == 0 {
		// If no plugin group was returned it probably means the cache is empty.
		// Try the call again but allow it to download the plugin DB.
		groups, err = pluginmanager.DiscoverPluginGroups()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	var comps []string
	for _, g := range groups {
		comps = append(comps, fmt.Sprintf("%s\t%s", plugininventory.PluginGroupToID(g), g.Description))
	}

	// Sort to allow for testing
	sort.Strings(comps)

	return comps, cobra.ShellCompDirectiveNoFileComp
}

func completeGroupsAndVersion(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var comps []string
	if idx := strings.Index(toComplete, ":"); idx != -1 {
		// The gID is already specified before the :
		// so now we should complete the gID version
		gID := toComplete[:idx]
		versionToComplete := toComplete[idx+1:]
		versions, directive := completeGroupVersions(cmd, []string{gID}, versionToComplete)
		for _, v := range versions {
			comps = append(comps, fmt.Sprintf("%s:%s", gID, v))
		}

		return comps, directive
	}

	// We need to complete a group name.
	// Don't add a space after the group name so the uer can add a : if
	// they want to specify a version.
	comps, _ = completeGroupNames(nil, nil, "")
	return comps, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

func completeGroupVersions(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveError
	}
	gID := args[0]
	groupIdentifier := plugininventory.PluginGroupIdentifierFromID(gID)
	if groupIdentifier == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	criteria := &discovery.GroupDiscoveryCriteria{
		Vendor:    groupIdentifier.Vendor,
		Publisher: groupIdentifier.Publisher,
		Name:      groupIdentifier.Name,
	}

	groups, err := pluginmanager.DiscoverPluginGroups(
		discovery.WithGroupDiscoveryCriteria(criteria),
		discovery.WithUseLocalCacheOnly())
	if err != nil || len(groups) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	// Since more recent versions are more likely to be
	// useful, we return the list of versions in reverse order
	// and tell the shell to preserve that order using
	// cobra.ShellCompDirectiveKeepOrder
	var versions []string
	for v := range groups[0].Versions {
		versions = append(versions, v)
	}
	// Sort in ascending order
	_ = utils.SortVersions(versions)

	// Create the completions in reverse order
	comps := make([]string, len(versions))
	for i := range versions {
		comps[len(versions)-1-i] = versions[i]
	}
	return comps, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

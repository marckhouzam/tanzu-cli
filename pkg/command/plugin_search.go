// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/tanzu-cli/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-cli/pkg/pluginmanager"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

var (
	useRegex bool
	// TODO(khouzam) implement this case
	listVersions bool
)

const searchLongDesc = `Search provides the ability to search for plugins available to be installed.
Without an argument, the command lists all plugins currently available.
The search command can also be used with a regular expression to filter the
list of available plugins. 
`

func newSearchPluginCmd() *cobra.Command {
	var searchCmd = &cobra.Command{
		Use:               "search [keyword]",
		Short:             "Search for a keyword in the list of available plugins",
		Long:              searchLongDesc,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			discoveredPlugins, err := pluginmanager.DiscoverStandalonePlugins()
			if err != nil {
				return err
			}

			var filter string
			if len(args) == 1 {
				filter = args[0]
			}
			filteredPlugins := filterPluginList(cmd, discoveredPlugins, filter)
			sort.Sort(discovery.DiscoveredSorter(filteredPlugins))
			displayPluginList(filteredPlugins, cmd.OutOrStdout())

			return nil
		},
	}

	f := searchCmd.Flags()
	f.BoolVarP(&useRegex, "regex", "r", false, "use regular expressions for searching for plugins")
	// TODO(khouzam)
	f.BoolVarP(&listVersions, "versions", "l", false, "show the long listing, with each available version of plugins")
	f.StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	// TODO(khouzam) does this command need to have a --local flag?

	// Shell completion for the flags
	// TODO(khouzam) move this to the component package to be used by others
	// We don't include the "listtable" format as it is more aimed at developers than users.
	err := searchCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{
			string(component.TableOutputType),
			string(component.JSONOutputType),
			string(component.YAMLOutputType)}, cobra.ShellCompDirectiveNoFileComp))
	if err != nil {
		// This will only happen if we have a coding error so panicking is ok
		panic(err)
	}

	return searchCmd
}

func filterPluginList(cmd *cobra.Command, allPlugins []discovery.Discovered, filter string) []discovery.Discovered {
	var filteredPlugins []discovery.Discovered
	var matcher *regexp.Regexp

	// Do case-insensitive matching to help the user
	filter = strings.ToLower(filter)

	if useRegex {
		var err error
		if matcher, err = regexp.Compile(filter); err != nil {
			return nil
		}
	}

	for _, plugin := range allPlugins {
		pluginDetails := []string{plugin.Name, plugin.Description, string(plugin.Target), plugin.Status}
		detailStr := strings.ToLower(strings.Join(pluginDetails, " "))

		if useRegex {
			// Only add plugin that match the regex filter
			if matcher.Match([]byte(detailStr)) {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		} else {
			// Only add plugin that match the keyword filter
			if strings.Contains(detailStr, filter) {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}
	}
	return filteredPlugins
}

func displayPluginList(plugins []discovery.Discovered, writer io.Writer) {
	var data [][]string
	output := component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Status")

	for _, p := range plugins {
		pluginDetails := []string{p.Name, p.Description, string(p.Target), p.Status}
		data = append(data, pluginDetails)
	}

	addDataToOutputWriter := func(output component.OutputWriter, data [][]string) {
		for _, row := range data {
			vals := make([]interface{}, len(row))
			for i, val := range row {
				vals[i] = val
			}
			output.AddRow(vals...)
		}
	}

	addDataToOutputWriter(output, data)
	output.Render()
}

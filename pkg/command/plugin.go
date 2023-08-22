// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/vmware-tanzu/tanzu-plugin-runtime/component"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config"
	configtypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"

	"github.com/vmware-tanzu/tanzu-cli/pkg/cli"
	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-cli/pkg/constants"
	"github.com/vmware-tanzu/tanzu-cli/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-cli/pkg/plugininventory"
	"github.com/vmware-tanzu/tanzu-cli/pkg/pluginmanager"
	"github.com/vmware-tanzu/tanzu-cli/pkg/pluginsupplier"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"
)

var (
	local        string
	version      string
	forceDelete  bool
	outputFormat string
	targetStr    string
	group        string
)

const (
	invalidTargetMsg                = "invalid target specified. Please specify a correct value for the `--target/-t` flag from '" + common.TargetList + "'"
	errorWhileDiscoveringPlugins    = "there was an error while discovering plugins, error information: '%v'"
	errorWhileGettingContextPlugins = "there was an error while getting installed context plugins, error information: '%v'"
	pluginNameCaps                  = "PLUGIN_NAME"
)

func newPluginCmd() *cobra.Command {
	var pluginCmd = &cobra.Command{
		Use:   "plugin",
		Short: "Manage CLI plugins",
		Long:  "Provides all lifecycle operations for plugins",
		Annotations: map[string]string{
			"group": string(plugin.SystemCmdGroup),
		},
	}

	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)

	listPluginCmd := newListPluginCmd()
	installPluginCmd := newInstallPluginCmd()
	upgradePluginCmd := newUpgradePluginCmd()
	describePluginCmd := newDescribePluginCmd()
	deletePluginCmd := newDeletePluginCmd()
	cleanPluginCmd := newCleanPluginCmd()
	syncPluginCmd := newSyncPluginCmd()
	discoverySourceCmd := newDiscoverySourceCmd()

	listPluginCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	describePluginCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")

	if !config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
		installPluginCmd.Flags().StringVar(&group, "group", "", "install the plugins specified by a plugin-group version")

		// --local is renamed to --local-source
		installPluginCmd.Flags().StringVarP(&local, "local", "", "", "path to local plugin source")
		msg := "this was done in the v1.0.0 release, it will be removed following the deprecation policy (6 months). Use the --local-source flag instead.\n"
		if err := installPluginCmd.Flags().MarkDeprecated("local", msg); err != nil {
			// Will only fail if the flag does not exist, which would indicate a coding error,
			// so let's panic so we notice immediately.
			panic(err)
		}

		// The --local-source flag for installing plugins is only used in development testing
		// and should not be used in production.  We mark it as hidden to help convey this reality.
		installPluginCmd.Flags().StringVarP(&local, "local-source", "l", "", "path to local plugin source")
		if err := installPluginCmd.Flags().MarkHidden("local-source"); err != nil {
			// Will only fail if the flag does not exist, which would indicate a coding error,
			// so let's panic so we notice immediately.
			panic(err)
		}
	} else {
		installPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local discovery/distribution source")
		listPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local plugin source")
	}
	installPluginCmd.Flags().StringVarP(&version, "version", "v", cli.VersionLatest, "version of the plugin")
	deletePluginCmd.Flags().BoolVarP(&forceDelete, "yes", "y", false, "delete the plugin without asking for confirmation")

	if config.IsFeatureActivated(constants.FeatureContextCommand) {
		targetFlagDesc := fmt.Sprintf("target of the plugin (%s)", common.TargetList)
		installPluginCmd.Flags().StringVarP(&targetStr, "target", "t", "", targetFlagDesc)
		upgradePluginCmd.Flags().StringVarP(&targetStr, "target", "t", "", targetFlagDesc)
		deletePluginCmd.Flags().StringVarP(&targetStr, "target", "t", "", targetFlagDesc)
		describePluginCmd.Flags().StringVarP(&targetStr, "target", "t", "", targetFlagDesc)
	}

	pluginCmd.AddCommand(
		listPluginCmd,
		installPluginCmd,
		upgradePluginCmd,
		describePluginCmd,
		deletePluginCmd,
		cleanPluginCmd,
		syncPluginCmd,
		discoverySourceCmd,
	)

	if !config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
		installPluginCmd.MarkFlagsMutuallyExclusive("group", "local")
		installPluginCmd.MarkFlagsMutuallyExclusive("group", "local-source")
		installPluginCmd.MarkFlagsMutuallyExclusive("group", "version")
		if config.IsFeatureActivated(constants.FeatureContextCommand) {
			installPluginCmd.MarkFlagsMutuallyExclusive("group", "target")
		}
		pluginCmd.AddCommand(
			newSearchPluginCmd(),
			newPluginGroupCmd(),
			newDownloadBundlePluginCmd(),
			newUploadBundlePluginCmd(),
		)
	}

	return pluginCmd
}

func newListPluginCmd() *cobra.Command {
	var listCmd = &cobra.Command{
		Use:               "list",
		Short:             "List installed plugins",
		Long:              "List installed standalone plugins or plugins recommended by the contexts being used",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			errorList := make([]error, 0)
			if !config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
				// List installed standalone plugins
				standalonePlugins, err := pluginsupplier.GetInstalledStandalonePlugins()
				if err != nil {
					errorList = append(errorList, err)
					log.Warningf("there was an error while getting installed standalone plugins, error information: '%v'", err.Error())
				}
				sort.Sort(cli.PluginInfoSorter(standalonePlugins))

				// List installed context plugins and also missing context plugins.
				// Showing missing ones guides the user to know some plugins are recommended for the
				// active contexts, but are not installed.
				installedContextPlugins, missingContextPlugins, pluginSyncRequired, err := getInstalledAndMissingContextPlugins()
				if err != nil {
					errorList = append(errorList, err)
					log.Warningf(errorWhileGettingContextPlugins, err.Error())
				}
				sort.Sort(discovery.DiscoveredSorter(installedContextPlugins))
				sort.Sort(discovery.DiscoveredSorter(missingContextPlugins))

				if config.IsFeatureActivated(constants.FeatureContextCommand) && (outputFormat == "" || outputFormat == string(component.TableOutputType)) {
					displayInstalledAndMissingSplitView(standalonePlugins, installedContextPlugins, missingContextPlugins, pluginSyncRequired, cmd.OutOrStdout())
				} else {
					displayInstalledAndMissingListView(standalonePlugins, installedContextPlugins, missingContextPlugins, cmd.OutOrStdout())
				}

				return kerrors.NewAggregate(errorList)
			}

			// Plugin listing before the Central Repository feature
			var err error
			var availablePlugins []discovery.Discovered
			if local != "" {
				// get absolute local path
				local, err = filepath.Abs(local)
				if err != nil {
					return err
				}
				availablePlugins, err = pluginmanager.AvailablePluginsFromLocalSource(local)
			} else {
				availablePlugins, err = pluginmanager.AvailablePlugins()
			}

			if err != nil {
				log.Warningf("there was an error while getting available plugins, error information: '%v'", err.Error())
			}
			sort.Sort(discovery.DiscoveredSorter(availablePlugins))

			if config.IsFeatureActivated(constants.FeatureContextCommand) && (outputFormat == "" || outputFormat == string(component.TableOutputType)) {
				displayPluginListOutputSplitViewContext(availablePlugins, cmd.OutOrStdout())
			} else {
				displayPluginListOutputListView(availablePlugins, cmd.OutOrStdout())
			}

			return err
		},
	}

	return listCmd
}

func newDescribePluginCmd() *cobra.Command {
	var describeCmd = &cobra.Command{
		Use:               "describe " + pluginNameCaps,
		Short:             "Describe a plugin",
		Long:              "Displays detailed information for a plugin",
		ValidArgsFunction: completeInstalledPlugins,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "name", "version", "status", "target", "description", "installationPath")
			if len(args) != 1 {
				return fmt.Errorf("must provide one plugin name as a positional argument")
			}
			pluginName := args[0]

			if !configtypes.IsValidTarget(targetStr, true, true) {
				return errors.New(invalidTargetMsg)
			}

			pd, err := pluginmanager.DescribePlugin(pluginName, getTarget())
			if err != nil {
				return err
			}
			output.AddRow(pd.Name, pd.Version, pd.Status, pd.Target, pd.Description, pd.InstallationPath)
			output.Render()
			return nil
		},
	}

	return describeCmd
}

// nolint: gocyclo
func newInstallPluginCmd() *cobra.Command {
	var installCmd = &cobra.Command{
		Use:               "install [" + pluginNameCaps + "]",
		Short:             "Install a plugin",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeAllPlugins,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var pluginName string

			if !configtypes.IsValidTarget(targetStr, true, true) {
				return errors.New(invalidTargetMsg)
			}
			if config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
				return legacyPluginInstall(cmd, args)
			}

			if group != "" {
				// We are installing from a group
				if len(args) == 0 {
					// Default to 'all' when installing from a group
					pluginName = cli.AllPlugins
				} else {
					pluginName = args[0]
				}

				groupWithVersion, err := pluginmanager.InstallPluginsFromGroup(pluginName, group)
				if err != nil {
					return err
				}

				if pluginName == cli.AllPlugins {
					log.Successf("successfully installed all plugins from group '%s'", groupWithVersion)
				} else {
					log.Successf("successfully installed '%s' from group '%s'", pluginName, groupWithVersion)
				}

				return nil
			}

			// Invoke install plugin from local source if local files are provided
			if local != "" {
				if len(args) == 0 {
					return fmt.Errorf("missing plugin name or '%s' as an argument", cli.AllPlugins)
				}
				pluginName = args[0]

				// get absolute local path
				local, err = filepath.Abs(local)
				if err != nil {
					return err
				}
				err = pluginmanager.InstallPluginsFromLocalSource(pluginName, version, getTarget(), local, false)
				if err != nil {
					return err
				}
				if pluginName == cli.AllPlugins {
					log.Success("successfully installed all plugins")
				} else {
					log.Successf("successfully installed '%s' plugin", pluginName)
				}
				return nil
			}

			if len(args) == 0 {
				return errors.New("missing plugin name as an argument or the use of '--group'")
			}
			pluginName = args[0]

			if pluginName == cli.AllPlugins {
				return fmt.Errorf("the '%s' argument can only be used with the '--group' flag", cli.AllPlugins)
			}

			pluginVersion := version
			err = pluginmanager.InstallStandalonePlugin(pluginName, pluginVersion, getTarget())
			if err != nil {
				return err
			}
			log.Successf("successfully installed '%s' plugin", pluginName)
			return nil
		},
	}
	if !config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
		installCmd.Example = `
    # Install all plugins of the vmware-tkg/default plugin group version v2.1.0
    tanzu plugin install --group vmware-tkg/default:v2.1.0

    # Install all plugins of the latest version of the vmware-tkg/default plugin group
    tanzu plugin install --group vmware-tkg/default

    # Install the latest version of plugin "myPlugin"
    # If the plugin exists for more than one target, an error will be thrown
    tanzu plugin install myPlugin

    # Install the latest version of plugin "myPlugin" for target kubernetes
    tanzu plugin install myPlugin --target k8s

    # Install version v1.0.0 of plugin "myPlugin"
    tanzu plugin install myPlugin --version v1.0.0`
		installCmd.Long = "Install a specific plugin by name or specify all to install all plugins of a group"
	}
	return installCmd
}

func legacyPluginInstall(cmd *cobra.Command, args []string) error {
	var err error
	if len(args) == 0 {
		return fmt.Errorf("missing plugin name or '%s' as an argument", cli.AllPlugins)
	}
	pluginName := args[0]

	// Invoke install plugin from local source if local files are provided
	if local != "" {
		// get absolute local path
		local, err = filepath.Abs(local)
		if err != nil {
			return err
		}
		err = pluginmanager.InstallPluginsFromLocalSource(pluginName, version, getTarget(), local, false)
		if err != nil {
			return err
		}
		if pluginName == cli.AllPlugins {
			log.Successf("successfully installed all plugins")
		} else {
			log.Successf("successfully installed '%s' plugin", pluginName)
		}
		return nil
	}

	// Invoke plugin sync if install all plugins is mentioned
	if pluginName == cli.AllPlugins {
		err = pluginmanager.SyncPlugins()
		if err != nil {
			return err
		}
		log.Successf("successfully installed all plugins")
		return nil
	}

	pluginVersion := version
	if pluginVersion == cli.VersionLatest {
		pluginVersion, err = pluginmanager.GetRecommendedVersionOfPlugin(pluginName, getTarget())
		if err != nil {
			return err
		}
	}

	err = pluginmanager.InstallStandalonePlugin(pluginName, pluginVersion, getTarget())
	if err != nil {
		return err
	}
	log.Successf("successfully installed '%s' plugin", pluginName)
	return nil
}

func newUpgradePluginCmd() *cobra.Command {
	var upgradeCmd = &cobra.Command{
		Use:               "upgrade " + pluginNameCaps,
		Short:             "Upgrade a plugin",
		Long:              "Installs the latest version available for the specified plugin",
		ValidArgsFunction: completeAllPlugins,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return fmt.Errorf("must provide plugin name as positional argument")
			}
			pluginName := args[0]

			if !configtypes.IsValidTarget(targetStr, true, true) {
				return errors.New(invalidTargetMsg)
			}

			var pluginVersion string
			if !config.IsFeatureActivated(constants.FeatureDisableCentralRepositoryForTesting) {
				// With the Central Repository feature we can simply request to install
				// the recommendedVersion.
				pluginVersion = cli.VersionLatest
			} else {
				pluginVersion, err = pluginmanager.GetRecommendedVersionOfPlugin(pluginName, getTarget())
				if err != nil {
					return err
				}
			}

			err = pluginmanager.UpgradePlugin(pluginName, pluginVersion, getTarget())
			if err != nil {
				return err
			}
			log.Successf("successfully upgraded plugin '%s'", pluginName)
			return nil
		},
	}

	return upgradeCmd
}

func newDeletePluginCmd() *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:               "delete " + pluginNameCaps,
		Short:             "Delete a plugin",
		Long:              "Uninstalls the specified plugin",
		ValidArgsFunction: completeInstalledPlugins,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return fmt.Errorf("must provide one plugin name as a positional argument")
			}
			pluginName := args[0]

			if !configtypes.IsValidTarget(targetStr, true, true) {
				return errors.New(invalidTargetMsg)
			}

			deletePluginOptions := pluginmanager.DeletePluginOptions{
				PluginName:  pluginName,
				Target:      getTarget(),
				ForceDelete: forceDelete,
			}

			err = pluginmanager.DeletePlugin(deletePluginOptions)
			if err != nil {
				return err
			}

			log.Successf("successfully deleted plugin '%s'", pluginName)
			return nil
		},
	}
	return deleteCmd
}

func newCleanPluginCmd() *cobra.Command {
	var cleanCmd = &cobra.Command{
		Use:               "clean",
		Short:             "Clean the plugins",
		Long:              "Remove all installed plugins from the system",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = pluginmanager.Clean()
			if err != nil {
				return err
			}
			log.Success("successfully cleaned up all plugins")
			return nil
		},
	}
	return cleanCmd
}

func newSyncPluginCmd() *cobra.Command {
	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Installs all plugins recommended by the active contexts",
		Long: `Installs all plugins recommended by the active contexts.
Plugins installed with this command will only be available while the context remains active.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = pluginmanager.SyncPlugins()
			if err != nil {
				return err
			}
			log.Success("Done")
			return nil
		},
	}
	return syncCmd
}

// getInstalledElseAvailablePluginVersion return installed plugin version if plugin is installed
// if not installed it returns available recommended plugin version
func getInstalledElseAvailablePluginVersion(p *discovery.Discovered) string {
	installedOrAvailableVersion := p.InstalledVersion
	if installedOrAvailableVersion == "" {
		installedOrAvailableVersion = p.RecommendedVersion
	}
	return installedOrAvailableVersion
}

// getInstalledAndMissingContextPlugins returns any context plugins that are not installed
func getInstalledAndMissingContextPlugins() (installed, missing []discovery.Discovered, pluginSyncRequired bool, err error) {
	errorList := make([]error, 0)
	serverPlugins, err := pluginmanager.DiscoverServerPlugins()
	if err != nil {
		errorList = append(errorList, err)
		log.Warningf(errorWhileDiscoveringPlugins, err.Error())
	}

	// Note that the plugins we get here don't know from which context they were installed.
	// We need to cross-reference them with the discovered plugins.
	installedPlugins, err := pluginsupplier.GetInstalledServerPlugins()
	if err != nil {
		errorList = append(errorList, err)
		log.Warningf(errorWhileGettingContextPlugins, err.Error())
	}

	for i := range serverPlugins {
		found := false
		for j := range installedPlugins {
			if serverPlugins[i].Name != installedPlugins[j].Name || serverPlugins[i].Target != installedPlugins[j].Target {
				continue
			}

			// Store the installed plugin, which includes the context from which it was installed
			found = true
			if serverPlugins[i].RecommendedVersion != installedPlugins[j].Version {
				serverPlugins[i].Status = common.PluginStatusUpdateAvailable
				pluginSyncRequired = true
			} else {
				serverPlugins[i].Status = common.PluginStatusInstalled
			}
			serverPlugins[i].InstalledVersion = installedPlugins[j].Version
			installed = append(installed, serverPlugins[i])
			break
		}
		if !found {
			// We have a server plugin that is not installed, include it in the list
			serverPlugins[i].Status = common.PluginStatusNotInstalled
			missing = append(missing, serverPlugins[i])
			pluginSyncRequired = true
		}
	}
	return installed, missing, pluginSyncRequired, kerrors.NewAggregate(errorList)
}

func displayPluginListOutputListView(availablePlugins []discovery.Discovered, writer io.Writer) {
	var data [][]string
	var output component.OutputWriter

	for index := range availablePlugins {
		data = append(data, []string{availablePlugins[index].Name, availablePlugins[index].Description, availablePlugins[index].Scope,
			availablePlugins[index].Source, getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status})
	}
	output = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Scope", "Discovery", "Version", "Status")

	for _, row := range data {
		vals := make([]interface{}, len(row))
		for i, val := range row {
			vals[i] = val
		}
		output.AddRow(vals...)
	}
	output.Render()
}

func displayPluginListOutputSplitViewContext(availablePlugins []discovery.Discovered, writer io.Writer) {
	var dataStandalone [][]string
	var outputStandalone component.OutputWriter
	dataContext := make(map[string][][]string)
	outputContext := make(map[string]component.OutputWriter)

	outputStandalone = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Discovery", "Version", "Status")

	for index := range availablePlugins {
		if availablePlugins[index].Scope == common.PluginScopeStandalone {
			newRow := []string{availablePlugins[index].Name, availablePlugins[index].Description, string(availablePlugins[index].Target),
				availablePlugins[index].Source, getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status}
			dataStandalone = append(dataStandalone, newRow)
		} else {
			newRow := []string{availablePlugins[index].Name, availablePlugins[index].Description, string(availablePlugins[index].Target),
				getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status}
			outputContext[availablePlugins[index].ContextName] = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Version", "Status")
			data := dataContext[availablePlugins[index].ContextName]
			data = append(data, newRow)
			dataContext[availablePlugins[index].ContextName] = data
		}
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

	cyanBold := color.New(color.FgCyan).Add(color.Bold)
	cyanBoldItalic := color.New(color.FgCyan).Add(color.Bold, color.Italic)

	_, _ = cyanBold.Println("Standalone Plugins")
	addDataToOutputWriter(outputStandalone, dataStandalone)
	outputStandalone.Render()

	for context, writer := range outputContext {
		fmt.Println("")
		_, _ = cyanBold.Println("Plugins from Context: ", cyanBoldItalic.Sprintf(context))
		data := dataContext[context]
		addDataToOutputWriter(writer, data)
		writer.Render()
	}
}

func displayInstalledAndMissingSplitView(installedStandalonePlugins []cli.PluginInfo, installedContextPlugins, missingContextPlugins []discovery.Discovered, pluginSyncRequired bool, writer io.Writer) {
	// List installed standalone plugins
	cyanBold := color.New(color.FgCyan).Add(color.Bold)
	_, _ = cyanBold.Println("Standalone Plugins")

	outputStandalone := component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Version", "Status")
	for index := range installedStandalonePlugins {
		outputStandalone.AddRow(
			installedStandalonePlugins[index].Name,
			installedStandalonePlugins[index].Description,
			string(installedStandalonePlugins[index].Target),
			installedStandalonePlugins[index].Version,
			common.PluginStatusInstalled,
		)
	}
	outputStandalone.Render()

	// List installed and missing context plugins in one list.
	// First group them by context.
	contextPlugins := installedContextPlugins
	contextPlugins = append(contextPlugins, missingContextPlugins...)
	sort.Sort(discovery.DiscoveredSorter(contextPlugins))

	ctxPluginsByContext := make(map[string][]discovery.Discovered)
	for index := range contextPlugins {
		ctx := contextPlugins[index].ContextName
		ctxPluginsByContext[ctx] = append(ctxPluginsByContext[ctx], contextPlugins[index])
	}

	cyanBoldItalic := color.New(color.FgCyan).Add(color.Bold, color.Italic)

	// sort contexts to maintain consistency in the plugin list output
	contexts := make([]string, 0, len(ctxPluginsByContext))
	for context := range ctxPluginsByContext {
		contexts = append(contexts, context)
	}
	sort.Strings(contexts)
	for _, context := range contexts {
		outputWriter := component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Version", "Status")

		fmt.Println("")
		_, _ = cyanBold.Println("Plugins from Context: ", cyanBoldItalic.Sprintf(context))
		for i := range ctxPluginsByContext[context] {
			version := ctxPluginsByContext[context][i].InstalledVersion
			if ctxPluginsByContext[context][i].Status == common.PluginStatusNotInstalled {
				version = ctxPluginsByContext[context][i].RecommendedVersion
			}
			outputWriter.AddRow(
				ctxPluginsByContext[context][i].Name,
				ctxPluginsByContext[context][i].Description,
				string(ctxPluginsByContext[context][i].Target),
				version,
				ctxPluginsByContext[context][i].Status,
			)
		}
		outputWriter.Render()
	}

	if pluginSyncRequired {
		// Print a warning to the user that some context plugins are not installed or outdated and plugin sync is required to install them
		fmt.Println("")
		log.Warningf("As shown above, some recommended plugins have not been installed or are outdated. To install them please run 'tanzu plugin sync'.")
	}
}

func displayInstalledAndMissingListView(installedStandalonePlugins []cli.PluginInfo, installedContextPlugins, missingContextPlugins []discovery.Discovered, writer io.Writer) {
	outputWriter := component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Version", "Status", "Context")
	for index := range installedStandalonePlugins {
		outputWriter.AddRow(
			installedStandalonePlugins[index].Name,
			installedStandalonePlugins[index].Description,
			string(installedStandalonePlugins[index].Target),
			installedStandalonePlugins[index].Version,
			installedStandalonePlugins[index].Status,
			"", // No context
		)
	}

	// List context plugins that are installed.
	for i := range installedContextPlugins {
		outputWriter.AddRow(
			installedContextPlugins[i].Name,
			installedContextPlugins[i].Description,
			string(installedContextPlugins[i].Target),
			installedContextPlugins[i].InstalledVersion,
			installedContextPlugins[i].Status,
			installedContextPlugins[i].ContextName,
		)
	}

	// List context plugins that are not installed.
	for i := range missingContextPlugins {
		outputWriter.AddRow(
			missingContextPlugins[i].Name,
			missingContextPlugins[i].Description,
			string(missingContextPlugins[i].Target),
			missingContextPlugins[i].RecommendedVersion,
			common.PluginStatusNotInstalled,
			missingContextPlugins[i].ContextName,
		)
	}
	outputWriter.Render()
}

func getTarget() configtypes.Target {
	return configtypes.StringToTarget(strings.ToLower(targetStr))
}

func completeInstalledPlugins(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	installedPlugins, err := pluginsupplier.GetInstalledPlugins()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var comps []string
	target := getTarget()
	if len(args) == 0 {
		// Complete all plugin names as long as the target matches and let the shell filter
		for i := range installedPlugins {
			if target == configtypes.TargetUnknown || target == installedPlugins[i].Target {
				comps = append(comps, fmt.Sprintf("%s\t%s", installedPlugins[i].Name, installedPlugins[i].Description))
			}
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func completeAllPlugins(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var err error
	var allPlugins []discovery.Discovered
	var comps []string
	if local != "" {
		// The user requested the list of plugins from a local path
		local, err = filepath.Abs(local)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		allPlugins, err = pluginmanager.DiscoverPluginsFromLocalSource(local)

		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		for i := range allPlugins {
			comps = append(comps, fmt.Sprintf("%s\t%s", allPlugins[i].Name, allPlugins[i].Description))
		}

		// When using the --local flag, the "all" keyword can be used
		comps = append(comps, fmt.Sprintf("%s\t%s", cli.AllPlugins, "All plugins of the local source"))
		return comps, cobra.ShellCompDirectiveNoFileComp
	}

	if group != "" {
		groupIdentifier := plugininventory.PluginGroupIdentifierFromID(group)
		if groupIdentifier == nil {
			return nil, cobra.ShellCompDirectiveError
		}

		if groupIdentifier.Version == "" {
			groupIdentifier.Version = cli.VersionLatest
		}

		groups, err := pluginmanager.DiscoverPluginGroups(discovery.WithGroupDiscoveryCriteria(&discovery.GroupDiscoveryCriteria{
			Vendor:    groupIdentifier.Vendor,
			Publisher: groupIdentifier.Publisher,
			Name:      groupIdentifier.Name,
			Version:   groupIdentifier.Version,
		}))
		if err != nil || len(groups) == 0 {
			return nil, cobra.ShellCompDirectiveError
		}

		for _, plugin := range groups[0].Versions[groups[0].RecommendedVersion] {
			if showNonMandatory || plugin.Mandatory {
				// To get the description we would need to query the central repo again.
				// Let's avoid that extra delay and simply not provide a description.
				comps = append(comps, plugin.Name)
			}
		}

		// When using the --group flag, the "all" keyword can be used
		comps = append(comps, cli.AllPlugins)
		return comps, cobra.ShellCompDirectiveNoFileComp
	}

	// Show plugins found in the central repos
	allPlugins, err = pluginmanager.DiscoverStandalonePlugins(discovery.WithPluginDiscoveryCriteria(&discovery.PluginDiscoveryCriteria{
		Name:   pluginName,
		Target: configtypes.StringToTarget(targetStr)}))

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for i := range allPlugins {
		// TODO(khouzam): zsh and fish when receiving two identical completions even with different
		// descriptions, will only show the first one. E.g.,
		// $ tanzu plugin install cluster<TAB>
		// cluster       -- A TMC managed Kubernetes cluster
		// clustergroup  -- A group of Kubernetes clusters
		//
		// The missing description for TKG can be confusing, as if there is no cluster plugin for tkg
		// maybe we should remove the description, or add both to the same completion?
		comps = append(comps, fmt.Sprintf("%s\t%s", allPlugins[i].Name, allPlugins[i].Description))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

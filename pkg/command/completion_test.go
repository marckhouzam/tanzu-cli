// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_runCompletion_MissingArg validates functionality when shell name is not provided.
func Test_runCompletion_MissingArg(t *testing.T) {
	var out bytes.Buffer
	var args []string
	err := runCompletion(&out, newCompletionCmd(), args)
	if err == nil {
		t.Error("Missing shell argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "not specified") {
		t.Errorf("Unexpected error returned for missing argument: %s", err.Error())
	}

	if out.String() != "" {
		t.Errorf("Unexpected output received: %s", out.String())
	}
}

// Test_runCompletion_InvalidArg validates functionality when shell name is invalid.
func Test_runCompletion_InvalidArg(t *testing.T) {
	var out bytes.Buffer
	args := []string{"cmd.exe"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err == nil {
		t.Error("Invalid shell argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unrecognized") {
		t.Errorf("Unexpected error returned for invalid shell argument: %s", err.Error())
	}

	if out.String() != "" {
		t.Errorf("Unexpected output received: %s", out.String())
	}
}

// Test_runCompletion_WrongArgs validates functionality with too many arguments.
func Test_runCompletion_WrongArgs(t *testing.T) {
	var out bytes.Buffer
	args := []string{"bash", "zsh"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err == nil {
		t.Error("Invalid shell argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "too many arguments") {
		t.Errorf("Unexpected error returned for invalid shell argument: %s", err.Error())
	}

	if out.String() != "" {
		t.Errorf("Unexpected output received: %s", out.String())
	}
}

// Test_runCompletion_Bash validates functionality for bash shell completion.
func Test_runCompletion_Bash(t *testing.T) {
	var out bytes.Buffer
	args := []string{"bash"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err != nil {
		t.Errorf("Unexpected error for valid shell: %v", err)
	}

	// Check for a snippet of the bash completion output
	// TODO make this test less brittle
	if !strings.Contains(out.String(), "# bash completion V2") {
		t.Errorf("Unexpected output for the bash shell script: %s", out.String())
	}
}

// Test_runCompletion_Zsh validates functionality for zsh shell completion.
func Test_runCompletion_Zsh(t *testing.T) {
	var out bytes.Buffer
	args := []string{"zsh"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err != nil {
		t.Errorf("Unexpected error for valid shell: %v", err)
	}

	// Check for a snippet of the zsh completion output
	if !strings.Contains(out.String(), "# For zsh, when completing a flag with an =") {
		t.Errorf("Unexpected output for the zsh shell script: %s", out.String())
	}
}

// Test_runCompletion_Fish validates functionality for fish shell completion.
func Test_runCompletion_Fish(t *testing.T) {
	var out bytes.Buffer
	args := []string{"fish"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err != nil {
		t.Errorf("Unexpected error for valid shell: %v", err)
	}

	// Check for a snippet of the fish completion output
	if !strings.Contains(out.String(), "# For Fish, when completing a flag with an =") {
		t.Errorf("Unexpected output for the fish shell script: %s", out.String())
	}
}

// Test_runCompletion_Pwsh validates functionality for powershell completion.
func Test_runCompletion_Pwsh(t *testing.T) {
	var out bytes.Buffer
	args := []string{"powershell"}
	err := runCompletion(&out, newCompletionCmd(), args)
	if err != nil {
		t.Errorf("Unexpected error for valid shell: %v", err)
	}

	// Check for a snippet of the powershell completion output
	if !strings.Contains(out.String(), "# PowerShell supports three different completion modes") {
		t.Errorf("Unexpected output for the powershell script: %s", out.String())
	}
}

func TestCompletionCompletion(t *testing.T) {
	// This is global logic and needs not be tested for each
	// command.  Let's deactivate it.
	os.Setenv("TANZU_ACTIVE_HELP", "no_short_help")

	tests := []struct {
		test     string
		args     []string
		expected string
	}{
		{
			test: "completion of supported shells as the arg to the completion command",
			args: []string{"__complete", "completion", ""},
			// ":4" is the value of the ShellCompDirectiveNoFileComp
			expected: strings.Join(completionShells, "\n") + "\n:4\n",
		},
		{
			test: "no completion after the first arg for the completion command",
			args: []string{"__complete", "completion", "fish", ""},
			// ":4" is the value of the ShellCompDirectiveNoFileComp
			expected: "_activeHelp_ " + compNoMoreArgsMsg + "\n:4\n",
		},
	}

	for _, spec := range tests {
		t.Run(spec.test, func(t *testing.T) {
			assert := assert.New(t)

			rootCmd, err := NewRootCmdForTest()
			assert.Nil(err)

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetArgs(spec.args)

			err = rootCmd.Execute()
			assert.Nil(err)

			assert.Equal(spec.expected, out.String())
		})
	}

	os.Unsetenv("TANZU_ACTIVE_HELP")
}

// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package centralconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
)

func TestGetCentralConfigEntry(t *testing.T) {
	// Create a timestamp in the RFC3339 format
	timestampStr := time.Now().Format(time.RFC3339)
	// Now convert it back to a time.Time object so the two can be compared
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	assert.Nil(t, err)

	tcs := []struct {
		name        string
		cfgContent  string
		nofile      bool
		key         CentralConfigKey
		expected    CentralConfigValue
		expectError bool
	}{
		{
			name:     "No file for central config",
			nofile:   true,
			key:      "testKey",
			expected: nil,
		},
		{
			name:       "Empty central config",
			cfgContent: "",
			key:        "testKey",
			expected:   nil,
		},
		{
			name:       "String value",
			cfgContent: "testKey: testValue",
			key:        "testKey",
			expected:   "testValue",
		},
		{
			name:       "Boolean true value",
			cfgContent: "testKey: true",
			key:        "testKey",
			expected:   true,
		},
		{
			name:       "Boolean TRUE value",
			cfgContent: "testKey: TRUE",
			key:        "testKey",
			expected:   true,
		},
		{
			name:       "Boolean FALSE value",
			cfgContent: "testKey: FALSE",
			key:        "testKey",
			expected:   false,
		},
		{
			name:       "Boolean int value",
			cfgContent: "testKey: 1",
			key:        "testKey",
			expected:   1,
		},
		{
			name:       "Timestamp value",
			cfgContent: "testKey: " + timestampStr,
			key:        "testKey",
			expected:   timestamp,
		},
		{
			name: "Complex map value",
			cfgContent: `testKey:
  testSubKey: testValue`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": "testValue"},
		},
		{
			name: "An array value",
			cfgContent: `testKey:
  - testValue1
  - testValue2`,
			key:      "testKey",
			expected: interface{}([]interface{}{"testValue1", "testValue2"}),
		},
		{
			name: "A map of array value",
			cfgContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": []interface{}{"testValue1", "testValue2"}},
		},
		{
			name: "A complex string",
			cfgContent: `testKey: |-
 {
   "testSubKey": [ 
     "testValue1",
     "testValue1"
   ]
 }`,
			key: "testKey",
			expected: `{
  "testSubKey": [ 
    "testValue1",
    "testValue1"
  ]
}`,
		}, {
			name: "Missing key",
			cfgContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "invalidKey",
			expected: nil,
		},
		{
			name: "Empty key",
			cfgContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "",
			expected: nil,
		},
		{
			name:       "Empty value",
			cfgContent: "testKey: ",
			key:        "testKey",
			expected:   interface{}(nil),
		},
		{
			name: "Invalid yaml",
			cfgContent: `testKey: testValue
- invalid`,
			key:         "testKey",
			expectError: true,
		},
	}

	dir, err := os.MkdirTemp("", "test-central-config")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	common.DefaultCacheDir = dir

	reader := NewCentralConfigReader(&types.PluginDiscovery{
		OCI: &types.OCIDiscovery{
			Name:  "my_discovery",
			Image: "image",
		},
	})

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.nofile {
				// Write the central config test content to the file
				path := reader.(*centralConfigYamlReader).configFile

				err = os.MkdirAll(filepath.Dir(path), 0755)
				assert.Nil(t, err)

				err = os.WriteFile(path, []byte(tc.cfgContent), 0644)
				assert.Nil(t, err)
			}

			value, err := reader.GetCentralConfigEntry(tc.key)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tc.expected == nil {
				assert.Nil(t, value)
			} else {
				assert.Equal(t, tc.expected, value)
			}
		})
	}
}

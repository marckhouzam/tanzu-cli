// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package datastore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDataStoreValue(t *testing.T) {
	// Create a timestamp in the RFC3339 format
	timestampStr := time.Now().Format(time.RFC3339)
	// Now convert it back to a time.Time object so the two can be compared
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	assert.Nil(t, err)

	tcs := []struct {
		name      string
		dsContent string
		nofile    bool
		nodir     bool
		key       DataStoreKey
		expected  DataStoreValue
	}{
		{
			name:     "No directory for data store",
			nodir:    true,
			key:      "testKey",
			expected: nil,
		},
		{
			name:     "No file for data store",
			nofile:   true,
			key:      "testKey",
			expected: nil,
		},
		{
			name:      "Empty data store",
			dsContent: "",
			key:       "testKey",
			expected:  nil,
		},
		{
			name:      "String value",
			dsContent: "testKey: testValue",
			key:       "testKey",
			expected:  "testValue",
		},
		{
			name:      "Boolean true value",
			dsContent: "testKey: true",
			key:       "testKey",
			expected:  true,
		},
		{
			name:      "Boolean TRUE value",
			dsContent: "testKey: TRUE",
			key:       "testKey",
			expected:  true,
		},
		{
			name:      "Boolean FALSE value",
			dsContent: "testKey: FALSE",
			key:       "testKey",
			expected:  false,
		},
		{
			name:      "Boolean int value",
			dsContent: "testKey: 1",
			key:       "testKey",
			expected:  1,
		},
		{
			name:      "Timestamp value",
			dsContent: "testKey: " + timestampStr,
			key:       "testKey",
			expected:  timestamp,
		},
		{
			name: "Complex map value",
			dsContent: `testKey:
  testSubKey: testValue`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": "testValue"},
		},
		{
			name: "More map of array value",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": []interface{}{"testValue1", "testValue2"}},
		},
		{
			name: "Missing key",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "invalidKey",
			expected: nil,
		},
		{
			name: "Empty key",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "",
			expected: nil,
		},
		{
			name:      "Empty value",
			dsContent: "testKey: ",
			key:       "testKey",
			expected:  nil,
		},
	}

	tmpDir, err := os.MkdirTemp("", "data_store_test")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDir)
	defer os.RemoveAll(tmpDir)

	tmpDSFile, err := os.CreateTemp(tmpDir, "data-store.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDSFile)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.nodir {
				// Set the environment variable to a nonexistent directory and file
				nonExistentDir := tmpDir + "_nonexistentdir"
				os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", filepath.Join(nonExistentDir, "data-store.yaml"))
				defer os.RemoveAll(nonExistentDir)
			} else {
				if tc.nofile {
					// Set the environment variable to a nonexistent file
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name()+"_nonexistent")
				} else {
					// Write the data store test content to the file
					err = os.WriteFile(tmpDSFile.Name(), []byte(tc.dsContent), 0644)
					assert.Nil(t, err)
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name())
				}
			}

			value, err := GetDataStoreValue(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, value)
		})
	}
	os.Unsetenv("TEST_CUSTOM_DATA_STORE_FILE")
}

func TestSetDataStoreValue(t *testing.T) {
	// Create a timestamp in the RFC3339 format
	timestampStr := time.Now().Format(time.RFC3339)
	// Now convert it back to a time.Time object so the two can be compared
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	assert.Nil(t, err)

	tcs := []struct {
		name   string
		nofile bool
		nodir  bool
		key    DataStoreKey
		value  DataStoreValue
		// expected is the expected value to be returned by GetDataStoreValue.
		// Normally, it should be the same as value, but in some cases, it can be different.
		expected DataStoreValue
	}{
		{
			name:  "No directory for data store",
			nodir: true,
			key:   "testKey",
			value: "testValue",
		},
		{
			name:   "No file for data store",
			nofile: true,
			key:    "testKey",
			value:  "testValue",
		},
		{
			name:  "String value",
			key:   "testKey",
			value: "testValue",
		},
		{
			name:  "Boolean true value",
			key:   "testKey",
			value: true,
		},
		{
			name:  "Boolean TRUE value",
			key:   "testKey",
			value: true,
		},
		{
			name:  "Boolean FALSE value",
			key:   "testKey",
			value: false,
		},
		{
			name:  "Boolean int value",
			key:   "testKey",
			value: 1,
		},
		{
			name:  "Timestamp value",
			key:   "testKey",
			value: timestamp,
		},
		{
			name:     "Complex map value",
			key:      "testKey",
			value:    map[string]string{"testSubKey": "testValue"},
			expected: map[string]interface{}{"testSubKey": "testValue"},
		},
		{
			name:     "More map of array value",
			key:      "testKey",
			value:    map[string][]string{"testSubKey": {"testValue1", "testValue2"}},
			expected: map[string]interface{}{"testSubKey": []interface{}{"testValue1", "testValue2"}},
		},
		{
			name:  "Empty key",
			key:   "",
			value: "testValue",
		},
		{
			name:  "Empty value",
			key:   "testKey",
			value: "",
		},
		{
			name:  "Both empty",
			key:   "",
			value: "",
		},
	}

	tmpDir, err := os.MkdirTemp("", "data_store_test")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDir)
	defer os.RemoveAll(tmpDir)

	tmpDSFile, err := os.CreateTemp(tmpDir, "data-store.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDSFile)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.nodir {
				// Set the environment variable to a nonexistent directory and file
				nonExistentDir := tmpDir + "_nonexistentdir2"
				os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", filepath.Join(nonExistentDir, "data-store.yaml"))
				defer os.RemoveAll(nonExistentDir)
			} else {
				if tc.nofile {
					// Set the environment variable to a nonexistent file
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name()+"_nonexistent2")
				} else {
					// Set the environment variable to the file we already created
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name())
				}
			}

			if tc.expected == nil {
				tc.expected = tc.value
			}

			err = SetDataStoreValue(tc.key, tc.value)
			assert.Nil(t, err)

			value, err := GetDataStoreValue(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestDeleteDataStoreValue(t *testing.T) {
	// Create a timestamp in the RFC3339 format
	timestampStr := time.Now().Format(time.RFC3339)
	// Now convert it back to a time.Time object so the two can be compared
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	assert.Nil(t, err)

	tcs := []struct {
		name        string
		dsContent   string
		nofile      bool
		nodir       bool
		key         DataStoreKey
		expected    DataStoreValue
		expectError bool
	}{
		{
			name:        "No directory for data store",
			nodir:       true,
			key:         "testKey",
			expectError: true,
		},
		{
			name:        "No file for data store",
			nofile:      true,
			key:         "testKey",
			expectError: true,
		},
		{
			name:        "Empty data store",
			dsContent:   "",
			key:         "testKey",
			expectError: true,
		},
		{
			name:      "String value",
			dsContent: "testKey: testValue",
			key:       "testKey",
			expected:  "testValue",
		},
		{
			name:      "Boolean true value",
			dsContent: "testKey: true",
			key:       "testKey",
			expected:  true,
		},
		{
			name:      "Boolean TRUE value",
			dsContent: "testKey: TRUE",
			key:       "testKey",
			expected:  true,
		},
		{
			name:      "Boolean FALSE value",
			dsContent: "testKey: FALSE",
			key:       "testKey",
			expected:  false,
		},
		{
			name:      "Boolean int value",
			dsContent: "testKey: 1",
			key:       "testKey",
			expected:  1,
		},
		{
			name:      "Timestamp value",
			dsContent: "testKey: " + timestampStr,
			key:       "testKey",
			expected:  timestamp,
		},
		{
			name: "Complex map value",
			dsContent: `testKey:
  testSubKey: testValue`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": "testValue"},
		},
		{
			name: "More map of array value",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:      "testKey",
			expected: map[string]interface{}{"testSubKey": []interface{}{"testValue1", "testValue2"}},
		},
		{
			name: "Missing key",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:         "invalidKey",
			expectError: true,
		},
		{
			name: "Empty key",
			dsContent: `testKey:
  testSubKey:
  - testValue1
  - testValue2`,
			key:         "",
			expectError: true,
		},
		{
			name:      "Empty value",
			dsContent: "testKey: ",
			key:       "testKey",
			expected:  nil,
		},
	}

	tmpDir, err := os.MkdirTemp("", "data_store_test")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDir)
	defer os.RemoveAll(tmpDir)

	tmpDSFile, err := os.CreateTemp(tmpDir, "data-store.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, tmpDSFile)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.nodir {
				// Set the environment variable to a nonexistent directory and file
				nonExistentDir := tmpDir + "_nonexistentdir3"
				os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", filepath.Join(nonExistentDir, "data-store.yaml"))
				defer os.RemoveAll(nonExistentDir)
			} else {
				if tc.nofile {
					// Set the environment variable to a nonexistent file
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name()+"_nonexistent3")
				} else {
					// Write the data store test content to the file
					err = os.WriteFile(tmpDSFile.Name(), []byte(tc.dsContent), 0644)
					assert.Nil(t, err)
					os.Setenv("TEST_CUSTOM_DATA_STORE_FILE", tmpDSFile.Name())
				}
			}

			value, err := DeleteDataStoreValue(tc.key)
			if tc.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, value)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, value)
			}

			// Make sure the key is deleted
			value, err = GetDataStoreValue(tc.key)
			assert.Nil(t, err)
			assert.Nil(t, value)
		})
	}
	os.Unsetenv("TEST_CUSTOM_DATA_STORE_FILE")
}

func TestGetDataStorePath(t *testing.T) {
	// Verify that the data store path is in the .config directory (not the .cache directory)
	path := getDataStorePath()
	assert.Contains(t, path, ".config")
}

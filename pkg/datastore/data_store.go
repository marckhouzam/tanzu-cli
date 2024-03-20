// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package datastore implements the use of a data store yaml file
// that can be used for the CLI to store and retrieve data that is not configuration.
package datastore

import (
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
)

// dataStoreFileName is the name of the data store yaml file
// that is stored in the .config/tanzu directory.
// It is a hidden file and should not be directly accessed by the user.
const dataStoreFileName = ".data-store.yaml"

var lockFile *lockedfile.File

type DataStoreValue interface{}
type DataStoreKey string

type dataStoreContent map[DataStoreKey]DataStoreValue

// GetDataStoreValue retrieves the value of the key from the data store.
func GetDataStoreValue(key DataStoreKey) (DataStoreValue, error) {
	content, err := getDataStoreContent(false)
	if err != nil || content == nil {
		return nil, err
	}

	return content[key], nil
}

// SetDataStoreValue sets the value of the key in the data store.
func SetDataStoreValue(key DataStoreKey, value DataStoreValue) error {
	content, err := getDataStoreContent(true)
	if err != nil {
		return err
	}

	if content == nil {
		content = make(dataStoreContent)
	}
	content[key] = value

	return saveAndClose(content)
}

// DeleteDataStoreValue deletes the key and value from the data store.
// The previous value of the key is returned.
func DeleteDataStoreValue(key DataStoreKey) (DataStoreValue, error) {
	content, err := getDataStoreContent(true)
	if err != nil {
		return nil, err
	}

	deletedValue, present := content[key]
	if !present {
		_ = saveAndClose(content)
		return nil, errors.New("key not found in data store")
	}

	delete(content, key)

	err = saveAndClose(content)
	return deletedValue, err
}

// getDataStore retrieves the data store from the config directory along with locking the file.
// If `setWriteLock` is false, it will read the data store file with a ReadLock and release the
// lock at the same time.
// If `setWriteLock` is true, it will apply a WriteLock to the data store file, read the file
// and keep the WriteLock on the file.  The function saveAndClose() should be called to save
// any changes and release the lock.
func getDataStoreContent(setWriteLock bool) (dataStoreContent, error) {
	var content dataStoreContent

	b, err := getDataStoreBytes(setWriteLock)
	if err != nil {
		if os.IsNotExist(err) {
			return content, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(b, &content)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode data store file")
	}

	return content, nil
}

func getDataStoreBytes(setWriteLock bool) ([]byte, error) {
	var err error
	var b []byte

	dsPath := getDataStorePath()
	if setWriteLock {
		dsDir := filepath.Dir(dsPath)
		if !utils.PathExists(dsDir) {
			// Create directory path if missing before locking the file
			_ = os.MkdirAll(dsDir, 0755)
		}
		lockFile, err = lockedfile.Edit(dsPath)
		if err != nil {
			return nil, err
		}

		b, err = io.ReadAll(lockFile)
	} else {
		b, err = lockedfile.Read(dsPath)
	}
	return b, err
}

// getDataStorePath gets the data store file path
func getDataStorePath() string {
	// NOTE: TEST_CUSTOM_DATA_STORE_FILE is only for test purpose
	customDSFile := os.Getenv("TEST_CUSTOM_DATA_STORE_FILE")
	if customDSFile != "" {
		return customDSFile
	}

	return filepath.Join(xdg.Home, ".config", "tanzu", dataStoreFileName)
}

// saveFile saves the data store file in the .config directory.
func saveAndClose(content dataStoreContent) error {
	if lockFile == nil {
		return errors.New("cannot save the data store file as it is not locked")
	}
	defer lockFile.Close()

	dsPath := getDataStorePath()
	_, err := os.Stat(dsPath)
	if err != nil {
		return errors.Wrap(err, "could not stat the data store file")
	}

	out, err := yaml.Marshal(content)
	if err != nil {
		return errors.Wrap(err, "failed to encode the data store file")
	}

	if err := lockFile.Truncate(0); err != nil {
		return errors.Wrap(err, "failed to truncate the data store file")
	}
	if _, err := lockFile.Seek(0, 0); err != nil {
		return errors.Wrap(err, "failed to reset the data store file")
	}
	if _, err := lockFile.Write(out); err != nil {
		return errors.Wrap(err, "failed to write the data store file")
	}
	return nil
}

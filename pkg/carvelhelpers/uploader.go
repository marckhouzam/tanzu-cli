// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"github.com/pkg/errors"
)

// UploadImage upload the image using imgpkg tools
func UploadImage(imageWithTag string, inputDir string) error {
	reg, err := newRegistry()
	if err != nil {
		return errors.Wrapf(err, "unable to initialize registry")
	}

	err = reg.UploadImage(imageWithTag, inputDir)
	if err != nil {
		return errors.Wrap(err, "error uploading image")
	}

	return nil
}

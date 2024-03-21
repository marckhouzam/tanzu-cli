// Copyright 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package centralconfig implements an interface to deal with the central configuration.
package centralconfig

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	stringType      = reflect.TypeOf("")
	boolType        = reflect.TypeOf(true)
	intType         = reflect.TypeOf(1)
	floatType       = reflect.TypeOf(1.0)
	stringArrayType = reflect.TypeOf([]string{})
	stringMapType   = reflect.TypeOf(map[string]string{})
	arrayType       = reflect.TypeOf([]interface{}{})
	mapType         = reflect.TypeOf(map[string]interface{}{})
	timeType        = reflect.TypeOf(time.Time{})
)

type centralConfigYamlReader struct {
	// configFile is the path to the central config file.
	configFile string
}

// Make sure centralConfigYamlReader implements CentralConfig
var _ CentralConfig = &centralConfigYamlReader{}

// parseConfigFile reads the central config file and returns the parsed yaml content.
// If the file does not exist, it does not return an error because some central repositories
// may choose not to have a central config file.
func (c *centralConfigYamlReader) parseConfigFile() (map[string]interface{}, error) {
	// Check if the central config file exists.
	if _, err := os.Stat(c.configFile); os.IsNotExist(err) {
		// The central config file is optional, don't return an error if it does not exist.
		return nil, nil
	}

	bytes, err := os.ReadFile(c.configFile)
	if err != nil {
		return nil, err
	}

	var content map[string]interface{}
	err = yaml.Unmarshal(bytes, &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (c *centralConfigYamlReader) GetCentralConfigEntry(key string, out interface{}) error {
	values, err := c.parseConfigFile()
	if err != nil {
		return err
	}

	ok, err := extractValue(out, values, key)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("key %s not found in central config", key)
	}

	return nil
}

//nolint:funlen,gocyclo
func extractValue(out interface{}, values map[string]interface{}, key string) (ok bool, err error) {
	res, ok := values[key]
	if !ok {
		return false, nil
	}

	v := reflect.ValueOf(out)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	} else {
		return false, fmt.Errorf("out must be a pointer to a value")
	}

	switch v.Type() {
	case stringType:
		var result string
		result, ok, err = unstructured.NestedString(values, key)
		if err == nil && ok {
			v.Set(reflect.ValueOf(result))
		}
	case boolType:
		var result bool
		result, ok, err = unstructured.NestedBool(values, key)
		if err == nil && ok {
			v.Set(reflect.ValueOf(result))
		}
	case intType:
		// The "unstructured" package only supports int64 but when parsing yaml, we get
		// an int type.  To deal with this we have to implement the support ourselves
		var result int
		var val interface{}
		val, ok, err = unstructured.NestedFieldNoCopy(values, key)
		if err == nil && ok {
			result, ok = val.(int)
			if !ok {
				err = fmt.Errorf("error: %v is of the type %T, expected int", val, val)
			} else {
				v.Set(reflect.ValueOf(result))
			}
		}
	case floatType:
		var result float64
		result, ok, err = unstructured.NestedFloat64(values, key)
		if err == nil && ok {
			v.Set(reflect.ValueOf(result))
		}
	case stringArrayType:
		var result []string
		result, ok, err = unstructured.NestedStringSlice(values, key)
		if err == nil && ok {
			v.Set(reflect.ValueOf(result))
		}
	// case stringMapType:
	// 	var result map[string]string
	// 	result, ok, err = unstructured.NestedStringMap(values, key)
	// 	if err == nil && ok {
	// 		v.Set(reflect.ValueOf(result))
	// 	}
	case arrayType: // generic array
		var result []interface{}
		result, ok, err = unstructured.NestedSlice(values, key)
		if err == nil && ok {
			v.Set(reflect.ValueOf(result))
		}
	// case mapType: // generic map
	// 	var result map[string]interface{}
	// 	result, ok, err = unstructured.NestedMap(values, key)
	// 	if err == nil && ok {
	// 		v.Set(reflect.ValueOf(result))
	// 	}
	case timeType:
		var result time.Time
		var val interface{}
		val, ok, err = unstructured.NestedFieldNoCopy(values, key)
		if err == nil && ok {
			result, ok = val.(time.Time)
			if !ok {
				err = fmt.Errorf("error: %v is of the type %T, expected time.Time", val, val)
			} else {
				v.Set(reflect.ValueOf(result))
			}
		}
	default:
		var yamlBytes []byte
		yamlBytes, err = yaml.Marshal(res)
		if err == nil {
			err = yaml.Unmarshal(yamlBytes, out)
		}
	}
	return ok, err
}

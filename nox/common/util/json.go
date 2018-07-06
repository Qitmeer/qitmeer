// Copyright (c) 2017-2018 The nox developers

package util

import (
	"bytes"
	"encoding/json"
	"reflect"
)


// The OrderedMap works as the order map[string]interface{}, which useful to
// marshal a Json map using a specific order
// Note:
// it works almost as the 'omitEmpty' set to true, empty value should be omitted when marshaling.
// only for empty string/array/map/slice/interface
type OrderedMap []KV

type KV struct {
	Key string
	Val interface{}
}

func (omap OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range omap {
		if isEmptyValue(reflect.ValueOf(kv.Val)) {
			continue
		}
		if i != 0 {
			buf.WriteString(",")
		}

		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

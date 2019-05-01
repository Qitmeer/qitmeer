// Copyright (c) 2017-2018 The nox developers

package json

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// The Result works as method calling response can be marshaled as json map
type Result map[string]interface{}

// The OrderedResult works as the order map[string]interface{}, which useful to
// marshal a Json map using a specific order
// Note:
// it works almost as the 'omitEmpty' set to true, empty value should be omitted when marshaling.
// only for empty string/array/map/slice/interface
type OrderedResult []KV

type KV struct {
	Key string
	Val interface{}
}

func (ores OrderedResult) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range ores {
		if isEmptyValue(reflect.ValueOf(kv.Val)) {
			continue  //omit empty
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

/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:file.go
 * Date:6/12/20 12:40 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import (
	"io/ioutil"
	"os"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func ReadFile(path string) ([]byte, error) {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}

	ba, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ba, nil
}

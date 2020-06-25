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
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetFileInfo(path string) os.FileInfo {
	fileInfo, e := os.Stat(path)
	if e != nil {
		return nil
	}
	return fileInfo
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func CopyFile(src, dst string) bool {
	if len(src) == 0 || len(dst) == 0 {
		return false
	}
	srcFile, e := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	if e != nil {
		log.Println("copyfile", e)
		return false
	}
	defer srcFile.Close()

	dst = strings.Replace(dst, "\\", "/", -1)
	dstPathArr := strings.Split(dst, "/")
	dstPathArr = dstPathArr[0 : len(dstPathArr)-1]
	dstPath := strings.Join(dstPathArr, "/")

	dstFileInfo := GetFileInfo(dstPath)
	if dstFileInfo == nil {
		if e := os.MkdirAll(dstPath, os.ModePerm); e != nil {
			log.Println("copyfile", e)
			return false
		}
	}
	dstFile, e := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if e != nil {
		log.Println("copyfile", e)
		return false
	}
	defer dstFile.Close()
	//fileInfo, e := srcFile.Stat()
	//fileInfo.Size() > 1024
	//byteBuffer := make([]byte, 10)
	if _, e := io.Copy(dstFile, srcFile); e != nil {
		log.Println("copyfile", e)
		return false
	} else {
		return true
	}

}

func CopyPath(src, dst string) bool {
	srcFileInfo := GetFileInfo(src)
	if srcFileInfo == nil || !srcFileInfo.IsDir() {
		return false
	}
	srcPath := strings.TrimRight(strings.TrimRight(src, "/"), "\\")
	dstPath := strings.TrimRight(strings.TrimRight(dst, "/"), "\\") + "/"

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("CopyPath", err)
			return err
		}
		relationPath := strings.Replace(path, srcPath, "", -1)
		relationPath = strings.TrimLeft(relationPath, "/")
		curdstPath := dstPath + relationPath
		if !info.IsDir() {
			if CopyFile(path, curdstPath) {
				return nil
			} else {
				return errors.New(path + " copy fail")
			}
		} else {
			if !Exists(curdstPath) {
				if err := os.MkdirAll(curdstPath, os.ModePerm); err != nil {
					log.Println("CopyPath", err)
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Println("CopyPath", err)
		return false
	}
	return true

}

func RemovePath(path string) error {
	// Remove the old path if it already exists.
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	log.Printf("Removing from '%s'", path)
	if fi.IsDir() {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	} else {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

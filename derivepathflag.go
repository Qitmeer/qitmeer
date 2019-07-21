// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import "github.com/HalalChain/qitmeer-lib/wallet"

type derivePathFlag struct {
	path wallet.DerivationPath
}

func (d *derivePathFlag) Set(s string) error {
	path, err := wallet.ParseDerivationPath(s)
	if err!=nil {
		return err
	}
	d.path = path
	return nil
}

func (d *derivePathFlag) String() string {
	return d.path.String()
}


package main

import "github.com/noxproject/nox/wallet"

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


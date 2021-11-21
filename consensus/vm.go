/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package consensus

type VM interface {
	GetID() string
	Initialize(ctx Context) error
	Bootstrapping() error
	Bootstrapped() error
	Shutdown() error
	Version() string
}

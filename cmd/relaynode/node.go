/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	cfg        *Config
	ctx        context.Context
	cancel     context.CancelFunc
	privateKey *ecdsa.PrivateKey
}

func (node *Node) init(cfg *Config) error {
	log.Info(fmt.Sprintf("Start relay node..."))
	node.ctx, node.cancel = context.WithCancel(context.Background())

	err := cfg.load()
	if err != nil {
		return err
	}
	node.cfg = cfg

	pk, err := p2p.PrivateKey(cfg.DataDir, cfg.PrivateKey, 0600)
	if err != nil {
		return err
	}
	node.privateKey = pk

	log.Info(fmt.Sprintf("Load config completed"))
	return nil
}

func (node *Node) exit() error {
	node.cancel()
	log.Info(fmt.Sprintf("Stop relay node"))
	return nil
}

func (node *Node) run() error {
	log.Info(fmt.Sprintf("Run relay node..."))

	var exip string
	if len(node.cfg.ExternalIP) > 0 {
		exip = node.cfg.ExternalIP
	} else {
		eip := p2p.IpAddr()
		if eip == nil {
			return fmt.Errorf("Can't get IP")
		}
		exip = eip.String()
	}

	eMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", exip, node.cfg.Port))
	if err != nil {
		log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	srcMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", defaultIP, node.cfg.Port))
	if err != nil {
		log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	opts := []libp2p.Option{
		libp2p.EnableRelay(relay.OptHop),
		libp2p.ListenAddrs(srcMAddr, eMAddr),
		libp2p.Identity(p2p.ConvertToInterfacePrivkey(node.privateKey)),
	}

	h, err := libp2p.New(
		node.ctx,
		opts...,
	)
	if err != nil {
		log.Error("Failed to create host %v", err)
		return err
	}

	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			log.Info(fmt.Sprintf("Connected:%s (%s)", remotePeer, conn.RemoteMultiaddr()))
		},
	})

	h.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			log.Info(fmt.Sprintf("Disconnected:%s (%s)", remotePeer, conn.RemoteMultiaddr()))
		},
	})

	log.Info(fmt.Sprintf("Relay Address: %s/p2p/%s\n", eMAddr.String(), h.ID()))
	log.Info("You can copy the relay address and configure it to the required Qitmeer-Node")

	interrupt := interruptListener()
	<-interrupt
	return nil
}

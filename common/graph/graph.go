// Copyright 2017 The nox-project Authors
// This file is part of the nox-project library.
//
// The nox-project library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The nox-project library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the nox-project library. If not, see <http://www.gnu.org/licenses/>.

// Package graph provides graph operation utilities.
package graph

type Graph struct {
	Nodes map[string]*Node
}

func New() *Graph {
	g := &Graph{Nodes: make(map[string]*Node)}
	return g
}

func (g *Graph) CreateNode(hash string) *Node{
	n, ok := g.Nodes[hash]
	if !ok {
		n = &Node{ID:len(g.Nodes)}
		g.Nodes[hash]=n
	}
	return n
}

type Node struct {
	ID   int
	In   []*Edge
	Out  []*Edge
}

// A Edge represents an edge in the graph.
type Edge struct {
	From *Node
	To *Node
}

// determines if the nox is included by at least one of earlier_noxes
// Leading zeros are accepted. The empty string parses as zero.

func IfIncluded(earlier_noxes []string, current_nox string) bool {
	for _, v := range earlier_noxes {
		if v == current_nox {
			return true
		}

	}
	return false
}

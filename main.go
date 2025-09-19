// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeNodeId = errors.New("node IDs nust be non-negative")
	ErrNodeOutOfBouns = errors.New("edge node ID out of bound")
	ErrNilEdges = errors.New("edges silce cannot be nil")
	ErrInvalidNumNodes = errors.New("numNodes must be positive ")
)

type Edge struct {
	Source int32
	Target int32
}

type CSR struct {
	rowPtr []int32 // Row Pointer
	colIdx []int32 // Column Index
	nNodes int32   // Number of nodes
	nEdges int32   // Number of edges
}

func NewCSR(edges []Edge, numNodes int32, reverse bool) (*CSR, error) {
	csr := &CSR{
		rowPtr: make([]int32, numNodes+1),
		colIdx: make([]int32, len(edges)),
		nNodes: numNodes,
		nEdges: int32(len(edges)),
	}

	// count outdegrees: count edges per each node and store them in the rowPtr
	// if reverse we change direction from target->source (top-down) else source->target(bottom-up)
	if reverse {
		for _, e := range edges {
			csr.rowPtr[e.Target+1]++
		}
	} else {
		for _, e := range edges {
			csr.rowPtr[e.Source+1]++
		}
	}

	// transform counts → offsets
	for i := int32(0); i < numNodes; i++ {
		csr.rowPtr[i+1] += csr.rowPtr[i]
	}

	// fill colIdx
	position := make([]int32, numNodes)
	copy(position, csr.rowPtr[:numNodes])
    if reverse {
        for _, e := range edges {
            csr.colIdx[position[e.Target]] = e.Source
            position[e.Target]++
        }
    } else {
        for _, e := range edges {
            csr.colIdx[position[e.Source]] = e.Target
            position[e.Source]++
        }
    }
	return csr, nil
}

func main() {
	fmt.Println("Hello from GBT")
}
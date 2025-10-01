// Copyright 2025 Piotr Wysocki
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graph

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeNodeID  = errors.New("node IDs must be non-negative")
	ErrNodeOutOfBounds = errors.New("edge node ID out of bound")
	ErrNilEdges        = errors.New("edges silce cannot be nil")
	ErrInvalidNumNodes = errors.New("numNodes must be non-negative ")
)

type Edge struct {
	Source int32
	Target int32
}

type CSR struct {
	RowPtr []int32 // Row Pointer
	ColIdx []int32 // Column Index
	NNodes int32   // Number of nodes
	NEdges int32   // Number of edges
}

func NewCSR(edges []Edge, numNodes int32, reverse bool) (*CSR, error) {
	if numNodes < 0 {
		return nil, ErrInvalidNumNodes
	}

	if err := ValidateEdges(edges, numNodes); err != nil {
		return nil, err
	}

	csr := &CSR{
		RowPtr: make([]int32, numNodes+1),
		ColIdx: make([]int32, len(edges)),
		NNodes: numNodes,
		NEdges: int32(len(edges)),
	}

	// count outdegrees: count edges per each node and store them in the rowPtr
	// if reverse we change direction from target->source (top-down) else source->target(bottom-up)
	if reverse {
		for _, e := range edges {
			csr.RowPtr[e.Target+1]++
		}
	} else {
		for _, e := range edges {
			csr.RowPtr[e.Source+1]++
		}
	}

	// transform counts → offsets
	for i := int32(0); i < numNodes; i++ {
		csr.RowPtr[i+1] += csr.RowPtr[i]
	}

	// fill colIdx
	position := make([]int32, numNodes)
	copy(position, csr.RowPtr[:numNodes])
	if reverse {
		for _, e := range edges {
			csr.ColIdx[position[e.Target]] = e.Source
			position[e.Target]++
		}
	} else {
		for _, e := range edges {
			csr.ColIdx[position[e.Source]] = e.Target
			position[e.Source]++
		}
	}
	return csr, nil
}

func (csr *CSR) Neighbors(node int32) []int32 {
	start := csr.RowPtr[node]
	end := csr.RowPtr[node+1]
	return csr.ColIdx[start:end]
}

func ValidateEdges(edges []Edge, numNodes int32) error {
	if edges == nil {
		return ErrNilEdges
	}

	for _, e := range edges {
		if e.Source < 0 || e.Target < 0 {
			return fmt.Errorf("%w: source=%d, target=%d", ErrNegativeNodeID, e.Source, e.Target)
		}
		if e.Source >= numNodes || e.Target >= numNodes {
			return fmt.Errorf("%w: source=%d, target=%d, max=%d", ErrNodeOutOfBounds, e.Source, e.Target, numNodes-1)
		}
	}
	return nil
}
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSR_ForwardEdges(t *testing.T) {
	// GIVEN: graph with 3 edges
	edges := []Edge{
		{Source: 0, Target: 1},
		{Source: 0, Target: 2},
		{Source: 2, Target: 3},
	}
	expectedRowPtr := []int32{0, 2, 2, 3, 3}
	expectedColIdx := []int32{1, 2, 3}
	n := int32(0)
	expectedNodeNeighbors := []int32{1, 2}

	// WHEN: build a forward CSR and call the node neighbors
	csr, err := NewCSR(edges, 4, false)
	require.NotNil(t, csr)
	require.NoError(t, err)
	nbh := csr.Neighbors(n)
	require.NotNil(t, nbh)

	// THEN: RowPtr, ColIdx, nbh match the expected
	assert.Equal(t, expectedRowPtr, csr.RowPtr)
	assert.ElementsMatch(t, expectedColIdx, csr.ColIdx)
	assert.ElementsMatch(t, expectedNodeNeighbors, nbh)
}

func TestCSR_ReverseEdges(t *testing.T) {
	// GIVEN: graph with 3 edges
	edges := []Edge{
		{Source: 0, Target: 1},
		{Source: 0, Target: 2},
		{Source: 2, Target: 3},
	}
	expectedRowPtr := []int32{0, 0, 1, 2, 3}
	expectedColIdx := []int32{0, 0, 2}
	n := int32(3)
	expectedNodeNeighbors := []int32{2}

	// WHEN: build a reverse CSR and call the node neighbors
	csr, err := NewCSR(edges, 4, true)
	require.NotNil(t, csr)
	require.NoError(t, err)
	nbh := csr.Neighbors(n)
	require.NotNil(t, nbh)

	// THEN: RowPtr and ColIdx match the expected
	assert.Equal(t, expectedRowPtr, csr.RowPtr)
	assert.ElementsMatch(t, expectedColIdx, csr.ColIdx)
	assert.ElementsMatch(t, expectedNodeNeighbors, nbh)
}

func TestCSR_NegativeNodeIdError(t *testing.T) {
	// GIVEN: Graph with negative node
	edges := []Edge{
		{Source: 0, Target: -1},
	}

	// WHEN: build a CSR
	csr, err := NewCSR(edges, 1, false)
	require.Nil(t, csr)

	// THEN: negative error node is required
	require.ErrorIs(t, err, ErrNegativeNodeID)
}

func TestCSR_ErrNodeOutOfBounds(t *testing.T) {
	//GIVEN: Graph with 2 nodes and numNode = 2
	edges := []Edge{
		{Source: 0, Target: 1},
		{Source: 0, Target: 2},
		{Source: 2, Target: 3},
	}
	numNode := int32(2)

	//WHEN: build a CSR
	csr, err := NewCSR(edges, numNode, false)
	require.Nil(t, csr)

	//THEN: node out of bounds error is required
	require.ErrorIs(t, err, ErrNodeOutOfBounds)
}

func TestCSR_ErrNilEdges(t *testing.T) {
	//GIVEN: nil edges graph
	var edges []Edge

	//WHEN: build a CSR
	csr, err := NewCSR(edges, 0, false)
	require.Nil(t, csr)

	//THEN: node out of bounds error is required
	require.ErrorIs(t, err, ErrNilEdges)
}

func TestCSR_ErrNumNodes(t *testing.T) {
	//GIVEN: negative numnodes
	edges := []Edge{
		{Source: 0, Target: 1},
	}
	numNodes := int32(-1)

	//WHEN: build as CSR
	csr, err := NewCSR(edges, numNodes, false)
	require.Nil(t, csr)

	//THEN: invalid number of nodes error is required
	require.ErrorIs(t, err, ErrInvalidNumNodes)
}

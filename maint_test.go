// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCSR_ForwardEdges(t *testing.T) {
	// GIVEN: Simple graph 
	edges := []Edge{
		{Source: 0, Target: 1},
		{Source: 0, Target: 2},
		{Source: 2, Target: 3},
	}
	expectedRowPtr := []int32{0,2,2,3,3}
	expectedColIdx := []int32{1,2,3}

	// WHEN: We build a forward CSR
	csr, err := NewCSR(edges, 4, false)
	require.NotNil(t, csr)
	require.NoError(t, err)

	// THEN: rowPtr and colIdx match the expected
	assert.Equal(t, expectedRowPtr, csr.rowPtr)
	assert.ElementsMatch(t, expectedColIdx, csr.colIdx)
}

func TestCSR_ReverseEdges(t *testing.T) {
	// GIVEN: Simple graph 
	edges := []Edge{
		{Source: 0, Target: 1},
		{Source: 0, Target: 2},
		{Source: 2, Target: 3},
	}
	expectedRowPtr := []int32{0,0,1,2,3}
	expectedColIdx := []int32{0,0,2}

	// WHEN: We build a reverse CSR
	csr, err := NewCSR(edges, 4, true)
	require.NotNil(t, csr)
	require.NoError(t, err)

	// THEN: rowPtr and colIdx match the expected
	assert.Equal(t, expectedRowPtr, csr.rowPtr)
	assert.ElementsMatch(t, expectedColIdx, csr.colIdx)
}
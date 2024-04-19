// Copyright 2024 PingCAP, Inc.
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"testing"

	"github.com/pingcap/tiproxy/pkg/balance/observer"
	"github.com/stretchr/testify/require"
)

func TestFactorHealth(t *testing.T) {
	factor := NewFactorHealth()
	tests := []struct {
		healthy       bool
		expectedScore uint64
	}{
		{
			healthy:       true,
			expectedScore: 0,
		},
		{
			healthy:       false,
			expectedScore: 1,
		},
	}
	backends := make([]*backendWrapper, 0, len(tests))
	for _, test := range tests {
		status := observer.StatusHealthy
		if !test.healthy {
			status = observer.StatusCannotConnect
		}
		backend := newBackendWrapper("", observer.BackendHealth{Status: status})
		backends = append(backends, backend)
	}
	factor.UpdateScore(backends)
	for i, test := range tests {
		require.Equal(t, test.expectedScore, backends[i].score())
	}
}

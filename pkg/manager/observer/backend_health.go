// Copyright 2024 PingCAP, Inc.
// SPDX-License-Identifier: Apache-2.0

package observer

import "fmt"

type BackendStatus int

func (bs BackendStatus) ToScore() int {
	return statusScores[bs]
}

func (bs BackendStatus) String() string {
	status, ok := statusNames[bs]
	if !ok {
		return "unknown"
	}
	return status
}

const (
	StatusHealthy BackendStatus = iota
	StatusCannotConnect
	StatusMemoryHigh
	StatusRunSlow
	StatusSchemaOutdated
)

var statusNames = map[BackendStatus]string{
	StatusHealthy:        "healthy",
	StatusCannotConnect:  "down",
	StatusMemoryHigh:     "memory high",
	StatusRunSlow:        "run slow",
	StatusSchemaOutdated: "schema outdated",
}

var statusScores = map[BackendStatus]int{
	StatusHealthy:        0,
	StatusCannotConnect:  10000000,
	StatusMemoryHigh:     5000,
	StatusRunSlow:        5000,
	StatusSchemaOutdated: 10000000,
}

type BackendHealth struct {
	Status BackendStatus
	// The error occurred when health check fails. It's used to log why the backend becomes unhealthy.
	PingErr error
	// The backend version that returned to the client during handshake.
	ServerVersion string
}

func (bh *BackendHealth) Equals(health BackendHealth) bool {
	return bh.Status == health.Status && bh.ServerVersion == health.ServerVersion
}

func (bh *BackendHealth) String() string {
	str := fmt.Sprintf("status: %s", bh.Status.String())
	if bh.PingErr != nil {
		str += fmt.Sprintf(", err: %s", bh.PingErr.Error())
	}
	return str
}

// BackendInfo stores the status info of each backend.
type BackendInfo struct {
	IP         string
	StatusPort uint
}

// HealthResult contains the health check results and is used to notify the routers.
// It's read-only for subscribers.
type HealthResult struct {
	// `backends` is empty when `err` is not nil. It doesn't mean there are no backends.
	backends map[string]*BackendHealth
	err      error
}

// NewHealthResult is used for testing in other packages.
func NewHealthResult(backends map[string]*BackendHealth, err error) HealthResult {
	return HealthResult{
		backends: backends,
		err:      err,
	}
}

func (hr HealthResult) Backends() map[string]*BackendHealth {
	newMap := make(map[string]*BackendHealth, len(hr.backends))
	for addr, health := range hr.backends {
		newMap[addr] = health
	}
	return newMap
}

func (hr HealthResult) Error() error {
	return hr.err
}

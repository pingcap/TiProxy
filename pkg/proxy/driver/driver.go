// Copyright 2020 Ipalfish, Inc.
// Copyright 2022 PingCAP, Inc.
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

package driver

import (
	"crypto/tls"
	"net"

	mgrns "github.com/pingcap/TiProxy/pkg/manager/namespace"
	"github.com/pingcap/TiProxy/pkg/proxy/backend"
	"github.com/pingcap/TiProxy/pkg/proxy/client"
)

type createClientConnFunc func(*client.QueryCtxImpl, net.Conn, uint64, *tls.Config, *tls.Config) *client.ClientConnection
type createBackendConnMgrFunc func(connectionID uint64) *backend.BackendConnManager

type Driver struct {
	nsmgr                    *mgrns.NamespaceManager
	createClientConnFunc     createClientConnFunc
	createBackendConnMgrFunc createBackendConnMgrFunc
}

func NewDriverImpl(nsmgr *mgrns.NamespaceManager, createClientConnFunc createClientConnFunc, createBackendConnMgrFunc createBackendConnMgrFunc) *Driver {
	return &Driver{
		nsmgr:                    nsmgr,
		createClientConnFunc:     createClientConnFunc,
		createBackendConnMgrFunc: createBackendConnMgrFunc,
	}
}

func (d *Driver) CreateClientConnection(conn net.Conn, connectionID uint64, serverTLSConfig, clusterTLSConfig *tls.Config) *client.ClientConnection {
	backendConnMgr := d.createBackendConnMgrFunc(connectionID)
	queryCtx := client.NewQueryCtxImpl(d.nsmgr, backendConnMgr, connectionID)
	return d.createClientConnFunc(queryCtx, conn, connectionID, serverTLSConfig, clusterTLSConfig)
}

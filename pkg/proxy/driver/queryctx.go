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
	"context"
	"crypto/tls"
	"errors"

	pnet "github.com/pingcap/TiProxy/pkg/proxy/net"
)

// Server information.
const (
	ServerStatusInTrans            uint16 = 0x0001
	ServerStatusAutocommit         uint16 = 0x0002
	ServerMoreResultsExists        uint16 = 0x0008
	ServerStatusNoGoodIndexUsed    uint16 = 0x0010
	ServerStatusNoIndexUsed        uint16 = 0x0020
	ServerStatusCursorExists       uint16 = 0x0040
	ServerStatusLastRowSend        uint16 = 0x0080
	ServerStatusDBDropped          uint16 = 0x0100
	ServerStatusNoBackslashEscaped uint16 = 0x0200
	ServerStatusMetadataChanged    uint16 = 0x0400
	ServerStatusWasSlow            uint16 = 0x0800
	ServerPSOutParams              uint16 = 0x1000
)

type QueryCtxImpl struct {
	connId  uint64
	nsmgr   NamespaceManager
	ns      Namespace
	connMgr BackendConnManager
}

func NewQueryCtxImpl(nsmgr NamespaceManager, backendConnMgr BackendConnManager, connId uint64) *QueryCtxImpl {
	return &QueryCtxImpl{
		connId:  connId,
		nsmgr:   nsmgr,
		connMgr: backendConnMgr,
	}
}

func (q *QueryCtxImpl) ExecuteCmd(ctx context.Context, request []byte, clientIO *pnet.PacketIO) error {
	return q.connMgr.ExecuteCmd(ctx, request, clientIO)
}

func (q *QueryCtxImpl) Close() error {
	if q.connMgr != nil {
		return q.connMgr.Close()
	}
	return nil
}

func (q *QueryCtxImpl) ConnectBackend(ctx context.Context, clientIO *pnet.PacketIO, serverTLSConfig, backendTLSConfig *tls.Config) error {
	ns, ok := q.nsmgr.GetNamespace("")
	if !ok {
		return errors.New("failed to find a namespace")
	}
	q.ns = ns
	router := ns.GetRouter()
	addr, err := router.Route(q.connMgr)
	if err != nil {
		return err
	}
	if err = q.connMgr.Connect(ctx, addr, clientIO, serverTLSConfig, backendTLSConfig); err != nil {
		return err
	}
	return nil
}

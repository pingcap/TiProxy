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

package backend

import (
	"crypto/tls"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"
	pnet "github.com/pingcap/TiProxy/pkg/proxy/net"
)

type proxyConfig struct {
	frontendTLSConfig *tls.Config
	backendTLSConfig  *tls.Config
	sessionToken      string
	waitRedirect      bool
}

func newProxyConfig() *proxyConfig {
	return &proxyConfig{
		sessionToken: mockToken,
	}
}

type mockProxy struct {
	*proxyConfig
	auth         *Authenticator
	cmdProcessor *CmdProcessor
	// outputs that received from the server.
	rs *gomysql.Result
	// execution results
	err         error
	holdRequest bool
}

func newMockProxy(cfg *proxyConfig) *mockProxy {
	return &mockProxy{
		proxyConfig:  cfg,
		auth:         new(Authenticator),
		cmdProcessor: NewCmdProcessor(),
	}
}

func (mp *mockProxy) authenticateFirstTime(clientIO, backendIO *pnet.PacketIO) error {
	_, err := mp.auth.handshakeFirstTime(clientIO, backendIO, mp.frontendTLSConfig, mp.backendTLSConfig)
	return err
}

func (mp *mockProxy) authenticateSecondTime(_, backendIO *pnet.PacketIO) error {
	return mp.auth.handshakeSecondTime(backendIO, mp.sessionToken)
}

func (mp *mockProxy) processCmd(clientIO, backendIO *pnet.PacketIO) error {
	clientIO.ResetSequence()
	request, err := clientIO.ReadPacket()
	if err != nil {
		return err
	}
	if mp.holdRequest, _, err = mp.cmdProcessor.executeCmd(request, clientIO, backendIO, mp.waitRedirect); err != nil {
		return err
	}
	// Pretend to redirect the held request to the new backend. The backend must respond for another loop.
	if mp.holdRequest {
		_, _, err = mp.cmdProcessor.executeCmd(request, clientIO, backendIO, false)
	}
	return err
}

func (mp *mockProxy) directQuery(_, backendIO *pnet.PacketIO) error {
	rs, _, err := mp.cmdProcessor.query(backendIO, mockCmdStr)
	mp.rs = rs
	return err
}

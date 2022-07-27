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

package net

import "github.com/pingcap/TiProxy/pkg/util/errors"

var (
	ErrExpectSSLRequest = errors.New("expect a SSLRequest packet")
	ErrReadConn         = errors.New("failed to read the connection")
	ErrWriteConn        = errors.New("failed to write the connection")
	ErrFlushConn        = errors.New("failed to flush the connection")
	ErrCloseConn        = errors.New("failed to close the connection")
	ErrHandshakeTLS     = errors.New("failed to complete tls handshake")
)

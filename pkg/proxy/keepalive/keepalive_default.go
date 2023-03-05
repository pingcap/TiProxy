//go:build !(linux || netbsd || freebsd || dragonfly || aix || windows || darwin)
// +build !(linux OR netbsd OR freebsd OR dragonfly OR aix OR windows OR darwin)

// Copyright 2023 PingCAP, Inc.
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

package keepalive

import (
	"syscall"

	"github.com/pingcap/TiProxy/lib/config"
)

func setKeepalive(syscn syscall.RawConn, cfg config.KeepAlive) error {
	var serr error
	return errors.Collect(ErrKeepAlive, serr, syscn.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)
		if serr != nil {
			return
		}
	}))
}

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

package errors_test

import (
	"testing"

	serr "github.com/pingcap/TiProxy/pkg/util/errors"
	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {
	e1 := serr.New("tt")
	e2 := serr.New("dd")
	e3 := serr.New("dd")
	e := serr.Collect(e1, e2, e3)

	require.ErrorIsf(t, e, e1, "equal to the external error")
	require.Equal(t, serr.Unwrap(e), e, "unwrapping is noop")
	require.Equal(t, e.(*serr.MError).Cause(), []error{e2, e3}, "get underlying errors")
	require.NoError(t, serr.Collect(e3), "nil if there is no underlying error")
}

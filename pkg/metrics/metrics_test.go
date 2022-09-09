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

package metrics

import (
	"testing"
)

func TestPushMetrics(t *testing.T) {
	SetupMetrics("127.0.0.1:9091", 15, "0.0.0.0:6000")
}

func TestNoPushMetrics(t *testing.T) {
	SetupMetrics("", 15, "0.0.0.0:6000")
}

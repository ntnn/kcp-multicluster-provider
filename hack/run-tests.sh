#!/usr/bin/env bash

# Copyright 2025 The kcp Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

cd $(dirname $0)/..
source hack/lib.sh

# export TEST_ASSET_SHARDED_TEST_SERVER="$(UGET_PRINT_PATH=absolute make --no-print-directory install-sharded-test-server)"
# export TEST_ASSET_KCP="$(UGET_PRINT_PATH=absolute make --no-print-directory install-kcp)"
# export TEST_ASSET_KCP_FRONT_PROXY="$(UGET_PRINT_PATH=absolute make --no-print-directory install-kcp-front-proxy)"
# export TEST_ASSET_CACHE_SERVER="$(UGET_PRINT_PATH=absolute make --no-print-directory install-cache-server)"

export TEST_KCP_NUM_SHARDS="${NUM_SHARDS:-1}"
export CGO_ENABLED=0

# TODO: remove once released artifacts are available upstream
export TEST_ASSET_SHARDED_TEST_SERVER="$(realpath "$(pwd)/_tools/sharded-test-server")"
export TEST_ASSET_KCP="$(realpath "$(pwd)/_tools/kcp")"
export TEST_ASSET_KCP_FRONT_PROXY="$(realpath "$(pwd)/_tools/kcp-front-proxy")"
export TEST_ASSET_CACHE_SERVER="$(realpath "$(pwd)/_tools/cache-server")"

go_test unit_tests -short -tags "unit" -timeout 20m -race -v ./...

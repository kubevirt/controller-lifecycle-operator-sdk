#!/bin/bash
#
# Copyright 2018-2020 Red Hat, Inc.
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

# Run functional tests
export KUBECONFIG=$(./cluster/kubeconfig.sh)
export KUBECTL=${KUBECTL:-'./cluster/kubectl.sh'}
NAMESPACE=${NAMESPACE:-'kubevirt'}


echo "Using: "
echo "  KUBECTL: $KUBECTL"
echo "  KUBECONFIG: $KUBECONFIG"

go test ./tests/operator --v -timeout 30m -kubeconfig "$KUBECONFIG" -namespace "$NAMESPACE"
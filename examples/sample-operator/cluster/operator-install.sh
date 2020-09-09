#!/bin/bash
#
# Copyright 2018-2019 Red Hat, Inc.
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

set -ex

export VERSION=v0.0.0
export SAMPLE_CONFIG_STATUS=${SAMPLE_CONFIG_STATUS:-''}

./cluster/kubectl.sh create -f _out/sample-operator/${VERSION}/operator.yaml
./cluster/kubectl.sh wait deploy/sample-operator -n kubevirt --for=condition=Available --timeout=600s

# Create and wait for the controller
./cluster/kubectl.sh create -f _out/sample-operator/${VERSION}/sampleconfig_cr.yaml

# When `kubectl wait` will support `--ignore-not` found parameter this `if` can be removed.
if [[ "$(./cluster/kubectl.sh get deploy/http-server -n kubevirt 2>&1)" =~ "not found" ]]; then
    sleep 10
fi
./cluster/kubectl.sh wait deploy/http-server -n kubevirt --for=condition=Available --timeout=600s

if [ ! -z "$SAMPLE_CONFIG_STATUS" ]
then
  ./cluster/kubectl.sh wait sampleconfig/example-sampleconfig -n kubevirt --timeout=600s --for=condition=$SAMPLE_CONFIG_STATUS
fi

#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(readlink -e $(dirname "$BASH_SOURCE[0]")/..)"
DEPLOY_DIR="${DEPLOY_DIR:-${PROJECT_ROOT}/manifests}"
CONTAINER_PREFIX="${CONTAINER_PREFIX:-quay.io/kubevirt}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:-Always}"
OPERATOR_IMAGE="${OPERATOR_IMAGE:-sample-operator}"
SERVER_IMAGE="${SERVER_IMAGE:-sample-http-server}"

templates=$(cd ${PROJECT_ROOT}/templates && find . -type f -name "*.yaml.in")
for template in $templates; do
	infile="${PROJECT_ROOT}/templates/${template}"

	dir="$(dirname ${DEPLOY_DIR}/${template})"
	dir=${dir/VERSION/$VERSION}
	mkdir -p ${dir}

	file="${dir}/$(basename -s .in $template)"
	echo $file
	file=${file/VERSION/$VERSION}
  echo $file
	sed -e "s#{{CONTAINER_PREFIX}}#$CONTAINER_PREFIX#g" \
		  -e "s/{{IMAGE_TAG}}/$IMAGE_TAG/g" \
		  -e "s/{{IMAGE_PULL_POLICY}}/$IMAGE_PULL_POLICY/g" \
      -e "s/{{OPERATOR_IMAGE}}/$OPERATOR_IMAGE/g" \
      -e "s/{{SERVER_IMAGE}}/$SERVER_IMAGE/g" \
	$infile > $file
done
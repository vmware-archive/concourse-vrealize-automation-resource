# Copyright 2020 program was created by VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash
VERSION=$1
REPO_USERNAME=$2
REPO_PASSWORD=$3

# Validate arguments
if [ $# -eq 0 ]; then
  echo "No arguments passed. Image version number and docker repo account password are expected"
  exit 1
fi

# Repo details
LABEL="concourse-vra-rt"
REPO_PATH="registry.docker.io"
TAG="$LABEL:$VERSION"

# Login to repo
docker login --username $REPO_USERNAME --password $REPO_PASSWORD

# Build the image
docker build -t cmbucicd/$TAG -f ../docker/Dockerfile ../../.

# Push the image
docker push "cmbucicd/$TAG"

# Remove the local image
docker rmi "cmbucicd/$TAG"

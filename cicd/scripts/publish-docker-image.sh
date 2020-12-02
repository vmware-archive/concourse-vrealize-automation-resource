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
LABEL="concourse-vra-resource"
REPO_PATH="projects.registry.vmware.com/concourse-vra-resource"
TAG="$LABEL:$VERSION"

# Login to repo
docker login $REPO_PATH --username $REPO_USERNAME --password $REPO_PASSWORD

# Build the image
docker build -t $REPO_PATH/$TAG -f ../docker/Dockerfile ../../.

# Push the image
docker push $REPO_PATH/$TAG

# Remove the local image
docker rmi $REPO_PATH/$TAG

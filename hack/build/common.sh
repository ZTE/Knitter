#!/bin/bash


# Copyright 2018 ZTE Corporation. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


export BUILDPKG="github.com/ZTE/Knitterr"
export OUT_DIR="_output/bin"
export DEFAULT_SUBJECT=" The result of ac-test with build log in attachments,please pay your attention."
#export GOROOT=$(echo $GOROOT |sed 's/\/bin//g')
export GO="$GOROOT/bin"

FILEDIR=$(cd "$(dirname "$0")";pwd)
export GOPATH="$(echo "$FILEDIR" | awk -F "/src/github.com/ZTE/" '{print $1}')"
echo "filedir=${FILEDIR}  gopath=${GOPATH}"

export KNITTERPATH="$GOPATH/src/github.com/ZTE/Knitter"
export VERTYPE="release"
export GITHASH=`git log -1|grep commit | head -1 |sed 's/commit //g'`
export BUILDTIME=`date  +%Y-%m-%d.%H:%M:%S`
##export BRANCH_VERSION=`curl -s http://10.62.40.90/shell/get_zart_info | xargs -0 -I {} bash -c "branch_name=master;info=version;{}"`
export BRANCH_VERSION=""

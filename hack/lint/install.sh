#!/usr/bin/env bash
# Copyright 2019 Iguazio
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
#

set -e

echo Installing linters...

FORCE_INSTALL=false
OS_NAME=$(uname -s)
OS_NAME_LOWERCASE=$(echo "${OS_NAME}" | tr "[:upper:]" "[:lower:]")

if [[ -z "${BIN_DIR}" ]]; then
  BIN_DIR=$(pwd)/.bin
fi

echo "Creating bin directory: ${BIN_DIR}"
mkdir -p "${BIN_DIR}"

if [[ ! -f ${BIN_DIR}/impi ]] ; then
    echo "impi binary does not exist. Fetching and installing..."
    curl -sSfL --retry 3 https://api.github.com/repos/pavius/impi/releases/latest \
      | grep -i "browser_download_url.*impi.*${OS_NAME}" \
      | cut -d : -f 2,3 \
      | tr -d '"' \
      | tr -d '[:space:]' \
      | xargs curl -sSL --output "${BIN_DIR}"/impi
    chmod +x "${BIN_DIR}/impi"
    echo "impi installed in: ${BIN_DIR}/impi"
fi

if [[ $# -ne 0 && "$1" == "force" ]]
  then
    echo "Force install golangci-lint requested"
    FORCE_INSTALL=true
fi

if [[ $FORCE_INSTALL = true || ! -f ${BIN_DIR}/golangci-lint ]] ; then
    echo "golangci-lint binary does not exist or force install requested. Fetching and installing..."
    curl -sSfL --retry 3 https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${BIN_DIR}" v1.50.1
    echo "golangci-lint installed in: ${BIN_DIR}/golangci-lint"
fi

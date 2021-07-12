#!/usr/bin/env bash
set -e

# NOTE: RUN THAT from repo root

if [[ -z "${BIN_DIR}" ]]; then
  BIN_DIR=$(pwd)/.bin
fi

# TODO: resolve import issues and enable
#echo Verifying imports...
#
#${BIN_DIR}/impi \
#  --local github.com/iguazio/provazio \
#  --skip pkg/controller/apis \
#  --skip pkg/controller/client \
#  --scheme stdLocalThirdParty \
#  ./pkg/... ./cmd/...

echo "Linting @"$(pwd)"..."
${BIN_DIR}/golangci-lint run -v
echo Done.

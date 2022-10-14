# Copyright Meshery Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: tidy check-lint check-clean-cache tidy verify error

error:
	go run github.com/layer5io/meshkit/cmd/errorutil -d . analyze -i ./helpers -o ./helpers

check-lint:
	$(GOBIN)/golangci-lint run ./...

check-clean-cache:
	go clean
	$(GOBIN)/golangci-lint cache clean

tidy:
	go mod tidy
	gofmt -w .
	$(GOBIN)/goimports -w .

verify:
	go mod verify
	go vet ./...
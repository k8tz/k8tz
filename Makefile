# Copyright © 2021 Yonatan Kahana
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

.DEFAULT_GOAL := build

# Build Variables
BINARY_NAME ?= k8tz
OUT_DIR ?= build/
VERSION = 0.5.0
VERSION_SUFFIX ?= -beta0
TARGET=/usr/local/bin
INSTALLCMD=install -v $(OUT_DIR)$(BINARY_NAME) $(TARGET)
BUILD_FLAGS ?= \
	-ldflags="-s -w \
	-X '$(MODULE)pkg/version.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE)pkg/version.AppVersion=$(VERSION)' \
	-X '$(MODULE)pkg/version.VersionSuffix=$(VERSION_SUFFIX)'\
	-X '$(MODULE)pkg/version.ImageRepository=$(IMAGE_REPOSITORY)'"

MODULE = github.com/k8tz/k8tz/
GIT_COMMIT ?= $(shell git rev-parse HEAD | tr -d "\n")

# Docker Image Variables
IMAGE_REPOSITORY ?= quay.io/k8tz/k8tz
IMAGE ?= $(IMAGE_REPOSITORY):$(VERSION)$(VERSION_SUFFIX)
IMAGE_EXTRA_TAGS ?=
IMAGE_LABELS ?= \
	--label gitCommit=$(GIT_COMMIT) \
	--label version=$(VERSION)$(VERSION_SUFFIX)
IMAGE_TZDATA_VERSION ?= 2022a

tzdata:
		(cd tzdata && make TZDATA_VERSION=$(IMAGE_TZDATA_VERSION) clean import)

# Targets
install: compile
		if [ -w $(TARGET) ]; then \
		$(INSTALLCMD); else \
		sudo $(INSTALLCMD); fi

clean:
		rm -rfv "$(OUT_DIR)"

test:
		go test -v ./...

coverage-report:
		go test -coverprofile build/coverage-report.html ./...
		go tool cover -html build/coverage-report.html

build: compile # Alias
compile:
		CGO_ENABLED=0 \
		go build \
		-v \
		-o $(OUT_DIR)$(BINARY_NAME) \
		$(BUILD_FLAGS) \
		.

docker: docker-build # Alias
docker-build: compile tzdata
		docker build \
		-t $(IMAGE) \
		--build-arg BINARY_LOCATION=$(OUT_DIR)$(BINARY_NAME) \
		$(IMAGE_LABELS)	\
		.
		$(foreach tag, $(IMAGE_EXTRA_TAGS), docker tag $(IMAGE) $(IMAGE_REPOSITORY):$(tag);)

docker-push: docker-build
		docker push $(IMAGE)
		$(foreach tag, $(IMAGE_EXTRA_TAGS), docker push $(IMAGE_REPOSITORY):$(tag);)


helm: helm-package # Alias
helm-package: helm-lint
		@rm -rfv $(OUT_DIR)k8tz-*.tgz
		helm package \
		-d $(OUT_DIR) \
		--app-version $(VERSION)$(VERSION_SUFFIX) \
		charts/k8tz/

helm-lint:
		helm lint charts/k8tz/

helm-install: helm-package helm-uninstall
		helm install k8tz $(OUT_DIR)k8tz-*.tgz

helm-uninstall:
		@helm status k8tz 2>&1 > /dev/null && echo Uninstalling helm package... && helm uninstall k8tz || true

release: test compile docker helm

# Phony Targets
.PHONY: install clean build test tzdata coverage-report compile docker docker-build docker-push helm-lint helm helm-package helm-install helm-uninstall release

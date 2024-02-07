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

IMAGE_REPOSITORY ?= quay.io/k8tz/tzdata
TZDATA_VERSION ?= 2024a
IMAGE = $(IMAGE_REPOSITORY):$(TZDATA_VERSION)

clean:
		rm -rfv zoneinfo

build:
		docker build -t $(IMAGE) --build-arg tzdbversion=$(TZDATA_VERSION) .

publish: build
		docker push $(IMAGE)

.PHONY: clean build publish

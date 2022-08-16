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

ARG TZDATA_IMAGE=quay.io/k8tz/tzdata:2022c
FROM $TZDATA_IMAGE

ARG BINARY_LOCATION=k8tz
COPY $BINARY_LOCATION /

USER 1000
ENTRYPOINT ["/k8tz"]

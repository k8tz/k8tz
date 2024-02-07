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

FROM alpine:3.13 AS src

ARG tzdbversion=2024a
RUN \
    apk add build-base lzip && \
    wget -O tzdb.tar.lz https://data.iana.org/time-zones/releases/tzdb-$tzdbversion.tar.lz && \
    lzip -cd tzdb.tar.lz | tar -xf - && \
    mv tzdb-* tzdb && \
    cd tzdb && \
    mkdir build && \
    make TOPDIR=build install

FROM scratch
COPY --from=src /tzdb/build/usr/share/zoneinfo /usr/share/zoneinfo

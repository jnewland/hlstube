# Dockerfile Copyright 2020 Seth Vargo
#                      2021 Jesse Newland
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Specify the version of Go to use
FROM golang:1.20.1@sha256:52921e63cc544c79c111db1d8461d8ab9070992d9c636e1573176642690c14b5 as builder

# Install upx (upx.github.io) to compress the compiled binary
RUN apt-get update && apt-get -y install upx

# Turn on Go modules support and disable CGO
ENV GO111MODULE=on CGO_ENABLED=0

# Get deps
WORKDIR /src
COPY go.* .
RUN go mod download

# Compile the binary - the added flags instruct Go to produce a
# standalone binary
COPY *.go .
RUN find /src > /tmp/manifest && go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/app \
  .

# Strip any symbols - this is not a library
RUN strip /bin/app

# Compress the compiled binary
RUN upx -q -9 /bin/app

# Use a pre-baked image with most of the tools we need
FROM mikenye/youtube-dl:2022.03.08.2@sha256:4a84039bfd156063a4acd8dd0ad42308a93c224da122a563198742161522abc5
RUN apt-get update && apt-get -y install procps lsof
RUN addgroup --system appgroup && adduser --system app && adduser app appgroup

# Ensure yt-dlp is up to date
# renovate: datasource=pip depName=yt-dlp
ENV YOUTUBE_DL_VERSION=2023.2.17
RUN python3 -m pip install --upgrade --no-cache-dir --force-reinstall yt-dlp==${YOUTUBE_DL_VERSION}

# Runtime configuration
WORKDIR /home/app
USER app
ENTRYPOINT ["/bin/app"]

# Copy over the compiled binary from the first step
COPY --from=builder /bin/app /bin/app
COPY --from=builder /tmp/manifest /tmp/manifest

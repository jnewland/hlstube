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
FROM golang:1.18.3@sha256:b203dc573d81da7b3176264bfa447bd7c10c9347689be40540381838d75eebef as builder

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

FROM mikenye/youtube-dl:2022.02.04@sha256:584aae5eaa719b51a1579eb598a6b6eac58493346221a558dd9849c67d137006
RUN apt-get update && apt-get -y install procps lsof
RUN addgroup --system appgroup && adduser --system app && adduser app appgroup
WORKDIR /home/app
USER app
ENTRYPOINT ["/bin/app"]
# Copy over the compiled binary from the first step
COPY --from=builder /bin/app /bin/app
COPY --from=builder /tmp/manifest /tmp/manifest
# Specify the container's entrypoint as the binary

# Dockerfile Copyright 2020 Seth Vargo
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
FROM golang:1.16 as builder

# Install upx (upx.github.io) to compress the compiled action
RUN apt-get update && apt-get -y install upx

# Turn on Go modules support and disable CGO
ENV GO111MODULE=on CGO_ENABLED=0

# Copy all the files from the host into the container
WORKDIR /src
COPY . .

# Compile the action - the added flags instruct Go to produce a
# standalone binary
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/app \
  .

# Strip any symbols - this is not a library
RUN strip /bin/app

# Compress the compiled action
RUN upx -q -9 /bin/app


FROM alpine:20200917
RUN apk add --no-cache \
        ca-certificates \
        curl \
        dumb-init \
        ffmpeg \
        gnupg \
        python3 \
    && ln -s /usr/bin/python3 /usr/bin/python 
RUN curl -sL https://yt-dl.org/downloads/2021.01.16/youtube-dl -o /usr/local/bin/youtube-dl && chmod a+rx /usr/local/bin/youtube-dl
# Copy over the compiled action from the first step
COPY --from=builder /bin/app /bin/app
# Specify the container's entrypoint as the action
RUN addgroup -S appgroup && adduser -S app -G appgroup
WORKDIR /home/app
USER app
ENTRYPOINT ["/bin/app"]
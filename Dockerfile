# Copyright 2016 The Kubernetes Authors.
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

# Stage 1: Build the Go application
FROM golang:1.10.0 AS builder

# Fetch dependencies
RUN go get github.com/urfave/negroni \
    && go get github.com/xyproto/simpleredis/v2 \
    && go get github.com/gorilla/mux

WORKDIR /app

# Copy the source code into the container
COPY main.go .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Create a minimal container to run the application
FROM scratch

WORKDIR /app

# Copy the built executable from the previous stage
COPY --from=builder /app/main .

# Assuming you have static files in a 'public' directory
COPY ./public/index.html public/index.html
COPY ./public/script.js public/script.js
COPY ./public/style.css public/style.css

# Expose the port your application listens on
EXPOSE 3000

# Command to run the application
CMD ["./main"]

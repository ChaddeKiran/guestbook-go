  
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

FROM scratch

WORKDIR /app

# Copy the built executable from the previous stage
COPY --from=builder /app/main .

COPY ./public/index.html public/index.html
COPY ./public/script.js public/script.js
COPY ./public/style.css public/style.css

# Expose the port your application listens on
EXPOSE 3000

# Command to run the application
CMD ["./main"]

# syntax=docker/dockerfile:1

###############
# Build stage #
###############
FROM golang:1.22.1-bullseye as builder

WORKDIR /app

# Add go module files
COPY go.mod go.sum ./

# Add source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build
RUN go build -o /app/app ./cmd/wonderful


#################
# Runtime stage #
#################

FROM ubuntu:22.04

ARG API_PORT=8888

# install ca-certificates so we can perform requests to https endpoints
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/app /app/

EXPOSE ${API_PORT}

ENTRYPOINT /app/app -port ${API_PORT}

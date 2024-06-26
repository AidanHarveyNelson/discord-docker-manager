# syntax=docker/dockerfile:1
# Build step
FROM golang:1.22.3 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-manager

# Copy to run step
FROM ubuntu

RUN apt-get update && apt-get install -y curl
ENV DOCKER_FILTER="replace_me"
ENV GUILD_ID="replace_me"
ENV DISCORD_TOKEN="replace_me"
ENV REMOVE_COMMANDS=true

COPY --from=build /docker-manager /docker-manager

SHELL ["/bin/sh", "-c"]
CMD "/docker-manager" "-token" $DISCORD_TOKEN "-guid" $GUILD_ID "-filter" $DOCKER_FILTER "-rmcmd" $REMOVE_COMMANDS

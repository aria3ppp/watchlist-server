ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS build

# install migrate which will be used by entrypoint.sh to perform DB migration
ARG MIGRATE_VERSION
ADD https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz /tmp
RUN tar -xzf /tmp/migrate.linux-amd64.tar.gz -C /usr/local/bin 

WORKDIR /app

# copy module files first so that they don't need to be downloaded again if no change
COPY go.* ./
RUN go mod download
RUN go mod verify

# copy source files and build the binary
COPY . .
RUN CGO_ENABLED=0 go build -o server .


FROM alpine:latest
RUN apk --no-cache add \
    ca-certificates \
    bash \
    curl

WORKDIR /app/

COPY --from=build /usr/local/bin/migrate /usr/local/bin
COPY --from=build /app/migrations ./migrations/
COPY --from=build /app/config.yml .
COPY --from=build /app/server .
COPY --from=build /app/entrypoint.sh .

ENTRYPOINT ["./entrypoint.sh"]
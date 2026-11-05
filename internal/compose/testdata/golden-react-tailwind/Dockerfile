# Builds a small container image of shop: the build stage compiles a
# static binary with cgo disabled, and the runtime stage holds only that
# binary on a distroless image, without a shell or package manager, run as
# the non-root user 65532.
#
#   docker build --tag shop .
#   docker run --rm --publish 8080:8080 shop
#
# The cache mounts need BuildKit, the default builder of Docker 23 and later.

# Build the frontend; the build stage copies only its output.
FROM node:26-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.27 AS build
WORKDIR /src

# Download the modules in a layer of their own, so that changing the code
# does not download them again.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
# The SQLite database lives in /data, which the runtime stage cannot create:
# it has no shell.
RUN mkdir -p /out/data
COPY --from=frontend /src/internal/views/dist/app internal/views/dist/app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/shop ./cmd/shop

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
COPY --from=build /out/shop /usr/local/bin/shop
COPY --from=build --chown=65532:65532 /out/data /data
ENV DATABASE_URL=/data/shop.db
VOLUME /data
ENV PORT=8080
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/shop"]

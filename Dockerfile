ARG TAG=1.20

FROM ghcr.io/anglo-korean/go-builder:$TAG as build
FROM ghcr.io/anglo-korean/go-scratch-kafka:$TAG

COPY --from=build /app/app /app

##############################   
# Builder
##############################

FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

WORKDIR /cmd

ARG TARGETOS TARGETARCH

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
        -o /out/hachyboop cmd/*.go

##############################
# Runtime image
##############################

FROM alpine

COPY --from=build /out/hachyboop /bin

ENTRYPOINT [ "/bin/hachyboop" ]

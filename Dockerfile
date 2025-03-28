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

# Envvars

ENV HACHYBOOP_OBSERVER_ID=anonymous-hachyfriend
ENV HACHYBOOP_OBSERVER_REGION=analog-nowhere
ENV HACHYBOOP_RESOLVERS=8.8.8.8:53
ENV HACHYBOOP_QUESTIONS=hachyderm.io
ENV HACHYBOOP_S3_WRITER_ENABLED=false
ENV HACHYBOOP_S3_ENDPOINT=replace-me.local
ENV HACHYBOOP_S3_BUCKET=replace-me
ENV HACHYBOOP_S3_PATH=replace-me/with-something
ENV HACHYBOOP_S3_ACCESS_KEY_ID=replace-me
ENV HACHYBOOP_S3_SECRET_ACCESS_KEY=replace-me
ENV HACHYBOOP_LOCAL_WRITER_ENABLED=true
ENV HACHYBOOP_LOCAL_RESULTS_PATH=data

ENTRYPOINT [ "/bin/hachyboop" ]

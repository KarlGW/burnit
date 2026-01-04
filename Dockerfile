# golang:1.25.5-alpine3.23 SHA1 digest.
FROM --platform=$BUILDPLATFORM golang@sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb as builder

ARG TARGETOS
ARG TARGETARCH
ARG BIN=burnit

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

ENV USER=${BIN}
ENV UID=10001

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nohome" \
  --no-create-home \
  --shell "/sbin/nologin" \
  --uid "${UID}" \
  "${USER}"


FROM scratch

ARG TARGETOS
ARG TARGETARCH
ARG BIN=burnit
ARG PORT=3000

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY ./build/$TARGETOS/$TARGETARCH/${BIN} /${BIN}

EXPOSE ${PORT}

USER ${BIN}:${BIN}

ENTRYPOINT [ "/burnit" ]

# golang:1.25.4-alpine3.22 SHA digest.
FROM --platform=$BUILDPLATFORM golang@sha256:d3f0cf7723f3429e3f9ed846243970b20a2de7bae6a5b66fc5914e228d831bbb AS builder

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

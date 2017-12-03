FROM golang:1.9

COPY . /go/src/github.com/arnaudmz/kaos

WORKDIR /go/src/github.com/arnaudmz/kaos/cmd/kaos-operator

RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch

COPY --from=0 /go/src/github.com/arnaudmz/kaos/cmd/kaos-operator/kaos-operator /

LABEL app.language=golang app.name=kaos-operator

ENTRYPOINT ["/kaos-operator", "-logtostderr",  "-v=2"]

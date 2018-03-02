FROM golang:1.9

COPY . /go/src/github.com/arnaudmz/kaos

WORKDIR /go/src/github.com/arnaudmz/kaos/cmd/kaos-operator

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=0.4 -X main.timestamp=1520033373"

FROM scratch

COPY --from=0 /go/src/github.com/arnaudmz/kaos/cmd/kaos-operator/kaos-operator /

LABEL app.language=golang app.name=kaos-operator

EXPOSE 8080

ENTRYPOINT ["/kaos-operator", "-logtostderr",  "-v=2"]

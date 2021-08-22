FROM golang:1.16-alpine as builder

WORKDIR /build 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pod-annotator -ldflags="-w -s" . 

FROM scratch

COPY --from=builder /build/pod-annotator /usr/bin/pod-annotator 

ENTRYPOINT ["pod-annotator"]
FROM golang:1.17-alpine AS build
COPY main.go /go/src/
WORKDIR /go/src/
RUN GO111MODULE=off CGO_ENABLED=0 go build -o /tmp/httpserver

FROM scratch
COPY --from=build /tmp/httpserver /httpserver
EXPOSE 80
ENTRYPOINT ["/httpserver"]

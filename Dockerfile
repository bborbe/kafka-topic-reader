FROM golang:1.25.2 AS build
COPY . /workspace
WORKDIR /workspace
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -ldflags "-s" -a -installsuffix cgo -o /main
CMD ["/bin/bash"]

FROM alpine:3.22 AS alpine
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=build /main /main
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/main"]

FROM golang:1.14-alpine AS build
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
WORKDIR /src/
COPY main.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/app

FROM scratch
COPY --from=build /bin/app /bin/app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/app"]

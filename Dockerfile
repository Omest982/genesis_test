FROM golang:1.22-alpine as build-stage

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY . .

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /awesomeProject .

#
# final build stage
#
FROM scratch

# Copy ca-certs for app web access
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /awesomeProject /awesomeProject
COPY --from=build-stage /app/.env .

ENTRYPOINT ["/awesomeProject"]
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/waf-admin ./cmd/waf-admin

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/waf-admin /usr/local/bin/waf-admin
COPY configs/config.example.yaml /app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/waf-admin","-config","/app/config.yaml"]

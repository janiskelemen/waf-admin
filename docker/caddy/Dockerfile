FROM caddy:builder AS builder
RUN xcaddy build \
    --with github.com/corazawaf/coraza-caddy \
    --with github.com/caddyserver/caddy/v2/modules/caddyhttp/tracing \
    --with github.com/stardothosting/caddy-altcha

FROM caddy:latest
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

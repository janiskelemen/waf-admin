FROM caddy:builder AS builder
RUN xcaddy build \
    --with github.com/corazawaf/coraza-caddy/v2 \
    --with github.com/caddyserver/caddy/v2/modules/caddyhttp/tracing \
    --with github.com/stardothosting/caddy-altcha

FROM caddy:2
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

FROM caddy:builder AS builder
COPY geoip/ /tmp/geoip/
RUN xcaddy build \
    --with github.com/corazawaf/coraza-caddy/v2 \
    --with github.com/caddyserver/caddy/v2/modules/caddyhttp/tracing \
    --with github.com/stardothosting/caddy-altcha \
    --with github.com/janiskelemen/waf-admin/docker/caddy/geoip=/tmp/geoip

FROM caddy:2
ADD https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb /usr/share/GeoIP/GeoLite2-Country.mmdb
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

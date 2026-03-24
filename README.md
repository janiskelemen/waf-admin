# waf-admin (OSS)

Admin API to manage **Caddy + Coraza** WAF sites & rules.
Validates via `caddy validate` and hot-reloads via Caddy Admin API (UNIX socket).
Includes optional daily S3 backups.

## Quick start
```bash
make run
```

## GeoIP Country Lookup

The Caddy image ships with the [coraza-geoip](https://github.com/corazawaf/coraza-geoip) plugin and a bundled [GeoLite2-Country](https://github.com/P3TERX/GeoLite.mmdb) database. This enables the `@geoLookup` operator in Coraza rules.

### Example rules

```
SecRule REMOTE_ADDR "@geoLookup" "id:1,phase:1,pass,nolog"
SecRule GEO:country_code "@pm CN RU" "id:2,phase:1,deny,status:403,msg:'Blocked country'"
```

The first rule performs the lookup and populates `GEO:country_code`, `GEO:country_name`, `GEO:continent_code`, and `GEO:country_continent`. The second rule matches against the result.

### Daily database updates

Enable the `geoip` section in your config to auto-update the database daily:

```yaml
geoip:
  enabled: true
  daily: "04:00"
  databaseURL: "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb"
  databaseDir: "/usr/share/GeoIP"
```

After downloading a new database, waf-admin stops Caddy via the admin socket. Docker's restart policy (`condition: any`) brings Caddy back up with the fresh database.

### Environment variables (Caddy container)

| Variable | Default | Description |
|---|---|---|
| `GEOIP_DB_PATH` | `/usr/share/GeoIP/GeoLite2-Country.mmdb` | Path to the `.mmdb` file |
| `GEOIP_DB_TYPE` | `country` | Database type: `country` or `city` |

## API
See `internal/api/openapi.yaml`.

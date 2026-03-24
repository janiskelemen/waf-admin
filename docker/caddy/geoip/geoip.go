package geoip

import (
	"log"
	"os"

	geo "github.com/corazawaf/coraza-geoip"
)

const defaultDBPath = "/usr/share/GeoIP/GeoLite2-Country.mmdb"

func init() {
	dbPath := os.Getenv("GEOIP_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}
	dbType := os.Getenv("GEOIP_DB_TYPE")
	if dbType == "" {
		dbType = "country"
	}
	if _, err := os.Stat(dbPath); err != nil {
		log.Printf("geoip: database not found at %s, skipping registration", dbPath)
		return
	}
	if err := geo.RegisterGeoDatabaseFromFile(dbPath, dbType); err != nil {
		log.Fatalf("geoip: failed to register database: %v", err)
	}
	log.Printf("geoip: registered %s database from %s", dbType, dbPath)
}

package dbs

import (
	"database/sql"
	"log"
	"time"

	"github.com/vkuznet/dbs2go/utils"
)

// MigrationProcessTimeout defines migration process timeout
var MigrationProcessTimeout int

// MigrationServerInterval defines migration process timeout
var MigrationServerInterval int

// MigrationDB points to migration DB
var MigrationDB *sql.DB

// MigrationServer represent migration server.
// it accepts migration process timeout used by ProcessMigration API and
// exit channel
func MigrationServer(interval, timeout int, ch <-chan bool) {
	log.Println("Start migration server")
	api := API{Api: "ProcessMigration"}

	for {
		select {
		case v := <-ch:
			if v == true {
				log.Println("Received notification to stop migration server")
				return
			}
		default:
			time.Sleep(time.Duration(interval) * time.Second)
			// look-up all available migration requests
			records, err := MigrationRequests(-1)
			if err != nil {
				log.Printf("fail to fetch migration records from %s, error %v", MigrateURL, err)
				continue
			}
			if utils.VERBOSE > 0 {
				log.Printf("found %d migration requests", len(records))
			}
			for _, r := range records {
				if utils.VERBOSE > 0 {
					log.Printf("process %+v", r)
				}
				params := make(map[string]interface{})
				params["migration_request_url"] = r.MIGRATION_URL
				params["migration_request_id"] = r.MIGRATION_REQUEST_ID
				api.Params = params
				//                 api := API{
				//                     Params: params,
				//                     Api:    "ProcessMigration",
				//                 }
				log.Printf("start new migration process with %+v", params)
				api.ProcessMigration(timeout, false)
				params = nil
			}
		}
	}
	log.Println("Exit migration server")
}

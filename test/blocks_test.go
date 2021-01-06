package main

import (
	"log"
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vkuznet/dbs2go/dbs"
)

// TestBlocksStatement API
func TestBlocksStatement(t *testing.T) {
	// initialize DB for testing
	db := initDB(true) // init DB with dryRun mode
	defer db.Close()
	var api dbs.API
	params := make(dbs.Record)                                // the Record is map[string]interface{}
	params["logical_file_name"] = []string{"/path/file.root"} // pass params as list as in HTTP
	var w http.ResponseWriter
	w = StdoutWriter("")
	log.Println("Test Blocks API statement with logical_file_name parameter")
	_, err := api.Blocks(params, w)
	if err != nil {
		t.Errorf("Fail in Blocks, error %v\n", err)
	}

	params = make(dbs.Record)
	params["dataset"] = []string{"/a/b/c"}
	log.Println("Test Blocks API statement with dataset parameter")
	_, err = api.Blocks(params, w)
	if err != nil {
		t.Errorf("Fail in Blocks, error %v\n", err)
	}
}
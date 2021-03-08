package dbs

import (
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/vkuznet/dbs2go/utils"
)

// FileOutputModConfigs DBS API
func (API) FileOutputModConfigs(params Record, w http.ResponseWriter) (int64, error) {
	var args []interface{}

	// get SQL statement from static area
	stm := getSQL("file_output_mod_configs")

	// use generic query API to fetch the results from DB
	return executeAll(w, stm, args...)
}

// FileOutputModConfigs
type FileOutputModConfigs struct {
	FILE_OUTPUT_CONFIG_ID int64 `json:"file_output_config_id"`
	FILE_ID               int64 `json:"file_id"`
	OUTPUT_MOD_CONFIG_ID  int64 `json:"output_mod_config_id"`
}

// Insert implementation of FileOutputModConfigs
func (r *FileOutputModConfigs) Insert(tx *sql.Tx) error {
	var tid int64
	var err error
	if r.FILE_OUTPUT_CONFIG_ID == 0 {
		if DBOWNER == "sqlite" {
			tid, err = LastInsertId(tx, "FILE_OUTPUT_MOD_CONFIGS", "file_output_config_id")
			r.FILE_OUTPUT_CONFIG_ID = tid + 1
		} else {
			tid, err = IncrementSequence(tx, "SEQ_FC")
			r.FILE_OUTPUT_CONFIG_ID = tid
		}
		if err != nil {
			return err
		}
	}
	// get SQL statement from static area
	stm := getSQL("insert_file_output_mod_configs")
	if DBOWNER == "sqlite" {
		stm = getSQL("insert_file_output_mod_configs_sqlite")
	}
	if utils.VERBOSE > 0 {
		log.Printf("Insert FileOutputModConfigs\n%s\n%+v", stm, r)
	}
	_, err = tx.Exec(stm, r.FILE_OUTPUT_CONFIG_ID, r.FILE_ID, r.OUTPUT_MOD_CONFIG_ID)
	return err
}

// Validate implementation of FileOutputModConfigs
func (r *FileOutputModConfigs) Validate() error {
	return nil
}

// Decode implementation for FileOutputModConfigs
func (r *FileOutputModConfigs) Decode(reader io.Reader) (int64, error) {
	// init record with given data record
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println("fail to read data", err)
		return 0, err
	}
	err = json.Unmarshal(data, &r)

	//     decoder := json.NewDecoder(r)
	//     err := decoder.Decode(&rec)
	if err != nil {
		log.Println("fail to decode data", err)
		return 0, err
	}
	size := int64(len(data))
	return size, nil
}

// FileOutputModConfigRecord
type FileOutputModConfigRecord struct {
	ReleaseVersion    string `json:"release_version"`
	PsetHash          string `json:"pset_hash"`
	Lfn               string `json:"lfn"`
	AppName           string `json:"app_name"`
	OutputModuleLabel string `json:"output_module_label"`
	GlobalTag         string `json:"global_tag"`
}

// InsertFileOutputModConfigs DBS API
func (API) InsertFileOutputModConfigs(tx *sql.Tx, r io.Reader) (int64, error) {
	// read given input
	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Println("fail to read data", err)
		return 0, err
	}
	size := int64(len(data))
	var rec FileOutputModConfigRecord
	err = json.Unmarshal(data, &rec)
	if err != nil {
		log.Println("fail to decode data", err)
		return 0, err
	}

	// start transaction
	//     tx, err := DB.Begin()
	//     if err != nil {
	//         msg := fmt.Sprintf("unable to get DB transaction %v", err)
	//         return 0, errors.New(msg)
	//     }
	//     defer tx.Rollback()

	fid, err := getTxtID(tx, "FILES", "file_id", "logical_file_name", rec.Lfn)
	if err != nil {
		log.Println("unable to find file_id for", rec.Lfn)
		return 0, err
	}
	// find output module config id
	var args []interface{}
	var conds []string
	params := make(Record)
	params["logical_file_name"] = rec.Lfn
	params["app_name"] = rec.AppName
	params["pset_hash"] = rec.PsetHash
	params["output_module_label"] = rec.OutputModuleLabel
	params["global_tag"] = rec.GlobalTag
	conds, args = AddParam("logical_file_name", "FS.LOGICAL_FILE_NAME", params, conds, args)
	conds, args = AddParam("app_name", "A.APP_NAME", params, conds, args)
	conds, args = AddParam("pset_hash", "P.PSET_HASH", params, conds, args)
	conds, args = AddParam("output_module_label", "O.OUTPUT_MODULE_LABEL", params, conds, args)
	conds, args = AddParam("global_tag", "O.GLOBAL_TAG", params, conds, args)
	tmpl := make(Record)
	tmpl["Owner"] = DBOWNER
	stm, err := LoadTemplateSQL("outputconfigs", tmpl)
	if err != nil {
		return 0, err
	}
	stm = WhereClause(stm, conds)
	var oid int64
	err = tx.QueryRow(stm, args...).Scan(&oid)
	if err != nil {
		log.Printf("unable to find output_mod_config_id for %+v", rec)
		return 0, err
	}

	// init all foreign Id's in output config record
	var rrr FileOutputModConfigs
	rrr.FILE_ID = fid
	rrr.OUTPUT_MOD_CONFIG_ID = oid
	err = rrr.Insert(tx)
	if err != nil {
		return 0, err
	}

	// commit transaction
	//     err = tx.Commit()
	//     if err != nil {
	//         log.Println("faile to insert_outputconfigs_sqlite", err)
	//         return 0, err
	//     }
	return size, err
}

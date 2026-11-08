package masterdatastore_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure/masterdatastore"
	"railkeeper/backend/internal/inventory/masterdata"
)

func TestSymbolSeedFailureKeepsCallerTransactionAndAllowsLosslessRetry(t *testing.T) {
	db, _ := masterDataDatabase(t)
	db.SetMaxOpenConns(1)
	const original = ` {"description":"local artwork context","ecosCode":1} `
	canonical := []byte(`{"description":"local artwork context","ecosCode":1}`)
	digest := fmt.Sprintf("%x", sha256.Sum256(canonical))
	manifest, err := json.Marshal(map[string]any{"symbols": []map[string]any{
		{"key": "seed-a", "previousSHA256": digest, "ecosCode": 2},
		{"key": "seed-b", "previousSHA256": digest, "ecosCode": 2},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"seed-a", "seed-b"} {
		if _, err := db.Exec(`INSERT INTO master_data_entries
(id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
VALUES(?,'symbols',?,'Original label',0,77,NULL,?,'original-created','original-updated','bundled')`,
			"symbols:"+key, key, original); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_second_symbol BEFORE UPDATE OF metadata_json ON master_data_entries
WHEN OLD.key='seed-b' BEGIN SELECT RAISE(FAIL,'second symbol failed'); END`); err != nil {
		t.Fatal(err)
	}
	before := masterDataRows(t, db)
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	err = masterdata.SeedFunctionSymbolCodes(t.Context(), masterdatastore.SeedStoreFor(tx), "new-time", manifest)
	if err == nil || !strings.Contains(err.Error(), "second symbol failed") {
		t.Fatal("seed failure not observed", err)
	}
	var firstCode int
	if err := tx.QueryRow(`SELECT json_extract(metadata_json,'$.ecosCode') FROM master_data_entries WHERE key='seed-a' AND type='symbols'`).Scan(&firstCode); err != nil || firstCode != 2 {
		t.Fatal("seed did not retain partial work inside the still caller-owned transaction", firstCode, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, masterDataRows(t, db)) {
		t.Fatal("caller rollback lost original seed metadata or settings")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_second_symbol`); err != nil {
		t.Fatal(err)
	}
	retry, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = retry.Rollback() }()
	if err := masterdata.SeedFunctionSymbolCodes(t.Context(), masterdatastore.SeedStoreFor(retry), "new-time", manifest); err != nil {
		t.Fatal(err)
	}
	if err := retry.Commit(); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"seed-a", "seed-b"} {
		var metadata, label, created, updated string
		var active, order int
		var source *string
		if err := db.QueryRow(`SELECT metadata_json,label,active,sort_order,source_url,created_at,updated_at
FROM master_data_entries WHERE type='symbols' AND key=?`, key).Scan(&metadata, &label, &active, &order, &source, &created, &updated); err != nil {
			t.Fatal(err)
		}
		if metadata != `{"description":"local artwork context","ecosCode":2}` || label != "Original label" || active != 0 ||
			order != 77 || source != nil || created != "original-created" || updated != "new-time" {
			t.Fatal("seed retry changed operator settings or failed to correct the code", key)
		}
	}
}

package overviewdata

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUnreadablePurchaseRowDoesNotReturnPartialPriceSource(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE accessory_purchases(quantity, unit_price, currency);
INSERT INTO accessory_purchases VALUES(1,'100.00','EUR'), ('unreadable','200.00','EUR');`); err != nil {
		t.Fatal(err)
	}
	values, err := New(db).AccessoryPurchases(t.Context())
	if err == nil || values != nil {
		t.Fatal("unreadable row produced an acknowledged partial price source", values, err)
	}
}

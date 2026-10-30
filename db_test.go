package supadb

import (
	"testing"
)

func TestA(t *testing.T) {
	db, err := Open[string, string]("tmp/testdb.json", Options{AutoCreate: true, FileMode: 0644})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Put("key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set key-value pair: %v", err)
	}

	value, err := db.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if value != "value1" {
		t.Fatalf("Expected value 'value1', got '%s'", value)
	}

	err = db.Delete("key1")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = db.Get("key1")
	if err == nil {
		t.Fatalf("Expected error for getting deleted key, but got none")
	}

	db.Close()
	_, err = db.Get("key1")
	if err == nil {
		t.Fatalf("Expected error for getting key after closing DB, but got none")
	}
}

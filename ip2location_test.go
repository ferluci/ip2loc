package ip2loc

import (
	"log"
	"testing"
)

var diskDB *DB
var inMemoryDB *DB

func init() {
	var err error
	diskDB, err = OpenDB("/path/to/db")
	if err != nil {
		log.Fatal(err)
	}

	inMemoryDB, err = OpenInMemoryDB("/path/to/db")
	if err != nil {
		log.Fatal(err)
	}
}

func BenchmarkDBGetAllFromDisk(b *testing.B) {
	for n := 0; n < b.N; n++ {
		diskDB.GetAll("8.8.8.8")
	}
}

func BenchmarkDBGetAllFromMemory(b *testing.B) {
	for n := 0; n < b.N; n++ {
		inMemoryDB.GetAll("8.8.8.8")
	}
}

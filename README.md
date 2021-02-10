IP2Location Go Package
======================

This package provides a fast lookup from IP address by using IP2Location database. 

To use this package tou need a file based database available at IP2Location.com. 

There are two different approaches to work with DB:
* Read database segments from file on every function call.
* Read db from disk to memory and work with in-memory copy.

You can choose an approach that suits your needs c:

Benchmarks
=======
```
BenchmarkDBGetAllFromDisk
BenchmarkDBGetAllFromDisk-8     	   56499	     20792 ns/op
BenchmarkDBGetAllFromMemory
BenchmarkDBGetAllFromMemory-8   	  282685	      3805 ns/op
```

Installation
=======

```
go get github.com/Ferluci/ip2loc
```

Examples
=======

InMemory
------

```go
package main

import (
	"fmt"
	"github.com/Ferluci/ip2loc"
)

func main() {
	db, err := ip2loc.OpenInMemoryDB("./IP-COUNTRY-REGION-CITY-LATITUDE-LONGITUDE-ZIPCODE-TIMEZONE-ISP-DOMAIN-NETSPEED-AREACODE-WEATHER-MOBILE-ELEVATION-USAGETYPE.BIN")
	
	if err != nil {
		return
	}
	ip := "8.8.8.8"
	record, err := db.GetAll(ip)
	
	if err != nil {
		fmt.Print(err)
		return
	}
	ip2loc.PrintRecord(record)
}
```
Disk
------
```go
package main

import (
	"fmt"
	"github.com/Ferluci/ip2loc"
)

func main() {
	db, err := ip2loc.OpenDB("./IP-COUNTRY-REGION-CITY-LATITUDE-LONGITUDE-ZIPCODE-TIMEZONE-ISP-DOMAIN-NETSPEED-AREACODE-WEATHER-MOBILE-ELEVATION-USAGETYPE.BIN")
	
	if err != nil {
		return
	}
	defer db.Close()

	ip := "8.8.8.8"
	record, err := db.GetAll(ip)
	
	if err != nil {
		fmt.Print(err)
		return
	}
	ip2loc.PrintRecord(record)
}
```
Copyright
=========

Copyright (C) 2020 by IP2Location.com, support@ip2location.com

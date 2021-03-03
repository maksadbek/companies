# companies

TLDR:
1. Compile with `make`
2. Run server with `./bin/server -addr=localhost:8080 -file=./testdata/companies.csv`
3. Run Postgresql and create datbase and run migrations from `migrations/initial_up.sql`
4. Run client with `./bin/client -serverAddr=http://localhost:8080 -updateInterval=1m -dsn=postgresql://docker:password@localhost/docker?sslmode=disable`
5. See Server API reference to make some changes.

## Server

Server reads records from csv file and fills up its cache.
Exposes HTTP API to list, add and remove companies.

CLI arg reference:
```
max@asylum companies % ./bin/server  --help           
Usage of ./bin/server:
  -addr string
    	address (default ":8080")
  -file string
    	file with companies (default "testdata/companies.csv")
```

API:

### List
```
GET http://localhost:8080/
```

### Add
```
POST http://localhost:8080/add
Content-type: application/json

{
	"name": "google",
	"inn": "12312313",
	"phone": "897442323",
	"address": "Moscow",
	"individual": "Larry Page"
}
```

### Remove

```
DELETE http://localhost:8080/delete?id=google
DELETE http://localhost:8080/delete?id=inn
```

Where id can be name or INN.


## Client

Client runs periodically and synchronizes data in a server with a database.

Client CLI reference:
```
max@asylum companies % ./bin/client -h            
Usage of ./bin/client:
 -dsn string
    database source (default "postgresql://docker:password@localhost/docker?sslmode=disable")
 -serverAddr string
    server address (default "http://localhost:8080")
 -updateInterval duration
    update interval (default 1m0s)

max@asylum companies % make client && ./bin/client
cd src && go build -o ../bin/client ./cmd/client
2021/03/03 13:57:31 initialized companies from database, size: 4
2021/03/03 13:57:31 change: prev: &{intuit inn2 897442325 moscow bond, james bond false} new: &{intuit inn22 897442325 moscow bond, james bond false}
2021/03/03 13:57:31 updating database: actual: 4, add: 0, del: 0, change: 1
2021/03/03 13:57:31 change: &{Name:intuit INN:inn22 Phone:897442325 Address:moscow Individual:bond, james bond Removed:false}
2021/03/03 13:57:31 successfully updated database
2021/03/03 13:58:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 13:59:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 14:00:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 14:01:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 14:02:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 14:03:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused
2021/03/03 14:04:31 failed to update database: Get "http://localhost:8080/": dial tcp [::1]:8080: connect: connection refused

```


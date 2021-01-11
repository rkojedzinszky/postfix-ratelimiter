package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	"github.com/namsral/flag"
	"github.com/rkojedzinszky/postfix-sasl-exporter/server"
)

func main() {
	defaultrate := flag.Float64("default-rate", 1, "Default rate for policing (recipient/seconds)")
	defaultburst := flag.Float64("default-burst", 60, "Default burst for policing")
	driver := flag.String("dbdriver", "", "Database type for dynamic rate/burst lookups (mysql or postgresql)")
	dsn := flag.String("dbdsn", "", "Database DSN for dynamic rate/burst lookup")
	querystring := flag.String("querystring", "", "SQL Query returning dynamic (rate, burst) settings for a (local_part, domain) lookup")

	flag.Parse()

	lis, err := net.Listen("tcp", ":10028")
	if err != nil {
		log.Fatal(err)
	}

	var stmt *sql.Stmt
	if *driver != "" && *querystring != "" {
		db, err := sql.Open(*driver, *dsn)
		if err != nil {
			log.Fatal(err)
		}

		if err = db.Ping(); err != nil {
			log.Fatal(err)
		}

		stmt, err = db.Prepare(*querystring)
		if err != nil {
			log.Fatal(err)
		}
	}

	rd := &ratelimiter{
		defaultrate:  *defaultrate,
		defaultburst: *defaultburst,
		dynstmt:      stmt,
		users:        make(map[string]*tbf),
	}

	server.Run(context.Background(), lis, rd)
}

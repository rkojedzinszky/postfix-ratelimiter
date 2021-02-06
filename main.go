package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rkojedzinszky/postfix-sasl-exporter/server"
)

var version = "devel"

func main() {
	defaultrate := flag.Float64("default-rate", 1, "Default rate for policing (recipient/seconds)")
	defaultburst := flag.Float64("default-burst", 60, "Default burst for policing")
	driver := flag.String("dbdriver", "", "Database type for dynamic rate/burst lookups (mysql or postgres)")
	dsn := flag.String("dbdsn", "", "Database DSN for dynamic rate/burst lookup")
	querystring := flag.String("querystring", "", "SQL Query returning dynamic (rate, burst) settings for a (local_part, domain) lookup")
	policyListenAddress := flag.String("policy-listen-address", ":10028", "Postfix Policy listen address")
	webListenAddress := flag.String("web-listen-address", ":9028", "Exporter WEB listen address")

	flag.Parse()

	policyListener, err := net.Listen("tcp", *policyListenAddress)
	if err != nil {
		log.Fatal(err)
	}

	webListener, err := net.Listen("tcp", *webListenAddress)
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

	log.Printf("postfix-ratelimiter version %s starting", version)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()

		server.Run(ctx, policyListener, rd)
	}()

	go func() {
		defer wg.Done()

		webListen(webListener)
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)
	<-sigchan
	cancel()
	policyListener.Close()
	webListener.Close()

	wg.Wait()
}

func webListen(l net.Listener) {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{Handler: mux}

	server.Serve(l)
}

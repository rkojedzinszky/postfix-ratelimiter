package main

import (
	"context"
	"database/sql"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rkojedzinszky/postfix-sasl-exporter/pkg/server"
)

var (
	rejected = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "postfix",
		Subsystem: "ratelimiter",
		Name:      "rejects",
		Help:      "Rejected recipient count",
	}, []string{"sasl_username"})
)

func init() {
	prometheus.MustRegister(rejected)
}

type ratelimiter struct {
	defaultrate  float64
	defaultburst float64

	dynstmt *sql.Stmt

	mu    sync.Mutex
	users map[string]*tbf
}

func (r *ratelimiter) Handle(ctx context.Context, req *server.Request) string {
	if req.SaslUsername == "" {
		return server.DUNNO
	}

	saslusername := strings.ToLower(req.SaslUsername)

	var user, domain string
	splitted := strings.SplitN(saslusername, "@", 2)
	user = splitted[0]
	if len(splitted) > 1 {
		domain = splitted[1]
	}

	t := r.gettbf(user, domain)

	recipientCount, err := strconv.ParseFloat(req.RecipientCount, 64)
	if err != nil {
		return server.REJECT + " Internal error occured"
	}

	rate, burst := r.defaultrate, r.defaultburst
	if r.dynstmt != nil {
		row := r.dynstmt.QueryRowContext(ctx, user, domain)
		var dynrate, dynburst sql.NullFloat64

		err := row.Scan(&dynrate, &dynburst)
		switch err {
		case nil:
			if dynrate.Valid {
				rate = dynrate.Float64
			} else {
				rate = math.Inf(1)
			}

			if dynburst.Valid {
				burst = dynburst.Float64
			} else {
				burst = math.Inf(1)
			}
		case sql.ErrNoRows:
		default:
			log.Print("SQL returned error:", err)
		}
	}

	if !t.get(rate, burst, recipientCount) {
		rejected.WithLabelValues(saslusername).Add(recipientCount)

		return server.REJECT + " Rate-limit exceeded"
	}

	return server.DUNNO
}

func (r *ratelimiter) gettbf(user, domain string) *tbf {
	key := user + "@" + domain

	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.users[key]
	if !ok {
		t = &tbf{
			ts:       time.Time{},
			capacity: r.defaultburst,
		}
		r.users[key] = t
	}

	return t
}

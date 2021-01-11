# postfix-ratelimiter

A simple policy daemon which rate-limits sending mails based on sasl_username. Recipient count is rate-limited.

## Usage

```shell
$ ./postfix-ratelimiter -h
Usage of ./postfix-ratelimiter:
  -dbdriver="": Database type for dynamic rate/burst lookups (mysql or postgresql)
  -dbdsn="": Database DSN for dynamic rate/burst lookup
  -default-burst=60: Default burst for policing
  -default-rate=1: Default rate for policing (recipient/seconds)
  -querystring="": SQL Query returning dynamic (rate, burst) settings for a (local_part, domain) pair
```

The policy daemon will create a token-bucket rate-limiter for each sasl authenticated user. Rate-limits against unauthenticated mails are not enforced. The token-buckets will have `default-rate` rate and `default-burst` burst settings.

The daemon listens on `:10028`.

A sample postfix configuration might look like:
```
smtpd_data_restrictions = ...
    check_policy_service { inet:127.0.0.1:10028, { default_action=dunno } },
    permit
```

## Dynamic rate/burst

You can specify a database to look up rate/burst settings dynamically.

For this, you'll have to specify `-dbdriver` (mysql or postgres), the DSN the driver uses ([mysql](https://github.com/go-sql-driver/mysql#dsn-data-source-name) or [postgres](https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters)), and the `-querystring` which must return one row with two columns: `(rate, burst)`. The querystring is prepared, and during lookup, `(local_part, domain)` is passed as an argument.

`Null` returned for any of the columns is treated as `Infinity`.

Example with postgresql, minimal schema:
```sql
create table rate_limits (
    local_part varchar(128),
    domain varchar(128),
    rate float,
    burst float,
    primary key (local_part, domain)
);
```

```shell
$ ./postfix-ratelimiter -dbdriver=postgres -dbdsn "postgres://localhost?sslmode=disable" -querystring='select rate, burst from rate_limits where local_part = $1 and domain = $2'
```

## Containerized deployment

Configuration arguments are parsed using [flag](github.com/namsral/flag), so they can be specified using capitalized environment variables too. For example, you can start the app as:

```shell
$ docker run -d --restart=always -p 10028:10028 -e DEFAULT_RATE=2 -e DEFAULT_BURST=100 ghcr.io/rkojedzinszky/postfix-ratelimiter
```

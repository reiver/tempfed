package dbsrv

import (
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"tempfed/cfg"
	"tempfed/lib/db"
)

var Conn driver.Conn

func init() {
	var host string = cfg.ClickHouseHost()

	var err error
	Conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDataBase(),
			Username: cfg.ClickHouseUserName(),
			Password: cfg.ClickHousePassWord(),
		},
		DialTimeout: 30 * time.Second,
	})
	if nil != err {
		panic(fmt.Errorf("failed to connect to ClickHouse database server (%q): %w", host, err))
	}

	err = libdb.Init(Conn)
	if nil != err {
		panic(fmt.Errorf("failed to create data_nodes table in ClickHouse database server (%q): %w", host, err))
	}
}

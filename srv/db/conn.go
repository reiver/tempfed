package dbsrv

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"tempfed/cfg"
)

var Conn driver.Conn

func init() {
	var err error
	Conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouseHost()},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDataBase(),
			Username: cfg.ClickHouseUserName(),
			Password: cfg.ClickHousePassWord(),
		},
	})
	if nil != err {
		panic(fmt.Errorf("failed to connect to ClickHouse database: %w", err))
	}

	err = Conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS data_nodes (
			-- identity
			id             String,
			type           Array(String),

			-- core Object fields
			name           Nullable(String),
			summary        Nullable(String),
			content        Nullable(String),
			media_type     Nullable(String),
			url            Nullable(String),

			-- authorship & addressing
			attributed_to  Array(String),
			to             Array(String),
			cc             Array(String),
			audience       Array(String),

			-- timestamps
			published      Nullable(DateTime64(3)),
			updated        Nullable(DateTime64(3)),
			start_time     Nullable(DateTime64(3)),
			end_time       Nullable(DateTime64(3)),
			duration       Nullable(String),

			-- threading / context
			in_reply_to    Array(String),
			also_known_as  Array(String),
			moved_to       Nullable(String),

			-- Link-specific fields
			href           Nullable(String),
			hreflang       Nullable(String),
			rel            Nullable(String),
			height         Nullable(UInt32),
			width          Nullable(UInt32),

			-- ingestion metadata
			db_when_created DateTime DEFAULT now()
		) ENGINE = ReplacingMergeTree(db_when_created)
		ORDER BY id
		TTL db_when_created + INTERVAL 30 DAY DELETE
		SETTINGS ttl_only_drop_parts = 1
	`)
	if nil != err {
		panic(fmt.Errorf("failed to create data_nodes table: %w", err))
	}
}

//@TODO: deal with "defer Conn.Close()"

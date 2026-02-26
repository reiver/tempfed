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
			-- internal
			db_row_id       UUID     DEFAULT generateUUIDv7(),
			db_row_when_created DateTime DEFAULT now(),

			-- identity
			as_id             String,
			as_type           Array(String),

			-- core Object fields
			as_name           Nullable(String),
			as_summary        Nullable(String),
			as_content        Nullable(String),
			as_media_type     Nullable(String),
			as_url            Nullable(String),

			-- authorship & addressing
			as_attributed_to  Array(String),
			as_to             Array(String),
			as_cc             Array(String),
			as_audience       Array(String),

			-- timestamps
			as_published      Nullable(DateTime64(3)),
			as_updated        Nullable(DateTime64(3)),
			as_start_time     Nullable(DateTime64(3)),
			as_end_time       Nullable(DateTime64(3)),
			as_duration       Nullable(String),

			-- tags
			hashtags          Array(String),

			-- threading / context
			as_in_reply_to    Array(String),
			as_also_known_as  Array(String),
			as_moved_to       Nullable(String),

			-- Link-specific fields
			as_href           Nullable(String),
			as_hreflang       Nullable(String),
			as_rel            Nullable(String),
			as_height         Nullable(UInt32),
			as_width          Nullable(UInt32)

		) ENGINE = ReplacingMergeTree(db_row_when_created)
		ORDER BY as_id
		TTL db_row_when_created + INTERVAL 30 DAY DELETE
		SETTINGS ttl_only_drop_parts = 1
	`)
	if nil != err {
		panic(fmt.Errorf("failed to create data_nodes table: %w", err))
	}
}

//@TODO: deal with "defer Conn.Close()"

package cfg

import (
	"codeberg.org/reiver/go-env"
)

func ClickHouseDataBase() string {
	return env.GetElse[string]("CLICKHOUSE_DATABASE", "default")
}

func ClickHouseHost() string {
	return env.GetElse[string]("CLICKHOUSE_HOST", "localhost:9000")
}

func ClickHousePassWord() string {
	return env.GetElse[string]("CLICKHOUSE_PASSWORD", "")
}

func ClickHouseUserName() string {
	return env.GetElse[string]("CLICKHOUSE_USERNAME", "default")
}

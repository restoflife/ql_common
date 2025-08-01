package db

type XORMConfigLite struct {
	Driver  string `toml:"driver" yaml:"driver" json:"driver"`       // 数据库驱动类型，例如：mysql、postgres、sqlite3
	Dsn     string `toml:"dsn" yaml:"dsn" json:"dsn"`                // 主数据库 DSN（数据源名称），如：user:pass@tcp(127.0.0.1:3306)/dbname
	MaxIdle int    `toml:"max_idle" yaml:"max_idle" json:"max_idle"` // 连接池中最大空闲连接数
	MaxOpen int    `toml:"max_open" yaml:"max_open" json:"max_open"` // 连接池中最大打开连接数
	ShowSql bool   `toml:"show_sql" yaml:"show_sql" json:"show_sql"` // 是否在日志中输出执行的 SQL 语句

	Slave []struct {
		Dsn string `toml:"dsn" yaml:"dsn" json:"dsn"` // 从库 DSN，支持多个从库，用于读写分离配置
	} `toml:"slave" yaml:"slave" json:"slave"`

	MaxLife         int  `toml:"max_life" yaml:"max_life" json:"max_life"`                      // 连接的最大生命周期（单位：秒），超时将重连
	Synchronization bool `toml:"synchronization" yaml:"synchronization" json:"synchronization"` // 是否自动同步数据库结构（建表、更新字段）
}

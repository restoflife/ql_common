package db

import "time"

// PoolConfig overrides one engine pool; Lifetime preserves sub-second precision.
type PoolConfig struct {
	MaxIdle, MaxOpen int
	Lifetime         time.Duration
}

type XORMConfigLite struct {
	// Optional per-engine overrides. SlavePools must match Slave when supplied.
	MasterPool *PoolConfig  `yaml:"-" json:"-" toml:"-"`
	SlavePools []PoolConfig `yaml:"-" json:"-" toml:"-"`
	Driver     string       `toml:"driver" yaml:"driver" json:"driver"`
	Dsn        string       `toml:"dsn" yaml:"dsn" json:"dsn"`
	MaxIdle    int          `toml:"max_idle" yaml:"max_idle" json:"max_idle"`
	MaxOpen    int          `toml:"max_open" yaml:"max_open" json:"max_open"`
	ShowSql    bool         `toml:"show_sql" yaml:"show_sql" json:"show_sql"`

	Slave []struct {
		Dsn string `toml:"dsn" yaml:"dsn" json:"dsn"`
	} `toml:"slave" yaml:"slave" json:"slave"`

	MaxLife         int  `toml:"max_life" yaml:"max_life" json:"max_life"`
	Synchronization bool `toml:"synchronization" yaml:"synchronization" json:"synchronization"`
}

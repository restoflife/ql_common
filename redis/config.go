package redis

import (
	"crypto/tls"
	"time"
)

type Config struct {
	// Additional pool and TLS settings shared by all modes; do not mutate TLSConfig after startup.
	MaxActiveConns  int           `yaml:"max_active_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `yaml:"-"`
	TLSConfig       *tls.Config   `yaml:"-" json:"-" toml:"-"`
	Mode            string        `toml:"mode" yaml:"mode" json:"mode"`
	Addr            string        `toml:"addr" yaml:"addr" json:"addr"`
	Password        string        `toml:"password" yaml:"password" json:"password"`
	DB              int           `toml:"db" yaml:"db" json:"db"`
	MasterName      string        `toml:"master_name" yaml:"master_name" json:"master_name"`
	Slaves          []string      `toml:"slaves" yaml:"slaves" json:"slaves"`
	PoolSize        int           `toml:"pool_size" yaml:"pool_size" json:"pool_size"`
	MinIdle         int           `toml:"min_idle" yaml:"min_idle" json:"min_idle"`
}

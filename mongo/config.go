package mongo

import "time"

type Config struct {
	MaxConnecting uint64        `yaml:"max_connecting"`
	Timeout       time.Duration `yaml:"-"`
	URI           string        `toml:"uri" yaml:"uri" json:"uri"`
	MaxPoolSize   uint64        `toml:"max_pool_size" yaml:"max_pool_size" json:"max_pool_size"`
	MinPoolSize   uint64        `toml:"min_pool_size" yaml:"min_pool_size" json:"min_pool_size"`
	CACertFile    string        `toml:"ca_cert_file" yaml:"ca_cert_file" json:"ca_cert_file"`
	Username      string        `toml:"username" yaml:"username" json:"username"`
	Password      string        `toml:"password" yaml:"password" json:"password"`
	Database      string        `toml:"database" yaml:"database" json:"database"`
	AuthSource    string        `toml:"auth_source" yaml:"auth_source" json:"auth_source"`
}

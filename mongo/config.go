package mongo

type Config struct {
	URI         string `toml:"uri" yaml:"uri" json:"uri"`                               // MongoDB 连接 URI，例如：mongodb://127.0.0.1:27017
	MaxPoolSize uint64 `toml:"max_pool_size" yaml:"max_pool_size" json:"max_pool_size"` // 最大连接池大小（单位：连接数）
	MinPoolSize uint64 `toml:"min_pool_size" yaml:"min_pool_size" json:"min_pool_size"` // 最小连接池大小（单位：连接数）
	CACertFile  string `toml:"ca_cert_file" yaml:"ca_cert_file" json:"ca_cert_file"`    // CA 证书文件路径，用于 TLS/SSL 连接（可选）
	Username    string `toml:"username" yaml:"username" json:"username"`                // 用户名（可选，连接需要认证时使用）
	Password    string `toml:"password" yaml:"password" json:"password"`                // 密码（与用户名配合使用）
	Database    string `toml:"database" yaml:"database" json:"database"`                // 默认数据库名称（如 admin、test、your_db_name）
	AuthSource  string `toml:"auth_source" yaml:"auth_source" json:"auth_source"`       // 认证数据库名（用户名密码验证时使用）
}

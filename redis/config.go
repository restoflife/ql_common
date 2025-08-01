package redis

type Config struct {
	Mode       string   `toml:"mode" yaml:"mode" json:"mode"`                      // Redis 模式：standalone（单机）、sentinel（哨兵）、cluster（集群）
	Addr       string   `toml:"addr" yaml:"addr" json:"addr"`                      // Redis 地址，单机模式下使用，例如：127.0.0.1:6379
	Password   string   `toml:"password" yaml:"password" json:"password"`          // Redis 认证密码
	DB         int      `toml:"db" yaml:"db" json:"db"`                            // Redis 数据库编号（仅单机和哨兵模式有效）
	MasterName string   `toml:"master_name" yaml:"master_name" json:"master_name"` // 哨兵模式下主节点名称
	Slaves     []string `toml:"slaves" yaml:"slaves" json:"slaves"`                // 哨兵或集群模式下的节点地址列表
	PoolSize   int      `toml:"pool_size" yaml:"pool_size" json:"pool_size"`       // 最大连接池大小
	MinIdle    int      `toml:"min_idle" yaml:"min_idle" json:"min_idle"`          // 最小空闲连接数
}

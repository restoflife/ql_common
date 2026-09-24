package logger

const (
	// XORM defines the prefix of the log entry from XORM
	XORM = "[xorm]  "
	// GORM defines the prefix of the log entry from GORM
	GORM = "[gorm]  "
	// SQL defines the prefix of the log entry from SQL
	SQL = "[sql]  "
	// REDIS defines the prefix of Redis command log entries.
	REDIS = "[redis]  "
	// MONGO defines the prefix of MongoDB command and driver log entries.
	MONGO = "[mongo]  "
)

package start

// types only used in init phase

type AppConfig struct {
	Server AppConfigServer `toml:"app"`
	GPU    AppConfigGPU    `toml:"gpu"`
	Log    AppConfigLog    `toml:"log"`
	PG     AppConfigPG     `toml:"pg"`
	Redis  AppConfigRedis  `toml:"redis"`
	MQ     AppConfigMQ     `toml:"mq"`
}

type AppConfigServer struct {
	Port        int    `toml:"port"`
	SessionName string `toml:"session_name"`
	SessionKey  string `toml:"session_key"`
	AESKey      string `toml:"aes_key"`
}

type AppConfigGPU struct {
	Enabled bool `toml:"enabled"`
}

type AppConfigLog struct {
	InfoLogFilePath    string `toml:"info_log_file_path"`
	WarningLogFilePath string `toml:"warning_log_file_path"`
	ErrorLogFilePath   string `toml:"error_log_file_path"`
}

type AppConfigPG struct {
	Host            string `toml:"host"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	DBName          string `toml:"db_name"`
	SSLMode         string `toml:"ssl_mode"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

type AppConfigRedis struct {
	Host            string `toml:"host"`
	Port            string `toml:"port"`
	Password        string `toml:"password"`
	DB              int    `toml:"db"`
	PoolSize        int    `toml:"pool_size"`
	DialTimeout     int    `toml:"dial_timeout"`
	ReadTimeout     int    `toml:"read_timeout"`
	WriteTimeout    int    `toml:"write_timeout"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

type AppConfigMQ struct {
	Protocol string `toml:"protocol"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

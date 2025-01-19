package start

// the types where only use in init phase

type AppConfigPG struct {
	Host            string `toml:"Host"`
	User            string `toml:"User"`
	Password        string `toml:"Password"`
	DBName          string `toml:"DBName"`
	SSLMode         string `toml:"SSLMode"`
	MaxOpenConns    int    `toml:"Max_open_conns"`
	MaxIdleConns    int    `toml:"Max_idle_conns"`
	ConnMaxLifetime string `toml:"Conn_max_lifetime"`
}

type AppConfig struct {
	Server AppConfigServer `toml:"App"`
	GPU    AppConfigGPU    `toml:"GPU"`
	Log    AppConfigLog    `toml:"Log"`
	PG     AppConfigPG     `toml:"PG"`
}

type AppConfigServer struct {
	Port        int    `toml:"Port"`
	SessionName string `toml:"SessionName"`
	SessionKey  string `toml:"SessionKey"`
	AESKey      string `toml:"AESKey"`
}

type AppConfigGPU struct {
	Enabled bool `toml:"Enabled"`
}

type AppConfigLog struct {
	InfoLogFilePath    string `toml:"InfoLogFilePath"`
	WarningLogFilePath string `toml:"WarningLogFilePath"`
	ErrorLogFilePath   string `toml:"ErrorLogFilePath"`
}

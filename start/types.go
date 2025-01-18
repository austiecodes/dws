package start

// the types where only use in init phase

type PGConfig struct {
	Host            string `toml:"host"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	DBName          string `toml:"dbname"`
	SSLMode         string `toml:"sslmode"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifetime string `toml:"conn_max_lifetime"`
}

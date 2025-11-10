package configs

type DatabaseConfig struct {
	Host     string `env:"DB_HOST,required"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
	Port     int    `env:"DB_PORT,required"`
	SslMode  string `env:"DB_SSL_MODE,required"`
}

type MigrationConfig struct {
	Path string `env:"MIGRATION_PATH,required"`
}

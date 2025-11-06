package configs

type RedisConfig struct {
	// Addr     string `env:"REDIS_ADDR,required"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB,required"`
}

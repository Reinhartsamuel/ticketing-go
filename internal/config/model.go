package config

type Config struct {
	Server   Server
	Database Database
}

type Server struct {
	Host string
	Port string
}

type Database struct {
	Host string `env:"DB_HOST"`
	Port string `env:"DB_PORT"`
	Name string `env:"DB_NAME"`
	User string `env:"DB_USER"`
	Pass string `env:"DB_PASS"`
	Tz   string `env:"DB_TZ"`
}

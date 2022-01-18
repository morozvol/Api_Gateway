package apiserver

// Config ...
type Config struct {
	BindAddr         string `toml:"bind_addr"`
	LogLevel         string `toml:"log_level"`
	JWT_Key          string `toml:"jwt_key"`
	AuthService_Addr string `toml:"authservice_addr"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "debug",
	}
}

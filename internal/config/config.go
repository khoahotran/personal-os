package config

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"app"`
	DB struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"db"`
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
	} `mapstructure:"redis"`
	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
	} `mapstructure:"kafka"`
	Auth struct {
		JWTSecret     string        `mapstructure:"jwt_secret"`
		TokenLifespan time.Duration `mapstructure:"token_lifespan"`
	} `mapstructure:"auth"`
	Cloudinary struct {
		CloudName string `mapstructure:"cloud_name"`
		ApiKey    string `mapstructure:"api_key"`
		ApiSecret string `mapstructure:"api_secret"`
	} `mapstructure:"cloudinary"`
}

func LoadConfig() (cfg Config, err error) {

	err = godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found, use default.")
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err = viper.ReadInConfig(); err != nil {
		log.Printf("note: config.yaml not found, read .env only. Error: %v", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("db.dsn", "DB_DSN")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	viper.BindEnv("auth.token_lifespan", "TOKEN_LIFESPAN")

	viper.BindEnv("cloudinary.cloud_name", "CLOUDINARY_CLOUD_NAME")
	viper.BindEnv("cloudinary.api_key", "CLOUDINARY_API_KEY")
	viper.BindEnv("cloudinary.api_secret", "CLOUDINARY_API_SECRET")

	err = viper.Unmarshal(&cfg)
	return
}

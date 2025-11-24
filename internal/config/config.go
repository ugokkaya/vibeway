package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig              `mapstructure:"server"`
	Routes    []RouteConfig             `mapstructure:"routes"`
	Upstreams map[string]UpstreamConfig `mapstructure:"upstreams"`
	Security  SecurityConfig            `mapstructure:"security"`
}

type ServerConfig struct {
	Port             int    `mapstructure:"port"`
	Mode             string `mapstructure:"mode"`
	RequestTimeoutMs int    `mapstructure:"request_timeout_ms"`
}

type RouteConfig struct {
	Path        string   `mapstructure:"path"`
	Methods     []string `mapstructure:"methods"`
	Upstream    string   `mapstructure:"upstream"`
	Middlewares []string `mapstructure:"middlewares"`
}

type UpstreamConfig struct {
	URLs           []string             `mapstructure:"urls"`
	LoadBalancer   string               `mapstructure:"load_balancer"`
	TimeoutMs      int                  `mapstructure:"timeout_ms"`
	Retry          RetryConfig          `mapstructure:"retry"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type RetryConfig struct {
	Count     int `mapstructure:"count"`
	BackoffMs int `mapstructure:"backoff_ms"`
}

type CircuitBreakerConfig struct {
	FailureThreshold int `mapstructure:"failure_threshold"`
	ResetTimeoutMs   int `mapstructure:"reset_timeout_ms"`
}

type SecurityConfig struct {
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type JWTConfig struct {
	Issuer        string `mapstructure:"issuer"`
	Audience      string `mapstructure:"audience"`
	PublicKeyPath string `mapstructure:"public_key_path"`
	Secret        string `mapstructure:"secret"`
}

type RateLimitConfig struct {
	GlobalPerMinute int `mapstructure:"global_per_minute"`
	PerIP           int `mapstructure:"per_ip"`
	PerRoute        int `mapstructure:"per_route"`
}

var AppConfig Config

func LoadConfig(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
		if err := viper.Unmarshal(&AppConfig); err != nil {
			log.Printf("Failed to unmarshal updated config: %v", err)
		}
	})

	return nil
}

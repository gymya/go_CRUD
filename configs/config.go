package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// AppConfig 定義應用程式設定
// jwt_expiry_minutes 由 configs/app.yaml 提供
// Secret Key 由環境變數提供
type AppConfig struct {
	JWTExpiryMinutes int `yaml:"jwt_expiry_minutes"`
}

func LoadConfig(path string) (AppConfig, error) {
	cfg := AppConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.JWTExpiryMinutes <= 0 {
		return cfg, fmt.Errorf("jwt_expiry_minutes must be greater than 0")
	}
	return cfg, nil
}

func (c AppConfig) JWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryMinutes) * time.Minute
}

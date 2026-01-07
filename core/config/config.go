package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"database"`
	SecurityEntrance string `mapstructure:"securityEntrance"`
	Initialized      bool   `mapstructure:"initialized"`
}

var AppConfig *Config

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("initialized", true)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, creating default config: %v", err)
		if err := createDefaultConfig(); err != nil {
			log.Printf("Failed to create default config: %v", err)
		}
		viper.ReadInConfig()
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return err
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.Server.Port = port
	}

	return nil
}

func createDefaultConfig() error {
	defaultConfig := `server:
  port: "8080"
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: ""
  name: "gpanel"

securityEntrance: "/"
initialized: false
`

	configPaths := []string{".", "./config"}
	for _, path := range configPaths {
		configFile := filepath.Join(path, "config.yaml")
		if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err == nil {
			log.Printf("Created default config file at: %s", configFile)
			return nil
		}
	}
	return os.ErrNotExist
}

func SaveConfig() error {
	viper.Set("server.port", AppConfig.Server.Port)
	viper.Set("server.mode", AppConfig.Server.Mode)
	viper.Set("database.host", AppConfig.Database.Host)
	viper.Set("database.port", AppConfig.Database.Port)
	viper.Set("database.user", AppConfig.Database.User)
	viper.Set("database.password", AppConfig.Database.Password)
	viper.Set("database.name", AppConfig.Database.Name)
	viper.Set("securityEntrance", AppConfig.SecurityEntrance)
	viper.Set("initialized", AppConfig.Initialized)

	// 获取当前配置文件路径
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// 如果没有找到配置文件，使用默认路径
		configFile = "config.yaml"
	}

	log.Printf("Attempting to save config to: %s", configFile)
	if err := viper.WriteConfigAs(configFile); err != nil {
		log.Printf("Failed to save config: %v", err)
		return err
	}

	log.Printf("Config saved successfully to: %s", configFile)
	return nil
}
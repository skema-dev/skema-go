package config

import (
	"bytes"
	"os"

	"github.com/skema-dev/skema-go/logging"
	"github.com/spf13/viper"
)

var (
	defaultPaths  = []string{"./", "config", "/config"}
	defaultConfig *Config
)

type Config struct {
	viperData *viper.Viper
}

func NewConfigWithFile(path string) *Config {
	logging.Init(logging.DebugLevel, "console")
	if _, err := os.Stat(path); err != nil {
		logging.Errorw("loading config from file failed", "path", path, "error", err.Error())
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logging.Errorw("failed reading config", "path", path, "error", err.Error())
		return nil
	}

	conf := NewConfigWithString(string(data))
	return conf
}

func NewConfigWithString(data string) *Config {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer([]byte(data)))
	if err != nil {
		logging.Errorw("loading config from raw bytes failed", "error", err.Error())
		return nil
	}
	return &Config{
		viperData: v,
	}
}

func (c *Config) GetSubConfig(key string) *Config {
	sub := c.viperData.Sub(key)
	if sub == nil {
		logging.Errorw("no config found", "key", key)
	}
	return &Config{
		viperData: sub,
	}
}

func (c *Config) GetValue(key string, target interface{}) error {
	sub := c.viperData.Sub(key)
	err := sub.Unmarshal(target)
	return err
}

func (c *Config) GetString(key string) string {
	return c.viperData.GetString(key)
}

func (c *Config) GetInt(key string) int {
	return c.viperData.GetInt(key)
}

func (c *Config) GetBool(key string) bool {
	return c.viperData.GetBool(key)
}

func (c *Config) GetFloat(key string) float64 {
	return c.viperData.GetFloat64(key)
}

func (c *Config) GetStringArray(key string) []string {
	return c.viperData.GetStringSlice(key)
}

func (c *Config) GetIntArray(key string) []int {
	return c.viperData.GetIntSlice(key)
}

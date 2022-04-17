package config

import (
	"bytes"
	"os"

	"github.com/skema-dev/skema-go/logging"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var (
	defaultPaths  = []string{"./", "config", "/config"}
	defaultConfig *Config
)

type Config struct {
	viperData *viper.Viper
}

func NewConfigWithFile(path string) *Config {
	logging.Init("debug", "console")
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

func NewConfigWithEtcd(url string, path string) *Config {
	v := viper.New()
	v.AddRemoteProvider("etcd", url, path)
	v.SetConfigType("yaml")
	err := v.ReadRemoteConfig()
	if err != nil {
		logging.Errorf(err.Error())
		return nil
	}

	return &Config{viperData: v}
}

func NewConfigWithConsul(endpoint string, key string) *Config {
	v := viper.New()
	v.AddRemoteProvider("consul", endpoint, key)
	v.SetConfigType("json")
	err := v.ReadRemoteConfig()
	if err != nil {
		logging.Errorf(err.Error())
		return nil
	}

	return &Config{viperData: v}
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

func (c *Config) GetString(key string, opts ...string) string {
	if !c.viperData.IsSet(key) && len(opts) > 0 {
		return opts[0]
	}
	return c.viperData.GetString(key)
}

func (c *Config) GetInt(key string, opts ...int) int {
	if !c.viperData.IsSet(key) && len(opts) > 0 {
		return opts[0]
	}
	return c.viperData.GetInt(key)
}

func (c *Config) GetBool(key string, opts ...bool) bool {
	if !c.viperData.IsSet(key) && len(opts) > 0 {
		return opts[0]
	}
	return c.viperData.GetBool(key)
}

func (c *Config) GetFloat(key string, opts ...float64) float64 {
	if !c.viperData.IsSet(key) && len(opts) > 0 {
		return opts[0]
	}
	return c.viperData.GetFloat64(key)
}

func (c *Config) GetStringArray(key string) []string {
	return c.viperData.GetStringSlice(key)
}

func (c *Config) GetIntArray(key string) []int {
	return c.viperData.GetIntSlice(key)
}

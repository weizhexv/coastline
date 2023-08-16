package vconfig

import (
	"encoding/json"
	"github.com/spf13/viper"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	App      map[string]string `json:"app"`
	Server   map[string]string `json:"server"`
	Db       map[string]string `json:"db"`
	Upstream *upstreamConfig   `json:"upstream"`
	Monitor  map[string]string `json:"monitor"`
}

type upstreamConfig struct {
	Url map[string]string `json:"url"`
}

const activeEnv = "profile"

var (
	ENV    string
	config *Config
)

func init() {
	config = initConfig()
	ENV = loadENV()
}

func loadENV() string {
	env, ok := os.LookupEnv(activeEnv)
	if !ok || len(env) == 0 {
		panic("missing os variable: profile")
	}
	return env
}

func initConfig() *Config {
	viper.AddConfigPath("./vconfig")
	viper.AddConfigPath("../vconfig")
	viper.SetConfigName(getConfigName())
	viper.SetConfigType("yml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	conf := new(Config)
	err = viper.Unmarshal(conf)
	if err != nil {
		panic(err)
	}

	bs, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}

	log.Printf("init viper config: %s\n", bs)
	return conf
}

func getConfigName() string {
	env, ok := os.LookupEnv(activeEnv)
	if !ok || len(env) == 0 {
		panic("missing os variable: profile")
	}
	name := strings.ToLower("app-" + env)

	log.Printf("active config filename: %s\n", name)
	return name
}

func IsProd() bool {
	return strings.EqualFold("prod", ENV)
}

func ServerPort() string {
	return config.Server["port"]
}

func DbType() string {
	return config.Db["type"]
}

func DbUrl() string {
	return config.Db["url"]
}

func DbMaxOpenConns() int {
	num, err := strconv.Atoi(config.Db["max-open-conns"])
	if err != nil {
		panic(err)
	}
	if num == 0 {
		return runtime.NumCPU()
	}
	return num
}

func DbMaxIdleConns() int {
	num, err := strconv.Atoi(config.Db["max-idle-conns"])
	if err != nil {
		panic(err)
	}
	if num == 0 {
		return runtime.NumCPU()
	}
	return num
}

func Upstream() *upstreamConfig {
	return config.Upstream
}

func (u *upstreamConfig) UrlAuth() string {
	return u.Url["auth"]
}

func (u *upstreamConfig) UrlUserInfo() string {
	return u.Url["user-info"]
}

func LoggingLevel() string {
	return config.Server["logging-level"]
}

func AppName() string {
	return config.App["name"]
}

func MonitorPort() int {
	port, err := strconv.Atoi(config.Monitor["port"])
	if err != nil {
		panic(err)
	}
	return port
}

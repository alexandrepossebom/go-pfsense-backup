package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	confOnce sync.Once
	conf     *ConfYaml
)

var defaultConf = []byte(`#Configurations
firewalls:
  - name: Tijolo
    username: alexandre
    password: verysecretpass
    url: https://192.168.100.254
	directory: '/root/backups/'
  - name: Reboco
    username: alexandre
    password: verysecretpass
    url: https://192.168.100.1
	directory: '/root/backups/'
`)

// ConfYaml is config structure
type ConfYaml struct {
	Firewalls []FirewallItem `yaml:"firewalls,mapstructure"`
}

// FirewallItem configuration for items
type FirewallItem struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Password  string `yaml:"password"`
	Username  string `yaml:"username"`
	Directory string `yaml:"directory"`
}

// Get is used to get all configurations
func Get() *ConfYaml {
	confOnce.Do(func() {
		loadConf()
	})
	return conf
}

func loadConf() {
	conf = &ConfYaml{}
	filename := "go-pfsense-backup.yaml"
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/root/")
	viper.AddConfigPath(".")
	viper.SetConfigName(filename)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if err := viper.ReadConfig(bytes.NewBuffer(defaultConf)); err != nil {
			panic(err)
		}
		err = os.WriteFile(filename, defaultConf, 0o600)
		if err != nil {
			panic(err)
		}
		fmt.Println("Config file not found, new one is generated.")
	}
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalln("Invalid configuration file")
	}
}

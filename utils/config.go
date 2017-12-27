package utils

import (
  "encoding/json"
  "io/ioutil"
  "log"

  // "github.com/jinzhu/gorm"
  // _ "github.com/jinzhu/gorm/dialects/postgres"
)

type ServiceConfig struct {
  Name        string  `json:"service"`
  ReadToken   string  `json:"read_token"`
}

func NewServiceConfigs(pathToConfig string) []ServiceConfig {
  configContent, err := ioutil.ReadFile(pathToConfig)
  if err != nil {
    log.Fatal(err)
  }

  var configs []ServiceConfig
  err = json.Unmarshal(configContent, &configs)
  if err != nil {
    log.Fatal(err)
  }

  return configs
}

type DatabaseConfig struct {
  Dialect   string  `json:"dialect"`
  Database  string  `json:"database"`
  User      string  `json:"user"`
  Password  string  `json:"password"`
}

func NewDatabaseConfig(pathToConfig string) DatabaseConfig {
  configContent, err := ioutil.ReadFile(pathToConfig)
  if err != nil {
    log.Fatal(err)
  }

  var config DatabaseConfig
  err = json.Unmarshal(configContent, &config)
  if err != nil {
    log.Fatal(err)
  }

  return config
}

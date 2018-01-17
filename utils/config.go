package utils

import (
  "encoding/json"
  "os"
  "fmt"
  "io/ioutil"
  "reflect"

  "github.com/spf13/viper"
)

func RaccoltaHome() string {
  raccoltaHome := "%s/.raccolta"
  return fmt.Sprintf(raccoltaHome, os.Getenv("HOME"))
}

type ServiceConfig struct {
  Name        string    `json:"service"`
  ReadToken   string    `json:"read_token"`
  APIs        []string  `json:"api"`
}

func NewServiceConfig(direction interface{}) (ServiceConfig, error) {
  var config ServiceConfig

  switch _direction := direction.(type) {
  case *viper.Viper:
    config.Name = _direction.Get("service").(string)
    config.ReadToken = _direction.Get("read_token").(string)

    apis := make([]string, len(_direction.Get("api").([]interface{})))
    for i, api := range _direction.Get("api").([]interface{}) {
      switch api.(type) {
      case string:
        apis[i] = api.(string)
      }
    }
    config.APIs = apis
  case string:
    configContent, err := ioutil.ReadFile(_direction)
    if err != nil {
      panic(fmt.Errorf("[Error] %s : Read file faild", err))
    }

    err = json.Unmarshal(configContent, &config)
    if err != nil {
      panic(fmt.Errorf("[Error] %s : Unmarshal faild", err))
    }
  default:
    return config, fmt.Errorf("Type Error: %s is not supported", reflect.TypeOf(_direction))
  }

  return config, nil
}

type DatabaseConfig struct {
  Dialect   string  `json:"dialect"`
  Database  string  `json:"database"`
  User      string  `json:"user"`
  Password  string  `json:"password"`
}

func NewDatabaseConfig(direction interface{}) (DatabaseConfig, error) {
  var config DatabaseConfig

  switch _direction := direction.(type) {
  case *viper.Viper:
    config.Dialect = _direction.Get("dialect").(string)
    config.Database = _direction.Get("database").(string)
    config.User = _direction.Get("user").(string)
    config.Password = _direction.Get("password").(string)
  case string:
    configContent, err := ioutil.ReadFile(_direction)
    if err != nil {
      panic(fmt.Errorf("[Error] %s : Read file faild", err))
    }

    err = json.Unmarshal(configContent, &config)
    if err != nil {
      panic(fmt.Errorf("[Error] %s : Unmarshal faild", err))
    }
  default:
    return config, fmt.Errorf("Type Error: %s is not supported", reflect.TypeOf(_direction))
  }

  return config, nil
}

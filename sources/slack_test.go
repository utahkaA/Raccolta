package sources

import (
  "testing"

  "github.com/utahkaA/Raccolta/utils"
)

func TestDesimationSlackHistory(t *testing.T) {
  configPath := utils.RaccoltaHome()
  slackInterface, err := NewSlackInterface(configPath)
  if err != nil {
    t.Fatalf("Faild Test: Desimation()")
  }
  slackInterface.DecimationSlackHistory("on:2018-01-16")
}

func TestConstructSlackSearchURL(t *testing.T) {
  configPath := utils.RaccoltaHome()
  slackInterface, err := NewSlackInterface(configPath)
  if err != nil {
    t.Fatalf("Faild Test: ConstructSlackSearchURL")
  }

  url := slackInterface.constructSlackSearchURL("on:2018-01-16")
  t.Log(url)
}

func TestConstructEndpoint(t *testing.T) {
  configPath := utils.RaccoltaHome()
  slackInterface, err := NewSlackInterface(configPath)
  if err != nil {
    t.Fatalf("Faild Test: ConstructEndpoint()")
  }

  url, err := slackInterface.constructEndpoint("hogehoge")
  if err != nil {
    t.Log(err)
  }
  t.Logf("url: %s", url)

  url, err = slackInterface.constructEndpoint("api.history")
  if err != nil {
    t.Log(err)
  }
  t.Logf("url: %s", url)
}

package sources

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/spf13/viper"
)

type SlackChannel struct {
  id    string  `json:"id"`
  Name  string  `json:"name"`
}

type SlackHistory struct {
  Ok        bool            `json:"ok"`
  Messages  []SlackMessage  `json:"messages"`
  HasMore   bool            `json:"has_more"`
}

type SlackMessage struct {
  Text      string            `json:"text"`
  Username  string            `json:"username"`
  BotId     string            `json:"bot_id"`
  Type      string            `json:"type"`
  Subtype   string            `json:"subtype`
  Timestamp string            `json:"ts"`
  Reactions []SlackReaction   `json:"reactions"`
}

type SlackReaction struct {
  Name  string    `json:"name"`
  Users []string  `json:"users"`
  Count int       `json:"count"`
}

type SlackSearchResult struct {
  Ok        bool                  `json:"ok"`
  Messages  SlackMatchMessages    `json:"messages"`
  Query     string                `json:"query"`
}

type SlackMatchMessages struct {
  Matches []SlackMatch  `json:"matches"`
  Total   int           `json:"total"`
}

type SlackMatch struct {
  SlackMessage
  Channel       SlackChannel  `json:"channel"`
  Permalink     string        `json:"permalink"`
}

type SlackInterface struct {
  config      *viper.Viper
  token       string
  channelId   string
  channel     string
}

func NewSlackInterface(configPath string) (*SlackInterface, error) {
  sl := new(SlackInterface)

  // loading slack configuration
  v := viper.New()
  v.SetConfigName("slack")
  v.AddConfigPath(configPath)
  if err := v.ReadInConfig(); err != nil {
    errMsg := "[Error] Reading configuration file faild.: %s\n"
    return sl, fmt.Errorf(errMsg, err)
  }
  sl.config = v

  sl.token = sl.config.GetString("token")
  sl.channelId = sl.config.GetString("channel_id")
  sl.channel = sl.config.GetString("channel")

  return sl, nil
}

func (sl *SlackInterface) constructEndpoint(apiMethod string) (string, error) {
  endpointTemp := sl.config.GetString(apiMethod)

  getQuery := url.Values{}
  getQuery.Set("token", sl.token)

  switch apiMethod {
  case "api.history":
    getQuery.Add("channel", sl.channelId)
    return fmt.Sprintf("%s?%s", endpointTemp, bytes.NewBufferString(getQuery.Encode())), nil
  case "api.search":
    return fmt.Sprintf("%s?%s", endpointTemp, bytes.NewBufferString(getQuery.Encode())), nil
  }

  return "", fmt.Errorf("[Error] Invalid argument `%s`", apiMethod)
}

func (sl *SlackInterface) constructSlackHistoryURL() string {
  url, _ := sl.constructEndpoint("api.history")
  return url
}

func (sl *SlackInterface) constructSlackSearchURL(query string) string {
  endpoint, _ := sl.constructEndpoint("api.search")
  getQuery := url.Values{}
  getQuery.Set("query", query)
  url := fmt.Sprintf("%s&%s", endpoint, bytes.NewBufferString(getQuery.Encode()))
  return url
}

func (sl *SlackInterface) getRequest(url string) ([]byte, error) {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    errMsg := "[Error] Creating a new request to Slack failed\n"
    return nil, fmt.Errorf(errMsg)
  }

  client := new(http.Client)
  resp, err := client.Do(req)
  if err != nil {
    errMsg := "[Error] GET request to Slack failed (send request)\n"
    return nil, fmt.Errorf(errMsg)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    errMsg := "[Error] A ploblem occurred while reading a body of the response.: %s\n"
    return nil, fmt.Errorf(errMsg, err)
  }
  return body, nil
}

func (sl *SlackInterface) GetSlackHistory() (*SlackHistory, error) {
  url := sl.constructSlackHistoryURL()

  // req, err := http.NewRequest("GET", url, nil)
  // if err != nil {
  //   errMsg := "[Error] Creating a new request to Slack failed\n"
  //   return nil, fmt.Errorf(errMsg)
  // }
  //
  // client := new(http.Client)
  // resp, err := client.Do(req)
  // if err != nil {
  //   errMsg := "[Error] GET request to Slack failed (send request)\n"
  //   return nil, fmt.Errorf(errMsg)
  // }
  // defer resp.Body.Close()
  //
  // history, err := ioutil.ReadAll(resp.Body)
  // if err != nil {
  //   errMsg := "[Error] A ploblem occurred while reading a body of the response.: %s\n"
  //   return nil, fmt.Errorf(errMsg, err)
  // }
  history, err := sl.getRequest(url)
  if err != nil {
    errMsg := "[Error] GET request to %s falid :\n\t%s"
    return nil, fmt.Errorf(errMsg, url, err)
  }

  var slackHist SlackHistory
  err = json.Unmarshal(history, &slackHist)
  if err != nil {
    errMsg := "[Error] A problem occurred while parsing data as JSON file.: %s\n"
    return nil, fmt.Errorf(errMsg, err)
  }

  return &slackHist, nil
}

func (sl *SlackInterface) Search(channel string, query string) (*SlackSearchResult, error) {
  url := sl.constructSlackSearchURL(query)

  matches, err := sl.getRequest(url)
  if err != nil {
    errMsg := "[Error] GET request to %s falid :\n\t%s"
    return nil, fmt.Errorf(errMsg, url, err)
  }

  var result SlackSearchResult
  err = json.Unmarshal(matches, &result)
  if err != nil {
    errMsg := "[Error] A problem occurred while parsing data as JSON file.: %s\n"
    return nil, fmt.Errorf(errMsg, err)
  }

  slackMatches := make([]SlackMatch, 0)
  for _, match := range result.Messages.Matches {
    if match.Channel.Name == channel {
      slackMatches = append(slackMatches, match)
    }
  }
  result.Messages.Matches = slackMatches

  return &result, nil
}

func (sl *SlackInterface) DecimationSlackHistory(query string) error {
  result, err := sl.Search(sl.channel, query)
  if err != nil {
    errMsg := "[Error] A problem occurred while searching: %s\n"
    return fmt.Errorf(errMsg, err)
  }

  for i, match := range result.Messages.Matches {
    fmt.Println(i, match.Text)
  }

  return nil
}

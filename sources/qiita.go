package sources

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "reflect"

  "github.com/utahkaA/Raccolta/utils"
)

type QiitaArticle struct {
  RenderedBody    string    `json:"rendered_body"`
  Body            string    `json:"body"`
  Coediting       bool      `json:"coediting"`
  CommentsCount	  int       `json:"comments_count"`
  CreatedAt       string    `json:"created_at"`
  Group           string    `json:"group"`
  Id              string    `json:"id"`
  LikesCount      int       `json:"likes_count"`
  Private         bool      `json:"private"`
  ReactionsCount  int       `json:"reactions_count"`
  Tags            []Tag     `json:"tags"`
  Title           string    `json:"title"`
  UpdatedAt       string    `json:"updated_at"`
  URL             string    `json:"url"`
  User            UserInfo  `json:"user"`
}

type Tag struct {
  Name      string    `json:"name"`
  Versions  []string  `json:"versions"`
}

type UserInfo struct {
  Description     string  `json:"description"`
  FacebookId      string  `json:"facebook_id"`
  FolloweesCount  int     `json:"followees_count"`
  FollowersCount  int     `json:"followers_count"`
  GithubLoginName string  `json:"github_login_name"`
  Id              string  `json:"id"`
  ItemsCount      int     `json:"items_count"`
  LinkedInId      string  `json:"linkedin_id"`
  Location        string  `json:"location"`
  Name            string  `json:"name"`
  Organization    string  `json:"organization"`
}

type QiitaInterface struct {
  apiEndPointTemp string
  urls            []string
  readToken       string
}

func NewQiitaInterface(config interface{}) (*QiitaInterface, error) {
  qi := &QiitaInterface{}
  qi.apiEndPointTemp = `https://qiita.com/api/v2/`
  // qi.url = fmt.Sprintf("%s%s", qi.apiEndPointTemp, api)

  switch _config := config.(type) {
  case utils.ServiceConfig:
    if _config.Name == "Qiita" {
      qi.readToken = _config.ReadToken
      for _, api := range _config.APIs {
        qi.urls = append(qi.urls, fmt.Sprintf("%s%s", qi.apiEndPointTemp, api))
      }
    }
  case []utils.ServiceConfig:
    for _, c := range _config {
      if c.Name == "Qiita" {
        qi.readToken = c.ReadToken
        for _, api := range c.APIs {
          qi.urls = append(qi.urls, fmt.Sprintf("%s%s", qi.apiEndPointTemp, api))
        }
      }
    }
  default:
    return qi, fmt.Errorf("Type Error: %s is not supported", reflect.TypeOf(_config))
  }

  return qi, nil
}

func (qi *QiitaInterface) SetReadToken(token string) {
  qi.readToken = token
}

func (qi *QiitaInterface) AppendAPI(api string) {
  qi.urls = append(qi.urls, fmt.Sprintf("%s%s", qi.apiEndPointTemp, api))
}

func (qi *QiitaInterface) Get() []QiitaArticle {
  allQiitaArticles := make([]QiitaArticle, 0)
  for _, url := range qi.urls {
    // Create a new request instance.
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
      errMsg := "Creating a new request to Qiita failed\n"
      log.Fatalf(errMsg)
    }
    // Add an authorization header to the request.
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", qi.readToken))

    // Create a HTTP client and send a GET request.
    client := new(http.Client)
    res, err := client.Do(req)
    if err != nil {
      errMsg := "GET request to Qiita failed (send request)\n"
      log.Fatalf(errMsg)
    }
    defer res.Body.Close()

    // Read a response from the Qiita API.
    articles, err := ioutil.ReadAll(res.Body)
    if err != nil {
      log.Fatal(err)
    }

    // Parse response data as a JSON format file.
    var qiitaArticles []QiitaArticle
    err = json.Unmarshal(articles, &qiitaArticles)
    if err != nil {
      log.Fatal(err)
    }
    allQiitaArticles = append(allQiitaArticles, qiitaArticles...)
  }
  return allQiitaArticles
}

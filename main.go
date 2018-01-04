package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "os"
  "log"
  "time"
  "net/http"
  "net/url"
  "strings"

  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"

  "github.com/spf13/viper"

  "github.com/utahkaA/Raccolta/sources"
  "github.com/utahkaA/Raccolta/models"
  "github.com/utahkaA/Raccolta/utils"
)

func ConfigPath() string {
  raccoltaHome := "%s/.raccolta"
  return fmt.Sprintf(raccoltaHome, os.Getenv("HOME"))
}

func ConnectDB(configPath string) (*gorm.DB, error) {
  var (
    db *gorm.DB
    err error
  )

  // Read database configuration file.
  v := viper.New()
  v.SetConfigName("database")
  v.AddConfigPath(configPath)
  if err = v.ReadInConfig(); err != nil {
    return nil, err
  }

  if v.Get("dialect") == "postgres" {
    info := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s",
                        v.Get("user"),
                        v.Get("database"),
                        v.Get("password"))
    // Connect to a PostgreSQL server.
    if db, err = gorm.Open(v.GetString("dialect"), info); err != nil {
      return db, err
    }
  }

  return db, nil
}

func CollectQiitaArticles(dbCh <-chan *gorm.DB, articleCh chan<- models.Article,
                          interval int, configPath string) {
  // load qiita config using viper
  v := viper.New()
  v.SetConfigName("qiita")
  v.AddConfigPath(configPath)
  if err := v.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }
  qiitaConfig, err := utils.NewServiceConfig(v)
  if err != nil {
    panic(fmt.Errorf("[Error] %s : Making a instance of ServiceConfig faild."))
  }

  // make a qiita interface instance
  qiitaInterface, err := sources.NewQiitaInterface(qiitaConfig)
  if err != nil {
    panic(fmt.Errorf("[Error] %s : Making a instance of QiitaInterface faild.\n"))
  }

  db := <-dbCh

  for {
    // get qiita articles
    articles := qiitaInterface.Get()
    if db.HasTable("articles") {
      for _, article := range articles {
        // parse article's createdAt string.
        createdAt, err := time.Parse(time.RFC3339, article.CreatedAt)
        if err != nil {
          panic(fmt.Errorf("[Error] Parsing a article.CreatedAt faild.\n"))
        }

        // make article's record
        articleRecord := models.Article{
          Title: article.Title,
          SiteName: "Qiita",
          Author: article.User.Id,
          Tags: models.NewTags(article.Tags),
          URL: article.URL,
          PostedAt: createdAt,
          LikeCount: uint(article.LikesCount),
          CommentsCount: uint(article.CommentsCount),
          IsFavorite: false,
        }

        if dbc := db.Create(&articleRecord); dbc.Error != nil {
          log.Printf("The article '%s' is already registered.\n\n", article.Title)
        } else {
          // share new articles through a article channel with different goroutine.
          articleCh <- articleRecord
        }
      }
      time.Sleep(time.Duration(interval) * time.Second)
    } else {
      log.Printf("Article table does not exist.")
    }
  }
}

func StoreInSlack(articleCh <-chan models.Article, configPath string) {
  v := viper.New()
  v.SetConfigName("slack")
  v.AddConfigPath(configPath)
  if err := v.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }

  for {
    article := <-articleCh

    showTags := "タグ : "
    for _, tag := range article.Tags {
      showTags = fmt.Sprintf("%s %s", showTags, tag.Name)
    }

    msg := fmt.Sprintf("%s (by %s)\n%s\n%s\n---\n",
      article.Title,
      article.Author,
      article.URL,
      showTags,
    )
    fmt.Printf(msg)

    postMsg := `{"channel":"%s","username":"%s","text":"%s"}`
    postMsg = fmt.Sprintf(postMsg, v.Get("channel"), v.Get("username"), msg)
    req, err := http.NewRequest("POST", v.Get("url").(string), bytes.NewBuffer([]byte(postMsg)))
    if err != nil {
      log.Fatal(err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := new(http.Client)
    resp, err := client.Do(req)
    if err != nil {
      log.Fatal(err)
    }
    defer resp.Body.Close()
  }
}

func CleanSlackChannel(dbCh <-chan *gorm.DB, interval int, configPath string) {
  db := <-dbCh

  // load slack config using viper
  v := viper.New()
  v.SetConfigName("slack")
  v.AddConfigPath(configPath)
  if err := v.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }

  chatDeleteApi := v.GetString("api.delete")

  for {
    slackHist := getSlackHistory(v)
    r := strings.NewReplacer("<", "", ">", "")
    for _, slackMsg := range slackHist.Messages {
      if len(slackMsg.Reactions) == 1 {
        label := slackMsg.Reactions[0].Name

        getQuery := url.Values{}
        getQuery.Set("token", v.GetString("token"))
        getQuery.Add("channel", v.GetString("channel_id"))
        getQuery.Add("ts", slackMsg.Timestamp)
        endpoint := fmt.Sprintf("%s?%s", chatDeleteApi, bytes.NewBufferString(getQuery.Encode()))

        msg := strings.Split(slackMsg.Text, "\n")
        title := msg[0]
        url := msg[1]
        url = r.Replace(url)

        if label == "heart" {
          // annotation
          articleRecord := models.Article{}
          db.First(&articleRecord, "url=?", url)
          newArticleRecord := articleRecord
          newArticleRecord.IsFavorite = true

          db.Model(&articleRecord).Update(newArticleRecord)

          // delete slack message
          resp, err := http.Get(endpoint)
          if err != nil {
            panic(fmt.Errorf("HTTP request faild"))
          }
          defer resp.Body.Close()

          fmt.Printf("%s [heart]\n", title)
        } else if label == "weary" {
          // delete slack message
          resp, err := http.Get(endpoint)
          if err != nil {
            panic(fmt.Errorf("HTTP request faild"))
          }
          defer resp.Body.Close()

          fmt.Printf("%s [weary]\n", title)
        }
      }
    }
    time.Sleep(time.Duration(interval) * time.Second)
  }
}

func constructEndpoint(v *viper.Viper) string {
  chatHistoryApi := v.GetString("api.history")
  getQuery := url.Values{}
  getQuery.Set("token", v.GetString("token"))
  getQuery.Add("channel", v.GetString("channel_id"))
  endpoint := fmt.Sprintf("%s?%s", chatHistoryApi, bytes.NewBufferString(getQuery.Encode()))
  return endpoint
}

func getSlackHistory(v *viper.Viper) *sources.SlackHistory {
  endpoint := constructEndpoint(v)

  req, err := http.NewRequest("GET", endpoint, nil)
  if err != nil {
    errMsg := "Creating a new request to Slack failed\n"
    panic(fmt.Errorf(errMsg))
  }

  client := new(http.Client)
  resp, err := client.Do(req)
  if err != nil {
    errMsg := "GET request to Slack failed (send request)\n"
    panic(fmt.Errorf(errMsg))
  }
  defer resp.Body.Close()

  var slackHist sources.SlackHistory
  history, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(fmt.Errorf("%s\n", err))
  }

  err = json.Unmarshal(history, &slackHist)
  if err != nil {
    panic(fmt.Errorf("%s\n", err))
  }

  return &slackHist
}

func main() {
  // preparation of a config directory.
  configPath := ConfigPath()
  _, err := os.Stat(configPath);
  if err != nil {
    if err = os.Mkdir(configPath, 0777); err != nil {
      fmt.Errorf("[Error] %s : Since Raccolta configuration directory did not exist, operation faild, please try again.\n", err)
    }
  }

  // connecting to a db
  db, err := ConnectDB(configPath)
  if err != nil {
    panic(fmt.Errorf("[Error] DB Connecting faild : %s", err))
  }
  defer db.Close()

  // make a channels
  dbInsertCh := make(chan *gorm.DB)
  dbUpdateCh := make(chan *gorm.DB)
  articleCh := make(chan models.Article, 256)

  // begin collecting qiita articles goroutine.
  go CollectQiitaArticles(dbInsertCh, articleCh, 1800, configPath)
  go StoreInSlack(articleCh, configPath)
  go CleanSlackChannel(dbUpdateCh, 900, configPath)

  dbInsertCh <- db
  dbUpdateCh <- db

  for {
    log.Printf("Raccolta is runnning...")
    time.Sleep(1800 * time.Second)
  }
}

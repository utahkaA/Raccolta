package main

import (
  "bytes"
  "fmt"
  "os"
  "log"
  "time"
  "net/http"

  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"

  "github.com/spf13/viper"

  "github.com/utahkaA/Raccolta/sources"
  "github.com/utahkaA/Raccolta/models"
  "github.com/utahkaA/Raccolta/utils"
)

const (
  qiitaConfig = "qiita"
  slackConfig = "slack"
  databaseConfig = "database"
)

func main() {
  configPath := fmt.Sprintf("%s/.raccolta", os.Getenv("HOME"))
  if _, err := os.Stat(configPath); err != nil {
    if err = os.Mkdir(configPath, 0777); err != nil {
      fmt.Errorf("[Error] %s : Since Raccolta configuration directory did not exist, operation faild, please try again.\n", err)
    }
  }

  // Read Database Config
  viper.SetConfigName(databaseConfig)
  viper.AddConfigPath(configPath)
  if err := viper.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }

  // databaseConfig := utils.NewDatabaseConfig(pathToDatabaseConfig)
  var db *gorm.DB
  if viper.Get("dialect") == "postgres" {
    info := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s",
                        viper.Get("user"),
                        viper.Get("database"),
                        viper.Get("password"))
    // Connect to a PostgreSQL server.
    var err error
    if db, err = gorm.Open(viper.Get("dialect").(string), info); err != nil {
      log.Fatal(err)
    }
  }
  defer db.Close()

  // Read Qiita Config
  viper.SetConfigName(qiitaConfig)
  if err := viper.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }
  qiitaConfig, err := utils.NewServiceConfig(viper.GetViper())
  qiitaIF, err := sources.NewQiitaInterface(qiitaConfig)
  if err != nil {
    log.Fatal(err)
  }

  // Get Qiita Articles
  articles := qiitaIF.Get()

  newArticles := make([]models.Article, 0)
  if db.HasTable("articles") {
    for _, article := range articles {
      var tags []models.Tag
      for _, t := range article.Tags {
        tags = append(tags, models.Tag{Name: t.Name})
      }

      createdAt, err := time.Parse(time.RFC3339, article.CreatedAt)
      if err != nil {
        log.Fatal(err)
      }

      showTags := make([]string, 0)
      for _, t := range article.Tags {
        showTags = append(showTags, t.Name)
      }
      fmt.Printf("%s,%v,%s,%d,%s\n",
        article.Title,
        showTags,
        article.URL,
        article.LikesCount,
        article.User.Id,
      )

      articleRecord := models.Article{
        Title: article.Title,
        SiteName: "Qiita",
        Author: article.User.Id,
        Tags: tags,
        URL: article.URL,
        PostedAt: createdAt,
        LikeCount: uint(article.LikesCount),
        CommentsCount: uint(article.CommentsCount),
        IsFavorite: false,
      }
      if dbc := db.Create(&articleRecord); dbc.Error != nil {
        log.Printf("The article '%s' is already registered.\n\n", article.Title)
      } else {
        newArticles = append(newArticles, articleRecord)
      }
    }
  } else {
    log.Printf("Article table does not exist.")
  }

  fmt.Println("--- Stock Articles using Slack ---")
  // Read Slack Config
  viper.SetConfigName(slackConfig)
  if err := viper.ReadInConfig(); err != nil {
    panic(fmt.Errorf("[Error] %s : Fatal error config file.\n", err))
  }

  for _, article := range newArticles {
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
    postMsg = fmt.Sprintf(postMsg, viper.Get("channel"), viper.Get("username"), msg)
    req, err := http.NewRequest("POST", viper.Get("url").(string), bytes.NewBuffer([]byte(postMsg)))
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

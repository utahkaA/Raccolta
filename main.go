package main

import (
  "fmt"
  "log"
  "time"

  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"

  "github.com/utahkaA/Raccolta/sources"
  "github.com/utahkaA/Raccolta/models"
  "github.com/utahkaA/Raccolta/utils"
)

const (
  pathToServiceConfig = ".config.json"
  pathToDatabaseConfig = ".database.json"
)

func main() {
  serviceConfigs := utils.NewServiceConfigs(pathToServiceConfig)

  qiitaIF, err := sources.NewQiitaInterface(serviceConfigs)
  if err != nil {
    log.Fatal(err)
  }
  articles := qiitaIF.Get()

  databaseConfig := utils.NewDatabaseConfig(pathToDatabaseConfig)
  var db *gorm.DB
  if databaseConfig.Dialect == "postgres" {
    info := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s",
                        databaseConfig.User,
                        databaseConfig.Database,
                        databaseConfig.Password)
    // Connect to a PostgreSQL server.
    db, err = gorm.Open(databaseConfig.Dialect, info)
    if err != nil {
      log.Fatal(err)
    }
  }
  defer db.Close()

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
      fmt.Println("---------------")

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
      db.Create(&articleRecord)
    }
  } else {
    log.Printf("Article table does not exist.")
  }
}

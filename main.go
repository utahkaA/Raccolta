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

// type Config struct {
//   Service     string  `json:"service"`
//   ReadToken   string  `json:"read_token"`
// }

func main() {
  serviceConfigs := utils.NewServiceConfigs(".config.json")

  api := "tags/python/items"
  qiitaIF, err := sources.NewQiitaInterface(api, serviceConfigs)
  if err != nil {
    log.Fatal(err)
  }
  articles := qiitaIF.Get()

  databaseConfig := utils.NewDatabaseConfig(".database.json")
  var db *gorm.DB
  if databaseConfig.Dialect == "postgres" {
    info := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s",
                        databaseConfig.User,
                        databaseConfig.Database,
                        databaseConfig.Password)
    db, err = gorm.Open(databaseConfig.Dialect, info)
    if err != nil {
      log.Fatal(err)
    }
  }
  defer db.Close()

  // db.DropTableIfExists(&models.Article{}, &models.Tag{}, "article_tags")
  // db.AutoMigrate(&models.Article{})
  // db.AutoMigrate(&models.Tag{})

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

      log.Printf("%s, Qiita, %s, %s, %s, %d, %s, %d\n",
        article.Title,
        tags,
        article.URL,
        createdAt,
        article.LikesCount,
        article.User.Id,
        article.CommentsCount,
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

      log.Println(articleRecord)

      db.Create(&articleRecord)
    }
  } else {
    log.Printf("Article table does not exist.")
  }
}

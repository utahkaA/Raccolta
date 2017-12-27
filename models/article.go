package models

import (
  "time"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

type Article struct {
  gorm.Model
  Title         string
  SiteName      string
  Author        string
  Tags          []Tag       `gorm:"many2many:article_tags;"`  // Many-To-Many relationship
  URL           string
  PostedAt      time.Time
  LikeCount     uint
  CommentsCount uint
  IsFavorite    bool
}

type Tag struct {
  gorm.Model
  Name  string
}

// func main() {
//   db, err := gorm.Open("postgres", "user=raccolta dbname=raccolta sslmode=disable password=WFMT3604")
//   if err != nil {
//     log.Fatal(err)
//   }
//   defer db.Close()
//
//   if db.HasTable(&Article{}) {
//     log.Printf("Article table exists.")
//
//     db.DropTable(&Article{})
//   } else {
//
//     db.CreateTable(&Article{})
//   }
// }

package models

import (
  "time"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

type Article struct {
  gorm.Model
  ArticleID     int
  Title         string
  SiteName      string
  Author        string
  Tags          []Tag       `gorm:"many2many:article_tags;"`  // Many-To-Many relationship
  URL           string      `gorm:"unique"`
  PostedAt      time.Time
  LikeCount     uint
  CommentsCount uint
  IsFavorite    bool
}

type Tag struct {
  gorm.Model
  Name  string
}

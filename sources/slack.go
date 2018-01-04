package sources

type SlackHistory struct {
  Ok        bool            `json:"ok"`
  Messages  []SlackMessage  `json:"messages"`
  HasMore   bool            `json:"has_more"`
}

type SlackMessage struct {
  Text      string      `json:"text"`
  Username  string      `json:"username"`
  BotId     string      `json:"bot_id"`
  Type      string      `json:"type"`
  Subtype   string      `json:"subtype`
  Timestamp string      `json:"ts"`
  Reactions []Reaction  `json:"reactions"`
}

type Reaction struct {
  Name  string    `json:"name"`
  Users []string  `json:"users"`
  Count int       `json:"count"`
}

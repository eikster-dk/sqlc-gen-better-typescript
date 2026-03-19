// Query: GetStories
// Command: :many
// SQL: SELECT sms_stories.id, name, sender, sms_episodes.id, story_id, send_date, singular_content FROM sms_stories 
inner join sms_episodes on sms_stories.id  = sms_episodes.id

// Results:
//   id: Schema.String
//   name: Schema.String
//   sender: Schema.String
//   id: Schema.String
//   story_id: Schema.String
//   send_date: Schema.Date
//   singular_content: Schema.String

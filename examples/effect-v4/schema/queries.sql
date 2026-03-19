-- name: GetSMSProducts :many
select id, name from sms_stories;

-- name: GetByIds :many
select id, name from sms_stories where id in (sqlc.slice('ids'));

-- name: GetStory :one
select id, name from sms_stories where id = $1;

-- name: AddStory :exec
insert into sms_stories (id, name) values ($1,$2);

-- name: GetStories :many
SELECT * FROM sms_stories 
inner join sms_episodes on sms_stories.id = sms_episodes.story_id;

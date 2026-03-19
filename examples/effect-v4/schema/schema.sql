create table sms_stories (
    id text primary key,
    name text not null,
    sender text not null
);

create table sms_episodes (
    id text not null,
    story_id text not null,
    send_date date not null,
    singular_content text not null
);

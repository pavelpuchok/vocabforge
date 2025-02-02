create table if not exists users (
    id integer primary key,
    name text not null,
    telegram_id integer not null,
    telegram_chat_id integer not null,
    created_at datetime not null,
    updated_at datetime not null
);

create unique index telegram_id_uniq on users (telegram_id, telegram_chat_id);

create table if not exists vocab_words(
    id integer primary key,
    preply_id text,
    spelling text not null,
    definition text not null,
    lexical_category text not null,
    lang text not null,
    translation_ru text not null,
    answered_count integer not null,
    viewed_count integer not null,
    user_id integer not null,
    added_at datetime not nulL,
    learned_at datetime,
    last_showed_at datetime,

    FOREIGN KEY (user_id) REFERENCES users(id)
);

create index unseen_by_user_id on vocab_words(user_id, viewed_count, learned_at);
create unique index preply_id_for_user_uniq on vocab_words (user_id, preply_id);

create table if not exists vocab_words_exercises(
  id integer primary key,
  word_id integer not null,
  user_id integer not null,
  question text not null,
  answer text not null,

  telegram_msg_id integer not null,

  answered boolean not null,
  answered_correctly boolean,

  created_at datetime not null,
  answered_at datetime ,

  FOREIGN KEY (word_id) REFERENCES vocab_words(id)
);

create index by_user_id on vocab_words_exercises (user_id, answered, telegram_msg_id);


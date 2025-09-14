CREATE TABLE households (
  checklist TEXT[] NOT NULL DEFAULT '{}',
  crontab TEXT NOT NULL,
  current_member_index INTEGER NOT NULL DEFAULT 0,
  telegram_id BIGINT PRIMARY KEY
);

CREATE TABLE members (
  household_telegram_id BIGINT NOT NULL REFERENCES households(telegram_id),
  name TEXT NOT NULL,
  "order" INTEGER NOT NULL,
  telegram_id BIGINT NOT NULL
);

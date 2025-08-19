CREATE TABLE households (
  checklist TEXT[] NOT NULL DEFAULT '{}',
  crontab TEXT NOT NULL,
  current_member_index INTEGER NOT NULL DEFAULT 0
  telegram_id INTEGER PRIMARY KEY,
);

CREATE TABLE members (
  household_telegram_id INTEGER NOT NULL REFERENCES households(telegram_id),
  name TEXT NOT NULL
  telegram_id INTEGER NOT NULL,
);

package testenv

const migration = `
CREATE TABLE IF NOT EXISTS fruits (
    id          SERIAL          NOT NULL PRIMARY KEY,
    name        VARCHAR         NOT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vegetables (
    id          SERIAL          NOT NULL PRIMARY KEY,
    name        VARCHAR         NOT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS vegetables_names ON vegetables (name);

INSERT INTO
    fruits (name)
VALUES
    ('apple'),
    ('orange'),
    ('pineapple'),
    ('melon'),
    ('grape'),
    ('lemon'),
    ('blueberry');

INSERT INTO
    vegetables (name)
VALUES
    ('lettuce'),
    ('carrot'),
    ('tomato'),
    ('potato'),
    ('onion'),
    ('parsnip'),
    ('spinach');
`

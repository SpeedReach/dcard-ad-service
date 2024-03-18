CREATE TABLE IF NOT EXISTS Ads (
    id uuid PRIMARY KEY,
    title TEXT NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS Conditions (
    id uuid PRIMARY KEY,
    ad_id uuid REFERENCES Ads(id) ON DELETE CASCADE NOT NULL ,
    min_age INT NOT NULL,
    max_age	INT NOT NULL,
    male BOOLEAN NOT NULL,
    female BOOLEAN NOT NULL,
    ios BOOLEAN   NOT NULL,
    android BOOLEAN NOT NULL,
    web BOOLEAN NOT NULL,
    jp BOOLEAN NOT NULL,
    tw BOOLEAN    NOT NULL
);


SELECT a.id, a.title, a.start_at, a.end_at, c.min_age, c.max_age, c.male, c.female, c.ios, c.android, c.web
FROM Ads a
         LEFT JOIN Conditions c ON a.id = c.ad_id
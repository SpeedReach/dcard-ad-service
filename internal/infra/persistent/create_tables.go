package persistent

import "database/sql"

func CreateTables(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Ads (
    id uuid PRIMARY KEY,
    title TEXT NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL
)`)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS Conditions (
    id uuid PRIMARY KEY,
    ad_id uuid NOT NULL ,
    min_age INT NOT NULL,
    max_age	INT NOT NULL,
    male BOOLEAN NOT NULL,
    female BOOLEAN NOT NULL,
    ios BOOLEAN   NOT NULL,
    android BOOLEAN NOT NULL,
    web BOOLEAN NOT NULL,
    jp BOOLEAN NOT NULL,
    tw BOOLEAN    NOT NULL,
    CONSTRAINT fk_ad
        FOREIGN KEY(ad_id)
        REFERENCES Ads(id)
)`)
	if err != nil {
		panic(err)
	}
}

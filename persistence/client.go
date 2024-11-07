package persistence

import (
	"database/sql"
	"os"
	"pal/constants"
	"pal/util"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseClient struct {
	Conn *sql.DB
}

func StartClient(projectPath string) (DatabaseClient, error) {
	internalDirPath := path.Join(projectPath, constants.AppDir)
	if err := os.MkdirAll(internalDirPath, 0755); err != nil {
		return DatabaseClient{}, err
	}

	dbFilePath := path.Join(internalDirPath, "db.sqlite")

	if !util.FileExists(dbFilePath) {
		_, err := os.Create(dbFilePath)
		if err != nil {
			return DatabaseClient{}, err
		}
	}

	connectionString := "file:" + dbFilePath + "?_foreign_keys=true"

	conn, err := sql.Open("sqlite3", connectionString)

	if err != nil {
		return DatabaseClient{}, err
	}

	client := DatabaseClient{
		Conn: conn,
	}

	if err = client.runMigrations(); err != nil {
		return client, err
	}

	return client, nil
}

var allMigrations map[int]string = map[int]string{
	1: `
		create table conversations(
			id integer primary key autoincrement,
			created_at datetime default current_timestamp
		);
		create table messages(
			id integer primary key autoincrement,
			conversation_id integer,
			role string,
			content string,
			created_at datetime default current_timestamp,
			foreign key(conversation_id) references conversations(id) on delete cascade
		);
	`,
}

func (c *DatabaseClient) runMigrations() error {
	// Ensure that the `migrations` table exists. The table lets us keep track of
	// the migrations that have already been applied in the database.
	createMigrationsTableStatement := `
		create table if not exists migrations(id integer primary key)
	`
	if _, err := c.Conn.Exec(createMigrationsTableStatement); err != nil {
		return err
	}

	// Now scan the `migrations` table for the id of the last migration that has
	// been applied. We will then sequentially apply migrations whose id is greater
	// than this id.
	var lastAppliedMigration *int
	result := c.Conn.QueryRow("select max(id) from migrations")
	if err := result.Scan(&lastAppliedMigration); err != nil {
		return err
	}

	if lastAppliedMigration == nil {
		lastAppliedMigration = new(int)
	}

	for id, sql := range allMigrations {
		if id > *lastAppliedMigration {
			// Transaction for applying the migration and recording in the `migrations`
			// table that it has been applied
			tx, err := c.Conn.Begin()

			// Apply the migration.
			_, err = tx.Exec(sql)
			if err != nil {
				return err
			}

			// Record in the database that it contains this migration.
			_, err = tx.Exec("insert into migrations(id) values(?)", id)
			if err != nil {
				return err
			}

			// Commit the two statements.
			err = tx.Commit()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

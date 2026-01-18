package models

import (
	"database/sql"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

func ConnectDB() {
	sqlDB, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	database, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	database.AutoMigrate(&Post{}, &URL{}, &User{}, &Click{})

	// Manually add user_id column if it doesn't exist
	migrateUserIDColumn(database)

	DB = database
}

func migrateUserIDColumn(db *gorm.DB) {
	// Check if user_id column exists in urls table
	var columnExists bool
	db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('urls') WHERE name = 'user_id'").Scan(&columnExists)

	if !columnExists {
		// Add user_id column
		db.Exec("ALTER TABLE urls ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0")
	}
}

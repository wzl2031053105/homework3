package gohomework3

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	envLoaded bool
	envOnce   sync.Once
)

// DBType represents the type of database
type DBType string

const (
	DBTypeSQLite DBType = "sqllite"
)

// // loadEnv loads environment variables from .env file in the examples directory
// // This function locates the .env file by finding the examples directory

func loadEnv() {
	envOnce.Do(func() {
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			return
		}

		testutiDir := filepath.Dir(currentFile)
		examplesDir := filepath.Dir(testutiDir)

		envPath := filepath.Join(examplesDir, ".env")
		if err := godotenv.Load(envPath); err != nil {
			return
		}
		envLoaded = true
	})
}

func getDBDir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrInvalid
	}
	examplesDir := filepath.Dir(currentFile)

	dbDir := filepath.Join(examplesDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", err
	}
	return dbDir, nil
}

func NewTestDB(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	var db *gorm.DB
	var err error
	loadEnv()
	db, err = newSQLiteDB(t, filename)

	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open database:%v", err)
	}
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return db
}

func newSQLiteDB(t *testing.T, filename string) (*gorm.DB, error) {
	dbDir, err := getDBDir()
	if err != nil {
		return nil, err
	}

	if filename == "" {
		filename = "test.sqlite.db"
	} else {
		ext := filepath.Ext(filename)
		if ext == "" {
			ext = ".db"
		}
		base := filename[:len(filename)-len(ext)]
		if base == "" {
			base = "test"
		}
		baseLower := strings.ToLower(base)
		if !strings.Contains(baseLower, "sqlite") {
			filename = base + "_sqlite" + ext
		} else {
			filename = base + ext
		}
	}
	dbPath := filepath.Join(dbDir, filename)
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
			NoLowerCase:   false,
			NameReplacer:  nil,
		},
	})
}

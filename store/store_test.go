package store

import (
	adminmodel "github.com/xsxs89757/base-kit/model/admin"

	"reflect"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDialectorSupportsConfiguredDrivers(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		dsn      string
		wantType string
	}{
		{name: "empty driver defaults to sqlite", driver: "", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "sqlite", driver: "sqlite", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "sqlite3 alias", driver: "sqlite3", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "mysql", driver: "mysql", dsn: "user:pass@tcp(localhost:3306)/app", wantType: "*mysql.Dialector"},
		{name: "mariadb alias", driver: "mariadb", dsn: "user:pass@tcp(localhost:3306)/app", wantType: "*mysql.Dialector"},
		{name: "postgres", driver: "postgres", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "postgresql alias", driver: "postgresql", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "pgsql alias", driver: "pgsql", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "sqlserver", driver: "sqlserver", dsn: "sqlserver://user:pass@localhost:1433?database=app", wantType: "*sqlserver.Dialector"},
		{name: "mssql alias", driver: "mssql", dsn: "sqlserver://user:pass@localhost:1433?database=app", wantType: "*sqlserver.Dialector"},
		{name: "normalizes case and whitespace", driver: " PostgreSQL ", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dial, err := dialector(tt.driver, tt.dsn)
			if err != nil {
				t.Fatalf("dialector() error = %v", err)
			}
			if got := reflect.TypeOf(dial).String(); got != tt.wantType {
				t.Fatalf("dialector() type = %s, want %s", got, tt.wantType)
			}
		})
	}
}

func TestDialectorRejectsUnsupportedDriver(t *testing.T) {
	_, err := dialector("oracle", "")
	if err == nil {
		t.Fatal("dialector() expected unsupported driver error")
	}
	if !strings.Contains(err.Error(), "supported drivers") {
		t.Fatalf("dialector() error = %q, want supported drivers hint", err.Error())
	}
}

func TestSeedMenusBackfillsAuthCodeForExistingMenus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&adminmodel.Role{}, &adminmodel.Menu{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	DB = db

	system := adminmodel.Menu{Name: "System", Path: "/system", Type: "catalog", Title: "system.title", Status: 1}
	if err := DB.Create(&system).Error; err != nil {
		t.Fatalf("create system menu: %v", err)
	}
	roleMenu := adminmodel.Menu{Name: "SystemRole", Path: "/system/role", Component: "/system/role/list", Type: "menu", Title: "system.role.title", Status: 1}
	if err := DB.Create(&roleMenu).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	seedMenus()

	var updated adminmodel.Menu
	if err := DB.Where("name = ?", "SystemRole").First(&updated).Error; err != nil {
		t.Fatalf("load updated role menu: %v", err)
	}
	if updated.AuthCode != "System:Role:List" {
		t.Fatalf("expected SystemRole auth code to be backfilled, got %q", updated.AuthCode)
	}
	if updated.ParentID != system.ID {
		t.Fatalf("expected SystemRole parent id %d, got %d", system.ID, updated.ParentID)
	}
}

func TestSeedUsersRenamesLegacyRootAccountToSuper(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&adminmodel.User{}, &adminmodel.Role{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	DB = db

	superRole := adminmodel.Role{Name: "超级管理员", Code: "super", Status: 1}
	if err := DB.Create(&superRole).Error; err != nil {
		t.Fatalf("create super role: %v", err)
	}
	legacyRoot := adminmodel.User{Username: "vben", Password: "legacy-hash", RealName: "Vben", Status: 1, Roles: []adminmodel.Role{superRole}}
	if err := DB.Create(&legacyRoot).Error; err != nil {
		t.Fatalf("create legacy root user: %v", err)
	}

	seedUsers()

	var root adminmodel.User
	if err := DB.First(&root, legacyRoot.ID).Error; err != nil {
		t.Fatalf("load root user: %v", err)
	}
	if root.Username != "super" {
		t.Fatalf("expected legacy root username to be super, got %q", root.Username)
	}
	if root.RealName != "Super" {
		t.Fatalf("expected legacy root real name to be Super, got %q", root.RealName)
	}

	var count int64
	DB.Model(&adminmodel.User{}).Where("username = ?", "super").Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly one super user, got %d", count)
	}
}

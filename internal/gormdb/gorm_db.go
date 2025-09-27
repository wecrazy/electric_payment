package gormdb

import "gorm.io/gorm"

type DBUsed struct {
	Web *gorm.DB
}

// Global databases
var Databases *DBUsed

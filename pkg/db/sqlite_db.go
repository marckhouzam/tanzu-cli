// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"

	"github.com/pkg/errors"

	// Import the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

type sqliteDB struct {
	dbFile string
}

// NewSQLiteDB returns a new SQLiteDB instance
func NewSQLiteDB(sqliteDBFilePath string) DB {
	return &sqliteDB{
		dbFile: sqliteDBFilePath,
	}
}

func (db *sqliteDB) ListPluginsRows() []PluginInventoryRow {
	return nil
}

func (db *sqliteDB) InsertPluginRow(row PluginInventoryRow) error {
	sdb, err := sql.Open("sqlite3", db.dbFile)
	if err != nil {
		return errors.Wrapf(err, "failed to open the DB from '%s' file", db.dbFile)
	}
	defer sdb.Close()

	_, err = sdb.Exec("INSERT INTO PluginBinaries VALUES(?,?,?,?,?,?,?,?,?,?,?,?);", row.Name, row.Target, row.RecommendedVersion, row.Version, row.Hidden, row.Description, row.Publisher, row.Vendor, row.OS, row.Arch, row.Digest, row.URI)
	if err != nil {
		return errors.Wrap(err, "unable to insert plugin row to the DB")
	}

	return nil
}

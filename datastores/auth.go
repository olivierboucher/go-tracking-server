package datastores

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)
//AuthDatastore represents a wrapper around sql.DB
type AuthDatastore struct {
  sql.DB
}

//NewAuthInstance returns a wrapped db connection
func NewAuthInstance(db *sql.DB) *AuthDatastore {
  return &AuthDatastore{*db}
}
//IsTokenAuthorized checks wheter a token is valid or not
func (d *AuthDatastore) IsTokenAuthorized(token string) (bool, error) {
  //TODO: Change table name
  var exists bool
  tableName := ""
  err := d.DB.QueryRow("SELECT EXISTS(SELECT id FROM ? WHERE token = ? ) as 'exists'", tableName, token).Scan(&exists)

  return exists, err
}

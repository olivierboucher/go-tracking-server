package datastores

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)
//AuthDatastore represents a wrapper around sql.DB
type AuthDatastore struct {
  sql.DB
  init bool
}

var (
  instance AuthDatastore
)

//GetAuthInstance returns the singleton
func GetAuthInstance() *AuthDatastore {
  if !instance.init {
    //TODO: Fix the connection string
    db, err := sql.Open("mysql", "")
    if err != nil {
      //TODO : Handle this
      panic(err)
    }
    instance = AuthDatastore{*db, true}
  }
  return &instance
}
//IsTokenAuthorized checks wheter a token is valid or not
func (d *AuthDatastore) IsTokenAuthorized(token string) (bool, error) {
  //TODO: Change table name
  var exists bool
  tableName := ""
  err := d.DB.QueryRow("SELECT EXISTS(SELECT id FROM ? WHERE token = ? ) as 'exists'", tableName, token).Scan(&exists)

  return exists, err
}

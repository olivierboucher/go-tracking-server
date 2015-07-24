package datastores

import (
  "github.com/gocql/gocql"
)

type StorageDatastore struct {
  *gocql.Session
}
//NewStorageInstance returns a new StorageDatastore containing a connection to the specified cluster
func NewStorageInstance(cluster *gocql.ClusterConfig) (*StorageDatastore, error) {
  session, err := cluster.CreateSession()
  if(err != nil){
    return nil, err
  }
  return &StorageDatastore{session}, nil
}
//Close closes the internal gocql.Session instance
func (d *StorageDatastore) Close() {
  d.Session.Close()
}

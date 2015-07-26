package datastores

import (
  "github.com/gocql/gocql"

  "github.com/OlivierBoucher/go-tracking-server/models"
)
//StorageDatastore wraps a gocql.Session
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
//StoreBatchEvents stores multiple events from models.EventTrackingPayload within a batch
func (d *StorageDatastore) StoreBatchEvents(p *models.EventTrackingPayload) error {
  batch := gocql.NewBatch(gocql.LoggedBatch)

  statement := `INSERT INTO tracking.events (id, client, name, date, properties) VALUES (?, ?, ?, ?, ?)`

  for _,e := range p.Events {
    //Generate a new UUID
    u4, err := gocql.RandomUUID()
    if err != nil {
      return err
    }
    //map properties
    var propMap map[string]string
    propMap = make(map[string]string)
    for _,property := range e.Properties {
      propMap[property.Name] = property.Value
    }
    //Add to batch
    batch.Query(statement, u4.String(), p.Token, e.Name, e.Date, propMap)
  }

  err := d.Session.ExecuteBatch(batch)
  if err != nil {
    return err
  }

  return nil
}

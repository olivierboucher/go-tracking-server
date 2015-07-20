package ctx

import (
  "github.com/OlivierBoucher/go-tracking-server/datastores"
)

type Context struct {
  AuthDb *datastores.AuthDatastore
}

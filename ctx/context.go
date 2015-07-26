package ctx

import (
	"os"

	"github.com/OlivierBoucher/go-tracking-server/datastores"
	"github.com/OlivierBoucher/go-tracking-server/queues"
	"github.com/OlivierBoucher/go-tracking-server/validators"
	"github.com/Sirupsen/logrus"
)

//Context a context that holds database and queue connections
type Context struct {
	AuthDb                     *datastores.AuthDatastore
	StorageDb                  *datastores.StorageDatastore
	Queue                      *queues.RabbitMQConnection
	JSONTrackingEventValidator *validators.JSONEventTrackingValidator
	Logger                     *logrus.Logger
}

//NewContext returns a new context from arguments
func NewContext(a *datastores.AuthDatastore, s *datastores.StorageDatastore, q *queues.RabbitMQConnection, jtv *validators.JSONEventTrackingValidator, env string) *Context {
	var logger *logrus.Logger
	if env == "PROD" {
		//TODO: Define a logger for production
	} else if env == "DEV" {
		logger = &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
	}

	return &Context{
		AuthDb:    a,
		StorageDb: s,
		Queue:     q,
		JSONTrackingEventValidator: jtv,
		Logger: logger,
	}
}

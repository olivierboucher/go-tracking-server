package validators

import (
	"github.com/xeipuuv/gojsonschema"
)

const schema string = `{"$schema":"http://json-schema.org/draft-04/schema#","id":"http://olivierboucher.com/go-tracker-server/draft-02","type":"object","properties":{"token":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/token","type":"string"},"events":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events","type":"array","items":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0","type":"object","properties":{"event":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/event","type":"string"},"date":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/date","type":"string"},"properties":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/properties","type":"array","items":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/properties/2","type":"object","properties":{"name":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/properties/2/name","type":"string"},"value":{"id":"http://olivierboucher.com/go-tracker-server/draft-02/events/0/properties/2/value","type":["string","number"]}},"additionalProperties":false}}},"additionalProperties":false}}},"additionalProperties":false,"required":["token","events"]}`

//JSONEventTrackingValidator a wrapper around a gojsonschema.Schema
type JSONEventTrackingValidator struct {
	Schema *gojsonschema.Schema
}

//NewJSONEventTrackingValidator returns a new JSONEventTrackingValidator with the parsed Schema
func NewJSONEventTrackingValidator() (*JSONEventTrackingValidator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	if err != nil {
		return nil, err
	}
	return &JSONEventTrackingValidator{schema}, err
}

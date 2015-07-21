package validators

import (
  "github.com/xeipuuv/gojsonschema"
)
const schema string = `{"$schema":"http://json-schema.org/draft-04/schema#","id":"http://olivierboucher.com/go-tracker-server/draft-01","type":"array","items":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0","type":"object","properties":{"event":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/event","type":"string"},"date":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/date","type":"string"},"properties":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/properties","type":"array","items":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/properties/0","type":"object","properties":{"name":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/properties/0/name","type":"string"},"value":{"id":"http://olivierboucher.com/go-tracker-server/draft-01/0/properties/0/value","type":["string","number"]}}}}},"required":["event","date","properties"]},"required":["0"]}`
//JSONEventTrackingValidator -
type JSONEventTrackingValidator struct {
  Schema gojsonschema.JSONLoader
}
//NewJSONEventTrackingValidator -
func NewJSONEventTrackingValidator() *JSONEventTrackingValidator {
  return &JSONEventTrackingValidator{gojsonschema.NewStringLoader(schema)}
}

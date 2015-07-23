package models

import (
  _"encoding/json"
)
type AuthenticatedEventTrackingPayload struct {
  ClientID string `json:"clientId"`
  Events []event `json:"events"`
}
//EventTrackingPayload a payload without authentication
type EventTrackingPayload struct {
  Events []event `json:"events"`
}
//Event represents an event in a tracking payload array
type event struct {
  Event string `json:"event"`
  Date string `json:"date"`
  Properties []property `json:"properties"`
}
type property struct {
  Name string `json:"name"`
  Value interface{} `json:"value"`
}

/*func (e *EventTrackingPayload) MarshalJSON() ([]byte, error) {
  //return json.Marshal
}*/

package models

//EventTrackingPayload is a complete payload with client link
type EventTrackingPayload struct {
  ClientID string `json:"clientid"`
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

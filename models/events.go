package models

//EventTrackingPayload represents an event tracking payload's json
type EventTrackingPayload struct {
  Token string `json:"token"`
  Events []Event `json:"events"`
}
//Event represent an event and its datapoints
type Event struct {
  Name string `json:"name"`
  Date string `json:"date"`
  Properties map[string]string `json:"properties"`
}

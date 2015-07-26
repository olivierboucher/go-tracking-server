package models
 import (
   "time"
 )
//EventTrackingPayload represents an event tracking payload's json
type EventTrackingPayload struct {
  Token string `json:"token"`
  Events []Event `json:"events"`
}
//Event represent an event and its datapoints
type Event struct {
  Name string `json:"name"`
  Date time.Time `json:"date"`
  Properties []EventProperty `json:"properties"`
}
type EventProperty struct {
  Name string `json:"name"`
  Value string `json:"value"`
}

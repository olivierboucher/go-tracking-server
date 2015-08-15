# iOS client

This is the first client for iOS written in Obj-C. It is based on Mixpanel's SDK

### Usage

First, initialize the singleton
```objective-c
GTSEventTracker *tracker = [GTSEventTracker sharedInstanceWithToken:@"xxxxxxxxxxx" andClientId:@"xxxxxxxxxxxx"];
```
Then you can start tracking events
```objective-c
//Create an event
NSString *eventName = @"SomeEvent";
NSDictionary *eventProps = @{
  @"prop1": "foo",
  @"prop2": "bar"
};
//Report it
[tracker reportEvent:eventName WithParams:eventProps];
```

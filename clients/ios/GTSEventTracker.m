//
//  GTSEventTracker.m
//
//
//  Created by Olivier Boucher on 2015-07-26.
//  Released under the MIT licence
//  Adaptation of the mixpanel's SDK

#import <CoreTelephony/CTCarrier.h>
#import <CoreTelephony/CTTelephonyNetworkInfo.h>
#import <AdSupport/AdSupport.h>
#import <SystemConfiguration/SystemConfiguration.h>
#import <UIKit/UIDevice.h>
#import <UIKit/UIScreen.h>
#import <UIKit/UIApplication.h>
#import <CoreGraphics/CoreGraphics.h>
#import <sys/sysctl.h>
#import "GTSEventTracker.h"


@interface GTSEventTracker() {
    NSUInteger _flushInterval;
}

@property (nonatomic, strong) dispatch_queue_t serialQueue;
@property (nonatomic, strong) NSString *token;
@property (nonatomic, strong) NSString *clientId;
@property (nonatomic, strong) NSString *serverURL;
@property (nonatomic, strong) NSMutableArray *eventsQueue;
@property (nonatomic, strong) NSDateFormatter *dateFormatter;
@property (nonatomic, strong) CTTelephonyNetworkInfo *telephonyInfo;
@property (nonatomic, strong) NSTimer *timer;
@property (nonatomic) BOOL flushOnBackground;
@property (nonatomic, assign) UIBackgroundTaskIdentifier taskId;
@property (atomic, strong) NSDictionary *automaticProperties;

@end

@implementation GTSEventTracker

// =========================    CLASS METHODS ==========================
+ (id)sharedInstanceWithToken:(NSString *)token andClientId:(NSString *)ClientId {
    static GTSEventTracker *instance;
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        instance = [[GTSEventTracker alloc] initWithToken:token andClientId:ClientId];
    });
    return instance;
}
+ (BOOL)inBackground {
    return [UIApplication sharedApplication].applicationState == UIApplicationStateBackground;
}
// =========================    PRIVATE METHODS =========================
- (instancetype)initWithToken:(NSString *)token andClientId:(NSString *)ClientId {
    if(self = [super init]){
        if (token == nil) {
            token = @"";
        }
        if ([token length] == 0) {
            NSLog(@"%@ warning empty api token", self);
        }
        //TODO: Enforce https once tests done
        self.serverURL = @"http://xx.xx.xx.xx";
        self.token = token;
        self.clientId = ClientId;
        self.flushOnBackground = YES;
        _flushInterval = 60; //In seconds, can change
        self.telephonyInfo = [[CTTelephonyNetworkInfo alloc] init];
        self.eventsQueue = [NSMutableArray array];
        self.taskId = UIBackgroundTaskInvalid;
        self.automaticProperties = [self collectAutomaticProperties];
        NSString *label = [NSString stringWithFormat:@"com.gts.tracking.%@.%p", ClientId, self];
        self.serialQueue = dispatch_queue_create([label UTF8String], DISPATCH_QUEUE_SERIAL);
        self.dateFormatter = [[NSDateFormatter alloc] init];
        [_dateFormatter setDateFormat:@"yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"]; //RFC3339
        [_dateFormatter setTimeZone:[NSTimeZone timeZoneWithAbbreviation:@"UTC"]];
        [_dateFormatter setLocale:[[NSLocale alloc] initWithLocaleIdentifier:@"en_US_POSIX"]];

        [self setUpListeners];
        [self unarchiveEvents];

    }
    return self;
}

- (void)dealloc {
    [[NSNotificationCenter defaultCenter] removeObserver:self];
}

- (void)setUpListeners {
    NSNotificationCenter *notificationCenter = [NSNotificationCenter defaultCenter];

    // Application lifecycle events
    [notificationCenter addObserver:self
                           selector:@selector(applicationWillTerminate:)
                               name:UIApplicationWillTerminateNotification
                             object:nil];
    [notificationCenter addObserver:self
                           selector:@selector(applicationWillResignActive:)
                               name:UIApplicationWillResignActiveNotification
                             object:nil];
    [notificationCenter addObserver:self
                           selector:@selector(applicationDidBecomeActive:)
                               name:UIApplicationDidBecomeActiveNotification
                             object:nil];
    [notificationCenter addObserver:self
                           selector:@selector(applicationDidEnterBackground:)
                               name:UIApplicationDidEnterBackgroundNotification
                             object:nil];
    [notificationCenter addObserver:self
                           selector:@selector(applicationWillEnterForeground:)
                               name:UIApplicationWillEnterForegroundNotification
                             object:nil];
}

#pragma mark - Device identifying

- (NSDictionary *)collectAutomaticProperties {
    NSMutableDictionary *p = [NSMutableDictionary dictionary];
    UIDevice *device = [UIDevice currentDevice];
    NSString *deviceModel = [self deviceModel];
    CGSize size = [UIScreen mainScreen].bounds.size;
    CTCarrier *carrier = [self.telephonyInfo subscriberCellularProvider];

    // Use setValue semantics to avoid adding keys where value can be nil.
    [p setValue:[[NSBundle mainBundle] infoDictionary][@"CFBundleVersion"] forKey:@"AppVersion"];
    [p setValue:[[NSBundle mainBundle] infoDictionary][@"CFBundleShortVersionString"] forKey:@"AppRelease"];
    [p setValue:[self IFA] forKey:@"IFA"];
    [p setValue:carrier.carrierName forKey:@"Carrier"];

    [p addEntriesFromDictionary:@{
                                  @"RivusLib": @"iphone",
                                  @"Manufacturer": @"Apple",
                                  @"OS": [device systemName],
                                  @"OSVersion": [device systemVersion],
                                  @"Model": deviceModel,
                                  @"ScreenHeight": @((NSInteger)size.height),
                                  @"ScreenWidth": @((NSInteger)size.width)
                                  }];
    return [p copy];
}

- (NSString *)deviceModel {
    size_t size;
    sysctlbyname("hw.machine", NULL, &size, NULL, 0);
    char answer[size];
    sysctlbyname("hw.machine", answer, &size, NULL, 0);
    NSString *results = @(answer);
    return results;
}

-(NSString *)IFA {
    NSString *ifa =
    [ASIdentifierManager sharedManager].advertisingTrackingEnabled ?
    [[ASIdentifierManager sharedManager].advertisingIdentifier UUIDString] : @"";

    return ifa;
}

#pragma mark - JSON encoding

- (NSData *)JSONSerializeObject:(id)obj {
    id coercedObj = [self JSONSerializableObjectForObject:obj];
    NSError *error = nil;
    NSData *data = nil;
    @try {
        data = [NSJSONSerialization dataWithJSONObject:coercedObj options:0 error:&error];
    }
    @catch (NSException *exception) {
        NSLog(@"%@ exception encoding json data: %@", self, exception);
    }
    if (error) {
        NSLog(@"%@ error encoding json data: %@", self, error);
    }
    return data;
}

- (id)JSONSerializableObjectForObject:(id)obj {
    //We want to convert everything to string values
    if ([obj isKindOfClass:[NSString class]] ||
        [obj isKindOfClass:[NSNull class]]){
        return obj;
    }
    else if ([obj isKindOfClass:[NSNumber class]]) {
        return [obj stringValue];
    }
    // recurse on containers
    if ([obj isKindOfClass:[NSArray class]]) {
        NSMutableArray *a = [NSMutableArray array];
        for (id i in obj) {
            [a addObject:[self JSONSerializableObjectForObject:i]];
        }
        return [NSArray arrayWithArray:a];
    }
    if ([obj isKindOfClass:[NSDictionary class]]) {
        NSMutableDictionary *d = [NSMutableDictionary dictionary];
        for (id key in obj) {
            NSString *stringKey;
            if (![key isKindOfClass:[NSString class]]) {
                stringKey = [key description];
                NSLog(@"%@ warning: property keys should be strings. got: %@. coercing to: %@", self, [key class], stringKey);
            } else {
                stringKey = [NSString stringWithString:key];
            }
            id v = [self JSONSerializableObjectForObject:obj[key]];
            d[stringKey] = v;
        }
        return [NSDictionary dictionaryWithDictionary:d];
    }
    // some common cases
    if ([obj isKindOfClass:[NSDate class]]) {
        return [self.dateFormatter stringFromDate:obj];
    } else if ([obj isKindOfClass:[NSURL class]]) {
        return [obj absoluteString];
    }
    // default to sending the object's description
    NSString *s = [obj description];
    NSLog(@"%@ warning: property values should be valid json types. got: %@. coercing to: %@", self, [obj class], s);
    return s;
}

#pragma mark - Flushes

- (void)flush {
    dispatch_async(self.serialQueue, ^{
        [self flushEvents];
    });
}

- (void)flushEvents {
    while ([self.eventsQueue count] > 0) {
        NSUInteger batchSize = ([self.eventsQueue count] > 50) ? 50 : [self.eventsQueue count];
        NSArray *batch = [self.eventsQueue subarrayWithRange:NSMakeRange(0, batchSize)];

        NSMutableDictionary *payload = [NSMutableDictionary dictionary];
        //Named token but is actually the client id, the real token is used for server auth
        [payload setValue:self.clientId forKey:@"token"];
        [payload setValue:batch forKey:@"events"];

        NSData *requestData = [self JSONSerializeObject:payload];

        NSLog(@"%@ json payload %@", self, [[NSString alloc]initWithData:requestData encoding:NSUTF8StringEncoding]);

        NSURL *URL = [NSURL URLWithString:[self.serverURL stringByAppendingString:@"/track"]];
        NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:URL];
        [request setValue:@"application/json; charset=UTF-8" forHTTPHeaderField:@"Content-Type"];
        [request setValue:self.token forHTTPHeaderField:@"Tracking-Token"];
        [request setHTTPMethod:@"POST"];
        [request setHTTPBody:requestData];

        NSError *error = nil;

        NSHTTPURLResponse *urlResponse = nil;
        [NSURLConnection sendSynchronousRequest:request returningResponse:&urlResponse error:&error];

        if (error) {
            NSLog(@"%@ network failure: %@", self, error);
            break;
        }

        if ([urlResponse statusCode] != 200) {
            NSLog(@"%@ %@ server rejected some items - status code %ld", self, @"/track", [urlResponse statusCode]);
        };

        [self.eventsQueue removeObjectsInArray:batch];
    }
}

#pragma mark - Timer related methods

- (NSUInteger)flushInterval
{
    @synchronized(self) {
        return _flushInterval;
    }
}

- (void)setFlushInterval:(NSUInteger)interval
{
    @synchronized(self) {
        _flushInterval = interval;
    }
    [self startFlushTimer];
}

- (void)startFlushTimer
{
    [self stopFlushTimer];
    dispatch_async(dispatch_get_main_queue(), ^{
        if (self.flushInterval > 0) {
            self.timer = [NSTimer scheduledTimerWithTimeInterval:self.flushInterval
                                                          target:self
                                                        selector:@selector(flush)
                                                        userInfo:nil
                                                         repeats:YES];
            NSLog(@"%@ started flush timer: %@", self, self.timer);
        }
    });
}

- (void)stopFlushTimer
{
    dispatch_async(dispatch_get_main_queue(), ^{
        if (self.timer) {
            [self.timer invalidate];
            NSLog(@"%@ stopped flush timer: %@", self, self.timer);
        }
        self.timer = nil;
    });
}

#pragma mark - Persistence

- (void)archiveEvents {
    NSString *filePath = [self eventsFilePath];
    NSMutableArray *eventsQueueCopy = [NSMutableArray arrayWithArray:[self.eventsQueue copy]];

    if (![NSKeyedArchiver archiveRootObject:eventsQueueCopy toFile:filePath]) {
        NSLog(@"%@ unable to archive events data", self);
    }
}
- (id)unarchiveFromFile:(NSString *)filePath {
    id unarchivedData = nil;
    @try {
        unarchivedData = [NSKeyedUnarchiver unarchiveObjectWithFile:filePath];
        NSLog(@"%@ unarchived data from %@: %@", self, filePath, unarchivedData);
    }
    @catch (NSException *exception) {
        NSLog(@"%@ unable to unarchive data in %@, starting fresh", self, filePath);
        unarchivedData = nil;
    }
    if ([[NSFileManager defaultManager] fileExistsAtPath:filePath]) {
        NSError *error;
        BOOL removed = [[NSFileManager defaultManager] removeItemAtPath:filePath error:&error];
        if (!removed) {
            NSLog(@"%@ unable to remove archived file at %@ - %@", self, filePath, error);
        }
    }
    return unarchivedData;
}

- (void)unarchiveEvents {
    self.eventsQueue = (NSMutableArray *)[self unarchiveFromFile:[self eventsFilePath]];
    if (!self.eventsQueue) {
        self.eventsQueue = [NSMutableArray array];
    }
}

- (NSString *)filePathForData:(NSString *)data {
    NSString *filename = [NSString stringWithFormat:@"tracking-%@-%@.plist", self.clientId, data];
    return [[NSSearchPathForDirectoriesInDomains(NSLibraryDirectory, NSUserDomainMask, YES) lastObject]
            stringByAppendingPathComponent:filename];
}

- (NSString *)eventsFilePath {
    return [self filePathForData:@"events"];
}

#pragma mark - UIApplication notifications

- (void)applicationDidBecomeActive:(NSNotification *)notification {
    [self startFlushTimer];

}
- (void)applicationWillResignActive:(NSNotification *)notification {
    [self stopFlushTimer];
}

- (void)applicationDidEnterBackground:(NSNotification *)notification {
    NSLog(@"%@ did enter background", self);

    self.taskId = [[UIApplication sharedApplication] beginBackgroundTaskWithExpirationHandler:^{
        NSLog(@"%@ flush %lu cut short", self, (unsigned long)self.taskId);
        [[UIApplication sharedApplication] endBackgroundTask:self.taskId];
        self.taskId = UIBackgroundTaskInvalid;
    }];

    NSLog(@"%@ starting background cleanup task %lu", self, (unsigned long)self.taskId);

    if (self.flushOnBackground) {
        [self flush];
    }

    dispatch_async(_serialQueue, ^{
        [self archiveEvents];
        NSLog(@"%@ ending background cleanup task %lu", self, (unsigned long)self.taskId);
        if (self.taskId != UIBackgroundTaskInvalid) {
            [[UIApplication sharedApplication] endBackgroundTask:self.taskId];
            self.taskId = UIBackgroundTaskInvalid;
        }
    });
}

- (void)applicationWillEnterForeground:(NSNotificationCenter *)notification {
    NSLog(@"%@ will enter foreground", self);
    dispatch_async(self.serialQueue, ^{
        if (self.taskId != UIBackgroundTaskInvalid) {
            [[UIApplication sharedApplication] endBackgroundTask:self.taskId];
            self.taskId = UIBackgroundTaskInvalid;
        }
    });
}

- (void)applicationWillTerminate:(NSNotification *)notification {
    dispatch_async(_serialQueue, ^{
        [self archiveEvents];
    });
}
// =========================    PUBLIC METHODS ==========================
- (void)reportEvent:(NSString*)name withParams:(NSDictionary*)params {
    dispatch_async(self.serialQueue, ^{
        NSDate *now = [NSDate date];
        //We initialize an array
        NSMutableArray *p = [[NSMutableArray alloc] init];
        //Append automatic properties
        for( id key in self.automaticProperties)
        {
            NSDictionary *prop = @{
                @"name": key,
                @"value" : [self.automaticProperties objectForKey:key]
            };

            [p addObject:prop];
        }
        //Append custom properties
        for( id key in params)
        {
            NSDictionary *prop = @{
                @"name": key,
                @"value" : [params objectForKey:key]
            };

            [p addObject:prop];
        }

        NSDictionary *event = @{@"name": name, @"date": [self.dateFormatter stringFromDate:now], @"properties": [NSArray arrayWithArray:p]};

        [self.eventsQueue addObject:event];

        if ([GTSEventTracker inBackground]) {
            [self archiveEvents];
        }
    });
}
@end

//
//  GTSEventTracker.h
//
//
//  Created by Olivier Boucher on 2015-07-26.
//  Released under the MIT licence
//

#import <Foundation/Foundation.h>

@interface GTSEventTracker : NSObject

+ (id)sharedInstanceWithToken:(NSString *)token andClientId:(NSString *)ClientId;
- (void) reportEvent:(NSString*)name WithParams:(NSDictionary*)params;
@end

//go:build darwin && cgo

// Keeps the wrapping secret in a device-only macOS Keychain item requiring user presence.
// Security.framework handles credential prompts and returns only the requested app-owned item.
package keydialog

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework Security
#import <Foundation/Foundation.h>
#import <Security/Security.h>
#include <stdlib.h>
#include <string.h>
static char *keychain(const char *operation, const char *value) {
    @autoreleasepool {
        NSString *op = [NSString stringWithUTF8String:operation];
        NSString *input = [NSString stringWithUTF8String:value];
        BOOL create = [op isEqualToString:@"store"];
        NSString *identifier = create ? NSUUID.UUID.UUIDString : input;
        NSMutableDictionary *query = [@{(__bridge id)kSecClass:(__bridge id)kSecClassGenericPassword,
            (__bridge id)kSecAttrService:@"org.apparat.node", (__bridge id)kSecAttrAccount:identifier} mutableCopy];
        if (create) {
            SecAccessControlRef control = SecAccessControlCreateWithFlags(NULL,kSecAttrAccessibleWhenUnlockedThisDeviceOnly,kSecAccessControlUserPresence,NULL);
            if (!control) return NULL;
            query[(__bridge id)kSecAttrAccessControl] = CFBridgingRelease(control);
            query[(__bridge id)kSecValueData] = [input dataUsingEncoding:NSUTF8StringEncoding];
            return SecItemAdd((__bridge CFDictionaryRef)query,NULL)==errSecSuccess ? strdup(identifier.UTF8String) : NULL;
        }
        if ([op isEqualToString:@"remove"]) {
            SecItemDelete((__bridge CFDictionaryRef)query);
            return strdup("");
        }
        query[(__bridge id)kSecReturnData] = @YES;
        query[(__bridge id)kSecUseOperationPrompt] = @"Unlock your Apparat node";
        CFTypeRef found = NULL;
        if (SecItemCopyMatching((__bridge CFDictionaryRef)query,&found)!=errSecSuccess) return NULL;
        NSString *answer = [[NSString alloc] initWithData:CFBridgingRelease(found) encoding:NSUTF8StringEncoding];
        return answer ? strdup(answer.UTF8String) : NULL;
    }
}
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

func macCredential(packet string) (string, bool) {
	var req struct{ Op, Value string }
	if json.Unmarshal([]byte(packet), &req) != nil {
		return "null", true
	}
	switch req.Op {
	case "available":
		return `"Use macOS Keychain"`, true
	case "store", "load", "remove":
	default:
		return "", false
	}
	op := C.CString(req.Op)
	val := C.CString(req.Value)
	defer C.free(unsafe.Pointer(op))
	defer C.free(unsafe.Pointer(val))
	res := C.keychain(op, val)
	if res == nil {
		return "null", true
	}
	defer C.free(unsafe.Pointer(res))
	data, _ := json.Marshal(C.GoString(res))
	return string(data), true
}

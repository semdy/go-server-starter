package constant

import (
	"fmt"
	"time"
)

const (
	MAX_PAGE_SIZE                = 500                    // max page size
	DEFAULT_PAGE_SIZE            = 20                     // default page size
	REDIS_EXPIRE_OF_AUTH_ROLES   = 5 * time.Minute        // redis expire of auth roles
	REDIS_KEY_OF_RATE_LIMIT      = "api:rate_limit:%s:%s" // redis key of rate limit: zone:ip
	REDIS_KEY_OF_AUTH_ROLES      = "auth:roles:%d:%s"     // redis key of auth roles: tenantID:uniCode
	REDIS_KEY_OF_VERIFY_CODE     = "verify:%s:%s"         // redis key of verify code: type:mobileOrEmail
	REDIS_EXPIRE_OF_VERIFY_CODE  = 5 * time.Minute        // redis expire of verify code
	REDIS_EXPIRE_OF_VERIFY_LIMIT = 60 * time.Second       // resend cooldown

	CTX_KEY_OF_LOCALE        = "ctx:locale"
	CTX_KEY_OF_TRANSLATOR    = "ctx:translator"
	CTX_KEY_OF_USER_UNI_CODE = "ctx:user_uni_code"
	CTX_KEY_OF_TENANT_ID     = "ctx:tenant_id"
)

func RedisKeyOfRateLimit(zone string, ip string) string {
	return fmt.Sprintf(REDIS_KEY_OF_RATE_LIMIT, zone, ip)
}

func RedisKeyOfAuthRoles(tenantID uint64, uniCode string) string {
	return fmt.Sprintf(REDIS_KEY_OF_AUTH_ROLES, tenantID, uniCode)
}

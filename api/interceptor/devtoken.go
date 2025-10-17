package interceptor

import (
	"strconv"
	"strings"
	"time"

	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/security"

	"github.com/gin-gonic/gin"
)

func DevTokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		allHeaders, ok := c.Get("HEADERS")
		if !ok {
			log.Info("unable to get headers")
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token check failed")
			return
		}
		allHeadersMap := allHeaders.(common.HeaderParam)
		// if allHeadersMap.XAuth == "123456" {
		// 	c.Set("user_wallet", "0x0")
		// 	c.Set("user_id", "1")
		// 	c.Next()
		// 	return
		// }
		log.Info("Token Parse:", allHeadersMap)
		token, err := security.Decrypt(c.Request.Header.Get("DAUTH"))
		log.Info("TOKENCHECK ", c.Request.Header.Get("DAUTH"), token)
		if err != nil {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token check failed")
			return
		}
		tokenArr := strings.Split(token, "|")
		if len(tokenArr) != 2 {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token length error")
			return
		}
		expireTs, err := strconv.ParseInt(tokenArr[1], 10, 64)
		if err != nil {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token format error")
			return
		}
		if time.Now().Unix()-expireTs > int64(common.TOKEN_DURATION.Seconds()) {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token expired error")
			return
		}

		c.Set("dev_id", tokenArr[0])

		c.Next()
	}
}

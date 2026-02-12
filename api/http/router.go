package http

import (
	"github.com/gin-gonic/gin"

	"chaos/api/api/http/controller/auth"
	"chaos/api/api/http/controller/developer"
	"chaos/api/api/http/controller/home"
	preauth "chaos/api/api/http/controller/preauth"
	"chaos/api/api/interceptor"
	"chaos/api/log"
)

func Routers(e *gin.RouterGroup) {

	homeGroup := e.Group("/")
	homeGroup.GET("public/config", home.Public)
	homeGroup.GET("public/assets", home.Assets)

	homeGroup.GET("public/game", home.Game)
	homeGroup.GET("public/game/:game_id", home.GameDetail)
	homeGroup.GET("home/season", home.Season)

	homeGroup.GET("leaderboard/:season_code", home.Leaderboard)
	homeGroup.GET("leaderboard/:season_code/game/:game_id", home.LeaderboardGame)

	homeGroup.GET("companies", home.CompanyList)
	homeGroup.GET("chart/:symbol", home.ChartBySymbol)

	preAuthGroup := e.Group("/preauth")
	preAuthGroup.POST("get_msg", preauth.GetAuthMsg)
	preAuthGroup.POST("verify_msg", preauth.VerifyMessage)
	preAuthGroup.POST("register", preauth.Register)
	preAuthGroup.POST("verify", preauth.Verify)
	preAuthGroup.POST("login", preauth.Login)
	preAuthGroup.POST("/developer/login", preauth.DeveloperLogin)
	preAuthGroup.POST("/developer/register", preauth.DeveloperRegister)
	preAuthGroup.POST("/developer/verify", preauth.DeveloperVerify)

	authGroup := e.Group("/auth", interceptor.TokenInterceptor())

	authGroup.GET("/user/profile", auth.Profile)
	authGroup.POST("/user/profile", auth.UpdateProfile)
	authGroup.POST("/user/profile/email/send", auth.SendEmail)
	authGroup.POST("/user/profile/email/verify", auth.VerifyEmail)
	authGroup.POST("/user/profile/wallet/verify", auth.VerifyWallet)

	authGroup.GET("/user/balance", auth.QueryBalance)
	authGroup.GET("/user/transactions", auth.FetchTransactions)
	authGroup.GET("/user/game/session", auth.QueryGameSession)
	authGroup.POST("/user/game/session", auth.ConfirmGameSession)

	authGroup.POST("/season/join", auth.JoinSeason)
	authGroup.GET("/season/joined", auth.CheckSeasonJoined)

	authGroup.POST("/account/balance/topup", auth.BalanceTopupReport)
	authGroup.GET("/account/balance/topup", auth.BalanceTopupList)
	authGroup.POST("/account/balance/withdraw/request", auth.BalanceWithdrawRequest)
	authGroup.GET("/account/balance/withdraw/check", auth.BalanceWithdrawCheck)

	// Twitter OAuth callback endpoint - requires authentication
	authGroup.GET("/thirdpart/x/callback", auth.XCallback)

	devAuthGroup := e.Group("/devauth", interceptor.DevTokenInterceptor())
	devAuthGroup.GET("/profile", developer.Profile)
	devAuthGroup.POST("/profile", developer.UpdateProfile)
	devAuthGroup.GET("/game/list", developer.ListGame)
	devAuthGroup.GET("/game/detail", developer.DetailGame)
	devAuthGroup.POST("/game/update", developer.UpdateGame)
	devAuthGroup.POST("/game/refresh_key", developer.RefreshGameKey)
	devAuthGroup.POST("/game/testing/start", developer.TestingStart)
	devAuthGroup.POST("/game/testing/finish", developer.TestingFinish)
	devAuthGroup.POST("/game/setting/save", developer.SaveSetting)
	devAuthGroup.POST("/game/setting/delete", developer.DeleteSetting)
	devAuthGroup.POST("/game/online/audit", developer.SubmitOnlineAudit)
	devAuthGroup.POST("/game/session/list", developer.GameSessionList)

	/***** Intend to use api in future ****/
	// authGroup.POST("ref_uri", auth.Ref)
	// authGroup.POST("/ref/stat", auth.RefCount)
	// authGroup.POST("/ref/list", auth.RefList)
	// authGroup.POST("/daily/checkin", auth.DailyCheckin)
	// authGroup.GET("/daily/checkin", auth.DailyCheckinRecord)
	// authGroup.POST("/attention/check", auth.PayAttention)
	// authGroup.POST("/trans/swap", auth.Trans)

	// authGroup.POST("/asset/board", auth.AssetView)
	// authGroup.POST("/asset/list", auth.AssetList)
	// authGroup.POST("/asset/trans", auth.AssetTrans)

	// v2Group := e.Group("/v2")
	// v2Group.POST("index", homev2.UpdateLeaderboard)
	// v2Group.GET("/k/:chain/:ca", homev2.K)
	// v2Group.GET("/token/holders/:chain/:ca", homev2.TokenHolders)
	// v2Group.GET("/search/:key", homev2.Search)
	// v2Group.GET("/token/:chain/:ca", homev2.TokenInfoV2)
	// v2Group.GET("/pair/:chain/:ca", homev2.PairFlowV2)
	// v2Group.GET("/token/:chain/newlist", homev2.TokenNewList)

	// v2Group.GET("/token/chgs/:chain/:ca", homev2.TokenChgV2)
	// v2AuthGroup := v2Group.Group("/auth", interceptor.TokenInterceptor())
	// v2AuthGroup.GET("/asset-token/trans", authv2.AssetTokenTrans)
	// v2AuthGroup.GET("/asset/list", authv2.AssetList)

	log.Info(preAuthGroup, authGroup)
}

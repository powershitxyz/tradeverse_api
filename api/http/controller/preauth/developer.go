package preauth

import (
	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/security"
	"chaos/api/system"
	"chaos/api/utils"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

func DeveloperLogin(c *gin.Context) {
	var req LoginRequest
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid request" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var userInfo model.GameDeveloper
	db.Model(&model.GameDeveloper{}).
		Where("email = ?", req.Email).
		First(&userInfo)

	if userInfo.ID == 0 || req.Password != userInfo.Password {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "email not found or password incorrect"
		c.JSON(http.StatusOK, res)
		return
	}

	expireTs := time.Now().Add(common.TOKEN_DURATION).Unix()

	tokenOrig := fmt.Sprintf("%d|%d", userInfo.ID, expireTs)
	tokenEnc, err := security.Encrypt([]byte(tokenOrig))
	if err != nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "token gen error:" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = gin.H{
		"dev_id":            userInfo.ID,
		"email":             userInfo.Email,
		"provisional_token": tokenEnc,
	}
	c.JSON(http.StatusOK, res)
}

func DeveloperRegister(c *gin.Context) {
	var req DeveloperRegisterRequest
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid request" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var userInfo model.GameDeveloper
	db.Model(&model.GameDeveloper{}).
		Where("email = ?", req.Email).
		First(&userInfo)

	if userInfo.ID > 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "email already registered"
		c.JSON(http.StatusOK, res)
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "Invalid email format"
		c.JSON(http.StatusOK, res)
		return
	}

	err := utils.SendVerifyCodeMailAPI(req.Email, "20")
	if err != nil {
		log.Error("send email err", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "send email failed"
		c.JSON(http.StatusOK, res)
		return
	}

	userInfo.Email = req.Email
	userInfo.Password = req.Password
	userInfo.AddTime = time.Now()

	db.Create(&userInfo)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func DeveloperVerify(c *gin.Context) {
	var req DeveloperVerifyRequest
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid request" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var userInfo model.GameDeveloper
	db.Model(&model.GameDeveloper{}).
		Where("email = ?", req.Email).
		First(&userInfo)

	if userInfo.ID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "email not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var verifyProcess model.VerificationProcess
	db.Model(&model.VerificationProcess{}).
		Where("target = ? and code = ? and type = ? and sort = ?", req.Email, req.Code, "10", "20").
		First(&verifyProcess)
	if verifyProcess.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "verification code not sent"
		c.JSON(http.StatusOK, res)
		return
	}

	if time.Now().After(verifyProcess.AddTime.Add(time.Duration(verifyProcess.ValidatePeriod) * time.Second)) {
		res.Code = codes.CODE_ERR_REQ_EXPIRED
		res.Msg = "verification code expired"
		c.JSON(http.StatusOK, res)
		return
	}

	userInfo.Status = "20"
	db.Save(&userInfo)
	verifyProcess.Status = "100"
	db.Save(&verifyProcess)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = gin.H{
		"dev_id": userInfo.ID,
		"email":  userInfo.Email,
	}
	c.JSON(http.StatusOK, res)
}

package auth

import (
	"chaos/api/api/common"
	"chaos/api/chain"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"chaos/api/tools"
	"context"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
)

var signInLocks sync.Map
var treeLocks sync.Map
var tree *tools.MerkleTree

func QueryAmount(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	userWallet, ok := c.Get("user_wallet")
	_, _ = c.Get("user_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "need login"
		c.JSON(http.StatusOK, res)
		return
	}

	targetWallet, exist := c.GetQuery("target_wallet")
	if !exist {
		targetWallet = userWallet.(string)
	}

	db := system.GetDb()
	var chaosHolder model.ChaosHolder
	var chaosTrans []model.ChaosTrans
	result := db.Model(&model.ChaosHolder{}).Where("wallet_address = ?", targetWallet).First(&chaosHolder)
	db.Model(&model.ChaosTrans{}).Where("user_wallet = ?", targetWallet).Where("status = ?", "20").Find(&chaosTrans)

	var balance uint64
	if result.Error == nil && chaosHolder.ID != 0 {
		if chaosHolder.SnapAmount > 0 && chaosHolder.AfterTx == 0 {
			balance = uint64(chaosHolder.SnapAmount)
		}
	}

	var submittedAmount uint64
	if balance > 0 {
		for _, chaos := range chaosTrans {
			submittedAmount += chaos.Amount
		}
	}

	var deci = decimal.NewFromInt(1000000)
	realBal := decimal.NewFromUint64(balance).Div(deci).Round(6)
	submittedReal := decimal.NewFromUint64(submittedAmount).Div(deci).Round(6)

	var bindWallet string = chaosHolder.BscWallet
	if chaosHolder.Step == 0 {
		bindWallet = ""
	}
	res.Data = map[string]interface{}{
		"snapshot_amount":  realBal.String(),
		"submitted_amount": submittedReal.String(),
		"has_tx":           chaosHolder.AfterTx,
		"step":             chaosHolder.Step,
		"bind_wallet":      bindWallet,
	}

	c.JSON(http.StatusOK, res)
}

func SubmitTrans(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	var req struct {
		TxHash string `json:"tx_hash"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(req.TxHash) <= 80 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "tx hash error"
		c.JSON(http.StatusOK, res)
		return
	}

	userWallet, ok := c.Get("user_wallet")
	_, _ = c.Get("user_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "need login"
		c.JSON(http.StatusOK, res)
		return
	}

	// 这里我需要判断是否是今天晚上9点，超过了这个时间要返回对应提示
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		now := time.Now().In(loc)
		deadline := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, loc)
		if now.After(deadline) {
			res.Code = codes.CODE_ERR_REQ_EXPIRED
			res.Msg = "service closed after 21:00 (Asia/Shanghai), please try again tomorrow"
			c.JSON(http.StatusOK, res)
			return
		}
	}

	_, loaded := signInLocks.LoadOrStore(req.TxHash, struct{}{})
	defer func() {
		signInLocks.Delete(req.TxHash)
	}()
	if loaded {
		res.Code = codes.CODE_SUCCESS
		res.Msg = "checking, please do not repeat the operation"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var chaosTrans model.ChaosTrans
	result := db.Model(&model.ChaosTrans{}).Where("tx_hash = ?", req.TxHash).First(&chaosTrans)
	if result.Error == nil && chaosTrans.ID > 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "tx already submitted"
		c.JSON(http.StatusOK, res)
		return
	}

	chaosTrans = model.ChaosTrans{
		UserWallet: userWallet.(string),
		TxHash:     req.TxHash,
		Amount:     0,
		AddTime:    time.Now(),
		Status:     "00",
	}
	db.Save(&chaosTrans)

	/**
	status:
	00: pending
	11: failed tx
	19: old tx than snap slot
	21: data pruned
	20: success
	*/
	go func(txHash string, chaosId uint64, wallet string) {
		db := system.GetDb()
		targetAddress := os.Getenv("TARGETWALLET")
		tx, err := chain.AnalyzeTxMintDelta(context.Background(), "https://solana.publicnode.com", txHash, "3ovR2CQczTA3T6t37UnyMj2pD3VgvJ6Dvec6rvFishot")
		if err != nil {
			if strings.HasPrefix(err.Error(), "data_pruned_error") {
				log.Info("data_pruned_error: ", txHash)
				db.Model(&model.ChaosTrans{}).Where("id = ?", chaosId).Update("status", "21")
			}
			return
		}

		if !tx.Success {
			db.Model(&model.ChaosTrans{}).Where("id = ?", chaosId).Update("status", "11")
		} else {
			if tx.Slot <= 357547209 {
				db.Model(&model.ChaosTrans{}).Where("id = ?", chaosId).Update("status", "19")
				return
			}

			var realAmount = decimal.NewFromUint64(0)
			var decimals = decimal.NewFromInt(1000000)

			var transferOut = decimal.NewFromUint64(0)
			var transferIn = decimal.NewFromUint64(0)
			for _, delta := range tx.Deltas {
				d, _ := decimal.NewFromString(delta.Delta)
				switch delta.Owner {
				case targetAddress:
					transferIn = transferIn.Add(d)
				case wallet:
					transferOut = transferOut.Add(d)
				}
			}

			if transferIn.Add(transferOut).Cmp(realAmount) == 0 {
				realAmount = transferIn.Mul(decimals).Abs().Round(0)
			}
			tx := db.Begin()
			err = tx.Model(&model.ChaosTrans{}).Where("id = ?", chaosId).Updates(map[string]interface{}{
				"status": "20",
				"amount": realAmount.BigInt().Uint64(),
			}).Error
			if err != nil {
				log.Error("[DB-TX] update chaos_trans error", err, txHash, wallet)
				tx.Rollback()
				return
			}
			err = tx.Model(&model.ChaosHolder{}).Where("wallet_address = ?", wallet).Update("step", 1).Error
			if err != nil {
				log.Error("[DB-TX] update chaos_holder error", err, txHash, wallet)
				tx.Rollback()
				return
			}
			tx.Commit()
		}
	}(req.TxHash, chaosTrans.ID, userWallet.(string))

	c.JSON(http.StatusOK, res)
}

func QueryN(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	userWallet, ok := c.Get("user_wallet")
	_, _ = c.Get("user_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please connect your wallet and sign in"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var chaosHolder model.NAirdropClaim
	result := db.Model(&model.NAirdropClaim{}).Where("bsc_wallet = ?", userWallet.(string)).First(&chaosHolder)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "system busy, please try again later"
		c.JSON(http.StatusOK, res)
		return
	}

	var airdropAmount uint64 = chaosHolder.SolSnapAmount
	if airdropAmount > chaosHolder.SolTransferAmount {
		airdropAmount = chaosHolder.SolTransferAmount
	}

	res.Data = map[string]interface{}{
		// "sol_snap_amount": chaosHolder.SolSnapAmount,
		"airdrop_amount": airdropAmount,
	}

	c.JSON(http.StatusOK, res)
}

func buildTree() error {
	db := system.GetDb()
	var wls []model.NAirdropClaim
	err := db.Find(&wls).Error
	if err != nil {
		return err
	}

	var contents []tools.TreeContent

	// var contentData []byte

	for _, vwl := range wls {
		var airdropAmount uint64 = vwl.SolSnapAmount
		if airdropAmount > vwl.SolTransferAmount {
			airdropAmount = vwl.SolTransferAmount
		}
		v := tools.EncodePack(vwl.BscWallet, big.NewInt(int64(airdropAmount)))

		// contentData = append(contentData, []byte(v+"\n")...)

		// fmt.Println(v)
		contents = append(contents, tools.DefaultCont{
			Data: v,
		})
	}

	tree, err = tools.NewTreeWithHashStrategySorted(contents, sha3.NewLegacyKeccak256, true)
	if err != nil {
		return err
	}
	return nil
}

func QueryNProof(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	userWallet, ok := c.Get("user_wallet")
	_, _ = c.Get("user_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please connect your wallet and sign in"
		c.JSON(http.StatusOK, res)
		return
	}

	if tree == nil {
		// 只有第一次成功存入的协程负责删除锁
		if _, loaded := signInLocks.LoadOrStore("tree-building", struct{}{}); loaded {
			res.Code = codes.CODE_SUCCESS
			res.Msg = "building proof, please do not repeat the operation"
			c.JSON(http.StatusOK, res)
			return
		}
		defer signInLocks.Delete("tree-building")

		// 再次校验，防止在拿到锁前已构建完成
		if tree == nil {
			if err := buildTree(); err != nil {
				res.Code = codes.CODE_ERR_UNKNOWN
				res.Msg = "system busy, please try again later"
				c.JSON(http.StatusOK, res)
				return
			}
		}
	}

	db := system.GetDb()
	var chaosHolder model.NAirdropClaim
	result := db.Model(&model.NAirdropClaim{}).Where("bsc_wallet = ?", userWallet.(string)).First(&chaosHolder)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "system busy, please try again later"
		c.JSON(http.StatusOK, res)
		return
	}

	var airdropAmount uint64 = chaosHolder.SolSnapAmount
	if airdropAmount > chaosHolder.SolTransferAmount {
		airdropAmount = chaosHolder.SolTransferAmount
	}

	v := tools.EncodePack(chaosHolder.BscWallet, big.NewInt(int64(airdropAmount)))
	custLeaf := tools.DefaultCont{
		Data: v,
	}

	merklePath, index, err := tree.GetMerklePathHex(custLeaf)
	log.Info("get merkle path: ", merklePath, index, err)

	res.Data = map[string]interface{}{
		"proof":          merklePath,
		"n_address":      "0xB82582bf335bc4f57ec3c536E67019e1FA263F81",
		"airdrop_amount": airdropAmount,
	}
	c.JSON(http.StatusOK, res)
}

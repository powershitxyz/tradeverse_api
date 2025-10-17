package thirdpart

import "encoding/json"

// 顶级响应结构
type GoplusApiResponse struct {
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Result  map[string]GoplusTokenInfo `json:"result"`
}

// Token 详细信息
type GoplusTokenInfo struct {
	BalanceMutableAuthority       GoplusAuthorityStatus `json:"balance_mutable_authority"`
	Closable                      GoplusAuthorityStatus `json:"closable"`
	Creators                      []GoplusAddressStatus `json:"creators"`
	DefaultAccountState           string                `json:"default_account_state"`
	DefaultAccountStateUpgradable GoplusAuthorityStatus `json:"default_account_state_upgradable"`
	Dex                           []GoplusDexInfo       `json:"dex"`
	Freezable                     GoplusAuthorityStatus `json:"freezable"`
	Holders                       []GoplusHolderInfo    `json:"holders"`
	LpHolders                     []GoplusHolderInfo    `json:"lp_holders"`
	Metadata                      GoplusMetadata        `json:"metadata"`
	MetadataMutable               GoplusAuthorityStatus `json:"metadata_mutable"`
	Mintable                      GoplusAuthorityStatus `json:"mintable"`
	NonTransferable               string                `json:"non_transferable"`
	TotalSupply                   string                `json:"total_supply"`
	TransferFee                   TransferFeeWrapper    `json:"transfer_fee"`
	TransferFeeUpgradable         GoplusAuthorityStatus `json:"transfer_fee_upgradable"`
	TransferHook                  []string              `json:"transfer_hook"`
	TransferHookUpgradable        GoplusAuthorityStatus `json:"transfer_hook_upgradable"`
	TrustedToken                  int                   `json:"trusted_token"`
}

// 授权状态
type GoplusAuthorityStatus struct {
	Authority []GoplusAddressStatus `json:"authority"`
	Status    string                `json:"status"`
}

// DEX 交易池信息
type GoplusDexInfo struct {
	Day      GoplusDexStats `json:"day"`
	DexName  string         `json:"dex_name"`
	FeeRate  string         `json:"fee_rate"`
	ID       string         `json:"id"`
	LpAmount string         `json:"lp_amount"` // 可能为 null
	Month    GoplusDexStats `json:"month"`
	OpenTime string         `json:"open_time"`
	Price    string         `json:"price"`
	Tvl      string         `json:"tvl"`
	Type     string         `json:"type"`
	Week     GoplusDexStats `json:"week"`
}

// DEX 统计数据
type GoplusDexStats struct {
	PriceMax string `json:"price_max"`
	PriceMin string `json:"price_min"`
	Volume   string `json:"volume"`
}

// Token 持有人信息
type GoplusHolderInfo struct {
	Account      string                   `json:"account"`
	Balance      string                   `json:"balance"`
	IsLocked     int                      `json:"is_locked"`
	LockedDetail []GoplusHolderLockDetail `json:"locked_detail"`
	Percent      string                   `json:"percent"`
	Tag          string                   `json:"tag"`
	TokenAccount string                   `json:"token_account"`
}

// Metadata 结构
type GoplusMetadata struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	URI         string `json:"uri"`
}

type GoplusAddressStatus struct {
	Address          string `json:"address"`
	MaliciousAddress int    `json:"malicious_address"`
}

type GoplusHolderLockDetail struct {
	Amount  string `json:"amount"`
	EndTime string `json:"end_time"`
	OptTime string `json:"opt_time"`
}

type GoplusTransferfeeCurrentRate struct {
	FeeRate    string `json:"fee_rate"`
	MaximumFee string `json:"maximum_fee"`
}
type GoplusTransferfeeScheduleRate struct {
	Epoch string `json:"epoch"`
	GoplusTransferfeeCurrentRate
}
type GoplusTransferfee struct {
	CurrentFeeRate   GoplusTransferfeeCurrentRate    `json:"current_fee_rate"`
	ScheduledFeeRate []GoplusTransferfeeScheduleRate `json:"scheduled_fee_rate"`
}

type TransferFeeWrapper struct {
	Data json.RawMessage `json:"transfer_fee"`
}

// 解析 transfer_fee 逻辑
func (t *TransferFeeWrapper) UnmarshalJSON(data []byte) error {
	// 先尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// 如果是字符串（""），直接返回，不解析
		return nil
	}

	// 否则解析为结构体
	var fee GoplusTransferfee
	if err := json.Unmarshal(data, &fee); err != nil {
		return err
	}

	// 成功解析后存入
	t.Data = data
	return nil
}

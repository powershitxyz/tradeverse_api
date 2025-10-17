package chain

import (
	topupabi "chaos/api/chain/abi"
	"chaos/api/config"
	"chaos/api/tools"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TopupTxInfo 解析后的充值交易信息
type TopupTxInfo struct {
	TxHash      string
	From        string
	To          string
	Contract    string
	Amount      *big.Int
	BlockNumber uint64
	BlockTime   time.Time
	Status      string // pending/success/failed
	UserFromLog string // Deposited 事件中的 user（与 From 一致时可作为校验）
	MainID      uint64
	RefFlowID   uint64
	Op          int
	ChainID     uint64
	LockID      [32]byte
}

type QueuePassObject struct {
	ChainID   uint64
	TxHash    string
	MainID    uint64
	RefFlowID uint64
	Op        int
}

// pickRPCByChainID 根据 EVM chainID 选择一个可用的 RPC（取对应配置的第一个）
func pickRPCByChainID(chainID uint64) (string, error) {
	// 常见链 ID 到配置名的简单映射（需要与配置文件中的 chain.name 对应）
	idToName := map[uint64]string{
		1:     "ETH",
		56:    "BSC",
		97:    "BSC_TESTNET",
		137:   "POLYGON",
		10:    "OPTIMISM",
		42161: "ARBITRUM",
	}
	if name, ok := idToName[chainID]; ok {
		if cfg := config.GetRpcConfig(name); cfg != nil {
			rpcs := cfg.GetRpc()
			if len(rpcs) > 0 {
				return rpcs[0], nil
			}
		}
	}
	// 回退：优先 BSC
	if cfg := config.GetRpcConfig("BSC"); cfg != nil {
		rpcs := cfg.GetRpc()
		if len(rpcs) > 0 {
			return rpcs[0], nil
		}
	}
	return "", fmt.Errorf("no rpc found for chainID %d", chainID)
}

// ParseTopupTx 解析 TopupLogic 的 deposit 交易，返回关键信息
// - 根据传入的 chainID 选择对应 RPC；若未匹配到则回退 BSC 配置
// - 优先从事件日志解码金额与用户；若无日志则从 input 解码
func ParseTopupTx(chainID uint64, txHash string) (*TopupTxInfo, error) {
	if txHash == "" {
		return nil, errors.New("empty tx hash")
	}

	// 根据 chainID 选择 RPC
	rpcURL, err := pickRPCByChainID(chainID)
	if err != nil {
		return nil, err
	}

	// 连接客户端（复用全局连接池）
	client, err := tools.GetGlobalClient().GetClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("get client error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h := common.HexToHash(txHash)
	tx, isPending, err := client.TransactionByHash(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("get tx error: %w", err)
	}

	info := &TopupTxInfo{TxHash: txHash, Status: "pending"}

	// from 需要通过签名者恢复
	nid, err := client.NetworkID(ctx)
	if err == nil {
		signer := types.LatestSignerForChainID(nid)
		// 手动从 RLP 解码 message（兼容 go-ethereum v1.16）
		if msg, e := types.Sender(signer, tx); e == nil {
			info.From = msg.Hex()
		}
	}
	if tx.To() != nil {
		info.To = tx.To().Hex()
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(topupabi.TopupLogicABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi error: %w", err)
	}

	// 如果已打包了 input，尝试解析 deposit(amount)
	if data := tx.Data(); len(data) >= 4 {
		if method, ok := parsedABI.Methods["deposit"]; ok {
			// 只在还未从日志得到 amount 时作为兜底解析
			if args, e := method.Inputs.Unpack(data[4:]); e == nil && len(args) == 1 {
				if v, ok := args[0].(*big.Int); ok {
					info.Amount = new(big.Int).Set(v)
				} else if v2, ok2 := args[0].(big.Int); ok2 {
					info.Amount = new(big.Int).Set(&v2)
				}
			}
		}
	}

	// 拿收据（可能 pending）
	if isPending {
		return info, nil
	}

	receipt, err := client.TransactionReceipt(ctx, h)
	if err != nil {
		// 可能未上链
		return info, nil
	}

	if receipt.Status == 1 {
		info.Status = "success"
	} else {
		info.Status = "failed"
	}
	if receipt.BlockNumber != nil {
		info.BlockNumber = receipt.BlockNumber.Uint64()
		if blk, e := client.BlockByNumber(ctx, receipt.BlockNumber); e == nil {
			info.BlockTime = time.Unix(int64(blk.Time()), 0)
		}
	}

	// 优先从 Deposited 事件读取
	if evt, ok := parsedABI.Events["Deposited"]; ok {
		topic0 := evt.ID
		for _, lg := range receipt.Logs {
			if len(lg.Topics) == 0 || lg.Topics[0] != topic0 {
				continue
			}
			// 索引: user (topic[1])，非索引: amount (data)
			if len(lg.Topics) > 1 {
				info.UserFromLog = common.HexToAddress(lg.Topics[1].Hex()).Hex()
			}
			// data 只包含非 indexed 参数：amount(uint256)
			vals, err := evt.Inputs.NonIndexed().Unpack(lg.Data)
			if err == nil && len(vals) == 1 {
				switch v := vals[0].(type) {
				case *big.Int:
					info.Amount = new(big.Int).Set(v)
				case big.Int:
					info.Amount = new(big.Int).Set(&v)
				}
			}
			info.Contract = lg.Address.Hex()
			break
		}
	}

	return info, nil
}

func ParseOpenLockTx(chainID uint64, txHash string) (*TopupTxInfo, error) {
	if txHash == "" {
		return nil, errors.New("empty tx hash")
	}

	// 根据 chainID 选择 RPC
	rpcURL, err := pickRPCByChainID(chainID)
	if err != nil {
		return nil, err
	}

	// 连接客户端（复用全局连接池）
	client, err := tools.GetGlobalClient().GetClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("get client error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h := common.HexToHash(txHash)
	tx, isPending, err := client.TransactionByHash(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("get tx error: %w", err)
	}

	info := &TopupTxInfo{TxHash: txHash, Status: "pending"}

	// from 需要通过签名者恢复
	nid, err := client.NetworkID(ctx)
	if err == nil {
		signer := types.LatestSignerForChainID(nid)
		// 手动从 RLP 解码 message（兼容 go-ethereum v1.16）
		if msg, e := types.Sender(signer, tx); e == nil {
			info.From = msg.Hex()
		}
	}
	if tx.To() != nil {
		info.To = tx.To().Hex()
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(topupabi.TopupLogicABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi error: %w", err)
	}

	// 如果已打包了 input，尝试解析 deposit(amount)
	if data := tx.Data(); len(data) >= 4 {
		if method, ok := parsedABI.Methods["openLock"]; ok {
			// 只在还未从日志得到 amount 时作为兜底解析
			if args, e := method.Inputs.Unpack(data[4:]); e == nil && len(args) == 5 {
				if v, ok := args[1].(*big.Int); ok {
					info.Amount = new(big.Int).Set(v)
				} else if v2, ok2 := args[1].(big.Int); ok2 {
					info.Amount = new(big.Int).Set(&v2)
				}
			}
		}
	}

	// 拿收据（可能 pending）
	if isPending {
		return info, nil
	}

	receipt, err := client.TransactionReceipt(ctx, h)
	if err != nil {
		// 可能未上链
		return info, nil
	}

	if receipt.Status == 1 {
		info.Status = "success"
	} else {
		info.Status = "failed"
	}
	if receipt.BlockNumber != nil {
		info.BlockNumber = receipt.BlockNumber.Uint64()
		if blk, e := client.BlockByNumber(ctx, receipt.BlockNumber); e == nil {
			info.BlockTime = time.Unix(int64(blk.Time()), 0)
		}
	}

	// 优先从 Deposited 事件读取
	if evt, ok := parsedABI.Events["LockOpened"]; ok {
		topic0 := evt.ID
		for _, lg := range receipt.Logs {
			if len(lg.Topics) == 0 || lg.Topics[0] != topic0 {
				continue
			}
			// 索引: user (topic[1])，非索引: amount (data)
			if len(lg.Topics) > 1 {
				info.UserFromLog = common.HexToAddress(lg.Topics[1].Hex()).Hex()
			}
			// data 只包含非 indexed 参数：amount(uint256)
			vals, err := evt.Inputs.NonIndexed().Unpack(lg.Data)
			if err == nil && len(vals) == 2 {
				switch v := vals[0].(type) {
				case *big.Int:
					info.Amount = new(big.Int).Set(v)
				case big.Int:
					info.Amount = new(big.Int).Set(&v)
				}
			}
			info.Contract = lg.Address.Hex()
			break
		}
	}

	return info, nil
}

func ParseClaimLockedTx(chainID uint64, txHash string) (*TopupTxInfo, error) {
	if txHash == "" {
		return nil, errors.New("empty tx hash")
	}

	// 根据 chainID 选择 RPC
	rpcURL, err := pickRPCByChainID(chainID)
	if err != nil {
		return nil, err
	}

	// 连接客户端（复用全局连接池）
	client, err := tools.GetGlobalClient().GetClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("get client error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h := common.HexToHash(txHash)
	tx, isPending, err := client.TransactionByHash(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("get tx error: %w", err)
	}

	info := &TopupTxInfo{TxHash: txHash, Status: "pending"}

	// from 需要通过签名者恢复
	nid, err := client.NetworkID(ctx)
	if err == nil {
		signer := types.LatestSignerForChainID(nid)
		// 手动从 RLP 解码 message（兼容 go-ethereum v1.16）
		if msg, e := types.Sender(signer, tx); e == nil {
			info.From = msg.Hex()
		}
	}
	if tx.To() != nil {
		info.To = tx.To().Hex()
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(topupabi.TopupLogicABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi error: %w", err)
	}

	// 如果已打包了 input，尝试解析 deposit(amount)
	if data := tx.Data(); len(data) >= 4 {
		if method, ok := parsedABI.Methods["claimLocked"]; ok {
			// 只在还未从日志得到 amount 时作为兜底解析
			if args, e := method.Inputs.Unpack(data[4:]); e == nil && len(args) == 2 {
				if v, ok := args[0].([32]byte); ok {
					info.LockID = v
				} else if v2, ok2 := args[1].(common.Address); ok2 {
					info.To = v2.Hex() //真实的转账地址
				}
			}
		}
	}

	// 拿收据（可能 pending）
	if isPending {
		return info, nil
	}

	receipt, err := client.TransactionReceipt(ctx, h)
	if err != nil {
		// 可能未上链
		return info, nil
	}

	if receipt.Status == 1 {
		info.Status = "success"
	} else {
		info.Status = "failed"
	}
	if receipt.BlockNumber != nil {
		info.BlockNumber = receipt.BlockNumber.Uint64()
		if blk, e := client.BlockByNumber(ctx, receipt.BlockNumber); e == nil {
			info.BlockTime = time.Unix(int64(blk.Time()), 0)
		}
	}

	// 优先从 LockClaimed 事件读取
	if evt, ok := parsedABI.Events["LockClaimed"]; ok {
		topic0 := evt.ID

		for _, lg := range receipt.Logs {
			if len(lg.Topics) == 0 || lg.Topics[0] != topic0 {
				continue
			}

			// ------- indexed 部分 -------
			// user: address（topic[1] 的后 20 字节）
			user := common.BytesToAddress(lg.Topics[1].Bytes())

			// lockId: bytes32（topic[2]，保留为 [32]byte 或 common.Hash）
			var lockId [32]byte
			copy(lockId[:], lg.Topics[2].Bytes())

			// ------- non-indexed 部分（data）-------
			// 只包含 to(address), amount(uint256)，按声明顺序
			nonIdx, err := evt.Inputs.NonIndexed().Unpack(lg.Data)
			if err != nil || len(nonIdx) != 2 {
				// 处理错误或日志格式不符
				continue
			}

			// to
			var to common.Address
			switch v := nonIdx[0].(type) {
			case common.Address:
				to = v
			case [20]byte:
				to = common.BytesToAddress(v[:])
			case []byte:
				to = common.BytesToAddress(v)
			default:
				// 类型不符
				continue
			}

			// amount
			var amount *big.Int
			switch v := nonIdx[1].(type) {
			case *big.Int:
				amount = new(big.Int).Set(v)
			case big.Int:
				amount = new(big.Int).Set(&v)
			default:
				continue
			}

			// 赋值到你的结构
			info.UserFromLog = user.Hex()
			info.LockID = lockId // 如果需要 hex: common.BytesToHash(lockId[:]).Hex()
			info.To = to.Hex()
			info.Amount = amount
			info.Contract = lg.Address.Hex()

			break // 找到一条就退出；如需多条，去掉 break
		}
	}

	return info, nil
}

func ParseCancelLockedTx(chainID uint64, txHash string) (*TopupTxInfo, error) {
	if txHash == "" {
		return nil, errors.New("empty tx hash")
	}

	// 根据 chainID 选择 RPC
	rpcURL, err := pickRPCByChainID(chainID)
	if err != nil {
		return nil, err
	}

	// 连接客户端（复用全局连接池）
	client, err := tools.GetGlobalClient().GetClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("get client error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h := common.HexToHash(txHash)
	tx, isPending, err := client.TransactionByHash(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("get tx error: %w", err)
	}

	info := &TopupTxInfo{TxHash: txHash, Status: "pending"}

	// from 需要通过签名者恢复
	nid, err := client.NetworkID(ctx)
	if err == nil {
		signer := types.LatestSignerForChainID(nid)
		// 手动从 RLP 解码 message（兼容 go-ethereum v1.16）
		if msg, e := types.Sender(signer, tx); e == nil {
			info.From = msg.Hex()
		}
	}
	if tx.To() != nil {
		info.To = tx.To().Hex()
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(topupabi.TopupLogicABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi error: %w", err)
	}

	// 如果已打包了 input，尝试解析 deposit(amount)
	if data := tx.Data(); len(data) >= 4 {
		if method, ok := parsedABI.Methods["cancelLock"]; ok {
			// 只在还未从日志得到 amount 时作为兜底解析
			if args, e := method.Inputs.Unpack(data[4:]); e == nil && len(args) == 1 {
				if v, ok := args[0].([32]byte); ok {
					info.LockID = v
				}
			}
		}
	}

	// 拿收据（可能 pending）
	if isPending {
		return info, nil
	}

	receipt, err := client.TransactionReceipt(ctx, h)
	if err != nil {
		// 可能未上链
		return info, nil
	}

	if receipt.Status == 1 {
		info.Status = "success"
	} else {
		info.Status = "failed"
	}
	if receipt.BlockNumber != nil {
		info.BlockNumber = receipt.BlockNumber.Uint64()
		if blk, e := client.BlockByNumber(ctx, receipt.BlockNumber); e == nil {
			info.BlockTime = time.Unix(int64(blk.Time()), 0)
		}
	}

	// 优先从 LockCanceled 事件读取
	// 期望事件签名：LockCanceled(address indexed user, bytes32 indexed lockId, uint256 amount)
	if evt, ok := parsedABI.Events["LockCanceled"]; ok {
		topic0 := evt.ID // 事件签名哈希

		for _, lg := range receipt.Logs {
			// 可选：如果只处理特定合约地址，先过滤一下
			// if lg.Address != targetContractAddress { continue }

			// 必须有 3 个 topic：topic0(事件ID) + user + lockId
			if len(lg.Topics) < 3 || lg.Topics[0] != topic0 {
				continue
			}

			// -------- indexed 部分 --------
			// user: topic[1] 的后 20 字节
			// BytesToAddress 会取最后 20 字节，但这里手动切片更直观
			t1 := lg.Topics[1].Bytes()
			user := common.BytesToAddress(t1[len(t1)-20:])

			// lockId: topic[2] 是 bytes32，直接用 common.Hash 或转 [32]byte
			lockHash := lg.Topics[2] // common.Hash
			var lockId [32]byte
			copy(lockId[:], lockHash.Bytes()) // 若你的结构用 [32]byte，就这样拷

			// -------- non-indexed 部分（data）--------
			// 只有 amount(uint256)，按声明顺序
			nonIdx, err := evt.Inputs.NonIndexed().Unpack(lg.Data)
			if err != nil || len(nonIdx) != 1 {
				// 数据格式不符，跳过
				continue
			}

			// amount
			var amount *big.Int
			switch v := nonIdx[0].(type) {
			case *big.Int:
				amount = new(big.Int).Set(v)
			case big.Int:
				amount = new(big.Int).Set(&v)
			default:
				continue
			}

			// -------- 赋值到你的结构 --------
			info.UserFromLog = user.Hex()
			info.LockID = lockId // 或者：common.BytesToHash(lockId[:]).Hex()
			info.Amount = amount
			info.Contract = lg.Address.Hex()

			// 如果你只需要第一条匹配日志，保留 break；需要全部则去掉
			break
		}
	}

	return info, nil
}

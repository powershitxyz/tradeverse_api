package chain

import (
	"context"
	"errors"
	"math/big"
	"strings"

	topupabi "chaos/api/chain/abi"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EIP712: LockAuth(address user,bytes32 lockId,uint256 amount,uint64 expiry,uint256 nonce)
var lockAuthTypeHash = crypto.Keccak256Hash([]byte("LockAuth(address user,bytes32 lockId,uint256 amount,uint64 expiry,uint256 nonce)"))

// leftPadBytes32 pads b to 32 bytes on the left (big-endian)
func leftPadBytes32(b []byte) []byte {
	if len(b) >= 32 {
		out := make([]byte, 32)
		copy(out, b[len(b)-32:])
		return out
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

// toUint256 encodes an integer into 32-byte big-endian
func toUint256(x *big.Int) []byte {
	if x == nil {
		x = big.NewInt(0)
	}
	return leftPadBytes32(x.Bytes())
}

// buildLockAuthStructHash builds keccak256(abi.encodePacked(typehash, fields...))
func buildLockAuthStructHash(user common.Address, lockId common.Hash, amount *big.Int, expiry uint64, nonce *big.Int) common.Hash {
	// address (20 -> 32)
	userPadded := leftPadBytes32(user.Bytes())
	// userPadded := leftPadBytes32(big.NewInt(int64(mainID)).Bytes())
	// bytes32 lockId is already 32
	lockIdBytes := lockId.Bytes()
	// uint256 amount
	amountBytes := toUint256(new(big.Int).Set(amount))
	// uint64 -> uint256
	expiryBI := new(big.Int).SetUint64(expiry)
	expiryBytes := toUint256(expiryBI)
	// uint256 nonce
	nonceBytes := toUint256(new(big.Int).Set(nonce))

	// pack
	enc := make([]byte, 0, 32*6)
	enc = append(enc, lockAuthTypeHash.Bytes()...)
	enc = append(enc, userPadded...)
	enc = append(enc, lockIdBytes...)
	enc = append(enc, amountBytes...)
	enc = append(enc, expiryBytes...)
	enc = append(enc, nonceBytes...)

	return crypto.Keccak256Hash(enc)
}

// fetchDomainSeparator calls domainSeparator() from the verifying contract
func fetchDomainSeparator(ctx context.Context, rpcURL string, contract common.Address) (common.Hash, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return common.Hash{}, err
	}
	defer client.Close()

	topupAbi, err := abi.JSON(strings.NewReader(topupabi.TopupLogicABI))
	if err != nil {
		return common.Hash{}, err
	}

	// encode call data for domainSeparator()
	data, err := topupAbi.Pack("domainSeparator")
	if err != nil {
		return common.Hash{}, err
	}

	// low-level eth_call
	var result []byte
	callMsg := ethereum.CallMsg{To: &contract, Data: data}
	result, err = client.CallContract(ctx, callMsg, nil)
	if err != nil {
		return common.Hash{}, err
	}

	// unpack
	out, err := topupAbi.Unpack("domainSeparator", result)
	if err != nil || len(out) != 1 {
		return common.Hash{}, errors.New("failed to unpack domainSeparator")
	}
	// return as bytes32
	switch v := out[0].(type) {
	case [32]byte:
		return common.BytesToHash(v[:]), nil
	case []byte:
		return common.BytesToHash(v), nil
	default:
		return common.Hash{}, errors.New("unexpected domainSeparator type")
	}
}

// BuildAndSignLockAuth computes the EIP-712 digest and signs it with the provided private key
// chainID is used only to select RPC (via pickRPCByChainID). The contract's domainSeparator is fetched on-chain to avoid domain mismatch.
func BuildAndSignLockAuth(ctx context.Context, chainID uint64, contractAddr string, privKeyHex string,
	// mainID uint64,
	userAddr string,
	lockIdHex string, amount *big.Int, expiry uint64, nonce *big.Int) (digest common.Hash, sig []byte, err error) {

	if contractAddr == "" || userAddr == "" || lockIdHex == "" || privKeyHex == "" {
		return common.Hash{}, nil, errors.New("missing required params")
	}

	rpcURL, err := pickRPCByChainID(chainID)
	if err != nil {
		return common.Hash{}, nil, err
	}

	contract := common.HexToAddress(contractAddr)
	user := common.HexToAddress(userAddr)

	// parse lockId (bytes32)
	lockIdBytes, err := hexutil.Decode(lockIdHex)
	if err != nil {
		return common.Hash{}, nil, err
	}
	if len(lockIdBytes) != 32 {
		return common.Hash{}, nil, errors.New("lockId must be 32 bytes")
	}
	var lockId common.Hash
	copy(lockId[:], lockIdBytes)

	if amount == nil {
		amount = big.NewInt(0)
	}
	if nonce == nil {
		nonce = big.NewInt(0)
	}

	// struct hash
	structHash := buildLockAuthStructHash(user, lockId, amount, expiry, nonce)

	// domain separator from chain
	domainSep, err := fetchDomainSeparator(ctx, rpcURL, contract)
	if err != nil {
		return common.Hash{}, nil, err
	}

	// EIP-191 prefix 0x1901
	prefix := []byte{0x19, 0x01}
	digestBytes := crypto.Keccak256(append(append(prefix, domainSep.Bytes()...), structHash.Bytes()...))
	copy(digest[:], digestBytes)

	// sign
	privKey, err := crypto.HexToECDSA(strings.TrimPrefix(privKeyHex, "0x"))
	if err != nil {
		return common.Hash{}, nil, err
	}

	signature, err := crypto.Sign(digest[:], privKey)
	if err != nil {
		return common.Hash{}, nil, err
	}

	// ensure v is 27/28
	if len(signature) == 65 {
		if signature[64] < 27 {
			signature[64] += 27
		}
	}

	return digest, signature, nil
}

// RecoverLockAuthSigner recovers signer address from digest and signature
func RecoverLockAuthSigner(digest common.Hash, sig []byte) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, errors.New("invalid signature length")
	}
	// normalize v to 0/1 for recovery
	sigCopy := make([]byte, len(sig))
	copy(sigCopy, sig)
	if sigCopy[64] >= 27 {
		sigCopy[64] -= 27
	}
	pub, err := crypto.SigToPub(digest.Bytes(), sigCopy)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

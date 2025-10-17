package chain

import (
	"fmt"
	"testing"
)

func TestParseTopupTx(t *testing.T) {
	txHash := "0x1126ee117dc6709b90ebf14599d0b264c55ce0220d7267537fdbc5dbdbe7123e"
	info, err := ParseTopupTx(97, txHash)
	if err != nil {
		t.Fatalf("failed to parse topup tx: %v", err)
	}
	fmt.Println(info)
}

func TestParseOpenLockTx(t *testing.T) {
	txHash := "0xb168f39ce3d66697ea15c721d2d91356a5f6106dac4e4b913700d34ceef2ce90"
	info, err := ParseOpenLockTx(97, txHash)
	if err != nil {
		t.Fatalf("failed to parse open lock tx: %v", err)
	}
	amountStr := info.Amount.String()
	fmt.Println(amountStr)
}

func TestParseClaimLockedTx(t *testing.T) {
	txHash := "0xadf6f497dadc92fe33d51643269107705d4e0649b71819617156803ca1a9fa27"
	info, err := ParseClaimLockedTx(97, txHash)
	if err != nil {
		t.Fatalf("failed to parse claim locked tx: %v", err)
	}
	fmt.Println(info)
}

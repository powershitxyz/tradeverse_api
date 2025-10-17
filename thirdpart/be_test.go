package thirdpart

import (
	"chaos/api/log"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_BE_Metadata(t *testing.T) {
	b, err := GetBirdClient()
	t.Log(b, err)
	var cas = []string{"So11111111111111111111111111111111111111112", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"}
	result, _ := b.Metadata(cas, "solana")
	spew.Dump(result)
	//t.Log(result)
}
func Test_BE_TokenOverview(t *testing.T) {
	b, err := GetBirdClient()
	if err != nil {
		log.Errorf("failed to get bird client: %v", err)
	}
	result, _ := b.TokenOverview("6NcdiK8B5KK2DzKvzvCfqi8EHaEqu48fyEzC8Mm9pump", "solana")
	spew.Dump(result)
}

func Test_BE_TokenTxsByTokenAddress(t *testing.T) {
	b, err := GetBirdClient()
	if err != nil {
		log.Errorf("failed to get bird client: %v", err)
	}
	result, _ := b.TokenTxsByTokenAddress("5nKvj6LofaRsUf47Tma658EqrJfZtYAVSHye3X67GKX1", "solana", "", "", "", "", 0, 0, 1, 50)
	spew.Dump(result)

}

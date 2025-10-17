package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"
)

func TestHash(t *testing.T) {
	password := "123456"
	hashPassByte := sha256.Sum256([]byte(password))
	hashPass := hex.EncodeToString(hashPassByte[:])

	fmt.Println(hashPass)
}

func TestFloat(t *testing.T) {
	v := 86
	vs := float64(v) / 100
	s := math.Round(vs*10) / 10
	fmt.Println(s)
}

func TestGoPlus(t *testing.T) {
	url := "https://api.gopluslabs.io/api/v1/solana/token_security?contract_addresses=Buoj8HCZMnLRwzDmjzaswhkVhLZD58PG4pZ7rnYp6pCr"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("accept", "*/*")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	fmt.Println(result)

}

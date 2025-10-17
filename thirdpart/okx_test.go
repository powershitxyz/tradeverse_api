package thirdpart

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"chaos/api/log"
)

func TestOkxTrade(t *testing.T) {
	okx, err := GetOKXClient()
	if err != nil {
		t.Log("create okx client failed ", err)
		t.Fail()
	}

	tinfo, err := okx.Quote("27d3hjmWQSTi7fJEvshn3gJbsR5KdEonTYjz1NtqpNTp", "So11111111111111111111111111111111111111112", "1000000", "0.5")
	t.Log(tinfo, err)

	v2, err := okx.Swap("27d3hjmWQSTi7fJEvshn3gJbsR5KdEonTYjz1NtqpNTp", "So11111111111111111111111111111111111111112", "1000000", "0.5", "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C")
	t.Log(v2, err)

	v3, err := okx.WalletValueSpec("0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8", "1,8453", "0", true)
	t.Log(v3, err)

	v4, err := okx.WalletAssets("0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8", "1,8453", "0")
	t.Log(v4, err)
}

func TestOkxPrice(t *testing.T) {
	okx, err := GetOKXClient()
	if err != nil {
		t.Log("create okx client failed ", err)
		t.Fail()
	}

	var tprs []TokenPriceRequest = make([]TokenPriceRequest, 0)
	tprs = append(tprs, TokenPriceRequest{
		ChainIndex:   "501",
		TokenAddress: "3iD8fEkkQT8JC9dbpyorVsxjsCzdim8NXr6XgERJpump",
	})
	tprs = append(tprs, TokenPriceRequest{
		ChainIndex:   "501",
		TokenAddress: "6p6xgHyF7AeE6TZkSmFsko444wqoP15icUSqi2jfGiPN",
	})
	result, err := okx.CoinPrice(tprs)
	t.Log(result, err)

	r, _ := okx.WalletTransactions("BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C", "501", "", 0, 0, 0)
	t.Log(r)

	// 	var jsontrans = `
	// 	{
	// 	"code": "0",
	// 	"msg": "success",
	// 	"data": [{
	// 		"cursor": "1740980200",
	// 		"transactionList": [{
	// 			"chainIndex": "501",
	// 			"txHash": "3jEZGpgARvwEuDKcfGBJXz5AUiQMofVxs9CpaPCwRLBZA5JKSiKtci1hoAmjRvKtS7Qd22sZjdD5bfPeq81A83s7",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741651586000",
	// 			"from": [{
	// 				"address": "5Hr7wZg7oBpVhH5nngRqzr5W7ZFUfCsfEhbziZJak7fr",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "4EUrALmajqN85qfhprRduKTYabXEF7NaQVGjopcW2AsZe2x7vmgE5tfLfHoyKMSwqNCJkpqVy7RjN8ma9MgadGSu",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741608262000",
	// 			"from": [{
	// 				"address": "5Hr7wZg7oBpVhH5nngRqzr5W7ZFUfCsfEhbziZJak7fr",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "4Jm9WHpJzALfjcr84F7XWcURjfzWTgL1wL9dfD2f51X5L5QLgoZSGeHFSEAWEpgfLf8hMzSai1nx7WZQ3gxFkZ4Z",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567497000",
	// 			"from": [{
	// 				"address": "5ifyfzJLkpThxrjvCmzTPRfpvUtBBkXLNb4URD7vq7Nm",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "5qd7Txv6UMyM9LgXvo6hUjLwouycBiYRyBTzUQhadG3ncykH5wedLBhdqFkiGmeAR1my6VHyEvCagVGFQSufB2tk",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567497000",
	// 			"from": [{
	// 				"address": "fLiPgg2yTvmgfhiPkKriAHkDmmXGP6CdeFX9UF5o7Zc",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3qF9PeubraiAYQqtzHv5aGY6RssoTKjrB7kFtByKK11fmBLCNXcnQrUmBdkhzitEzQBcFUA76gCvabD4ZAZfKV8L",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567492000",
	// 			"from": [{
	// 				"address": "Habp5bncMSsBC3vkChyebepym5dcTNRYeg2LVG464E96",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3sfeJw5ktwACUASRVmvEwmV5mSSGaPcdTQBHcLG8tQMJ5RH4MsPEQ3iZhXBkYrpYoG8RnDiARNP4Qdh7JXdspBMg",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567491000",
	// 			"from": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
	// 			"amount": "0.05",
	// 			"symbol": "USDT",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3sfeJw5ktwACUASRVmvEwmV5mSSGaPcdTQBHcLG8tQMJ5RH4MsPEQ3iZhXBkYrpYoG8RnDiARNP4Qdh7JXdspBMg",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567491000",
	// 			"from": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
	// 			"amount": "99.95",
	// 			"symbol": "USDT",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3sfeJw5ktwACUASRVmvEwmV5mSSGaPcdTQBHcLG8tQMJ5RH4MsPEQ3iZhXBkYrpYoG8RnDiARNP4Qdh7JXdspBMg",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741567491000",
	// 			"from": [{
	// 				"address": "BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
	// 			"amount": "200.694901",
	// 			"symbol": "JUP",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "4qTwVkskBtFRk1cCvhBR1HBosjJofZgNfChQrFJKsRVjMYsqVjptrEyqMdAySg649gDo5YCkwRm15EQqsB7GJqwt",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741305853000",
	// 			"from": [{
	// 				"address": "5Hr7wZg7oBpVhH5nngRqzr5W7ZFUfCsfEhbziZJak7fr",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "XwxQ6vcY2ZTUj1hikbYVkTGd3gHjdmucnUU7hoPysCFVwzmrEnYnXxxWdvAWxNbKXJqw9QShcLS776wBBJoPsjD",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741262573000",
	// 			"from": [{
	// 				"address": "5Hr7wZg7oBpVhH5nngRqzr5W7ZFUfCsfEhbziZJak7fr",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "2nia4TR4q1T25L8EWdbnmrbqpeavjy4iRaTuL4nx3iJ7HQx3umKEj3CVERCajWXBUhDaves6jJMazELd86eVLLbM",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224561000",
	// 			"from": [{
	// 				"address": "5ifyfzJLkpThxrjvCmzTPRfpvUtBBkXLNb4URD7vq7Nm",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "4cvhfzUNmrxDY9S6SFRJ7PoUUAPm6GJPKjen7q3uFfe9xrS3fGjk4iD18n2p9WcGhthnyxk5spJpnZZ7wegyg5w9",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224561000",
	// 			"from": [{
	// 				"address": "fLiPgg2yTvmgfhiPkKriAHkDmmXGP6CdeFX9UF5o7Zc",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3zL6QSm8EbuSjaPpuYdxQHyG9JitJqdu4uZmZuWqM4WEHpBXHEUKZgjcKX8xjV4v9Ne9D2vXYZw4MuTG6RvUHnT8",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224561000",
	// 			"from": [{
	// 				"address": "Habp5bncMSsBC3vkChyebepym5dcTNRYeg2LVG464E96",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3NJPJqsEcUm3LDZXw1PDewvHU5B7VQQ1F87gjZiG4CmpKwquNP9U3PUe5dEzjrkzVQuKn1FuNBJBmvZvErmkup2i",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224557000",
	// 			"from": [{
	// 				"address": "Habp5bncMSsBC3vkChyebepym5dcTNRYeg2LVG464E96",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3zSvd6UWpXw29ex4b5YrF5CmfG32HKumY29Tgc4aG4CoLW5trjCfggDZzrx6uf3psP82yBh2BBQkpkUUbcdSoYRu",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224557000",
	// 			"from": [{
	// 				"address": "fLiPgg2yTvmgfhiPkKriAHkDmmXGP6CdeFX9UF5o7Zc",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "2JMjR2Lb2VJX9HSYubeqbGq1wmNYkJypvWhqoA5uB6f58hYJbUSLVnF9tn2Veg6Zg2bZ4AAMUD2scEWas6SJX9Kp",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224552000",
	// 			"from": [{
	// 				"address": "EyqYPFiSL3LXBjhiaLZ44kfKnstGsv2ZaggXzVMxDx9X",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "",
	// 			"amount": "0.000000001",
	// 			"symbol": "SOL",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "0"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "dcmoPwJHXu8U2Xd49oDAbcMSBerLwyo6rwBT6F4WA72gkdL7z7GRj26W4Jqb2WHstyiV4gkg4EYz62WREvUdMsq",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224551000",
	// 			"from": [{
	// 				"address": "9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
	// 			"amount": "481.939471",
	// 			"symbol": "JUP",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "dcmoPwJHXu8U2Xd49oDAbcMSBerLwyo6rwBT6F4WA72gkdL7z7GRj26W4Jqb2WHstyiV4gkg4EYz62WREvUdMsq",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224551000",
	// 			"from": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "DSN3j1ykL3obAVNv7ZX49VsFCPe4LqzxHnmtLiPwY6xg",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
	// 			"amount": "0.15",
	// 			"symbol": "USDT",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "dcmoPwJHXu8U2Xd49oDAbcMSBerLwyo6rwBT6F4WA72gkdL7z7GRj26W4Jqb2WHstyiV4gkg4EYz62WREvUdMsq",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1741224551000",
	// 			"from": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
	// 			"amount": "299.85",
	// 			"symbol": "USDT",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}, {
	// 			"chainIndex": "501",
	// 			"txHash": "3W77UryTeiQHFjUBEuM2Sb4SX5wLkTPRLvEs3DugcRYRNmt6SmrizgRdM7EPriXxtg3xGDFLXdbdbfnxTRKssAt1",
	// 			"methodId": "",
	// 			"nonce": "",
	// 			"txTime": "1740980200000",
	// 			"from": [{
	// 				"address": "uSbhsXoy67MLZ7dvCsvhyv41iv4F6SvhU3npvfDZufx",
	// 				"amount": ""
	// 			}],
	// 			"to": [{
	// 				"address": "BYk8eddhd9kNvq7dbeSp7WRnwd5sgNKZnaxsaAGDuw9C",
	// 				"amount": ""
	// 			}],
	// 			"tokenAddress": "Ds9FdU48nD34dgtfeSLtqvc5L9LoPig7cwb7kcxbmoon",
	// 			"amount": "1000",
	// 			"symbol": "PDAO",
	// 			"txFee": "",
	// 			"txStatus": "success",
	// 			"hitBlacklist": false,
	// 			"tag": "",
	// 			"itype": "2"
	// 		}]
	// 	}]
	// }
	// 	`
	// 	var response struct {
	// 		Code string                    `json:"code"`
	// 		Msg  string                    `json:"msg"`
	// 		Data []TransactionListResponse `json:"data"`
	// 	}
	// 	json.Unmarshal([]byte(jsontrans), &response)

	// fmt.Println(response)
}

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	// 运行测试
	os.Exit(m.Run())
}

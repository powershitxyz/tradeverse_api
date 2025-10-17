package home

import (
	"chaos/api/model"
	"chaos/api/system"
	"fmt"
	"strings"
	"testing"
)

func TestJoin(t *testing.T) {
	s := []string{
		"Grass7B4RdKfBCjTKgSqnXkqjwiGvQyFbuSCUJr3XXjs",
		"cbbtcf3aa214zXHbiAZQwf4122FBYbraNdFqgw4iMij",
		"LAYER4xPpTCb3QL8S9u41EAhAX7mhBn8Q6xMTwY2Yzc",
		"J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn",
		"SonicxvLud67EceaEzCLRnMTBqzYUUYNr93DBkBdDES",
		"TNSRxcUxoT9xBG3de7PiJyTDYu7kskLqcpddxnEJAS6",
	}

	sql := fmt.Sprintf("insert into token_list_temp(token0) values%s", "('"+strings.Join(s, `'),('`)+"')")
	fmt.Println(sql)
}

//insert into token_list_temp(token0) values('Grass7B4RdKfBCjTKgSqnXkqjwiGvQyFbuSCUJr3XXjs'),cbbtcf3aa214zXHbiAZQwf4122FBYbraNdFqgw4iMij'),LAYER4xPpTCb3QL8S9u41EAhAX7mhBn8Q6xMTwY2Yzc'),J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn'),SonicxvLud67EceaEzCLRnMTBqzYUUYNr93DBkBdDES'),TNSRxcUxoT9xBG3de7PiJyTDYu7kskLqcpddxnEJAS6'

// insert into token_list_temp(token0) values('Grass7B4RdKfBCjTKgSqnXkqjwiGvQyFbuSCUJr3XXjs'),('cbbtcf3aa214zXHbiAZQwf4122FBYbraNdFqgw4iMij'),('LAYER4xPpTCb3QL8S9u41EAhAX7mhBn8Q6xMTwY2Yzc'),('J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn'),('SonicxvLud67EceaEzCLRnMTBqzYUUYNr93DBkBdDES'),('TNSRxcUxoT9xBG3de7PiJyTDYu7kskLqcpddxnEJAS6'

func TestLeaderboardSql(t *testing.T) {
	var db = system.GetDb()
	var seasonGames []model.SeasonGame
	db.Model(&model.SeasonGame{}).Where("season_id = ?", 10001).Find(&seasonGames)
	if len(seasonGames) == 0 {
		t.Fatal("season game not found")
	}

	var scoreCaseSql = "(CASE p.game_id\n"
	var amountCaseSql = "(CASE p.game_id\n"
	for _, seasonGame := range seasonGames {
		scoreCaseSql += fmt.Sprintf("WHEN %d THEN %d\n", seasonGame.GameID, seasonGame.WeightScore)
		amountCaseSql += fmt.Sprintf("WHEN %d THEN %d\n", seasonGame.GameID, seasonGame.WeightAmount)
	}
	scoreCaseSql += "ELSE 0\n"
	amountCaseSql += "ELSE 0\n"
	scoreCaseSql += "END) * p.sum_score \n"
	amountCaseSql += "END) * p.sum_amount\n"

	sql := fmt.Sprintf(seasonLeaderboardSql, scoreCaseSql, amountCaseSql)
	fmt.Println(sql)
}

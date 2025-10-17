package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/wsm"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsmclient = wsm.RetrieveWsManager()

func Chat(c *gin.Context) {
	ws, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// check parameters
	chain := c.Query("chain")
	ca := c.Query("ca")
	if chain != "solana" || len(ca) == 0 {
		return
	}

	wsmclient.AddClient(chain, ca, ws)
	defer wsmclient.RemoveClient(ws)

	go func() {
		<-c.Done()
		log.Info("ws lost connection")
		wsmclient.RemoveClient(ws)
	}()

	timeNowHs := time.Now().UnixNano() / int64(time.Millisecond)

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Error("read error", err)
			break
		}
		if string(message) == "ping" { //heart beat
			message = []byte("pong")
			err = ws.WriteMessage(mt, message)
			if err != nil {
				log.Info(err)
				break
			}
		} else {
			requestModel, err := parseRequestMsg(message)
			log.Info(requestModel, err, timeNowHs)
			// if err != nil {
			// 	rp := makeReply(codes.CODE_ERR_REQFORMAT, err.Error(), timeNowHs, "", requestModel.Timestamp, "")
			// 	ws.WriteJSON(rp)
			// 	return
			// }

			if requestModel.Method == common.METHOD_GPT {
				// RequestGPT(ws, mt, requestModel, timeNowHs)
				log.Info("准备请求chat gpt")
			} else {
				rp := makeReply(codes.CODE_ERR_METHOD_UNSUPPORT, err.Error(), timeNowHs, "", requestModel.Timestamp, "")
				ws.WriteJSON(rp)
			}
		}

	}
}

func parseRequestMsg(body []byte) (c common.Request, e error) {

	defer func() {
		if r := recover(); r != nil {
			e = errors.New("invalid request data format")
		}
	}()

	log.Info("socket : ", string(body))

	err := json.Unmarshal(body, &c)
	if err != nil {
		log.Error(err)
		return common.Request{}, err
	}

	return c, nil
}

func makeReply(code int64, msg string, timeHs int64, chatId string, replyTs int64, content string) *common.Response {
	e := common.Response{
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
		Timestamp: replyTs,
		Data:      "hi",
	}

	return &e
}

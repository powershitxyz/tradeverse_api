package wsm

import (
	"errors"
	"fmt"
	"sync"

	"chaos/api/log"

	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	clients map[*websocket.Conn]string
	manset  map[string]map[*websocket.Conn]struct{} // 使用 map 代替切片，提高查找和删除效率
	mu      sync.RWMutex                            // 读写锁
}

// 全局 WebSocketManager
var wsManager = &WebSocketManager{
	clients: make(map[*websocket.Conn]string),
	manset:  make(map[string]map[*websocket.Conn]struct{}),
}

// 获取 WebSocketManager 单例
func RetrieveWsManager() *WebSocketManager {
	return wsManager
}

func (wm *WebSocketManager) Stat() (allConn, auditAllConn int) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	allConn = len(wm.clients)
	total := 0
	for _, connections := range wm.manset {
		total += len(connections)
	}
	auditAllConn = total
	return allConn, auditAllConn
}

// 添加 WebSocket 连接
func (wm *WebSocketManager) AddClient(chain, ca string, ws *websocket.Conn) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	key := buildKey(chain, ca)

	// 如果 key 不存在，初始化 map
	if _, exists := wm.manset[key]; !exists {
		wm.manset[key] = make(map[*websocket.Conn]struct{})
	}
	wm.manset[key][ws] = struct{}{} // 使用 map 进行去重
	wm.clients[ws] = key
}

// 删除 WebSocket 连接
func (wm *WebSocketManager) RemoveClient(ws *websocket.Conn) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// 获取 key
	key, exists := wm.clients[ws]
	if !exists {
		return
	}

	// 从 manset[key] 删除 WebSocket 连接
	if connections, ok := wm.manset[key]; ok {
		delete(connections, ws)
		if len(connections) == 0 {
			delete(wm.manset, key) // 清理空 key
		}
	}

	// 删除 clients 映射
	delete(wm.clients, ws)
}

// 广播消息到指定链和 CA
func (wm *WebSocketManager) Broadcast(chain, ca string, msg string) {
	wm.mu.RLock() // 只需要读锁
	defer wm.mu.RUnlock()
	key := buildKey(chain, ca)

	// 获取该 key 下所有的 WebSocket 连接
	connections, exists := wm.manset[key]
	if !exists {
		return
	}

	// 发送消息
	for ws := range connections {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Error("Failed to send message:", err)
			ws.Close()
			wm.RemoveClient(ws) // 关闭连接后移除
		}
	}
}

// 发送单独消息给指定 WebSocket 连接
func (wm *WebSocketManager) SendToClient(ws *websocket.Conn, msg string) error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// 检查连接是否存在
	if _, exists := wm.clients[ws]; !exists {
		return errors.New("WebSocket client not found")
	}
	return ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (wm *WebSocketManager) BroadcastToAll(msg []byte) error {
	wm.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(wm.clients))
	for ws := range wm.clients {
		connections = append(connections, ws) // 复制连接
	}
	wm.mu.RUnlock()

	var failedConnections []*websocket.Conn

	for _, ws := range connections {
		err := ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			failedConnections = append(failedConnections, ws) // 记录失败连接
		}
	}

	if len(failedConnections) > 0 {
		wm.mu.Lock()
		for _, ws := range failedConnections {
			delete(wm.clients, ws) // 从 clients 中删除失效的 WebSocket 连接
			ws.Close()
		}
		wm.mu.Unlock()
	}

	if len(failedConnections) > 0 {
		return fmt.Errorf("failed to send message to %d clients", len(failedConnections))
	}

	return nil
}

// 生成 key
func buildKey(chain, ca string) string {
	return fmt.Sprintf("%s-%s", chain, ca)
}

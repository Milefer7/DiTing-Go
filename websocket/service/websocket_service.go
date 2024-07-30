package service

import (
	global2 "DiTing-Go/global"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/websocket/global"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO:连接断开处理
// 定义一个升级器，将普通的http连接升级为websocket连接
var upgrader = &websocket.Upgrader{
	//定义读写缓冲区大小
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	//校验请求
	CheckOrigin: func(r *http.Request) bool {
		//如果不是get请求，返回错误
		if r.Method != "GET" {
			fmt.Println("请求方式错误")
			return false
		}
		//还可以根据其他需求定制校验规则
		return true
	},
}

// Connect 建立WebSocket连接
func Connect(w http.ResponseWriter, r *http.Request) {
	// 从请求的URL中获取token参数
	params := r.URL.Query()
	token := params.Get("token")

	// 解析token并获取用户信息
	tokenInfo, err := utils.ParseToken(token)
	if err != nil {
		global2.Logger.Errorf("无权限访问: %v", err)
		return
	}
	uid := &tokenInfo.Uid

	// 将HTTP连接升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("连接升级时出错:", err)
		return
	}
	// 连接关闭时执行清理操作
	defer conn.Close()

	// 连接成功后注册用户
	// 将uid转换为字符串形式
	stringUid := strconv.FormatInt(*uid, 10) // 转换为10进制表示

	// 初始化用户频道信息
	userChannel := global.Channels{
		Uid:         *uid,
		ChannelList: make([]*websocket.Conn, 0),
		Mu:          new(sync.RWMutex),
	}
	user := global.User{
		Uid:     *uid,
		Channel: conn,
	}

	// 将用户频道信息存储到全局用户频道映射表中
	global.UserChannelMap.Set(stringUid, &userChannel)
	userChannelPtr, _ := global.UserChannelMap.Get(stringUid)

	// 将连接加入到用户的频道列表中
	userChannelPtr.Mu.Lock()
	userChannelPtr.ChannelList = append(userChannelPtr.ChannelList, conn)
	userChannelPtr.Mu.Unlock()

	// 开始定时发送心跳消息以保持连接
	go heatBeat(&user)

	// 监听WebSocket连接上的消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// 连接断开时进行处理
			disConnect(&user)
			break
		}
	}
}

// Send 发送空消息代表有新消息，前端收到消息后再去后端拉取消息
func Send(uid int64, value []byte) error {
	stringUid := strconv.FormatInt(uid, 10)
	channels, _ := global.UserChannelMap.Get(stringUid)
	// 用户不在线，直接返回
	if channels == nil {
		return nil
	}
	for _, conn := range channels.ChannelList {
		// 发送空消息，代表有新消息
		err := conn.WriteMessage(websocket.TextMessage, value)
		if err != nil {
			global2.Logger.Errorf("发送消息失败: %v", err)
			return errors.New("Business Error")
		}
	}
	return nil
}

// 移除连接
func disConnect(user *global.User) {
	// 将用户的 UID 转换为字符串形式
	stringUid := strconv.FormatInt(user.Uid, 10)

	// 获取用户的 WebSocket 连接
	conn := user.Channel

	// 从全局用户通道映射中获取用户的频道信息
	userChannel, _ := global.UserChannelMap.Get(stringUid)

	// 锁定用户频道以进行安全的并发操作
	userChannel.Mu.Lock()

	// 遍历用户的频道列表，查找并移除指定的 WebSocket 连接
	for i, item := range userChannel.ChannelList {
		if item == conn {
			// 移除匹配的连接
			userChannel.ChannelList = append(userChannel.ChannelList[:i], userChannel.ChannelList[i+1:]...)
			break // 找到连接并移除后退出循环
		}
	}

	// 关闭 WebSocket 连接
	err := conn.Close()
	if err != nil {
		// 如果关闭连接失败，记录错误并返回
		return
	}

	// 解锁用户频道
	userChannel.Mu.Unlock()
}

// 解析jwt
func parseJwt(r *http.Request) (*int64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("无权限访问")
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("无权限访问")
	}
	// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
	token, err := utils.ParseToken(parts[1])
	if err != nil {
		return nil, errors.New("无权限访问")
	}
	return &token.Uid, nil
}

// 心跳检测
func heatBeat(user *global.User) {
	conn := user.Channel // 获取用户的 WebSocket 连接

	// 创建一个定时器，每 5 秒触发一次
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // 确保在函数退出时停止定时器

	for {
		select {
		case <-ticker.C:
			// 每次定时器触发时，发送一个 WebSocket Ping 消息作为心跳
			err := conn.WriteMessage(websocket.PingMessage, []byte("heartbeat"))
			if err != nil {
				log.Println(err) // 记录发送心跳消息时的错误
				return           // 如果发送心跳消息失败，退出函数
			}

			// TODO: 心跳时间从配置文件中读取
			// 设置读取消息的截止时间，防止长时间没有接收到消息
			// 目前设置了一个非常长的超时时间
			conn.SetReadDeadline(time.Now().Add(24 * 360 * time.Hour))

			// 尝试读取来自客户端的响应消息
			_, _, err = conn.ReadMessage()
			if err != nil {
				// 如果读取消息失败，断开连接并记录错误
				disConnect(user)
				log.Println("heartbeat response error:", err)
				return
			}
		}
	}
}

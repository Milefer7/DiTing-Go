package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils/jsonUtils"
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/spf13/viper"
)

// init 初始化 RocketMQ 消费者并订阅 FriendApplyTopic 主题。
func init() {
	// 从配置中获取 RocketMQ 的主机地址
	host := viper.GetString("rocketmq.host")

	// 创建一个新的 RocketMQ 推送消费者
	rocketConsumer, _ := rocketmq.NewPushConsumer(
		// 指定消费者组
		consumer.WithGroupName(enum.FriendApplyTopic),
		// 设置 namesrv 地址
		consumer.WithNameServer([]string{host}),
	)

	// 订阅主题 FriendApplyTopic 并设置处理函数
	err := rocketConsumer.Subscribe(enum.FriendApplyTopic, consumer.MessageSelector{}, friendApplyEvent)
	if err != nil {
		// 如果订阅失败，记录错误并终止程序
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}

	// 启动消费者
	err = rocketConsumer.Start()
	if err != nil {
		// 如果启动失败，记录错误并终止程序
		global.Logger.Panicf(": %s", err.Error())
	}
}

// FriendApplyEvent 好友申请事件处理函数
func friendApplyEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码消息
		userApplyR := model.UserApply{}
		if err := jsonUtils.UnmarshalMsg(&userApplyR, ext[i]); err != nil {
			// 如果解码失败，记录错误并稍后重试
			global.Logger.Errorf("jsonUtils unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}

		// 处理好友申请
		if err := friendApply(userApplyR); err != nil {
			// 如果处理失败，记录错误并稍后重试
			global.Logger.Errorf("friendApply error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
	}
	// 消费成功
	return consumer.ConsumeSuccess, nil
}

// friendApply 处理好友申请的具体逻辑
func friendApply(apply model.UserApply) error {
	// 发送新消息事件（示例代码，实际实现需根据具体业务逻辑）
	//service.Send(apply.TargetID)
	return nil
}

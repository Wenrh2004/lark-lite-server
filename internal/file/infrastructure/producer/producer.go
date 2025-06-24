package producer

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/bytedance/sonic"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/event"
)

type DelayHandle struct {
	MsgID      string
	BrokerName string
	QueueId    int
}

type Producer struct {
	client     rocketmq.Producer
	topic      string
	delayLevel int
}

func NewProducer(conf *viper.Viper) (*Producer, func()) {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(conf.GetStringSlice("app.mq.path")),
		producer.WithGroupName(conf.GetString("app.mq.producer.group")),
	)
	if err != nil {
		panic(err)
	}

	return &Producer{
			client:     p,
			topic:      conf.GetString("app.mq.topic"),
			delayLevel: conf.GetInt("app.mq.delay_level"),
		}, func() {
			p.Shutdown()
		}
}

func (p *Producer) SendExpiryMessage(ctx context.Context, fileID uint64) error {
	bytes, err := sonic.Marshal(&event.UploadEvent{
		Type:   event.Failed,
		FileID: fileID,
	})
	if err != nil {
		return fmt.Errorf("[Infrastructure.Producer.SendExpiryMessage]marshal failed event: %w", err)
	}
	msg := &primitive.Message{
		Topic: p.topic,
		Body:  bytes,
	}
	// delay level 4 ≈ 15 分钟  [oai_citation:0‡阿里云](https://www.alibabacloud.com/blog/rocketmq-message-integration-multi-type-business-message-scheduled-messages_599697?utm_source=chatgpt.com)
	msg.WithDelayTimeLevel(4)
	msg.WithTag("FAILED_UPLOAD")

	_, err = p.client.SendSync(ctx, msg)
	if err != nil {
		return fmt.Errorf("[Infrastructure.Producer.SendExpiryMessage]failed to send message to %s, err: %v", p.topic, err)
	}

	return nil
}

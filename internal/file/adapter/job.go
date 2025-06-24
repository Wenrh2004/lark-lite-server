package adapter

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/Wenrh2004/lark-lite-server/internal/file/domain"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/event"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

type FileJob struct {
	srv *adapter.Service
	fs  domain.FileService
}

func NewFileJob(srv *adapter.Service, fs domain.FileService) *FileJob {
	return &FileJob{
		srv: srv,
		fs:  fs,
	}
}

func (f *FileJob) UploadFailed(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		var e event.UploadEvent
		if err := sonic.Unmarshal(msg.Body, &e); err != nil {
			return consumer.ConsumeRetryLater, err
		}
		file := &domain.File{
			ID: e.FileID,
		}
		if err := f.fs.UploadFailed(ctx, file); err != nil {
			f.srv.Logger.Error("[Adapter.FileJob.UploadFailed]upload failed", zap.Uint64("file_id", file.ID), zap.Error(err))
			return consumer.ConsumeRetryLater, fmt.Errorf("[Adapter.FileJob.UploadFailed]file id:%d : %w", file.ID, err)
		}
	}
	return consumer.ConsumeSuccess, nil
}

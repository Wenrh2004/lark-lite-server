package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Wenrh2004/lark-lite-server/internal/file/domain"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/model"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/producer"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/repository/query"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/third/oss"
)

type FileRepository struct {
	db  *query.Query
	rdb *redis.Client
	p   *producer.Producer
	oss oss.Service
}

func (f *FileRepository) GetPreUploadURL(ctx context.Context, file *domain.File) (*domain.File, error) {
	ext, err := sonic.Marshal(file.ExtJSON)
	if err != nil {
		return nil, fmt.Errorf("[Infrastructure.FileRepository.GetPreUploadURL]marshal extension failed: %w", err)
	}
	uploadResp, err := f.oss.PreUpload(ctx, &oss.Object{
		Bucket: file.Domain,
		Key:    strconv.FormatUint(file.ID, 10),
	})
	if err != nil {
		return nil, fmt.Errorf("[Infrastructure.FileRepository.GetPreUploadURL]oss pre upload failed: %w", err)
	}
	file.AccessURL = uploadResp.AccessURL
	if err := f.db.WithContext(ctx).File.Create(&model.File{
		ID:       file.ID,
		Domain:   file.Domain,
		FileName: file.Name,
		FilePath: uploadResp.AccessURL,
		FileSize: uint64(file.Size),
		FileType: file.Type,
		FileHash: file.Hash,
		ExtJSON:  &ext,
	}); err != nil {
		return nil, fmt.Errorf("[Infrastructure.FileRepository.GetPreUploadURL]create file failed: %w", err)
	}
	return &domain.File{}, nil
}

func (f *FileRepository) PendingUpload(ctx context.Context, fileId uint64) error {
	if err := f.rdb.Set(ctx, fmt.Sprintf("FILE:%d", fileId), fileId, 0).Err(); err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.PendingUpload]set file cache failed: %w", err)
	}
	if err := f.p.SendExpiryMessage(ctx, fileId); err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.PendingUpload]send expiry message failed: %w", err)
	}
	return nil
}

func (f *FileRepository) PreUpload(ctx context.Context, file *domain.File) (*domain.File, error) {
	fileInfo, err := f.db.WithContext(ctx).File.Where(query.File.FileHash.Eq(file.Hash)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return f.GetPreUploadURL(ctx, file)
		}
		return nil, fmt.Errorf("[Infrastructure.FileRepository.PreUpload]query file %v failed: %w", file.ID, err)
	}
	switch fileInfo.Status {
	case 0:
		return &domain.File{
			ID:        fileInfo.ID,
			Exists:    false,
			AccessURL: fileInfo.FilePath,
		}, nil
	case 1:
		return &domain.File{
			ID:        fileInfo.ID,
			Exists:    true,
			AccessURL: fileInfo.FilePath,
		}, nil
	case 2:
		return f.GetPreUploadURL(ctx, file)
	default:
		return nil, errors.New("[Infrastructure.FileRepository.PreUpload]file status error")
	}
}

func (f *FileRepository) CompleteUpload(ctx context.Context, file *domain.File) error {
	exists, err := f.oss.CheckFileExists(ctx, file.Domain, file.Name)
	if err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.CompleteUpload]check file exists failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("[Infrastructure.FileRepository.CompleteUpload]file %d not exists: %w", file.ID, err)
	}
	if err = f.SetFileStatus(ctx, file.ID, domain.FileStatusSuccess); err != nil {
		return err
	}
	if err := f.rdb.Del(ctx, fmt.Sprintf("FILE:%d", file.ID)).Err(); err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.CompleteUpload]delete file cache failed: %w", err)
	}

	return nil
}

func (f *FileRepository) CreateFileByUploadIDMapping(ctx context.Context, file *domain.File) error {
	if err := f.db.WithContext(ctx).FileUser.Create(&model.FileUser{
		FileID: file.ID,
		UserID: file.UploadBy,
	}); err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.CompleteUpload]create file user mapping failed: %w", err)
	}

	return nil
}

func (f *FileRepository) SetFileStatus(ctx context.Context, fileId uint64, status int) error {
	resultInfo, err := f.db.WithContext(ctx).File.Where(query.File.ID.Eq(fileId)).Update(query.File.Status, status)
	if err != nil {
		return fmt.Errorf("[Infrastructure.FileRepository.SetFileStatus]update status %d failed: %w", status, err)
	}
	if resultInfo.RowsAffected == 0 {
		return fmt.Errorf("[Infrastructure.FileRepository.SetFileStatus]file %d not found: %w", fileId, err)
	}
	return nil
}

func (f *FileRepository) GetFile(ctx context.Context, file *domain.File) (*domain.File, error) {
	// TODO implement me
	panic("implement me")
}

func NewFileRepository(
	rdb *redis.Client,
	oss oss.Service,
) domain.FileRepository {
	return &FileRepository{
		db:  query.Q,
		rdb: rdb,
		oss: oss,
	}
}

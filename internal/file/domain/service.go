package domain

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/pkg/domain"
)

type FileService interface {
	GetPreUploadURL(ctx context.Context, file *File) (*File, error)
	CompleteUpload(ctx context.Context, file *File) error
	UploadFailed(ctx context.Context, file *File) error
	GetFile(ctx context.Context, file *File) (*File, error)
}

type fileService struct {
	srv  *domain.Service
	repo FileRepository
}

func (f *fileService) GetPreUploadURL(ctx context.Context, file *File) (*File, error) {
	uploadInfo, err := f.repo.PreUpload(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("[Domain.FileService.GetPreUploadURL]pre upload: %w", err)
	}
	return uploadInfo, nil
}

func (f *fileService) CompleteUpload(ctx context.Context, file *File) error {
	if err := f.srv.Tx.Transaction(ctx, func(ctx context.Context) error {
		if err := f.repo.CompleteUpload(ctx, file); err != nil {
			return fmt.Errorf("[Domain.FileService.CompleteUpload]complete upload failed: %w", err)
		}
		if err := f.repo.CreateFileByUploadIDMapping(ctx, file); err != nil {
			return fmt.Errorf("[Domain.FileService.CompleteUpload]create mapping failed: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (f *fileService) UploadFailed(ctx context.Context, file *File) error {
	if err := f.repo.SetFileStatus(ctx, file.ID, FileStatusFailed); err != nil {
		return fmt.Errorf("[Domain.FileService.UploadFailed]upload failed: %w", err)
	}
	return nil
}

func (f *fileService) GetFile(ctx context.Context, file *File) (*File, error) {
	// TODO implement me
	panic("implement me")
}

func NewFileService(srv *domain.Service, repo FileRepository) FileService {
	return &fileService{
		srv:  srv,
		repo: repo,
	}
}

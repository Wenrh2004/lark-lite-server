package domain

import "context"

type FileRepository interface {
	PreUpload(ctx context.Context, file *File) (*File, error)
	CompleteUpload(ctx context.Context, file *File) error
	CreateFileByUploadIDMapping(ctx context.Context, file *File) error
	SetFileStatus(ctx context.Context, fileId uint64, status int) error
	GetFile(ctx context.Context, file *File) (*File, error)
}

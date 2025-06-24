package adapter

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/internal/file/domain"
	"github.com/Wenrh2004/lark-lite-server/kitex_gen/file"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

type FileService struct {
	srv *adapter.Service
	fs  domain.FileService
}

func NewFileService(srv *adapter.Service, fs domain.FileService) *FileService {
	return &FileService{
		srv: srv,
		fs:  fs,
	}
}

func (f *FileService) PrepareUpload(ctx context.Context, req *file.PrepareUploadReq) (res *file.PrepareUploadResp, err error) {
	uploadURL, err := f.fs.GetPreUploadURL(ctx, &domain.File{
		Domain: req.GetDomain(),
		Name:   req.GetFileName(),
		Size:   req.GetSize(),
		Hash:   req.GetMd5(),
		Type:   req.GetContentType(),
	})
	if err != nil {
		return nil, err
	}
	return &file.PrepareUploadResp{
		Exists:    uploadURL.Exists,
		FileId:    uploadURL.ID,
		UploadUrl: uploadURL.UploadURL,
		AccessUrl: uploadURL.AccessURL,
	}, nil
}

func (f *FileService) CompleteUpload(ctx context.Context, req *file.CompleteUploadReq) (res *file.CompleteUploadResp, err error) {
	if err = f.fs.CompleteUpload(ctx, &domain.File{
		ID:       req.FileId,
		UploadBy: req.UploadBy,
	}); err != nil {
		return nil, fmt.Errorf("[Adapter.FileService.CompleteUpload] Completeupload failed: %w", err)
	}
	return &file.CompleteUploadResp{
		Success: true,
	}, nil
}

func (f *FileService) GetFileStatus(ctx context.Context, req *file.GetFileStatusReq) (res *file.GetFileStatusResp, err error) {
	// TODO implement me
	panic("implement me")
}

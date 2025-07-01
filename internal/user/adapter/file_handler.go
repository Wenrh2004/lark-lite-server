package adapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	v1 "github.com/Wenrh2004/lark-lite-server/common/api/v1"
	"github.com/cloudwego/hertz/pkg/app"
)

// FileUploadHandler 处理文件上传
func FileUploadHandler(ctx context.Context, c *app.RequestContext) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	if fileHeader.Size > 10*1024*1024 {
		v1.HandlerError(c, v1.Error{Code: 400, Message: "文件过大，最大10MB"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	defer file.Close()

	saveDir := "static/upload"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	// ext := filepath.Ext(fileHeader.Filename) // 已不再使用，移除
	// 可选：为文件名加上时间戳或随机前缀防止覆盖
	filename := fmt.Sprintf("%d_%s", ctx.Value("ts"), fileHeader.Filename)
	filePath := filepath.Join(saveDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	url := "/static/upload/" + filename
	c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "上传成功",
		"url":     url,
	})
}

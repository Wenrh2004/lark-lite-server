package domain

import (
	"strconv"
	"strings"
	"time"
)

const (
	FileStatusPending = iota
	FileStatusSuccess
	FileStatusFailed
)

type File struct {
	ID        uint64
	Domain    string
	Name      string
	Size      int64
	Hash      string
	Type      string
	ExtJSON   string
	Exists    bool
	UploadURL string
	AccessURL string
	ExpiresAt time.Time
	UploadBy  uint64
}

func (f *File) GetFileKey() string {
	return strings.Join([]string{strconv.FormatUint(f.ID, 10), f.Name}, ":")
}

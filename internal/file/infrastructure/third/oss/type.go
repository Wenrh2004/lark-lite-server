package oss

import "time"

type Object struct {
	Bucket string
	Key    string
}

type UploadResponse struct {
	UploadURL string
	AccessURL string
	ExpiresAt time.Time
}

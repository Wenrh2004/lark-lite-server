package event

const (
	Success = iota + 1
	Failed
)

type UploadEvent struct {
	Type   int    `json:"type"`
	FileID uint64 `json:"file_id"`
}

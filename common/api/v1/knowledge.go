package v1

type Knowledge struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     uint64 `json:"owner_id"`
}

type CreateKnowledgeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateKnowledgeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateDocumentRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content"`
	KnowledgeID uint64 `json:"knowledge_id" binding:"required"`
}

type UpdateDocumentRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
}

type AddRecentAccessRequest struct {
	ResourceID   uint64 `json:"resource_id" binding:"required"`
	ResourceType int    `json:"resource_type" binding:"required"`
}

type KnowledgeListResponse struct {
	List  []*Knowledge `json:"list"`
	Total int64        `json:"total"`
}

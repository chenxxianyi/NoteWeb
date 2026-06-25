package models

// DashboardSummary represents the aggregated data for the dashboard
type DashboardSummary struct {
	DocumentCount        int               `json:"documentCount"`
	AnnotationCount      int               `json:"annotationCount"`
	NoteCount            int               `json:"noteCount"`
	WeeklyReadingSeconds int               `json:"weeklyReadingSeconds"`
	ContinueReading      []ContinueReading `json:"continueReading"`
	RecentActivities     []RecentActivity  `json:"recentActivities"`
}

type ContinueReading struct {
	ID           uint    `json:"id"`
	Title        string  `json:"title"`
	FileType     string  `json:"fileType"`
	ReadProgress float64 `json:"readProgress"`
	UpdatedAt    string  `json:"updatedAt"`
}

type RecentActivity struct {
	Type        string `json:"type"`        // "highlight", "note", "upload"
	ID          uint   `json:"id"`
	DocumentID  uint   `json:"documentId"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	Document    string `json:"document"`
	Page        string `json:"page"`
	CreatedAt   string `json:"createdAt"`
}

// Dashboard repository interface
type DashboardRepo interface {
	GetSummary(userID uint) (*DashboardSummary, error)
}

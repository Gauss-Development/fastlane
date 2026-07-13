package entities

import "time"

type Project struct {
	ID           string    `json:"id"`
	OwnerID      string    `json:"owner_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	Status       string    `json:"status"`
	OwnerEmail   string    `json:"owner_email"`
	OwnerCompany string    `json:"owner_company"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DesignFile struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Kind          string    `json:"kind"`
	Filename      string    `json:"filename"`
	Version       int32     `json:"version"`
	ContentSHA256 string    `json:"content_sha256"`
	ObjectKey     string    `json:"object_key"`
	SizeBytes     int64     `json:"size_bytes"`
	ContentType   string    `json:"content_type"`
	UploadedBy    string    `json:"uploaded_by"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type NDA struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	ManufacturerID string     `json:"manufacturer_id"`
	Status         string     `json:"status"`
	NDAVersion     string     `json:"nda_version"`
	AcceptedIP     string     `json:"accepted_ip"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

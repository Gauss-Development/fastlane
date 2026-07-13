package entities

import "time"

type Manufacturer struct {
	ID              string
	UserID          string
	Name            string
	NameZh          string
	City            string
	Country         string
	Cluster         string
	Description     string
	Website         string
	ServiceTypes    []string
	AssemblyTypes   []string
	MinLayers       int32
	MaxLayers       int32
	Materials       []string
	SurfaceFinishes []string
	MinOrderQty     int32
	MaxOrderQty     int32
	LeadTimeDays    int32
	MonthlyCapacity int32
	SmallestPackage string
	Certifications  []string
	Verified        bool
	VerifiedAt      *time.Time
	Rating          float64
	OrderCount      int32
	OnTimeRate      float64
	ContactEmail    string
	ContactWechat   string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

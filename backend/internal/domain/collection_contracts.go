package domain

type VehicleSetInput struct {
	Name             string `json:"name"`
	Manufacturer     string `json:"manufacturer"`
	ArticleNumber    string `json:"articleNumber"`
	ArticleSourceURL string `json:"articleSourceUrl"`
	Gauge            string `json:"gauge"`
	Epoch            string `json:"epoch"`
	RailwayCompany   string `json:"railwayCompany"`
	Category         string `json:"category"`
	Gattung          string `json:"gattung"`
	Description      string `json:"description"`
	EAN              string `json:"ean"`
	ProductionPeriod string `json:"productionPeriod"`
	ListPrice        string `json:"listPrice"`
	AcquisitionType  string `json:"acquisitionType"`
	AcquiredFrom     string `json:"acquiredFrom"`
	PurchasePrice    string `json:"purchasePrice"`
	PurchaseDate     string `json:"purchaseDate"`
	StorageLocation  string `json:"storageLocation"`
	StorageDetails   string `json:"storageDetails"`
	Condition        string `json:"condition"`
	ConditionDetails string `json:"conditionDetails"`
	Packaging        string `json:"packaging"`
}

type ExhibitionStatus string

const (
	ExhibitionStatusDraft     ExhibitionStatus = "draft"
	ExhibitionStatusOpen      ExhibitionStatus = "open"
	ExhibitionStatusLocked    ExhibitionStatus = "locked"
	ExhibitionStatusRunning   ExhibitionStatus = "running"
	ExhibitionStatusCompleted ExhibitionStatus = "completed"
	ExhibitionStatusArchived  ExhibitionStatus = "archived"
)

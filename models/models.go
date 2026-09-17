package models

type Inventory struct {
	SerialNum   string
	ItemName    string
	Category    string
	Qty         string
	Location    string
	Status      string
	Regional    string
	NetworkType string
	Mbps        string
	MSISDN      string
	Owner       string
	DateUpdate  string
	Note        string
}

type Ticket struct {
	ID              int
	Customer        string
	Product         string
	ReportedToIOH   string
	DateTicket      string
	TicketNumber    string
	LogIssue        string
	IdentifiedIssue string
	ResolveDate     string
	SLA             string
	Note            string
}

type DashboardData struct {
	TotalTickets      int
	TicketsOnProgress int
	TicketsResolved   int
	TotalInventory    int
	CategoryStats     map[string]int
}

type PageData struct {
	ActiveTab     string
	Inventories   []Inventory
	Tickets       []Ticket
	Visits     []SiteVisit
	SearchQuery   string
	EditInventory *Inventory
	EditTicket    *Ticket
	Dashboard     DashboardData
	IsAdmin       bool
	ErrorLogin    string
}
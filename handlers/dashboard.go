package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"noc-app/models"
	"github.com/xuri/excelize/v2"
)

const ExcelFileName = "data_inventaris.xlsx"

func HandleDashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Wajib login (cek apakah ada cookie session_role)
	if _, err := r.Cookie("session_role"); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var totalTickets, onProgress, resolved int
	db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&totalTickets)
	db.QueryRow("SELECT COUNT(*) FROM tickets WHERE sla = 'On Progress'").Scan(&onProgress)
	resolved = totalTickets - onProgress

	var totalInv int
	catStats := make(map[string]int)

	if _, err := os.Stat(ExcelFileName); err == nil {
		f, err := excelize.OpenFile(ExcelFileName)
		if err == nil {
			rows, err := f.GetRows("Sheet1")
			f.Close()
			if err == nil {
				for i, row := range rows {
					if i == 0 || len(row) == 0 {
						continue
					}
					totalInv++
					category := "Uncategorized"
					if len(row) > 2 && row[2] != "" {
						category = row[2]
					}
					catStats[category]++
				}
			}
		}
	}

	// === AMBIL DATA JADWAL VISIT DARI DATABASE ===
	visits := GetVisits(db)

	errMsg := r.URL.Query().Get("error")

	data := models.PageData{
		ActiveTab: "dashboard",
		Dashboard: models.DashboardData{
			TotalTickets:      totalTickets,
			TicketsOnProgress: onProgress,
			TicketsResolved:   resolved,
			TotalInventory:    totalInv,
			CategoryStats:     catStats,
		},
		Visits:     visits, // <--- Data visit dimasukkan ke sini
		IsAdmin:    IsAdmin(r),
		ErrorLogin: errMsg,
	}
	
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}
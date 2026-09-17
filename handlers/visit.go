package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"noc-app/models"
)

// HandleVisits menampilkan halaman khusus jadwal visit
func HandleVisits(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if _, err := r.Cookie("session_role"); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	visits := GetVisits(db)

	data := models.PageData{
		ActiveTab: "visits",
		Visits:    visits,
		IsAdmin:   IsAdmin(r),
	}

	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

func HandleSaveVisit(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/visits", http.StatusSeeOther)
		return
	}

	siteName := r.FormValue("site_name")
	visitDate := r.FormValue("visit_date")
	status := r.FormValue("status")
	reportLink := r.FormValue("report_link")

	_, err := db.Exec("INSERT INTO site_visits (site_name, visit_date, status, report_link) VALUES (?, ?, ?, ?)",
		siteName, visitDate, status, reportLink)
	if err != nil {
		http.Error(w, "Gagal menyimpan data visit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/visits", http.StatusSeeOther)
}

// HandleDeleteVisit untuk menghapus data visit (Admin Only)
func HandleDeleteVisit(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id != "" {
		db.Exec("DELETE FROM site_visits WHERE id = ?", id)
	}
	http.Redirect(w, r, "/visits", http.StatusSeeOther)
}

func GetVisits(db *sql.DB) []models.SiteVisit {
	rows, err := db.Query("SELECT id, site_name, visit_date, status, report_link FROM site_visits ORDER BY visit_date DESC")
	if err != nil {
		return []models.SiteVisit{}
	}
	defer rows.Close()

	var visits []models.SiteVisit
	for rows.Next() {
		var v models.SiteVisit
		if err := rows.Scan(&v.ID, &v.SiteName, &v.VisitDate, &v.Status, &v.ReportLink); err != nil {
			continue
		}
		visits = append(visits, v)
	}
	return visits
}
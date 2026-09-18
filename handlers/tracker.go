package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"noc-app/models"
)

func HandleTracker(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if _, err := r.Cookie("session_role"); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	searchQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	query := "SELECT id, customer, product, reported_to_ioh, date_ticket, ticket_number, log_issue, identified_issue, resolve_date, sla, note FROM tickets"

	var rows *sql.Rows
	var errQuery error

	if searchQuery != "" {
		query += " WHERE ticket_number LIKE ? OR customer LIKE ? OR product LIKE ? ORDER BY id DESC"
		likeParam := "%" + searchQuery + "%"
		rows, errQuery = db.Query(query, likeParam, likeParam, likeParam)
	} else {
		query += " ORDER BY id DESC"
		rows, errQuery = db.Query(query)
	}

	if errQuery != nil {
		http.Error(w, errQuery.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		err := rows.Scan(&t.ID, &t.Customer, &t.Product, &t.ReportedToIOH, &t.DateTicket, &t.TicketNumber, &t.LogIssue, &t.IdentifiedIssue, &t.ResolveDate, &t.SLA, &t.Note)
		if err == nil {
			tickets = append(tickets, t)
		}
	}

	data := models.PageData{
		ActiveTab:   "tracker",
		Tickets:     tickets,
		SearchQuery: searchQuery,
		IsAdmin:     IsAdmin(r),
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

func HandleSaveTicket(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tracker", http.StatusSeeOther)
		return
	}

	customer, product, dateTicket, ticketNumber, logIssue, identifiedIssue, note, reportedFormatted, resolveFormatted, slaResult := parseTicketForm(r)

	insertSQL := `INSERT INTO tickets (customer, product, reported_to_ioh, date_ticket, ticket_number, log_issue, identified_issue, resolve_date, sla, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, customer, product, reportedFormatted, dateTicket, ticketNumber, logIssue, identifiedIssue, resolveFormatted, slaResult, note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tracker", http.StatusSeeOther)
}

func HandleEditTicket(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/tracker", http.StatusSeeOther)
		return
	}

	var t models.Ticket
	err = db.QueryRow("SELECT id, customer, product, reported_to_ioh, date_ticket, ticket_number, log_issue, identified_issue, resolve_date, sla, note FROM tickets WHERE id = ?", id).
		Scan(&t.ID, &t.Customer, &t.Product, &t.ReportedToIOH, &t.DateTicket, &t.TicketNumber, &t.LogIssue, &t.IdentifiedIssue, &t.ResolveDate, &t.SLA, &t.Note)

	if err != nil {
		http.Redirect(w, r, "/tracker", http.StatusSeeOther)
		return
	}

	rows, _ := db.Query("SELECT id, customer, product, reported_to_ioh, date_ticket, ticket_number, log_issue, identified_issue, resolve_date, sla, note FROM tickets ORDER BY id DESC")
	defer rows.Close()
	var tickets []models.Ticket
	for rows.Next() {
		var tk models.Ticket
		rows.Scan(&tk.ID, &tk.Customer, &tk.Product, &tk.ReportedToIOH, &tk.DateTicket, &tk.TicketNumber, &tk.LogIssue, &tk.IdentifiedIssue, &tk.ResolveDate, &tk.SLA, &tk.Note)
		tickets = append(tickets, tk)
	}

	data := models.PageData{
		ActiveTab:  "tracker",
		Tickets:    tickets,
		EditTicket: &t,
		IsAdmin:    IsAdmin(r),
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

func HandleUpdateTicket(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/tracker", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	customer, product, dateTicket, ticketNumber, logIssue, identifiedIssue, note, reportedFormatted, resolveFormatted, slaResult := parseTicketForm(r)

	updateSQL := `UPDATE tickets SET customer=?, product=?, reported_to_ioh=?, date_ticket=?, ticket_number=?, log_issue=?, identified_issue=?, resolve_date=?, sla=?, note=? WHERE id=?`
	_, err := db.Exec(updateSQL, customer, product, reportedFormatted, dateTicket, ticketNumber, logIssue, identifiedIssue, resolveFormatted, slaResult, note, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tracker", http.StatusSeeOther)
}

func HandleDeleteTicket(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id != "" {
		db.Exec("DELETE FROM tickets WHERE id = ?", id)
	}
	http.Redirect(w, r, "/tracker", http.StatusSeeOther)
}

func parseTicketForm(r *http.Request) (string, string, string, string, string, string, string, string, string, string) {
    customer := r.FormValue("customer")
    product := r.FormValue("product")
    dateTicket := r.FormValue("date_ticket")
    ticketNumber := r.FormValue("ticket_number")
    logIssue := r.FormValue("log_issue")
    identifiedIssue := r.FormValue("identified_issue")
    note := r.FormValue("note")

    reportedStr := r.FormValue("reported_to_ioh")
    resolveStr := r.FormValue("resolve_date")

    var reportedFormatted, resolveFormatted, slaResult string
    layoutInput := "2006-01-02T15:04"
    layoutDisplay := "15:04, 02/01/2006"

    // Simpan nilai mentah input (YYYY-MM-DDTHH:MM) agar bisa dibaca kembali oleh input datetime-local saat Edit
    reportedFormatted = reportedStr 
    resolveFormatted = resolveStr

    if tReported, err := time.Parse(layoutInput, reportedStr); err == nil {
        if tResolve, err := time.Parse(layoutInput, resolveStr); err == nil {
            duration := tResolve.Sub(tReported)
            hours := int(duration.Hours())
            minutes := int(duration.Minutes()) % 60
            slaResult = fmt.Sprintf("%d Jam %d Menit", hours, minutes)
            
            // Opsional: Jika Anda ingin tabel menampilkan format jam yang rapi, 
            // Anda bisa menyimpannya di kolom terpisah atau menyesuaikan cara tampilnya di HTML.
            // Saat ini kita simpan format ISO agar form edit tidak kosong.
        } else {
            slaResult = "On Progress"
        }
    } else {
        slaResult = "-"
    }

    return customer, product, dateTicket, ticketNumber, logIssue, identifiedIssue, note, reportedFormatted, resolveFormatted, slaResult
}
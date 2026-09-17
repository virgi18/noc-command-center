package handlers

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"noc-app/models"

	"github.com/xuri/excelize/v2"
)

var ExcelMutex sync.Mutex

func HandleInventory(w http.ResponseWriter, r *http.Request) {
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
	var items []models.Inventory

	ExcelMutex.Lock()
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
					getCol := func(index int) string {
						if index < len(row) {
							return row[index]
						}
						return ""
					}
					inv := models.Inventory{
						SerialNum:   getCol(0),
						ItemName:    getCol(1),
						Category:    getCol(2),
						Qty:         getCol(3),
						Location:    getCol(4),
						Status:      getCol(5),
						Regional:    getCol(6),
						NetworkType: getCol(7),
						Mbps:        getCol(8),
						MSISDN:      getCol(9),
						Owner:       getCol(10),
						DateUpdate:  getCol(11),
						Note:        getCol(12),
					}

					if searchQuery != "" {
						match := strings.Contains(strings.ToLower(inv.SerialNum), strings.ToLower(searchQuery)) ||
							strings.Contains(strings.ToLower(inv.ItemName), strings.ToLower(searchQuery)) ||
							strings.Contains(strings.ToLower(inv.Location), strings.ToLower(searchQuery)) ||
							strings.Contains(strings.ToLower(inv.Regional), strings.ToLower(searchQuery))
						if !match {
							continue
						}
					}

					items = append(items, inv)
				}
			}
		}
	}
	ExcelMutex.Unlock()

	data := models.PageData{
		ActiveTab:   "inventory",
		Inventories: items,
		SearchQuery: searchQuery,
		IsAdmin:     IsAdmin(r),
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

func HandleSaveInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	values := getInventoryFormValues(r)
	if values[0] == "" {
		http.Redirect(w, r, "/inventory?error=Serial+Number+tidak+boleh+kosong", http.StatusSeeOther)
		return
	}

	ExcelMutex.Lock()
	defer ExcelMutex.Unlock()

	var f *excelize.File
	sheetName := "Sheet1"

	if _, err := os.Stat(ExcelFileName); os.IsNotExist(err) {
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", sheetName)
		headers := []string{"Serial Number", "Item Name", "Category", "Qty", "Location", "In/Out", "Regional", "Network Type", "Mbps", "MSISDN", "Owner", "Date Update", "Notes"}
		for idx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
			f.SetCellValue(sheetName, cell, h)
		}
	} else {
		var openErr error
		f, openErr = excelize.OpenFile(ExcelFileName)
		if openErr != nil {
			http.Error(w, openErr.Error(), http.StatusInternalServerError)
			return
		}
	}
	defer f.Close()

	rows, _ := f.GetRows(sheetName)
	for i, row := range rows {
		if i > 0 && len(row) > 0 && row[0] == values[0] {
			http.Redirect(w, r, "/inventory?error=Serial+Number+sudah+terdaftar!", http.StatusSeeOther)
			return
		}
	}

	nextRow := len(rows) + 1
	for idx, val := range values {
		cell, _ := excelize.CoordinatesToCellName(idx+1, nextRow)
		f.SetCellValue(sheetName, cell, val)
	}

	f.SaveAs(ExcelFileName)
	http.Redirect(w, r, "/inventory", http.StatusSeeOther)
}

func HandleEditInventory(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	targetSerial := r.URL.Query().Get("serial")
	if targetSerial == "" {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	var editItem models.Inventory
	var items []models.Inventory

	ExcelMutex.Lock()
	f, err := excelize.OpenFile(ExcelFileName)
	if err == nil {
		rows, err := f.GetRows("Sheet1")
		f.Close()
		if err == nil {
			for i, row := range rows {
				if i == 0 || len(row) == 0 {
					continue
				}
				getCol := func(index int) string {
					if index < len(row) {
						return row[index]
					}
					return ""
				}
				inv := models.Inventory{
					SerialNum:   getCol(0),
					ItemName:    getCol(1),
					Category:    getCol(2),
					Qty:         getCol(3),
					Location:    getCol(4),
					Status:      getCol(5),
					Regional:    getCol(6),
					NetworkType: getCol(7),
					Mbps:        getCol(8),
					MSISDN:      getCol(9),
					Owner:       getCol(10),
					DateUpdate:  getCol(11),
					Note:        getCol(12),
				}
				if inv.SerialNum == targetSerial {
					editItem = inv
				}
				items = append(items, inv)
			}
		}
	}
	ExcelMutex.Unlock()

	data := models.PageData{
		ActiveTab:     "inventory",
		Inventories:   items,
		EditInventory: &editItem,
		IsAdmin:       IsAdmin(r),
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

func HandleUpdateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	targetSerial := r.FormValue("original_serial")
	if targetSerial == "" {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	values := getInventoryFormValues(r)

	ExcelMutex.Lock()
	defer ExcelMutex.Unlock()

	f, err := excelize.OpenFile(ExcelFileName)
	if err != nil {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	targetRow := -1
	for i, row := range rows {
		if i > 0 && len(row) > 0 && row[0] == targetSerial {
			targetRow = i + 1
			break
		}
	}

	if targetRow == -1 {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	for idx, val := range values {
		cell, _ := excelize.CoordinatesToCellName(idx+1, targetRow)
		f.SetCellValue("Sheet1", cell, val)
	}

	f.SaveAs(ExcelFileName)
	http.Redirect(w, r, "/inventory", http.StatusSeeOther)
}

func HandleDeleteInventory(w http.ResponseWriter, r *http.Request) {
	targetSerial := r.URL.Query().Get("serial")
	if targetSerial == "" {
		http.Redirect(w, r, "/inventory", http.StatusSeeOther)
		return
	}

	ExcelMutex.Lock()
	defer ExcelMutex.Unlock()

	f, err := excelize.OpenFile(ExcelFileName)
	if err == nil {
		rows, err := f.GetRows("Sheet1")
		if err == nil {
			for i, row := range rows {
				if i > 0 && len(row) > 0 && row[0] == targetSerial {
					f.RemoveRow("Sheet1", i+1)
					f.SaveAs(ExcelFileName)
					break
				}
			}
		}
		f.Close()
	}
	http.Redirect(w, r, "/inventory", http.StatusSeeOther)
}

func getInventoryFormValues(r *http.Request) []string {
	serialNum := strings.TrimSpace(r.FormValue("serial_num"))
	itemName := strings.TrimSpace(r.FormValue("item_name"))
	category := strings.TrimSpace(r.FormValue("category"))
	qty := r.FormValue("qty")
	location := strings.TrimSpace(r.FormValue("location"))
	status := r.FormValue("status")
	regional := strings.TrimSpace(r.FormValue("regional"))
	networkType := strings.TrimSpace(r.FormValue("network_type"))
	mbps := r.FormValue("mbps")
	msisdn := strings.TrimSpace(r.FormValue("msisdn"))
	owner := strings.TrimSpace(r.FormValue("owner"))
	note := strings.TrimSpace(r.FormValue("note"))
	dateUpdate := time.Now().Format("15:04 02/01/2006")

	return []string{serialNum, itemName, category, qty, location, status, regional, networkType, mbps, msisdn, owner, dateUpdate, note}
}
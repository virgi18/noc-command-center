package main

import (
	"fmt"
	"net/http"
	"os" // Tambahkan package os
	"noc-app/database"
	"noc-app/handlers"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	// Routing Auth
	http.HandleFunc("/login", handlers.HandleLogin)
	http.HandleFunc("/logout", handlers.HandleLogout)

	// Routing Utama & Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleDashboard(w, r, db)
	})

	// Routing Inventaris
	http.HandleFunc("/inventory", handlers.HandleInventory)
	http.HandleFunc("/save-inventory", handlers.GuardAdmin(handlers.HandleSaveInventory))
	http.HandleFunc("/edit-inventory", handlers.GuardAdmin(handlers.HandleEditInventory))
	http.HandleFunc("/update-inventory", handlers.GuardAdmin(handlers.HandleUpdateInventory))
	http.HandleFunc("/delete-inventory", handlers.GuardAdmin(handlers.HandleDeleteInventory))

	// Routing Troubleshoot Tracker
	http.HandleFunc("/tracker", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleTracker(w, r, db)
	})
	http.HandleFunc("/save-ticket", handlers.GuardAdmin(func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleSaveTicket(w, r, db)
	}))
	http.HandleFunc("/edit-ticket", handlers.GuardAdmin(func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleEditTicket(w, r, db)
	}))
	http.HandleFunc("/update-ticket", handlers.GuardAdmin(func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleUpdateTicket(w, r, db)
	}))
	http.HandleFunc("/delete-ticket", handlers.GuardAdmin(func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleDeleteTicket(w, r, db)
	}))

	// Ambil Port dari sistem Cloud, jika kosong gunakan 8080 untuk lokal
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server berjalan di port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Gagal memulai server: %v\n", err)
	}
}
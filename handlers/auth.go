package handlers

import (
	"html/template"
	"net/http"
)

func IsAdmin(r *http.Request) bool {
	cookie, err := r.Cookie("session_role")
	if err != nil {
		return false
	}
	return cookie.Value == "admin"
}

func GuardAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAdmin(r) {
			http.Redirect(w, r, "/dashboard?error=Akses+ditolak.+Anda+harus+login+sebagai+Admin.", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		action := r.FormValue("action")

		// Jika tombol "Masuk sebagai Guest" ditekan
		if action == "guest" {
			cookie := http.Cookie{
				Name:     "session_role",
				Value:    "guest",
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// Proses Login Admin
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "noc12345" {
			cookie := http.Cookie{
				Name:     "session_role",
				Value:    "admin",
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		} else {
			http.Redirect(w, r, "/login?error=Username+atau+Password+salah!", http.StatusSeeOther)
			return
		}
	}

	// Tampilkan halaman login khusus
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	errMsg := r.URL.Query().Get("error")
	tmpl.ExecuteTemplate(w, "login.html", map[string]string{"ErrorLogin": errMsg})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "session_role",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
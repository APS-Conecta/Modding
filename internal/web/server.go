// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/APS-Conecta/Modding/internal/catalog"
	"github.com/APS-Conecta/Modding/internal/config"
	"github.com/APS-Conecta/Modding/internal/state"
)

//go:embed templates/*.gohtml static
var assets embed.FS

type Options struct {
	Host   string
	Secure bool
}
type session struct {
	CSRF    string
	Expires time.Time
}
type App struct {
	catalog      *catalog.Catalog
	store        *state.Store
	template     *template.Template
	options      Options
	mu           sync.Mutex
	setupHash    [32]byte
	setupUsed    bool
	sessions     map[[32]byte]session
	failures     int
	blockedUntil time.Time
}

func randomToken() string {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(data[:])
}

func New(c *catalog.Catalog, s *state.Store, options Options) (*App, string, error) {
	if options.Host == "" {
		return nil, "", fmt.Errorf("an explicit HTTP host is required")
	}
	t, err := template.New("page").Funcs(template.FuncMap{"join": func(names []string) string { return strings.Join(names, "\n") }}).ParseFS(assets, "templates/*.gohtml")
	if err != nil {
		return nil, "", err
	}
	token := randomToken()
	return &App{catalog: c, store: s, template: t, options: options, setupHash: sha256.Sum256([]byte(token)), sessions: make(map[[32]byte]session)}, token, nil
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/establishments", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("GET /establishments", a.protect(a.establishments))
	mux.HandleFunc("POST /select", a.protect(a.selectEstablishment))
	mux.HandleFunc("GET /configure", a.protect(a.configure))
	mux.HandleFunc("POST /configure", a.protect(a.saveConfiguration))
	mux.HandleFunc("GET /preview", a.protect(a.preview))
	mux.HandleFunc("POST /confirm", a.protect(a.confirm))
	mux.HandleFunc("GET /configuration.json", a.protect(a.export))
	static, _ := fs.Sub(assets, "static")
	fileServer := http.StripPrefix("/assets/", http.FileServer(http.FS(static)))
	mux.HandleFunc("GET /assets/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/assets/")
		info, err := fs.Stat(static, name)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Preserve same-origin form origins; no-referrer can turn them into "null".
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; img-src 'self'; font-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
		if r.Host != a.options.Host {
			http.Error(w, "Dirección de acceso no permitida.", http.StatusMisdirectedRequest)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			scheme := "http"
			if a.options.Secure {
				scheme = "https"
			}
			origin := r.Header.Get("Origin")
			if (origin != "" && origin != scheme+"://"+a.options.Host) || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				http.Error(w, "Origen no permitido.", http.StatusForbidden)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
			if err := r.ParseForm(); err != nil {
				http.Error(w, "El formulario es demasiado grande o no es válido.", http.StatusBadRequest)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

type page struct {
	View           string
	Title          string
	Error          string
	CSRF           string
	Snapshot       state.Snapshot
	Query, Region  string
	Regions        []catalog.Region
	Results        []catalog.Entry
	Total, Count   int
	CatalogVersion string
	CatalogDigest  string
	Stale          bool
	Plan           config.Plan
	Apps           []config.App
}

func (a *App) render(w http.ResponseWriter, status int, p page) {
	p.CatalogVersion = a.catalog.Version
	p.CatalogDigest = a.catalog.Digest
	p.Count = a.catalog.Count()
	p.Stale = p.Snapshot.Config.Establishment.Code != "" && p.Snapshot.Config.CatalogDigest != a.catalog.Digest
	var output bytes.Buffer
	if err := a.template.ExecuteTemplate(&output, "page", p); err != nil {
		http.Error(w, "No se pudo mostrar la página.", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, 200, page{View: "login", Title: "Acceso a la configuración"})
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if time.Now().Before(a.blockedUntil) {
		http.Error(w, "Espere un minuto antes de volver a intentarlo.", 429)
		return
	}
	got := sha256.Sum256([]byte(r.PostForm.Get("token")))
	if a.setupUsed || subtle.ConstantTimeCompare(got[:], a.setupHash[:]) != 1 {
		a.failures++
		if a.failures >= 5 {
			a.blockedUntil = time.Now().Add(time.Minute)
			a.failures = 0
		}
		a.render(w, 401, page{View: "login", Title: "Acceso a la configuración", Error: "El código no es válido o ya fue utilizado. Solicite el código de la sesión actual al administrador del servidor."})
		return
	}
	a.setupUsed = true
	token := randomToken()
	a.sessions[sha256.Sum256([]byte(token))] = session{CSRF: randomToken(), Expires: time.Now().Add(8 * time.Hour)}
	http.SetCookie(w, &http.Cookie{Name: "modding_session", Value: token, Path: "/", HttpOnly: true, Secure: a.options.Secure, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60})
	http.Redirect(w, r, "/establishments", http.StatusSeeOther)
}

func (a *App) protect(next func(http.ResponseWriter, *http.Request, session)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("modding_session")
		if err != nil {
			http.Redirect(w, r, "/login", 303)
			return
		}
		a.mu.Lock()
		s, ok := a.sessions[sha256.Sum256([]byte(cookie.Value))]
		a.mu.Unlock()
		if !ok || time.Now().After(s.Expires) {
			http.Redirect(w, r, "/login", 303)
			return
		}
		if r.Method == "POST" && subtle.ConstantTimeCompare([]byte(r.PostForm.Get("csrf")), []byte(s.CSRF)) != 1 {
			http.Error(w, "La sesión del formulario no es válida. Recargue la página.", 403)
			return
		}
		next(w, r, s)
	}
}

func (a *App) load(w http.ResponseWriter, r *http.Request) (state.Snapshot, bool) {
	snapshot, err := a.store.Load(r.Context())
	if err != nil {
		http.Error(w, "No se pudo leer la configuración guardada.", 500)
		return snapshot, false
	}
	return snapshot, true
}

func (a *App) establishments(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	region := r.URL.Query().Get("region")
	if len(q) > 200 {
		http.Error(w, "Acorte el texto de búsqueda.", 400)
		return
	}
	rows, total := a.catalog.Search(q, region, 30)
	a.render(w, 200, page{View: "select", Title: "Elija su establecimiento", CSRF: s.CSRF, Snapshot: snapshot, Query: q, Region: region, Regions: a.catalog.Regions(), Results: rows, Total: total})
}

func revision(r *http.Request) (int64, error) {
	value, err := strconv.ParseInt(r.PostForm.Get("revision"), 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid revision")
	}
	return value, nil
}

func (a *App) save(w http.ResponseWriter, r *http.Request, value config.Config) bool {
	expected, err := revision(r)
	if err != nil {
		http.Error(w, "Recargue el formulario antes de guardar.", 400)
		return false
	}
	_, err = a.store.Save(r.Context(), expected, value)
	if errors.Is(err, state.ErrConflict) {
		http.Error(w, "La configuración cambió en otra pestaña. Recargue la página para revisar la versión actual.", 409)
		return false
	}
	if err != nil {
		http.Error(w, "No se pudo guardar la configuración.", 500)
		return false
	}
	return true
}

func (a *App) selectEstablishment(w http.ResponseWriter, r *http.Request, s session) {
	if r.PostForm.Get("catalog_digest") != a.catalog.Digest {
		http.Error(w, "El registro DEIS cambió. Recargue la búsqueda y confirme el establecimiento.", 409)
		return
	}
	entry, ok := a.catalog.Get(r.PostForm.Get("code"))
	if !ok {
		http.Error(w, "Seleccione un código válido del registro DEIS.", 400)
		return
	}
	snapshot, ok := a.load(w, r)
	if !ok {
		return
	}
	value := config.Config{SchemaVersion: 1, Status: "draft"}
	// Preserve choices only when explicitly revalidating the same establishment.
	if snapshot.Config.Establishment.Code == entry.Code {
		value = snapshot.Config
		value.Status = "draft"
	}
	value.Establishment = entry
	value.CatalogVersion = a.catalog.Version
	value.CatalogDigest = a.catalog.Digest
	if a.save(w, r, value) {
		http.Redirect(w, r, "/configure", 303)
	}
}

func (a *App) configure(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok {
		return
	}
	if snapshot.Config.Establishment.Code == "" {
		http.Redirect(w, r, "/establishments", 303)
		return
	}
	a.render(w, 200, page{View: "configure", Title: "Configure su establecimiento", CSRF: s.CSRF, Snapshot: snapshot, Apps: config.Apps()})
}

func (a *App) current(w http.ResponseWriter, snapshot state.Snapshot) bool {
	if snapshot.Config.Establishment.Code == "" {
		http.Error(w, "Seleccione un establecimiento DEIS antes de continuar.", 400)
		return false
	}
	if snapshot.Config.CatalogDigest != a.catalog.Digest {
		http.Error(w, "El registro DEIS cambió. Vuelva a seleccionar el establecimiento y revise su identidad.", 409)
		return false
	}
	return true
}

func (a *App) saveConfiguration(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok || !a.current(w, snapshot) {
		return
	}
	expected, revisionErr := revision(r)
	if revisionErr != nil {
		http.Error(w, "Recargue el formulario antes de guardar.", 400)
		return
	}
	if expected != snapshot.Revision {
		http.Error(w, "La configuración cambió en otra pestaña. Recargue la página antes de editar.", 409)
		return
	}
	value := snapshot.Config
	value.Domain = strings.ToLower(strings.TrimSpace(r.PostForm.Get("domain")))
	value.OptionalApps = r.PostForm["apps"]
	value.Status = "draft"
	var err error
	value.Sectors, err = config.Names(r.PostForm.Get("sectors"), "sector")
	if err == nil {
		value.Programs, err = config.Names(r.PostForm.Get("programs"), "programa")
	}
	if err == nil {
		value.Units, err = config.Names(r.PostForm.Get("units"), "unidad")
	}
	if err == nil {
		err = value.Validate()
	}
	if err != nil {
		// Keep the submitted fields available for correction, including invalid lists.
		value.Sectors = strings.Split(r.PostForm.Get("sectors"), "\n")
		value.Programs = strings.Split(r.PostForm.Get("programs"), "\n")
		value.Units = strings.Split(r.PostForm.Get("units"), "\n")
		snapshot.Config = value
		a.render(w, 400, page{View: "configure", Title: "Revise la configuración", Error: err.Error(), CSRF: s.CSRF, Snapshot: snapshot, Apps: config.Apps()})
		return
	}
	if a.save(w, r, value) {
		http.Redirect(w, r, "/preview", 303)
	}
}

func (a *App) preview(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok {
		return
	}
	if snapshot.Config.Validate() != nil {
		http.Redirect(w, r, "/configure", 303)
		return
	}
	a.render(w, 200, page{View: "preview", Title: "Revise su plan", CSRF: s.CSRF, Snapshot: snapshot, Plan: snapshot.Config.Plan()})
}

func (a *App) confirm(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok || !a.current(w, snapshot) {
		return
	}
	if err := snapshot.Config.Validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	value := snapshot.Config
	value.Status = "confirmed"
	if a.save(w, r, value) {
		http.Redirect(w, r, "/preview", 303)
	}
}

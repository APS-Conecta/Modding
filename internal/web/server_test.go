package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/APS-Conecta/Modding/internal/catalog"
	"github.com/APS-Conecta/Modding/internal/config"
	"github.com/APS-Conecta/Modding/internal/state"
)

const testCatalog = `codigo,tipo,nombre,direccion,comuna_codigo,comuna,region_codigo,region,servicio_salud,dependencia,latitud,longitud
900001,CESFAM,Centro Árbol,Calle Uno,99901,Comuna Sur,99,Región Prueba,Servicio Prueba,Municipal,-33,-70
900002,PSR,Posta Norte,Calle Dos,99902,Comuna Norte,98,Región Otra,Servicio Otro,Municipal,-34,-71
`

type harness struct {
	t           *testing.T
	server      *httptest.Server
	app         *App
	store       *state.Store
	client      *http.Client
	token, csrf string
}

func start(t *testing.T, store *state.Store, c *catalog.Catalog) *harness {
	t.Helper()
	ts := httptest.NewUnstartedServer(nil)
	app, token, err := New(c, store, Options{Host: ts.Listener.Addr().String()})
	if err != nil {
		t.Fatal(err)
	}
	ts.Config.Handler = app.Handler()
	ts.Start()
	t.Cleanup(ts.Close)
	jar, _ := cookiejar.New(nil)
	h := &harness{t: t, server: ts, app: app, store: store, token: token, client: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	return h
}

func fixture(t *testing.T) (*state.Store, *catalog.Catalog) {
	t.Helper()
	s, err := state.Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	c, err := catalog.Load([]byte(testCatalog), "test-v1")
	if err != nil {
		t.Fatal(err)
	}
	return s, c
}

func (h *harness) request(method, path string, data url.Values, headers map[string]string) (int, string) {
	h.t.Helper()
	body := ""
	if data != nil {
		body = data.Encode()
	}
	req, err := http.NewRequest(method, h.server.URL+path, strings.NewReader(body))
	if err != nil {
		h.t.Fatal(err)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for key, value := range headers {
		if key == "Host" {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}
	response, err := h.client.Do(req)
	if err != nil {
		h.t.Fatal(err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		h.t.Fatal(err)
	}
	return response.StatusCode, string(content)
}

func (h *harness) expect(method, path string, data url.Values, status int) string {
	h.t.Helper()
	got, body := h.request(method, path, data, nil)
	if got != status {
		h.t.Fatalf("%s %s: got %d want %d: %s", method, path, got, status, body)
	}
	return body
}

func (h *harness) login() {
	h.t.Helper()
	h.expect("POST", "/login", url.Values{"token": {h.token}}, 303)
	body := h.expect("GET", "/establishments?q=900001", nil, 200)
	match := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindStringSubmatch(body)
	if len(match) != 2 {
		h.t.Fatal("missing CSRF field")
	}
	h.csrf = match[1]
}

func (h *harness) form() url.Values {
	h.t.Helper()
	snapshot, err := h.store.Load(context.Background())
	if err != nil {
		h.t.Fatal(err)
	}
	return url.Values{"csrf": {h.csrf}, "revision": {strconv.FormatInt(snapshot.Revision, 10)}}
}

func (h *harness) selectCode(code string) {
	data := h.form()
	data.Set("code", code)
	data.Set("catalog_digest", h.app.catalog.Digest)
	data.Set("name", "forged identity")
	h.expect("POST", "/select", data, 303)
}

func (h *harness) configure() {
	data := h.form()
	data.Set("domain", "gestion.example.cl")
	data.Set("sectors", "Norte\nSur")
	data.Set("programs", "Salud mental")
	data.Set("units", "Dirección")
	data.Add("apps", "farmacia")
	h.expect("POST", "/configure", data, 303)
}

func TestWizardPersistsPlanAndIsolatesEstablishments(t *testing.T) {
	s, c := fixture(t)
	h := start(t, s, c)
	h.expect("GET", "/configuration.json", nil, 303)
	h.login()
	h.expect("GET", "/assets/brand/NunitoSans.woff2", nil, 200)
	data := h.form()
	data.Set("code", "999999")
	data.Set("catalog_digest", c.Digest)
	h.expect("POST", "/select", data, 400)
	h.selectCode("900001")
	h.configure()
	body := h.expect("GET", "/preview", nil, 200)
	for _, want := range []string{"Centro Árbol", "Sectores/Norte", "sector-norte", "nextcloud/IntraVox", "Confirmar plan"} {
		if !strings.Contains(body, want) {
			t.Errorf("preview missing %q", want)
		}
	}
	h.expect("POST", "/confirm", h.form(), 303)
	var saved config.Config
	if err := json.Unmarshal([]byte(h.expect("GET", "/configuration.json", nil, 200)), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Status != "confirmed" || saved.Establishment.Name != "Centro Árbol" || saved.CatalogDigest != c.Digest {
		t.Fatalf("saved identity/plan: %#v", saved)
	}
	// A new server session reuses persisted configuration without trusting an old token.
	h.server.Close()
	again := start(t, s, c)
	again.login()
	if !strings.Contains(again.expect("GET", "/preview", nil, 200), "Plan confirmado y guardado") {
		t.Fatal("lost confirmation")
	}
	again.selectCode("900002")
	snapshot, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Config.Establishment.Code != "900002" || len(snapshot.Config.Sectors) != 0 || len(snapshot.Config.OptionalApps) != 0 || snapshot.Config.Domain != "" || snapshot.Config.Status != "draft" {
		t.Fatalf("old establishment choices leaked: %#v", snapshot)
	}
	other, c := fixture(t)
	second := start(t, other, c)
	second.login()
	snapshot, _ = other.Load(context.Background())
	if snapshot.Config.Establishment.Code != "" {
		t.Fatal("separate installation inherited selection")
	}
}

func TestCatalogUpdateRequiresExplicitReview(t *testing.T) {
	s, c := fixture(t)
	h := start(t, s, c)
	h.login()
	h.selectCode("900001")
	h.configure()
	h.expect("POST", "/confirm", h.form(), 303)
	h.server.Close()
	updated, err := catalog.Load([]byte(strings.Replace(testCatalog, "Centro Árbol", "Centro Renombrado", 1)), "test-v2")
	if err != nil {
		t.Fatal(err)
	}
	h = start(t, s, updated)
	h.login()
	body := h.expect("GET", "/preview", nil, 200)
	if !strings.Contains(body, "El registro DEIS cambió") || !strings.Contains(body, "Centro Árbol") {
		t.Fatal("catalog update silently changed saved identity")
	}
	h.expect("POST", "/confirm", h.form(), 409)
	h.selectCode("900001")
	snapshot, _ := s.Load(context.Background())
	if snapshot.Config.Establishment.Name != "Centro Renombrado" || snapshot.Config.CatalogVersion != "test-v2" || len(snapshot.Config.Sectors) != 2 || snapshot.Config.Status != "draft" {
		t.Fatalf("explicit revalidation failed: %#v", snapshot)
	}
	h.expect("POST", "/confirm", h.form(), 303)
}

func TestRequestBoundariesAndStaleTabs(t *testing.T) {
	s, c := fixture(t)
	h := start(t, s, c)
	h.login()
	h.expect("POST", "/login", url.Values{"token": {h.token}}, 401)
	data := h.form()
	data.Set("code", "900001")
	data.Set("catalog_digest", c.Digest)
	for name, headers := range map[string]map[string]string{"origin": {"Origin": "https://attacker.invalid"}, "opaque-origin": {"Origin": "null"}, "fetch-site": {"Sec-Fetch-Site": "cross-site"}} {
		status, _ := h.request("POST", "/select", data, headers)
		if status != 403 {
			t.Errorf("%s: %d", name, status)
		}
	}
	bad := h.form()
	bad.Set("csrf", "wrong")
	h.expect("POST", "/select", bad, 403)
	if status, _ := h.request("GET", "/preview", nil, map[string]string{"Host": "attacker.invalid"}); status != 421 {
		t.Fatalf("host: %d", status)
	}
	h.selectCode("900001")
	stale := h.form()
	h.configure()
	stale.Set("domain", "not valid")
	h.expect("POST", "/configure", stale, 409)
	data = h.form()
	data.Set("domain", "gestion.example.cl")
	data.Set("sectors", "Norte\nnórte")
	h.expect("POST", "/configure", data, 400)
	snapshot, _ := s.Load(context.Background())
	if len(snapshot.Config.Sectors) != 2 {
		t.Fatal("invalid form altered saved plan")
	}
	data = h.form()
	data.Set("domain", "gestion.example.cl")
	data.Set("units", `<img src=x onerror=alert(1)>`)
	// HTML-like names can appear as data, but the template must always escape them.
	h.expect("POST", "/configure", data, 303)
	body := h.expect("GET", "/preview", nil, 200)
	if strings.Contains(body, `<img src=x`) || !strings.Contains(body, "&lt;img") {
		t.Fatal("organization data was not escaped")
	}
}

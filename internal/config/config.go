// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/APS-Conecta/Modding/internal/catalog"
)

type Config struct {
	SchemaVersion  int           `json:"schema_version"`
	Establishment  catalog.Entry `json:"establishment"`
	CatalogVersion string        `json:"catalog_version"`
	CatalogDigest  string        `json:"catalog_sha256"`
	Domain         string        `json:"domain"`
	Sectors        []string      `json:"sectors"`
	Programs       []string      `json:"programs"`
	Units          []string      `json:"units"`
	OptionalApps   []string      `json:"optional_apps"`
	Status         string        `json:"status"`
}

type App struct {
	ID, Name, Source string
	Optional         bool
}

// App choices describe the plan; this milestone does not install them.
func Apps() []App {
	return []App{
		{"intravox", "Inicio · IntraVox", "nextcloud/IntraVox", false},
		{"groupfolders", "Documentos compartidos", "Nextcloud App Store", false},
		{"eurooffice", "Oficina · EuroOffice", "AIO / Nextcloud App Store", false},
		{"calendar", "Calendario", "Nextcloud App Store", false},
		{"contacts", "Contactos", "Nextcloud App Store", false},
		{"epidemiologia", "Epidemiología", "APS-Conecta/epidemiologia", false},
		{"territorio", "Territorio", "APS-Conecta/territorio", false},
		{"spreed", "Talk", "AIO / Nextcloud App Store", false},
		{"notify_push", "Client Push", "AIO / Nextcloud App Store", false},
		{"fulltextsearch", "Búsqueda de texto completo", "Nextcloud App Store", false},
		{"fulltextsearch_elasticsearch", "Conector Elasticsearch", "Nextcloud App Store", false},
		{"files_fulltextsearch", "Búsqueda en archivos", "nextcloud/files_fulltextsearch", false},
		{"files_upload_guard", "Protección de cargas", "APS-Conecta/files-upload-guard", false},
		{"files_antivirus", "Antivirus", "AIO / Nextcloud App Store", false},
		{"side_menu", "Menú de navegación", "Nextcloud App Store", false},
		{"estadistica", "Estadística", "APS-Conecta/estadistica", true},
		{"farmacia", "Farmacia", "APS-Conecta/farmacia", true},
		{"totem", "Tótem", "APS-Conecta/totem", true},
	}
}

func (c Config) HasApp(id string) bool {
	for _, value := range c.OptionalApps {
		if value == id {
			return true
		}
	}
	return false
}

func (c Config) Validate() error {
	if c.Establishment.Code == "" || c.CatalogDigest == "" {
		return fmt.Errorf("Seleccione un establecimiento del registro DEIS.")
	}
	if net.ParseIP(c.Domain) != nil || len(c.Domain) > 253 || !strings.Contains(c.Domain, ".") {
		return fmt.Errorf("Ingrese el nombre de dominio de la suite, por ejemplo gestion.ejemplo.cl; no una dirección IP.")
	}
	label := regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	for _, part := range strings.Split(c.Domain, ".") {
		if !label.MatchString(part) {
			return fmt.Errorf("Ingrese un dominio válido, sin protocolo, puerto ni ruta.")
		}
	}
	allowed := make(map[string]bool)
	for _, app := range Apps() {
		if app.Optional {
			allowed[app.ID] = true
		}
	}
	seen := make(map[string]bool)
	for _, id := range c.OptionalApps {
		if !allowed[id] || seen[id] {
			return fmt.Errorf("La selección de aplicaciones opcionales no es válida.")
		}
		seen[id] = true
	}
	return nil
}

func Names(raw, prefix string) ([]string, error) {
	var names []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if strings.HasPrefix(catalog.Fold(name), prefix+" ") {
			name = strings.TrimSpace(name[len(prefix)+1:])
		}
		if name == "" || utf8.RuneCountInString(name) > 80 || strings.ContainsAny(name, "/\\|") || strings.IndexFunc(name, unicode.IsControl) >= 0 || name == "." || name == ".." {
			return nil, fmt.Errorf("Use nombres de hasta 80 caracteres, sin barras, saltos internos ni separadores.")
		}
		id := Slug(name)
		if id == "" {
			return nil, fmt.Errorf("El nombre «%s» debe contener letras o números.", name)
		}
		if seen[id] {
			return nil, fmt.Errorf("El nombre «%s» coincide con otro nombre de la misma lista.", name)
		}
		seen[id] = true
		names = append(names, name)
		if len(names) > 30 {
			return nil, fmt.Errorf("Ingrese hasta 30 nombres en cada lista.")
		}
	}
	return names, nil
}

func Slug(value string) string {
	var result strings.Builder
	separator := false
	for _, r := range catalog.Fold(value) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	return strings.TrimRight(result.String()[:min(result.Len(), 48)], "-")
}

type Group struct{ ID, Name string }
type Folder struct{ Path, Group, Access string }
type Plan struct {
	Groups  []Group
	Folders []Folder
	Apps    []App
}

func (c Config) Plan() Plan {
	plan := Plan{Groups: []Group{{"all-staff", "Todo el personal"}}, Folders: []Folder{{"Transversal", "all-staff", "Lectura"}}}
	for _, set := range []struct {
		prefix, folder string
		names          []string
	}{{"sector-", "Sectores", c.Sectors}, {"prog-", "Programas", c.Programs}, {"unidad-", "Unidades", c.Units}} {
		for _, name := range set.names {
			id := set.prefix + Slug(name)
			plan.Groups = append(plan.Groups, Group{id, name})
			plan.Folders = append(plan.Folders, Folder{set.folder + "/" + name, id, "Lectura y escritura, sin compartir"})
		}
	}
	for _, app := range Apps() {
		if !app.Optional || c.HasApp(app.ID) {
			plan.Apps = append(plan.Apps, app)
		}
	}
	return plan
}

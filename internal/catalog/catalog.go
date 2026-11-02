// SPDX-License-Identifier: AGPL-3.0-or-later
package catalog

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed data/*.csv
var files embed.FS

const DefaultVersion = "2026-07-23"

type Entry struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Address       string `json:"address"`
	Commune       string `json:"commune"`
	CommuneCode   string `json:"commune_code"`
	Region        string `json:"region"`
	RegionCode    string `json:"region_code"`
	HealthService string `json:"health_service"`
	Dependency    string `json:"dependency"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
}

type Region struct{ Code, Name string }

type Catalog struct {
	Version string
	Digest  string
	entries []Entry
	byCode  map[string]Entry
}

func Default() (*Catalog, error) {
	data, err := files.ReadFile("data/establecimientos-deis-" + DefaultVersion + ".csv")
	if err != nil {
		return nil, err
	}
	return Load(data, DefaultVersion)
}

// Fold follows gestion's accent- and case-insensitive search behavior.
func Fold(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return unicode.ToLower(r)
	}, norm.NFD.String(value))
}

func Load(data []byte, version string) (*Catalog, error) {
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("catalog version is required")
	}
	reader := csv.NewReader(bytes.NewReader(data))
	head, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("catalog header: %w", err)
	}
	positions := make(map[string]int)
	for i, key := range head {
		key = strings.TrimPrefix(key, "\ufeff")
		if _, exists := positions[key]; exists {
			return nil, fmt.Errorf("duplicate catalog column %s", key)
		}
		positions[key] = i
	}
	for _, key := range []string{"codigo", "tipo", "nombre", "direccion", "comuna_codigo", "comuna", "region_codigo", "region", "servicio_salud", "dependencia", "latitud", "longitud"} {
		if _, ok := positions[key]; !ok {
			return nil, fmt.Errorf("missing catalog column %s", key)
		}
	}
	hash := sha256.Sum256(data)
	result := &Catalog{Version: version, Digest: hex.EncodeToString(hash[:]), byCode: make(map[string]Entry)}
	codePattern := regexp.MustCompile(`^[0-9]{4,6}$`)
	for line := 2; ; line++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("catalog row %d: %w", line, err)
		}
		get := func(key string) string { return strings.TrimSpace(row[positions[key]]) }
		entry := Entry{Code: get("codigo"), Name: get("nombre"), Type: get("tipo"), Address: get("direccion"), Commune: get("comuna"), CommuneCode: get("comuna_codigo"), Region: get("region"), RegionCode: get("region_codigo"), HealthService: get("servicio_salud"), Dependency: get("dependencia"), Latitude: get("latitud"), Longitude: get("longitud")}
		if !codePattern.MatchString(entry.Code) || entry.Name == "" || entry.Commune == "" || entry.RegionCode == "" {
			return nil, fmt.Errorf("invalid establishment at row %d", line)
		}
		if _, exists := result.byCode[entry.Code]; exists {
			return nil, fmt.Errorf("duplicate DEIS code %s", entry.Code)
		}
		result.entries = append(result.entries, entry)
		result.byCode[entry.Code] = entry
	}
	if len(result.entries) == 0 {
		return nil, fmt.Errorf("empty catalog")
	}
	sort.Slice(result.entries, func(i, j int) bool { return Fold(result.entries[i].Name) < Fold(result.entries[j].Name) })
	return result, nil
}

func (c *Catalog) Get(code string) (Entry, bool) { e, ok := c.byCode[code]; return e, ok }
func (c *Catalog) Count() int                    { return len(c.entries) }

func (c *Catalog) Regions() []Region {
	values := make(map[string]string)
	for _, e := range c.entries {
		values[e.RegionCode] = e.Region
	}
	result := make([]Region, 0, len(values))
	for code, name := range values {
		result = append(result, Region{code, name})
	}
	sort.Slice(result, func(i, j int) bool { return Fold(result[i].Name) < Fold(result[j].Name) })
	return result
}

func (c *Catalog) Search(query, region string, limit int) ([]Entry, int) {
	terms := strings.Fields(Fold(query))
	if len(terms) == 0 && region == "" {
		return nil, 0
	}
	if limit < 1 || limit > 50 {
		limit = 50
	}
	var result []Entry
	total := 0
	for _, e := range c.entries {
		if region != "" && e.RegionCode != region {
			continue
		}
		text := Fold(strings.Join([]string{e.Code, e.Name, e.Type, e.Address, e.Commune, e.Region, e.HealthService, e.Dependency}, " "))
		match := true
		for _, term := range terms {
			if !strings.Contains(text, term) {
				match = false
				break
			}
		}
		if match {
			total++
			if len(result) < limit {
				result = append(result, e)
			}
		}
	}
	return result, total
}

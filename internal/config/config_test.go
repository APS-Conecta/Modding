package config

import (
	"github.com/APS-Conecta/Modding/internal/catalog"
	"testing"
)

func TestOrganizationValidation(t *testing.T) {
	names, err := Names("Sector Norte\nSalud Mental", "sector")
	if err != nil || len(names) != 2 || names[0] != "Norte" {
		t.Fatalf("normalization: %v %v", names, err)
	}
	for _, value := range []string{"Norte\nSECTOR nórte", "../Datos", "A|B", "A\x00B", ".."} {
		if _, err := Names(value, "sector"); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	c := Config{Establishment: catalog.Entry{Code: "900001"}, CatalogDigest: "fixture", Domain: "gestion.example.cl", Sectors: names, OptionalApps: []string{"farmacia"}}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	plan := c.Plan()
	if len(plan.Groups) != 3 || plan.Folders[1].Group != "sector-norte" || len(plan.Apps) != 16 {
		t.Fatalf("plan: %#v", plan)
	}
	for _, domain := range []string{"192.168.1.2", "http://gestion.example.cl", "local", "bad-.example.cl", "host:443", "a..cl"} {
		c.Domain = domain
		if c.Validate() == nil {
			t.Errorf("accepted domain %q", domain)
		}
	}
	c.Domain = "gestion.example.cl"
	c.OptionalApps = []string{"unreviewed-app"}
	if c.Validate() == nil {
		t.Fatal("accepted unknown optional app")
	}
}

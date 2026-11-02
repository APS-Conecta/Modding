package catalog

import (
	"strings"
	"testing"
)

const fixture = `codigo,tipo,nombre,direccion,comuna_codigo,comuna,region_codigo,region,servicio_salud,dependencia,latitud,longitud
900001,CESFAM,Centro Árbol,"Calle Uno, 12",99901,Comuna Sur,99,Región Prueba,Servicio Prueba,Municipal,-33,-70
900002,PSR,Posta Norte,Calle Dos,99902,Comuna Norte,98,Región Otra,Servicio Otro,Municipal,-34,-71
`

func TestSearchAndCatalogIntegrity(t *testing.T) {
	c, err := Load([]byte(fixture), "fixture-1")
	if err != nil {
		t.Fatal(err)
	}
	rows, total := c.Search("ARBOL sur", "99", 30)
	if total != 1 || len(rows) != 1 || rows[0].Code != "900001" || rows[0].Address != "Calle Uno, 12" {
		t.Fatalf("unexpected search: %#v %d", rows, total)
	}
	if _, total = c.Search("arbol", "98", 30); total != 0 {
		t.Fatal("region filter ignored")
	}
	if rows, total = c.Search("", "", 30); total != 0 || len(rows) != 0 {
		t.Fatal("empty search must not choose or list an establishment")
	}
	for name, text := range map[string]string{"duplicate": fixture + strings.Split(fixture, "\n")[1] + "\n", "bad code": strings.Replace(fixture, "900001", "oops", 1), "missing column": strings.Replace(fixture, "comuna_codigo", "unknown", 1)} {
		t.Run(name, func(t *testing.T) {
			if _, err := Load([]byte(text), "v1"); err == nil {
				t.Fatal("accepted invalid catalog")
			}
		})
	}
	changed, err := Load([]byte(strings.Replace(fixture, "Centro Árbol", "Centro Nuevo", 1)), "fixture-1")
	if err != nil {
		t.Fatal(err)
	}
	if c.Digest == changed.Digest {
		t.Fatal("catalog contents changed without digest change")
	}
}

func TestEmbeddedCatalog(t *testing.T) {
	c, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if c.Count() != 2655 {
		t.Fatalf("review imported catalog count: %d", c.Count())
	}
	if c.Version != DefaultVersion || len(c.Digest) != 64 {
		t.Fatal("missing catalog provenance")
	}
}

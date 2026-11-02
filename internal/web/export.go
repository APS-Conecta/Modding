// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"encoding/json"
	"net/http"
)

func (a *App) export(w http.ResponseWriter, r *http.Request, s session) {
	snapshot, ok := a.load(w, r)
	if !ok {
		return
	}
	if snapshot.Config.Establishment.Code == "" {
		http.Error(w, "Seleccione un establecimiento antes de exportar.", 400)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="modding-configuration.json"`)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(snapshot.Config)
}

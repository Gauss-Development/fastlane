package embedding

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ProductForEmbedding is the minimal view of a catalog product the text
// builder needs. We pass it explicitly (rather than the sqlc row type) so
// callers from different layers — seed CLI, search service, future re-index
// jobs — can construct it without importing sqlc types.
type ProductForEmbedding struct {
	Name            string
	NameZh          string
	Category        string
	Specs           []byte // raw jsonb
	SupplierName    string
	SupplierCity    string
	SupplierCluster string
}

// BuildEmbeddingText flattens a product + its supplier context into a single
// natural-language string suitable for embedding. Stable across re-runs (no
// timestamps, no random IDs) so re-embedding the same product produces the
// same vector when content doesn't change.
//
// Example output:
//
//	"100G QSFP28 LR4 Transceiver. Transceiver. 100G data rate, QSFP28 form
//	factor, 10km reach over single-mode fiber, 1310nm wavelength, LC duplex
//	connector. DDM supported. Compatible with Cisco Nexus 9000, Arista 7060X,
//	Juniper QFX5200. Operating temperature 0–70°C. Hot-pluggable. 4 lanes.
//	Supplier: Gigalight (Shenzhen, Shenzhen cluster)."
func BuildEmbeddingText(p ProductForEmbedding) string {
	var b strings.Builder

	// Product name + category.
	b.WriteString(p.Name)
	b.WriteString(". ")
	if p.NameZh != "" {
		b.WriteString(p.NameZh)
		b.WriteString(". ")
	}
	b.WriteString(strings.Title(p.Category))
	b.WriteString(". ")

	specs := parseSpecs(p.Specs)

	if v, ok := specs["data_rate"].(string); ok {
		b.WriteString(v)
		b.WriteString(" data rate, ")
	}
	if v, ok := specs["form_factor"].(string); ok {
		b.WriteString(v)
		b.WriteString(" form factor, ")
	}
	// reach_km can be int or float
	if reach, ok := numeric(specs["reach_km"]); ok {
		fmt.Fprintf(&b, "%gkm reach", reach)
		if fiber, ok := specs["fiber_type"].(string); ok {
			fmt.Fprintf(&b, " over %s fiber", fiber)
		}
		b.WriteString(", ")
	}
	if wl, ok := wavelengthString(specs["wavelength_nm"]); ok {
		fmt.Fprintf(&b, "%s wavelength, ", wl)
	}
	if conn, ok := specs["connector"].(string); ok {
		fmt.Fprintf(&b, "%s connector. ", conn)
	}
	if ddm, ok := specs["ddm"].(bool); ok && ddm {
		b.WriteString("DDM supported. ")
	}
	if lanes, ok := numeric(specs["lanes"]); ok {
		fmt.Fprintf(&b, "%g lanes. ", lanes)
	}
	if compat, ok := stringSlice(specs["compatibility"]); ok && len(compat) > 0 {
		fmt.Fprintf(&b, "Compatible with %s. ", strings.Join(compat, ", "))
	}
	if tmp, ok := tempRange(specs["operating_temp_c"]); ok {
		fmt.Fprintf(&b, "Operating temperature %s. ", tmp)
	}
	if hot, ok := specs["hot_pluggable"].(bool); ok && hot {
		b.WriteString("Hot-pluggable. ")
	}
	if bidi, ok := specs["bidi"].(bool); ok && bidi {
		b.WriteString("Bidirectional single-fiber. ")
	}
	if pon, ok := specs["pon"].(string); ok {
		fmt.Fprintf(&b, "%s PON. ", pon)
	}
	if industrial, ok := specs["industrial_grade"].(bool); ok && industrial {
		b.WriteString("Industrial-grade. ")
	}
	if tunable, ok := specs["tunable"].(bool); ok && tunable {
		b.WriteString("Tunable wavelength. ")
	}

	// Supplier context — last so the model still picks up product-side
	// salience first.
	b.WriteString("Supplier: ")
	b.WriteString(p.SupplierName)
	if p.SupplierCity != "" {
		fmt.Fprintf(&b, " (%s", p.SupplierCity)
		if p.SupplierCluster != "" && p.SupplierCluster != p.SupplierCity {
			fmt.Fprintf(&b, ", %s cluster", p.SupplierCluster)
		}
		b.WriteString(")")
	}
	b.WriteString(".")
	return b.String()
}

// ----- spec helpers -----

func parseSpecs(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

// numeric converts the various number shapes that come out of jsonb into a
// float64. Returns false if the field is missing or non-numeric.
func numeric(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}

func stringSlice(v any) ([]string, bool) {
	xs, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(xs))
	for _, e := range xs {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out, true
}

// wavelengthString handles both scalar (1310) and array ([1271,1291,1311,1331]
// for CWDM/LWDM lanes) forms, plus the special "tunable_C_band" string.
func wavelengthString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return strings.ReplaceAll(x, "_", " "), true
	case float64:
		return fmt.Sprintf("%gnm", x), true
	case int:
		return fmt.Sprintf("%dnm", x), true
	case []any:
		nums := make([]string, 0, len(x))
		for _, e := range x {
			if f, ok := numeric(e); ok {
				nums = append(nums, fmt.Sprintf("%g", f))
			}
		}
		if len(nums) == 0 {
			return "", false
		}
		sort.Strings(nums)
		return fmt.Sprintf("%snm lanes", strings.Join(nums, "/")), true
	}
	return "", false
}

func tempRange(v any) (string, bool) {
	pair, ok := v.([]any)
	if !ok || len(pair) != 2 {
		return "", false
	}
	lo, lok := numeric(pair[0])
	hi, hok := numeric(pair[1])
	if !lok || !hok {
		return "", false
	}
	return fmt.Sprintf("%g–%g°C", lo, hi), true
}

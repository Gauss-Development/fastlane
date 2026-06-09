package anthropic

import (
	"context"
	"fmt"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
)

const specSystemPrompt = `You are a sourcing assistant for fiber-optic transceivers (SFP, SFP+, QSFP, QSFP28).
Extract structured search specs from the engineer's natural-language request.
Only fill a field if the request clearly implies it; otherwise leave it empty/zero.
Never invent specs that were not stated or strongly implied.`

// specTool is the forced tool schema. Every field is required so the response
// always carries all keys (empty/zero where unknown), which keeps parsing
// branch-free.
var specTool = Tool{
	Name:        "extract_transceiver_specs",
	Description: "Extract structured fiber-optic transceiver specs from a buyer's natural-language query.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"data_rate":     map[string]any{"type": "string", "description": "Line rate, e.g. '100G', '10G', '1G'. Empty if unstated."},
			"form_factor":   map[string]any{"type": "string", "description": "One of QSFP28, QSFP, SFP+, SFP. Empty if unstated."},
			"reach_km":      map[string]any{"type": "number", "description": "Required link distance in kilometers (e.g. 10, 0.3, 80). 0 if unstated."},
			"wavelength_nm": map[string]any{"type": "integer", "description": "Wavelength in nm, e.g. 1310, 850, 1550. 0 if unstated."},
			"compatibility": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Vendor/platform compatibility, e.g. ['Cisco Nexus']. Empty array if unstated."},
			"fiber_type":    map[string]any{"type": "string", "description": "'single-mode' or 'multi-mode'. Empty if unstated."},
			"qty_estimated": map[string]any{"type": "integer", "description": "Quantity if mentioned, else 0."},
			"free_text":     map[string]any{"type": "string", "description": "Any residual intent not captured by the other fields."},
		},
		"required": []string{"data_rate", "form_factor", "reach_km", "wavelength_nm", "compatibility", "fiber_type", "qty_estimated", "free_text"},
	},
}

// ExtractSpecs runs Claude tool-use and maps the result into the proto type.
func (c *Client) ExtractSpecs(ctx context.Context, query string) (*searchv1.ParsedSpecs, error) {
	in, err := c.ToolUse(ctx, specSystemPrompt, query, specTool)
	if err != nil {
		return nil, fmt.Errorf("extract specs: %w", err)
	}
	return &searchv1.ParsedSpecs{
		DataRate:      getString(in, "data_rate"),
		FormFactor:    getString(in, "form_factor"),
		ReachKm:       getFloat(in, "reach_km"),
		WavelengthNm:  int32(getFloat(in, "wavelength_nm")),
		Compatibility: getStringSlice(in, "compatibility"),
		FiberType:     getString(in, "fiber_type"),
		QtyEstimated:  int32(getFloat(in, "qty_estimated")),
		FreeText:      getString(in, "free_text"),
	}, nil
}

func getString(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]any, k string) float64 {
	if v, ok := m[k].(float64); ok {
		return v
	}
	return 0
}

func getStringSlice(m map[string]any, k string) []string {
	raw, ok := m[k].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

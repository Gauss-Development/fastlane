package anthropic

import (
	"context"
	"fmt"
	"strings"

	"search-service/internal/domain"
)

const explainSystemPrompt = `You explain why a fiber-optic transceiver matches an engineer's request.
Reply with ONE sentence, under 25 words, concrete and specific (cite the spec that matches).
No preamble, no markdown, no quotes.`

// Explain returns a one-line "why this matches" rationale for a single hit.
// Best-effort: the orchestrator treats an error/empty string as "no rationale".
func (c *Client) Explain(ctx context.Context, query string, hit domain.CatalogHit) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Engineer's request:\n%s\n\n", query)
	fmt.Fprintf(&b, "Candidate part: %s — %s (%s)\n", hit.SKU, hit.Name, hit.Category)
	if len(hit.SpecsJSON) > 0 {
		fmt.Fprintf(&b, "Specs: %s\n", string(hit.SpecsJSON))
	}
	fmt.Fprintf(&b, "Supplier: %s, %s\n", hit.SupplierName, hit.SupplierCity)

	text, err := c.Text(ctx, explainSystemPrompt, b.String(), 80)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

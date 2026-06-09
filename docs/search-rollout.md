# Fiberlane Hybrid Search Rollout

This document describes the current Fiberlane MVP search path. The old
microblog/OpenSearch/Kafka search has been replaced by hybrid AI product search:

```
Next.js buyer UI
  → POST /api/bff/search
  → api-gateway POST /api/v1/search
  → search-service gRPC Search
  → Claude spec extraction + query embedding + pgvector catalog ranking
```

## Runtime architecture

1. **Buyer input** — the dashboard (`/dashboard`) and results page (`/search?q=...`)
   submit a natural-language photonics query.
2. **BFF** — `frontend/src/app/api/bff/search/route.ts` accepts:

   ```json
   {
     "query": "100G QSFP28 LR4 Cisco compatible",
     "limit": 20,
     "spec_overrides": null
   }
   ```

   It forwards the authenticated request to the gateway.
3. **Gateway** — `POST /api/v1/search` maps the JSON body into
   `search.v1.SearchRequest`.
4. **Search service** pipeline:
   - Claude tool-use extracts structured transceiver specs.
   - Query embeddings use Voyage `voyage-3` first, OpenAI fallback if configured.
   - `products.embedding <=> query_embedding` ranks catalog candidates in pgvector.
   - Extracted specs boost/re-rank the vector candidates.
   - Claude generates one-line match explanations for the top hits.
5. **Frontend results** — response contains `parsed_specs`, `results[]`, and
   `query_id`. The results page renders editable spec chips; removing/adding a
   chip re-runs search via `spec_overrides`.

## Required services

- `postgres_post` — owns suppliers/products/embeddings.
- `redis` — optional cache for spec extraction, embeddings, and explanations.
- `search-service` — gRPC on `:50054`.
- `api-gateway` — HTTP edge gateway.
- `frontend` — Next.js app.

OpenSearch and Kafka are not required for the Fiberlane search MVP.

## Environment variables

### search-service

- `CATALOG_DATABASE_URL` — read-only catalog DB connection. In compose this points
  to `postgres_post`.
- `REDIS_URL` — optional cache.
- `ANTHROPIC_API_KEY` — enables Claude spec extraction and match explanations.
- `ANTHROPIC_MODEL` — defaults to `claude-haiku-4-5`.
- `VOYAGE_API_KEY` — primary embedding provider.
- `OPENAI_API_KEY` — embedding fallback.

### api-gateway

- `SEARCH_SERVICE_GRPC_ADDR` — usually `search-service:50054` in compose or
  `localhost:50054` for local hybrid runs.

### frontend

- `BACKEND_API_URL` — points the BFF at api-gateway.

## Local setup

From the repo root:

```bash
make infra-up
make seed
make embed-fake
```

`make embed-fake` is sufficient for UI smoke testing because it populates the
pgvector column deterministically. It does not prove semantic ranking quality.
Use `make embed` with `VOYAGE_API_KEY` or `OPENAI_API_KEY` for real ranking.

For a hybrid local run:

```bash
cd services/search-service
CATALOG_DATABASE_URL="postgres://postgres:${POSTGRES_POST_PASSWORD}@localhost:${POSTGRES_POST_HOST_PORT:-15432}/postdb?sslmode=disable" \
REDIS_URL=localhost:6379 \
go run .
```

Then run the gateway with `SEARCH_SERVICE_GRPC_ADDR=localhost:50054`, and run
the frontend.

## Degraded mode

The service is intentionally usable without paid AI keys:

- Missing embedding key: search-service uses the deterministic fake embedder.
- Missing Anthropic key: spec extraction and match explanations are skipped.
- Missing Redis: the pipeline runs with a no-op cache.

This mode is for local development and CI only. It verifies that the UI and
transport path work, but it is not evidence of search relevance.

## Smoke test checklist

1. Register or log in as a buyer.
2. Confirm authenticated users land at `/dashboard`.
3. Submit `100G QSFP28 LR4 Cisco compatible`.
4. Confirm navigation to `/search?q=100G...`.
5. Confirm `/api/bff/search` sends a `POST` request.
6. Confirm api-gateway receives `POST /api/v1/search`.
7. Confirm the results page renders:
   - original query prompt,
   - parsed spec chips when Anthropic is enabled,
   - ranked product rows,
   - match explanations when Anthropic is enabled.
8. Remove or add a chip and confirm the next request includes
   `spec_overrides`.
9. Open the Quote modal and confirm it is UI-only until GAU-249 implements RFQ
   persistence and supplier magic-link email.

## Relevance validation

GAU-245 remains In Progress until it is validated with real keys and the seeded
catalog. GAU-258 owns the full 20-query bank. Minimum live validation before
marking search Done:

- real embeddings generated for all seeded products,
- real Claude spec extraction enabled,
- top-5 includes an appropriate part for the representative query set,
- p95 end-to-end latency below 4 seconds.

## Rollback

- Frontend: redirect `/dashboard` back to the previous app surface or hide the
  search results route.
- Gateway: disable/protect `POST /api/v1/search` or point
  `SEARCH_SERVICE_GRPC_ADDR` back to a known-good search-service version.
- Search service: stop the service; gateway returns search errors while other
  buyer app surfaces remain available.

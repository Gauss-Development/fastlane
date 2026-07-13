# Design Brief — Fiberlane / Fastlane Cross-Border Sourcing Platform

> **Purpose of this document.** A self-contained specification of the product's
> visual language, components, screens, and flows, precise enough that Claude
> (or a designer) can **rebuild the design from scratch and have it match.** It
> documents the design *system* verbatim (tokens, component recipes) and every
> screen as it exists today, then flags the in-progress pivot so new screens are
> built in the same language.

---

## 0. Read this first — the pivot

The repo was forked from a photonics-transceiver sourcing template and is **mid-pivot** to a **custom PCB / PCBA sourcing platform** (US hardware startups ↔ verified Chinese manufacturers for custom board design & assembly).

| Layer | State today |
|---|---|
| **Design system** (tokens, components, `RouteIndicator`) | Stable, vertical-agnostic. **Reuse verbatim.** |
| **Visible screens** (landing, search, RFQ, suppliers, orders) | Still the **photonics** surface — "Fiberlane", transceivers, QSFP28, optical specs. |
| **Backend + BFF** | Already **pivoted**: `design-service` with Projects → DesignFiles → NDA → manufacturer access. Roles are now `startup` / `manufacturer` / `admin`. |
| **Project / file / NDA UI** | **Does not exist yet.** BFF routes are wired (`/api/bff/projects/*`, `/api/bff/files/*`) but there are no pages consuming them. |

**Implication for a rebuild:** keep the industrial cross-border *look* exactly. The vertical vocabulary (transceivers → PCB/PCBA), the actors (buyer → **startup**, supplier → **manufacturer**), and the core object (RFQ/search → **Project + design files + NDA**) are what change. §10 gives the adaptation map. If the goal is a pixel-faithful clone of what's on screen now, build §6–§8 as written. If the goal is the product's real direction, build §6–§8's *shells* but swap the domain per §10.

---

## 1. Positioning & voice

- **What it is:** a cross-border B2B sourcing marketplace. A US buyer describes what they need; the platform matches verified Chinese factories and runs the quote → order flow. Trust and provenance are the whole pitch.
- **Signature idea:** **cross-border**. Shenzhen → San Francisco is the brand. It appears as a route mark on essentially every screen.
- **Tone:** industrial, technical, terse. Reads like a Bloomberg terminal or an EDA tool, not a consumer SaaS. Part numbers, coordinates, IDs, and prices are first-class citizens rendered in monospace.
- **Trust pillars** (recur in copy): **VERIFIED** (factories audited on-site), **ESCROWED** (funds held), **INSPECTED** (third-party QC).
- **Bilingual where the supplier/manufacturer is Chinese** — English label over 中文 label (see §4 bilingual pattern). Used on the magic-link response page.

---

## 2. Design principles (non-negotiable)

1. **No gradients. No shadows** except one modal backdrop (`black/60`). Depth comes from **borders**, not elevation.
2. **No emojis, no marketing illustrations, no stock imagery.** Structure and type carry the design.
3. **Monospace headlines.** Headings are `JetBrains Mono`, not a display sans. This is the single most distinctive choice — keep it.
4. **Sharp corners.** Base radius **4px**, never above ~8px. No pill-shaped buttons.
5. **Restraint in scale.** Largest headline is **32px** (`h1`), not a 60px hero. Density over drama.
6. **Uppercase + letter-spacing** for all labels, badges, IDs, nav. Tracking `0.02em`–`0.16em`.
7. **Tabular figures** for every number (prices, quantities, coordinates, lead times).
8. **The route mark is the seal.** `SHENZHEN ─────► SAN FRANCISCO` (or parametric endpoints) appears on most screens.
9. **Two themes, systematically inverted.** Dark is the default. No hue drift between them; the orange simply lifts a notch in dark mode.

---

## 3. Design tokens (verbatim)

Defined in `frontend/src/app/globals.css` via Tailwind 4 `@theme inline` with a `.dark` custom variant. Copy these exactly.

### 3.1 Colors

| Token | Light | Dark | Usage |
|---|---|---|---|
| `--background` | `#f7f5f0` | `#0d0d0e` | Page background (off-white linen / deep ink) |
| `--foreground` | `#0a0a0a` | `#ededeb` | Primary text |
| `--card` | `#ffffff` | `#141416` | Surface backgrounds |
| `--card-foreground` | `#0a0a0a` | `#ededeb` | Text on surfaces |
| `--primary` | `#d54e20` | `#e25a2c` | **International orange** — primary action/accent |
| `--primary-foreground` | `#ffffff` | `#0d0d0e` | Text on primary |
| `--secondary` | `#ffffff` | `#1c1c1f` | Secondary button bg |
| `--secondary-foreground` | `#0a0a0a` | `#ededeb` | |
| `--muted` | `#efece5` | `#1c1c1f` | Muted fills (badges, chips) |
| `--muted-foreground` | `#5b5950` | `#8a8980` | Labels, secondary text |
| `--accent` | `#efece5` | `#222226` | Hover fills |
| `--accent-foreground` | `#0a0a0a` | `#ededeb` | |
| `--destructive` | `#c8312d` | `#d8403b` | Errors / delete |
| `--destructive-foreground` | `#ffffff` | `#ffffff` | |
| `--success` | `#3f6f3f` | `#5a9a5a` | OK / verified / accepted |
| `--success-foreground` | `#ffffff` | `#0d0d0e` | |
| `--warning` | `#c28a2c` | `#d4a04a` | Caution / pending |
| `--warning-foreground` | `#ffffff` | `#0d0d0e` | |
| `--border` | `#d8d4cc` | `#2a2a2d` | Default border (concrete / steel) |
| `--border-strong` | `#262626` | `#4a4a4d` | Prominent borders, route connector line/arrow |
| `--input` | `#d8d4cc` | `#2a2a2d` | Input border |
| `--input-background` | `#ffffff` | `#141416` | Input fill |
| `--ring` | `#d54e20` | `#e25a2c` | Focus ring (== primary) |
| `--marker-cn` | `#c8312d` | `#d8403b` | **China marker** (origin dot, `RFQ-` prefix) |
| `--marker-us` | `#1a2b4a` | `#4a6fa5` | **US marker** (destination dot, `ORD-` prefix, `info` tone) |
| `--chart-1..5` | orange / navy / red / green / gold | (lifted) | Data viz |
| `--sidebar*` | mirrors card/primary/accent/border | | Sidebar surfaces |

**Reserved semantics:** `--marker-cn` and `--marker-us` are *only* for cross-border provenance (route dots, ID prefixes, origin/destination). Don't use them as generic accents.

### 3.2 Typography

- **Fonts loaded:** `@fontsource/jetbrains-mono` (400/500/700), `@fontsource/space-grotesk` (imported, largely unused), plus Google `Inter` (400/500/600) and `JetBrains Mono`.
- `--font-sans`: `"Inter", "Avenir Next", "Segoe UI", sans-serif`
- `--font-mono` / `--font-display`: `"JetBrains Mono", "SFMono-Regular", monospace`

| Role | Font | Size | Weight | Tracking | Treatment |
|---|---|---|---|---|---|
| `h1` | JetBrains Mono | 2rem (32px) | 500 | -0.01em | line-height 1.2 |
| `h2` | JetBrains Mono | 1.5rem (24px) | 500 | -0.01em | |
| `h3` | JetBrains Mono | 1.125rem (18px) | 500 | -0.01em | |
| `h4` | JetBrains Mono | 1rem (16px) | 500 | -0.01em | |
| Body | Inter | 15px | 400 | — | |
| **Label** | Inter | 0.8125rem (13px) | 500 | 0.02em | **UPPERCASE**, `--muted-foreground` |
| Code / IDs / numbers | JetBrains Mono | xs–base | — | 0.04em–0.1em | uppercase, `tabular-nums` |

### 3.3 Radius & spacing

- `--radius: 0.25rem` (4px). `sm` ≈ 0, `md` = 4px, `lg` = 6px, `xl` = 8px. **That's the ceiling.**
- Tailwind default spacing scale. Conventions: cards `px-5 py-4`; inputs `px-3 h-9`; buttons `h-9` (default) / `h-8` (sm) / `h-11` (lg); **table rows/headers `h-10` (40px, dense).**
- Borders: 1px everywhere. Focus: `2px solid var(--ring)` + `1px` offset.

### 3.4 Texture

Light theme only: an embedded SVG fractal-noise **paper grain** at ~4.5% opacity (asset-free data URI). Dark theme stays flat.

---

## 4. Component library

All under `frontend/src/components/ui/`. Recipes are the actual class strings — reproduce them.

**Button** (`button.tsx`) — `cva`, base `inline-flex items-center justify-center gap-2 rounded-sm text-sm font-medium transition-colors`, focus `ring-2 ring-ring ring-offset-2`, disabled `opacity-50 pointer-events-none`, icons `size-4`.
- variants: `default` (`bg-primary text-primary-foreground hover:bg-primary/90`), `secondary` (`bg-secondary border border-border hover:bg-accent`), `outline` (`border border-border bg-transparent hover:bg-accent`), `ghost` (`hover:bg-accent`), `destructive` (`bg-destructive text-destructive-foreground hover:bg-destructive/90`).
- sizes: `default h-9 px-4` · `sm h-8 px-3 text-xs` · `lg h-11 px-6 text-base` · `icon h-9 w-9`.

**Card** (`card.tsx`) — `rounded-md border border-border bg-card text-card-foreground`. `CardHeader px-5 py-4 border-b border-border`, `CardTitle` = `h3 text-base`, `CardDescription` = `p text-sm text-muted-foreground`, `CardContent px-5 py-4`, `CardFooter px-5 py-4 border-t border-border`.

**Input** (`input.tsx`) — `h-9 w-full px-3`, 1px `--input` border, 4px radius, 15px, `placeholder:text-muted-foreground`, focus swaps border to ring + 2px ring. **PasswordInput** adds a right-aligned eye/eye-off toggle (`pr-10`).

**Label** (`label.tsx`) — 13px, 500, `0.02em`, **UPPERCASE**, `--muted-foreground`.

**Badge** (`badge.tsx`) — `inline-flex items-center gap-1 rounded-sm border px-2 py-0.5 text-xs font-medium uppercase tracking-[0.08em]`. Tinted variants use 10% fill + 40% border: `primary` `bg-primary/10 border-primary/40 text-primary`; same shape for `success` / `warning` / `destructive`; `default` = `bg-muted`; `outline` = transparent.

**StatusPill** (`pill.tsx`) — `inline-flex items-center gap-1.5 rounded-sm border border-border bg-card px-2 py-0.5 font-mono text-xs uppercase tracking-[0.1em]` + a `size-1.5 rounded-full` **dot** colored by tone: `neutral`→muted-fg, `success`→success, `warning`→warning, `destructive`→destructive, `info`→marker-us.

**CodeId** (`code-id.tsx`) — renders system IDs like `RFQ-20260429-0142-SZX`. `font-mono uppercase tracking-[0.04em] tabular-nums`. **Prefix colored by domain:** `RFQ`→primary, `ORD`→marker-us, `SUP`→success, `QUOTE`→warning. Sizes `sm/md/lg` = `text-xs/sm/base`. Optional `copyable` (Copy→Check icon).

**RouteIndicator** (`route-indicator.tsx`) — **the brand mark.** Renders `ORIGIN ──────►  DESTINATION` with a marker dot per endpoint and optional coordinate subtitle. Defaults `SHENZHEN (22.54°N 114.06°E)` → `SAN FRANCISCO (37.77°N 122.42°W)`.
- Endpoint: city `font-medium uppercase font-mono leading-none` + `size-1.5` dot (cn / us colored) + optional coords `text-[0.78em] normal-case text-muted-foreground`.
- Connector: `h-px bg-border-strong` + `►` arrow in `--border-strong`. Widths sm `w-6` / md `w-10` / lg `w-20`.
- Sizes sm `text-[11px] tracking-[0.1em]` / md `text-xs tracking-[0.12em]` / lg `text-base tracking-[0.16em]`. Coords default on only for `lg`.

**Modal** (`modal.tsx`) — `fixed inset-0 z-50 flex items-center justify-center p-4`; backdrop `absolute inset-0 bg-black/60`; box `w-full max-w-lg rounded-md border border-border bg-card`; header `border-b border-border px-5 py-4` (title `h2 text-base`, desc `text-sm text-muted-foreground`); body `px-5 py-4`. Closes on Escape + backdrop click; locks body scroll.

**Table** (`table.tsx`) — "Bloomberg-terminal" density. `w-full rounded-md border border-border`. Rows `h-10 border-b border-border hover:bg-muted/60` (no zebra). `Th` sticky, `h-10 px-3 bg-card text-xs font-medium uppercase tracking-[0.08em] text-muted-foreground`; sortable shows a chevron (`ChevronUp/Down/ChevronsUpDown`, hidden until hover when unsorted); numeric `text-right`. `Td h-10 px-3`; numeric `text-right font-mono tabular-nums`.

**Bilingual pattern** (magic-link page) — `BiLabel`: English (small mono uppercase) stacked over 中文 (smaller, muted). `BiValueRow`: `EN / 中文` key with value below. Use wherever a Chinese counterpart reads the screen.

---

## 5. Layout & navigation

**Root layout** (`app/layout.tsx`): `ThemeScript` (hydration-safe) → `AppProviders` (TanStack Query + devtools) → `ThemeProvider`. Auth state in a Zustand `auth-store` (`accessToken`, `user`, …). Dark theme default.

**Protected shell** (`(protected)/app/layout.tsx`): `SessionBootstrap` (calls `refreshSessionOnce()`, spinner while refreshing) → `AppShell` = `AppSidebar` + `<main class="flex-1">`.

**AppSidebar** (`app-sidebar.tsx`): collapsible **240px ↔ 56px**. Logo `FIBERLANE` (mono, bold) → `/dashboard`. Numbered nav:
`01 Search → /dashboard` · `02 RFQs → /rfqs` · `03 Orders → /orders` · `04 Suppliers → /suppliers` · `05 Settings → /app/profile`.
Active item: left-border highlight + accent bg. Footer: user email (when open) + Sign out (red hover).

**Auth shell** (`auth-shell.tsx`): centered `Card max-w-sm` on dark bg, header logo (`MICROBLOG` — legacy, rebrand to Fiberlane), footer version string.

**Route protection:** `middleware.ts` guards protected paths on the refresh-token cookie; missing → `/auth/login?next=<path>` (open-redirect-guarded).

---

## 6. Screen inventory

Route groups in `()` don't appear in the URL. `[x]` = dynamic segment.

| Path | Access | Purpose |
|---|---|---|
| `/` | public | Landing (redirects authed → `/dashboard`) |
| `/auth/login` `/auth/register` `/auth/callback` | public | Email+password / Google OAuth / OAuth code exchange |
| `/q/[token]` | public (magic link) | Supplier/manufacturer quote response — no account |
| `/dashboard` | protected | Search workbench (default post-login) |
| `/search?q=` | protected | Search results + editable spec chips + quote modal |
| `/rfqs` · `/rfqs/[id]` | protected | RFQ list · RFQ detail with polling quotes table |
| `/orders` | protected | Order tracking table (placeholder data) |
| `/suppliers` | protected | Verified supplier directory (card grid) |
| `/products/[id]` | protected | Product detail (stub) |
| `/app/profile` | protected | Account + theme preference |
| `/app/post/[id]` | protected | Legacy blog-post detail (template leftover) |
| `/dev` | public | Token/primitive showcase (internal) |

---

## 7. Screen specs

### 7.1 Landing (`/`)
Ported from a Figma design; monospace, off-white, orange, zero imagery.
- **TopBar** (sticky): `FIBERLANE` + `CN ───► US` route mark; nav `Suppliers · How it Works · Pricing`; sign-in button.
- **Hero:** mono headline *"Photonics components, sourced direct from verified Chinese factories"*; search input; CTAs *Find parts* / *Browse suppliers*; a live RFQ ticker rotating every ~3.5s.
- **Route section:** Shenzhen → US lanes (SFO/LAX/DFW/ORD/JFK); live counter *"2847+ components shipped"* (+~4.2s).
- **Pillars (3 cards):** VERIFIED (247 audited) · ESCROWED ($14.2M held) · INSPECTED (1,847 inspections).
- **Featured suppliers (4-col):** code `SUP-CN-SZX-0024`, English + 中文 name, city, capability tags, verified date.
- **Timeline (6 stages):** RFQ → QUOTE → PAYMENT → PRODUCTION → QC → DELIVERY, with CN/US side markers and day offsets (Day 0 → Day 14).
- **Footer (dark):** office locations (Shenzhen & SF with coords), nav, legal.

### 7.2 Auth (`/auth/*`)
Login: email, password (`PasswordInput`), *Sign In* (spinner), *Continue with Google*, error banner (motion fade), link to register. Register adds a Name field. Callback: exchanges code, validates CSRF from sessionStorage, redirects to `?next` (guarded) or `/app`.

### 7.3 Dashboard (`/dashboard`)
Header: `BUYER • <domain> • USD • PST` + `RouteIndicator`.
- **Hybrid search card** (Section A): big mono input with `>` prefix, placeholder *"Describe the part you need…"*, *Search catalog* button, example-query chips (e.g. *100G QSFP28 LR4 Cisco compatible*).
- **Recent RFQs** (table): RFQ (`CodeId`) · Request (mono, truncated) · Qty · Status (`StatusPill`) · Age. Rows → `/rfqs/[id]`.
- **Live quote feed** (≈360px sidebar): supplier, city, RFQ id, part, price, lead time, timestamp.
- **Operating snapshot** (2×2 KPIs): verified suppliers (7) · seeded SKUs (80+) · median lead time (10d) · avg on-time (95.9%).
- **Featured suppliers** carousel → *View supplier surface →* `/suppliers`.
- *Currently demo data.*

### 7.4 Search results (`/search?q=`)
- **Header:** `> <query>` (mono, tinted `>`), small `RouteIndicator`, query id, a re-search input.
- **Spec chips card:** AI-extracted specs as removable chips `label: value ×`. Fields: `data_rate, form_factor, reach_km, wavelength_nm, compatibility[], fiber_type, qty_estimated, free_text`. Add-filter row (field dropdown + value + *+ add filter*). Status line: *AI extracted specs active* vs *Manual spec override active*; *Reset AI specs*.
- **Results:** count *"{n} ranked matches"*; states = 4 pulsing skeletons / empty (*"No products found. Remove a chip or broaden the query."*) / error card / list.
- **`ProductResultRow`:** SKU (mono) + *Verified* pill + `SCORE {n}`; name + 中文; spec line (form factor, wavelength, reach, fiber, connector); compatibility line; italic match explanation; supplier name/city + stock + lead-time pills; price (mono, XL, tabular) + MOQ; *Quote →*.
- **Quote modal:** title *Request quote*; copy *"The supplier receives a magic-link email and responds without an account. You are emailed as quotes arrive."*; product summary; fields Quantity (default `max(moq,100)`), Target date, Shipping address, Notes; *Request Quote* → `createRFQ` → redirect `/rfqs/[id]`.
- **Interaction:** editing/removing a chip or adding a filter **re-runs the search** (React Query refetch keyed on query + spec overrides).

### 7.5 RFQ list & detail (`/rfqs`, `/rfqs/[id]`)
- **List:** table RFQ · Request · Qty · Status · Age; empty *"No RFQs yet. Run a search and hit Quote → on a result."* Status tones: open→info, quoted→warning, accepted→success, closed→neutral. Age `{m|h|d}`.
- **Detail:** header `CodeId` (copyable, lg) + `RouteIndicator`. **Overview card:** `> {query_text}` + status pill; grid of Quantity / Target date / Ship to / Created / extracted specs (chips) / Notes. **Quotes card:** *"Suppliers respond through their magic link… refreshes automatically"*; **polls every 30s** while open; table Quote(`CodeId`) · Status · Unit price · Lead time · Valid until · Notes (quote tones: pending→neutral, submitted→info, accepted→success, rejected→destructive); empty *"No supplier responses yet. Suppliers are in CST (UTC+8); first quotes typically arrive within 24 hours."*

### 7.6 Supplier magic-link response (`/q/[token]`) — public, bilingual
Header `FIBERLANE` + `CN ───► US`. **RFQ summary** in `BiLabel`/`BiValueRow`: heading *"NEW RFQ FROM {company} (United States)"* + `CodeId`; rows Part / 询盘部件, Quantity / 数量, Delivery To / 交货地点, Target Date / 目标日期, Notes / 备注.
- **Form:** *YOUR QUOTE / 您的报价* — Unit price USD (req), Lead time days (req), Valid until (opt), Notes (opt); submit *Submit Quote / 提交报价* (full-width, mono, bold, uppercase); help *"No login required. This link is unique to {supplier}."*
- **Success:** pill *Submitted / 已提交*, *"QUOTE RECEIVED — THANK YOU / 报价已收到，谢谢"*, *"The buyer has been notified by email. / 买家已收到邮件通知。"* + quote `CodeId`.
- **Invalid/expired token:** *"THIS LINK IS INVALID OR HAS EXPIRED / 此链接无效或已过期"*.

### 7.7 Orders / Suppliers / Product / Profile
- **Orders:** table Order(`CodeId`) · Part · Status · Total (right) · Location (mono muted); demo rows (`ORD-…-SFO` shipped/qc passed, `-LAX` in production); tones shipped→info, qc passed→success, in production→warning. Placeholder pending order-timeline work.
- **Suppliers:** card grid (3-col xl / 2 md); per card `CodeId` + *Verified* pill, name, 中文 + city (mono muted), capability (uppercase mono), On-time % (green) + Orders count.
- **Product detail:** stub — title + product id + `RouteIndicator` + note that full spec table/datasheet/sticky quote panel is pending; *Back to search*.
- **Profile:** sticky *Profile* bar; Account card (Email/Name/User ID mono); Appearance card with Dark/Light/System buttons (current highlighted).

---

## 8. Core flows

1. **Search → RFQ:** dashboard/search input → results with editable chips → *Quote →* opens modal → `createRFQ` → `/rfqs/[id]` → detail polls quotes every 30s.
2. **Supplier quote (no login):** magic-link email → `/q/[token]` → loads RFQ (token is the only credential) → submit price + lead time → success card. Buyer notified by email.
3. **Auth:** password or Google OAuth → callback exchanges code → cookies + Zustand session → middleware gates protected routes.

---

## 9. Data shapes driving the UI

**Photonics surface (current frontend types):**
- `ParsedSpecs { data_rate?, form_factor?, reach_km?, wavelength_nm?, compatibility?[], fiber_type?, qty_estimated?, free_text? }`
- `ProductHit { id, sku, name, name_zh, supplier_name, supplier_city, supplier_verified, price_usd, moq, stock_qty, lead_time_days, match_score, match_explanation, specs }`
- `RFQ { id, query_text, parsed_specs, status(open|quoted|accepted|closed), qty, target_date, shipping_address, notes, created_at, buyer_company, matched_product_ids }`
- `Quote { id, supplier_id, price_usd, lead_time_days, validity_date, supplier_notes, status(pending|submitted|accepted|rejected), submitted_at }`

**PCB/PCBA surface (backend + BFF, no UI yet):**
- `Project { id: PRJ-YYYYMMDD-NNNN, owner_id, title, description, category(pcb|pcba|cable_assembly|enclosure|other), status(draft|active|archived), owner_email, owner_company, created_at }`
- `DesignFile { id: FILE-…, project_id, kind(gerber|bom|assembly_drawing|pick_place|datasheet|nda|other), filename, version, content_sha256, object_key, size_bytes, content_type, uploaded_by, status(pending|committed) }`
- `NDA { id, project_id, manufacturer_id, status(pending|accepted), nda_version, accepted_ip, accepted_at }`
- `User { id, email, name, picture, bio, location, website, role(startup|manufacturer|admin), company, is_active }`
- BFF: `GET/POST /api/bff/projects`, `GET /api/bff/projects/:id`, `POST …/files/upload-url`, `GET …/files`, `POST /api/bff/files/:id/confirm`, `GET /api/bff/files/:id/download-url`, `GET …/nda`, `POST …/nda/accept`.
- **Rule enforced server-side:** a manufacturer can list/download a project's files only if they own it **or** have an accepted NDA. Files never transit the service — S3 + short-TTL presigned URLs only.

---

## 10. Pivot adaptation map (for a forward-looking rebuild)

Keep §2–§5 unchanged. Swap vocabulary and objects:

| Photonics (on screen now) | PCB/PCBA (real direction) |
|---|---|
| Fiberlane / transceivers / QSFP28 | custom PCB, PCBA, cable assembly, enclosure |
| Buyer | **Startup** (project owner) |
| Supplier | **Manufacturer** |
| Search a catalog → `ProductHit` | Create a **Project** → upload **design files** |
| RFQ (text query + specs) | Project + attached files (Gerber, BOM, assembly, pick-and-place, datasheet) |
| `RFQ-…` id prefix | `PRJ-…`, `FILE-…` (extend `CodeId` prefix map + add a marker color) |
| Quote via magic link | NDA acceptance → then quote; magic-link infra still applies |
| `SHENZHEN ─────► SAN FRANCISCO` | **unchanged** — still the seal |

**New screens to design in the same language** (none exist yet):
- **Projects list** — table like `/rfqs`: `PRJ-…` id, title, category badge, status pill (draft/active/archived), file count, updated age; empty-state CTA *New project*.
- **Project detail** — overview card (title, category, description, owner company) + **files table** (kind badge, filename mono, version, size tabular, status pill pending/committed, download) + **NDA panel** (accept-to-unlock gate for manufacturers; owner sees invite/manage).
- **New-project + file upload** — form (title, description, category select) then a **drag-drop uploader** driving the presigned-URL flow (request upload-url → PUT to S3 → confirm). Show per-file states pending → committed.
- **NDA gate** — a manufacturer viewing a project they haven't signed sees a locked files table + *Accept NDA to view design files* (records IP + timestamp; captures `nda_version`).
- **Manufacturer role** in nav/profile; keep bilingual pattern (§4) on any manufacturer-facing surface.

Reuse: `Card`, `Table`, `StatusPill`, `Badge` (category/kind), `CodeId` (add `PRJ`/`FILE`), `RouteIndicator`, `Modal` (NDA + upload), the drag-drop should degrade to a normal `<input type="file">`.

---

## 11. Rebuild checklist / build order

1. **Tokens** — port §3 into `globals.css` (`@theme inline` + `.dark`), fonts (JetBrains Mono + Inter), paper-grain on light. Verify both themes via a `/dev` page.
2. **Primitives** — §4, in order: Button, Card, Input/PasswordInput, Label, Badge, StatusPill, CodeId, RouteIndicator, Modal, Table.
3. **Shells** — root providers + theme, protected `AppShell`/`AppSidebar` (collapsible, numbered nav), auth shell, middleware gate.
4. **Screens** — §7 in demo order: Landing → Auth → Dashboard → Search+chips → RFQ list/detail → Supplier magic-link → Suppliers/Orders/Product/Profile.
5. **Pivot (if forward-looking)** — build §10's Projects/Project detail/Upload/NDA screens with the same primitives; extend `CodeId` prefixes.

**Acceptance:** no gradients/shadows (except modal backdrop); every heading monospace; every id/number tabular-mono; a route mark on landing, search, RFQ detail, suppliers, orders, product, magic-link; radius ≤ 8px; dark default with a clean light inversion.

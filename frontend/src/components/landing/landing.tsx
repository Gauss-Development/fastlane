"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

// Industrial / cross-border landing page. Ported from the Fiberlane design
// source (Figma "Ravo"): monospace headlines, the SHENZHEN ─── SAN FRANCISCO
// route mark, off-white linen, international orange. No gradients, no imagery.

const PAPER_TEXTURE =
  "url(\"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='200' height='200'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 0  0 0 0 0 0  0 0 0 0 0  0 0 0 0.04 0'/></filter><rect width='100%' height='100%' filter='url(%23n)'/></svg>\")";

const MONO = "'JetBrains Mono', monospace";
const SANS = "'Inter', sans-serif";

function RouteMark({ className = "" }: { className?: string }) {
  return (
    <span
      className={`inline-flex items-center gap-2 tracking-wider ${className}`}
      style={{ fontFamily: MONO }}
    >
      <span style={{ color: "#C8312D" }}>CN</span>
      <span style={{ color: "#262626" }}>───</span>
      <span style={{ color: "#1A2B4A" }}>US</span>
    </span>
  );
}

function TopBar() {
  return (
    <header
      className="sticky top-0 z-50 w-full border-b"
      style={{ background: "#F7F5F0", borderColor: "#0A0A0A" }}
    >
      <div className="mx-auto flex max-w-[1400px] items-center justify-between px-8 py-4">
        <div className="flex items-center gap-4">
          <span
            className="tracking-[0.18em]"
            style={{ fontFamily: MONO, fontWeight: 700, fontSize: "16px", color: "#0A0A0A" }}
          >
            FIBERLANE
          </span>
          <span style={{ color: "#D8D4CC" }}>|</span>
          <RouteMark className="text-[11px]" />
        </div>
        <nav
          className="hidden items-center gap-8 md:flex"
          style={{ fontFamily: MONO, fontSize: "12px", letterSpacing: "0.08em" }}
        >
          <a href="#suppliers" style={{ color: "#0A0A0A" }} className="uppercase hover:opacity-60">
            Suppliers
          </a>
          <a href="#how" style={{ color: "#0A0A0A" }} className="uppercase hover:opacity-60">
            How it works
          </a>
          <a href="#pricing" style={{ color: "#0A0A0A" }} className="uppercase hover:opacity-60">
            Pricing
          </a>
        </nav>
        <Link
          href="/auth/login"
          className="px-4 py-2 uppercase tracking-wider hover:opacity-90"
          style={{ fontFamily: MONO, fontSize: "12px", fontWeight: 700, background: "#D54E20", color: "#FFFFFF" }}
        >
          Авторизоваться
        </Link>
      </div>
    </header>
  );
}

const TICKER = [
  { id: "RFQ-20260429-0142-SZX", part: "100G QSFP28 LR4", time: "18 min ago", quotes: 3 },
  { id: "RFQ-20260429-0138-SZX", part: "1550nm DFB Laser Diode", time: "32 min ago", quotes: 5 },
  { id: "ORD-20260429-0089-SFO", part: "OM4 MTP-12 Cassette", time: "47 min ago", quotes: 2 },
  { id: "RFQ-20260429-0131-SZX", part: "400G OSFP DR4", time: "1 hr ago", quotes: 4 },
  { id: "RFQ-20260429-0124-SZX", part: "10G SFP+ 1310nm 10km", time: "2 hr ago", quotes: 7 },
];

function Hero() {
  const [i, setI] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setI((n) => (n + 1) % TICKER.length), 3500);
    return () => clearInterval(id);
  }, []);
  const item = TICKER[i];

  return (
    <section className="border-b" style={{ background: "#F7F5F0", borderColor: "#262626" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-20">
        <div
          className="mb-10 flex items-center gap-3 uppercase tracking-[0.2em]"
          style={{ fontFamily: MONO, fontSize: "11px", color: "#262626" }}
        >
          <span className="inline-block h-2 w-2" style={{ background: "#3F6F3F", borderRadius: "1px" }} />
          <span>SHENZHEN → SAN FRANCISCO • LIVE</span>
        </div>

        <h1
          className="max-w-[920px]"
          style={{ fontFamily: MONO, fontSize: "40px", fontWeight: 700, lineHeight: 1.25, letterSpacing: "-0.01em", color: "#0A0A0A" }}
        >
          Photonics components,
          <br />
          sourced direct from
          <br />
          verified Chinese factories.
        </h1>

        <div className="mt-12 max-w-[920px]">
          <div className="flex items-center border" style={{ borderColor: "#262626", background: "#FFFFFF" }}>
            <span className="pl-5 pr-3" style={{ fontFamily: MONO, color: "#D54E20", fontWeight: 700 }}>
              {">"}
            </span>
            <input
              type="text"
              defaultValue="100G QSFP28 10km, Cisco compatible"
              className="flex-1 bg-transparent py-5 outline-none"
              style={{ fontFamily: MONO, fontSize: "15px", color: "#0A0A0A" }}
            />
            <span className="pr-5 uppercase tracking-wider" style={{ fontFamily: MONO, fontSize: "11px", color: "#717182" }}>
              ⌘ K
            </span>
          </div>

          <div className="mt-5 flex flex-wrap items-center gap-3">
            <Link
              href="/auth/login"
              className="px-6 py-3 uppercase tracking-wider"
              style={{ fontFamily: MONO, fontSize: "12px", fontWeight: 700, background: "#D54E20", color: "#FFFFFF" }}
            >
              Find parts →
            </Link>
            <Link
              href="/auth/login"
              className="border px-6 py-3 uppercase tracking-wider"
              style={{ fontFamily: MONO, fontSize: "12px", fontWeight: 700, borderColor: "#262626", color: "#0A0A0A", background: "transparent" }}
            >
              Browse suppliers
            </Link>
          </div>
        </div>

        <div
          className="mt-16 flex flex-wrap items-center gap-4 border px-5 py-3"
          style={{ borderColor: "#D8D4CC", background: "#FFFFFF", fontFamily: MONO, fontSize: "12px", color: "#0A0A0A" }}
        >
          <span className="flex h-2 w-2 shrink-0" style={{ background: "#D54E20", borderRadius: "1px" }} />
          <span style={{ color: "#717182" }} className="uppercase tracking-wider">LIVE</span>
          <span style={{ color: "#D8D4CC" }}>│</span>
          <span style={{ color: "#262626" }}>{item.id}</span>
          <span style={{ color: "#D8D4CC" }}>•</span>
          <span>{item.part}</span>
          <span style={{ color: "#D8D4CC" }}>•</span>
          <span style={{ color: "#717182" }}>{item.time}</span>
          <span style={{ color: "#D8D4CC" }}>•</span>
          <span style={{ color: "#3F6F3F" }}>{item.quotes} quotes received</span>
        </div>
      </div>
    </section>
  );
}

const CITIES = [
  { code: "SFO", x: 88, y: 38 },
  { code: "LAX", x: 90, y: 50 },
  { code: "DFW", x: 78, y: 56 },
  { code: "ORD", x: 76, y: 36 },
  { code: "JFK", x: 84, y: 32 },
];

function RouteSection() {
  const [count, setCount] = useState(2847);
  useEffect(() => {
    const id = setInterval(() => setCount((c) => c + 1), 4200);
    return () => clearInterval(id);
  }, []);

  return (
    <section className="border-b" style={{ background: "#F7F5F0", borderColor: "#262626" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-16">
        <div className="mb-10 flex items-end justify-between">
          <div>
            <div className="mb-3 uppercase tracking-[0.2em]" style={{ fontFamily: MONO, fontSize: "11px", color: "#717182" }}>
              Section A — The Route
            </div>
            <h2 style={{ fontFamily: MONO, fontSize: "24px", fontWeight: 700, color: "#0A0A0A" }}>
              Shenzhen ─── United States
            </h2>
          </div>
          <div className="text-right">
            <div style={{ fontFamily: MONO, fontSize: "32px", fontWeight: 700, color: "#0A0A0A" }}>
              {count.toLocaleString()}
            </div>
            <div className="uppercase tracking-wider" style={{ fontFamily: MONO, fontSize: "10px", color: "#717182" }}>
              components shipped • live
            </div>
          </div>
        </div>

        <div className="relative w-full border" style={{ borderColor: "#D8D4CC", background: "#FFFFFF", aspectRatio: "16 / 6" }}>
          <svg viewBox="0 0 100 38" preserveAspectRatio="none" className="absolute inset-0 h-full w-full">
            {[8, 18, 28].map((y) => (
              <line key={y} x1="0" y1={y} x2="100" y2={y} stroke="#D8D4CC" strokeWidth="0.05" strokeDasharray="0.4 0.4" />
            ))}
            <circle cx="10" cy="22" r="0.7" fill="#C8312D" />
            {CITIES.map((c) => (
              <g key={c.code}>
                <line x1="10" y1="22" x2={c.x} y2={c.y * 0.38} stroke="#262626" strokeWidth="0.12" />
                <circle cx={c.x} cy={c.y * 0.38} r="0.5" fill="#1A2B4A" />
              </g>
            ))}
          </svg>

          <div
            className="absolute"
            style={{ left: "10%", top: "58%", transform: "translate(-50%, 0)", fontFamily: MONO, fontSize: "10px", color: "#0A0A0A" }}
          >
            <div className="flex items-center gap-1">
              <span className="inline-block h-2 w-2" style={{ background: "#C8312D" }} />
              <span style={{ fontWeight: 700 }}>SHENZHEN — SZX</span>
            </div>
            <div style={{ color: "#717182" }}>22.54°N 114.05°E</div>
          </div>

          {CITIES.map((c) => (
            <div
              key={c.code}
              className="absolute"
              style={{ left: `${c.x}%`, top: `${c.y - 8}%`, fontFamily: MONO, fontSize: "9px", color: "#1A2B4A", fontWeight: 700 }}
            >
              {c.code}
            </div>
          ))}
        </div>

        <div className="mt-6 grid grid-cols-2 gap-px border md:grid-cols-4" style={{ borderColor: "#D8D4CC", background: "#D8D4CC" }}>
          {[
            { l: "ORIGIN", v: "Shenzhen, CN" },
            { l: "AVG TRANSIT", v: "11.4 days" },
            { l: "ON-TIME RATE", v: "94.7%" },
            { l: "ACTIVE LANES", v: "37" },
          ].map((s) => (
            <div key={s.l} className="px-5 py-4" style={{ background: "#FFFFFF" }}>
              <div className="uppercase tracking-wider" style={{ fontFamily: MONO, fontSize: "10px", color: "#717182" }}>
                {s.l}
              </div>
              <div className="mt-1" style={{ fontFamily: MONO, fontSize: "16px", fontWeight: 700, color: "#0A0A0A" }}>
                {s.v}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

const PILLARS = [
  { n: "01", title: "VERIFIED", body: "Every supplier audited on-site by our team in Shenzhen.", stat: "247 factories audited" },
  { n: "02", title: "ESCROWED", body: "Funds released only after delivery confirmed by buyer.", stat: "$14.2M held in escrow" },
  { n: "03", title: "INSPECTED", body: "Third-party QC on every shipment over $5K.", stat: "1,847 inspections" },
];

function Pillars() {
  return (
    <section id="pricing" className="border-b" style={{ background: "#F7F5F0", borderColor: "#262626" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-16">
        <div className="mb-10 uppercase tracking-[0.2em]" style={{ fontFamily: MONO, fontSize: "11px", color: "#717182" }}>
          Section B — Operating principles
        </div>
        <div className="grid grid-cols-1 gap-px md:grid-cols-3" style={{ background: "#D8D4CC" }}>
          {PILLARS.map((p) => (
            <div key={p.n} className="relative px-8 py-10" style={{ background: "#FFFFFF" }}>
              <div className="absolute left-0 top-10 h-2 w-2" style={{ background: "#C8312D" }} />
              <div className="pl-5">
                <div className="flex items-baseline gap-4" style={{ fontFamily: MONO }}>
                  <span style={{ fontSize: "14px", color: "#717182", fontWeight: 700 }}>{p.n}</span>
                  <span className="tracking-[0.15em]" style={{ fontSize: "20px", fontWeight: 700, color: "#0A0A0A" }}>
                    {p.title}
                  </span>
                </div>
                <p className="mt-5 max-w-[280px]" style={{ fontFamily: SANS, fontSize: "15px", lineHeight: 1.55, color: "#262626" }}>
                  {p.body}
                </p>
                <div
                  className="mt-10 border-t pt-4 uppercase tracking-wider"
                  style={{ borderColor: "#D8D4CC", fontFamily: MONO, fontSize: "11px", color: "#0A0A0A", fontWeight: 700 }}
                >
                  ▪ {p.stat}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

const SUPPLIERS = [
  { code: "SUP-CN-SZX-0024", name: "Gigalight Technology", cn: "光迅科技", city: "Shenzhen", caps: ["Transceivers", "QSFP28", "DWDM"], verified: "Mar 2026" },
  { code: "SUP-CN-DGN-0011", name: "InnoLight Optoelectronics", cn: "旭创科技", city: "Dongguan", caps: ["400G OSFP", "DR4", "Coherent"], verified: "Feb 2026" },
  { code: "SUP-CN-WUH-0007", name: "Hisense Broadband", cn: "海信宽带", city: "Wuhan", caps: ["DFB Lasers", "PON", "Modules"], verified: "Jan 2026" },
  { code: "SUP-CN-SZX-0019", name: "Accelink Technologies", cn: "光迅通讯", city: "Shenzhen", caps: ["MTP/MPO", "OM4", "Cassettes"], verified: "Mar 2026" },
];

function Suppliers() {
  return (
    <section id="suppliers" className="border-b" style={{ background: "#F7F5F0", borderColor: "#262626" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-16">
        <div className="mb-10 flex items-end justify-between">
          <div>
            <div className="mb-3 uppercase tracking-[0.2em]" style={{ fontFamily: MONO, fontSize: "11px", color: "#717182" }}>
              Section C — Featured Suppliers
            </div>
            <h2 style={{ fontFamily: MONO, fontSize: "24px", fontWeight: 700, color: "#0A0A0A" }}>
              247 audited factories. Browse a sample.
            </h2>
          </div>
          <Link
            href="/auth/login"
            className="uppercase tracking-wider hover:opacity-60"
            style={{ fontFamily: MONO, fontSize: "11px", color: "#D54E20", fontWeight: 700 }}
          >
            View all suppliers →
          </Link>
        </div>

        <div className="grid grid-cols-1 gap-px md:grid-cols-2 lg:grid-cols-4" style={{ background: "#D8D4CC" }}>
          {SUPPLIERS.map((s) => (
            <div key={s.code} className="px-6 py-6" style={{ background: "#FFFFFF" }}>
              <div
                className="mb-4 flex items-center justify-between uppercase tracking-wider"
                style={{ fontFamily: MONO, fontSize: "10px", color: "#717182" }}
              >
                <span>{s.code}</span>
                <span style={{ color: "#3F6F3F" }}>● VERIFIED</span>
              </div>
              <div style={{ fontFamily: MONO, fontSize: "16px", fontWeight: 700, color: "#0A0A0A" }}>{s.name}</div>
              <div className="mt-1" style={{ fontFamily: MONO, fontSize: "13px", color: "#717182" }}>{s.cn}</div>
              <div className="mt-4 flex items-center gap-2" style={{ fontFamily: MONO, fontSize: "11px", color: "#262626" }}>
                <span className="inline-block h-2 w-2" style={{ background: "#C8312D" }} />
                <span className="uppercase tracking-wider">{s.city}, CN</span>
              </div>
              <div className="mt-5 flex flex-wrap gap-1.5">
                {s.caps.map((c) => (
                  <span
                    key={c}
                    className="border px-2 py-1 uppercase tracking-wider"
                    style={{ fontFamily: MONO, fontSize: "10px", borderColor: "#D8D4CC", color: "#262626" }}
                  >
                    {c}
                  </span>
                ))}
              </div>
              <div
                className="mt-6 border-t pt-3 uppercase tracking-wider"
                style={{ borderColor: "#D8D4CC", fontFamily: MONO, fontSize: "10px", color: "#717182" }}
              >
                Audited {s.verified}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

const STAGES = [
  { n: "01", label: "RFQ", time: "Day 0", side: "US" },
  { n: "02", label: "QUOTE", time: "Day 1", side: "CN" },
  { n: "03", label: "PAYMENT", time: "Day 2", side: "US" },
  { n: "04", label: "PRODUCTION", time: "Day 3–9", side: "CN" },
  { n: "05", label: "QC", time: "Day 10", side: "CN" },
  { n: "06", label: "DELIVERY", time: "Day 14", side: "US" },
];

function TimelineSection() {
  return (
    <section id="how" className="border-b" style={{ background: "#F7F5F0", borderColor: "#262626" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-16">
        <div className="mb-3 uppercase tracking-[0.2em]" style={{ fontFamily: MONO, fontSize: "11px", color: "#717182" }}>
          Section D — How orders flow
        </div>
        <h2 className="mb-12 max-w-[760px]" style={{ fontFamily: MONO, fontSize: "24px", fontWeight: 700, color: "#0A0A0A" }}>
          From RFQ to delivery, every step timestamped and attributable.
        </h2>

        <div className="border" style={{ borderColor: "#D8D4CC", background: "#FFFFFF" }}>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
            {STAGES.map((s, i) => (
              <div
                key={s.n}
                className="px-5 py-8"
                style={{ borderRight: i < STAGES.length - 1 ? "1px solid #D8D4CC" : "none" }}
              >
                <div
                  className="mb-4 flex items-center gap-2 uppercase tracking-wider"
                  style={{ fontFamily: MONO, fontSize: "10px", color: "#717182" }}
                >
                  <span className="inline-block h-2 w-2" style={{ background: s.side === "CN" ? "#C8312D" : "#1A2B4A" }} />
                  {s.side}-side
                </div>
                <div style={{ fontFamily: MONO, fontSize: "12px", color: "#717182", fontWeight: 700 }}>{s.n}</div>
                <div className="mt-1 tracking-[0.1em]" style={{ fontFamily: MONO, fontSize: "16px", fontWeight: 700, color: "#0A0A0A" }}>
                  {s.label}
                </div>
                <div className="mt-4" style={{ fontFamily: MONO, fontSize: "11px", color: "#262626" }}>{s.time}</div>
              </div>
            ))}
          </div>
          <div
            className="border-t px-6 py-4"
            style={{ borderColor: "#D8D4CC", background: "#F7F5F0", fontFamily: MONO, fontSize: "11px", color: "#717182" }}
          >
            ORD-20260429-0089-SFO • SAMPLE ORDER • TYPICAL LEAD TIME 14 DAYS
          </div>
        </div>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer style={{ background: "#0A0A0A", color: "#F7F5F0" }}>
      <div className="mx-auto max-w-[1400px] px-8 py-16">
        <div className="grid grid-cols-1 gap-12 md:grid-cols-3">
          <div>
            <div className="tracking-[0.18em]" style={{ fontFamily: MONO, fontSize: "18px", fontWeight: 700 }}>
              FIBERLANE
            </div>
            <RouteMark className="mt-3 text-[12px]" />
            <p className="mt-6 max-w-[280px]" style={{ fontFamily: SANS, fontSize: "14px", lineHeight: 1.55, color: "#D8D4CC" }}>
              Verified photonics sourcing, Shenzhen ─── San Francisco.
            </p>
          </div>

          <div className="grid grid-cols-2 gap-8" style={{ fontFamily: MONO, fontSize: "12px" }}>
            <div>
              <div className="mb-4 uppercase tracking-[0.2em]" style={{ color: "#717182", fontSize: "10px" }}>Platform</div>
              <ul className="space-y-3">
                <li>Suppliers</li>
                <li>How it works</li>
                <li>Pricing</li>
                <li>Catalog</li>
              </ul>
            </div>
            <div>
              <div className="mb-4 uppercase tracking-[0.2em]" style={{ color: "#717182", fontSize: "10px" }}>Company</div>
              <ul className="space-y-3">
                <li>About</li>
                <li>Trust &amp; Compliance</li>
                <li>Careers</li>
                <li>Press</li>
              </ul>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-6" style={{ fontFamily: MONO, fontSize: "11px", lineHeight: 1.6 }}>
            <div>
              <div className="mb-3 flex items-center gap-2 uppercase tracking-wider" style={{ color: "#717182", fontSize: "10px" }}>
                <span className="inline-block h-2 w-2" style={{ background: "#C8312D" }} />
                CN OFFICE
              </div>
              <div style={{ color: "#F7F5F0" }}>
                Bao&apos;an District<br />
                518101 Shenzhen, CN<br />
                22.54°N 114.05°E
              </div>
            </div>
            <div>
              <div className="mb-3 flex items-center gap-2 uppercase tracking-wider" style={{ color: "#717182", fontSize: "10px" }}>
                <span className="inline-block h-2 w-2" style={{ background: "#1A2B4A" }} />
                US OFFICE
              </div>
              <div style={{ color: "#F7F5F0" }}>
                340 Brannan St<br />
                94107 San Francisco, US<br />
                37.77°N 122.40°W
              </div>
            </div>
          </div>
        </div>

        <div
          className="mt-16 flex flex-wrap items-center justify-between gap-4 border-t pt-6 uppercase tracking-wider"
          style={{ borderColor: "#262626", fontFamily: MONO, fontSize: "10px", color: "#717182" }}
        >
          <span>© 2026 FIBERLANE LOGISTICS, INC.</span>
          <span>FBL-MARK-20260429-V1.0</span>
        </div>
      </div>
    </footer>
  );
}

export function Landing() {
  return (
    <div
      className="min-h-screen w-full"
      style={{ background: "#F7F5F0", backgroundImage: PAPER_TEXTURE, color: "#0A0A0A", fontFamily: SANS }}
    >
      <TopBar />
      <main>
        <Hero />
        <RouteSection />
        <Pillars />
        <Suppliers />
        <TimelineSection />
      </main>
      <Footer />
    </div>
  );
}

export default Landing;

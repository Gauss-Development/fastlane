import type { StatusTone } from "@/components/ui/pill";

export const exampleQueries = [
  "100G QSFP28 LR4 Cisco compatible",
  "10G SFP+ LR Cisco Nexus",
  "400G QSFP-DD DR4 single-mode",
  "25G SFP28 SR multimode",
  "80km DWDM SFP+ 1550nm",
] as const;

export interface RecentRFQ {
  id: string;
  query: string;
  qty: number;
  status: string;
  tone: StatusTone;
  age: string;
}

export const recentRFQs: RecentRFQ[] = [
  {
    id: "RFQ-20260429-0142-SZX",
    query: "100G QSFP28 LR4, Cisco Nexus 9000",
    qty: 100,
    status: "quoted",
    tone: "info",
    age: "18m",
  },
  {
    id: "RFQ-20260429-0138-SZX",
    query: "10G SFP+ LR, 1310nm, single-mode",
    qty: 250,
    status: "open",
    tone: "success",
    age: "32m",
  },
  {
    id: "RFQ-20260429-0131-SZX",
    query: "400G QSFP-DD DR4, 500m",
    qty: 40,
    status: "supplier review",
    tone: "warning",
    age: "1h",
  },
  {
    id: "RFQ-20260429-0124-SZX",
    query: "25G SFP28 SR, Arista compatible",
    qty: 500,
    status: "quoted",
    tone: "info",
    age: "2h",
  },
  {
    id: "RFQ-20260428-0117-SZX",
    query: "80km DWDM SFP+ 1550nm C-band",
    qty: 60,
    status: "accepted",
    tone: "success",
    age: "1d",
  },
  {
    id: "RFQ-20260428-0109-SZX",
    query: "40G QSFP+ LR4, LC duplex",
    qty: 120,
    status: "closed",
    tone: "neutral",
    age: "1d",
  },
];

export interface LiveQuote {
  supplier: string;
  city: string;
  rfqId: string;
  part: string;
  price: string;
  leadTime: string;
  receivedAt: string;
}

export const liveQuotes: LiveQuote[] = [
  {
    supplier: "Eoptolink",
    city: "Chengdu",
    rfqId: "RFQ-20260429-0142-SZX",
    part: "100G QSFP28 LR4",
    price: "$182.00",
    leadTime: "10 days",
    receivedAt: "3 min ago",
  },
  {
    supplier: "Accelink",
    city: "Wuhan",
    rfqId: "RFQ-20260429-0138-SZX",
    part: "10G SFP+ LR",
    price: "$76.00",
    leadTime: "7 days",
    receivedAt: "9 min ago",
  },
  {
    supplier: "Gigalight",
    city: "Shenzhen",
    rfqId: "RFQ-20260429-0131-SZX",
    part: "400G QSFP-DD DR4",
    price: "$498.00",
    leadTime: "14 days",
    receivedAt: "21 min ago",
  },
  {
    supplier: "Source Photonics",
    city: "Chengdu",
    rfqId: "RFQ-20260429-0124-SZX",
    part: "25G SFP28 SR",
    price: "$44.00",
    leadTime: "7 days",
    receivedAt: "36 min ago",
  },
];

export interface FeaturedSupplier {
  code: string;
  name: string;
  nameZh: string;
  city: string;
  capability: string;
  onTimeRate: string;
  orders: string;
}

export const featuredSuppliers: FeaturedSupplier[] = [
  {
    code: "SUP-CN-SZX-0024",
    name: "Gigalight",
    nameZh: "易飞扬通信",
    city: "Shenzhen",
    capability: "400G / coherent optics",
    onTimeRate: "96.2%",
    orders: "1,847",
  },
  {
    code: "SUP-CN-CDU-0017",
    name: "Eoptolink",
    nameZh: "新易盛通信",
    city: "Chengdu",
    capability: "QSFP28 / SFP28",
    onTimeRate: "97.1%",
    orders: "2,103",
  },
  {
    code: "SUP-CN-WUH-0008",
    name: "Accelink",
    nameZh: "光迅科技",
    city: "Wuhan",
    capability: "telecom transceivers",
    onTimeRate: "95.4%",
    orders: "3,412",
  },
  {
    code: "SUP-CN-SUZ-0031",
    name: "InnoLight Technology",
    nameZh: "旭创科技",
    city: "Suzhou",
    capability: "800G / silicon photonics",
    onTimeRate: "96.8%",
    orders: "4,920",
  },
  {
    code: "SUP-CN-QDO-0011",
    name: "Hisense Broadband",
    nameZh: "海信宽带",
    city: "Qingdao",
    capability: "access network optics",
    onTimeRate: "94.9%",
    orders: "2,775",
  },
  {
    code: "SUP-CN-CHD-0014",
    name: "Source Photonics",
    nameZh: "索尔思光电",
    city: "Chengdu",
    capability: "data center optics",
    onTimeRate: "95.8%",
    orders: "1,562",
  },
];

export interface Manufacturer {
  id: string;
  user_id?: string;
  name: string;
  name_zh: string;
  city: string;
  cluster: string;
  description: string;
  website: string;
  service_types: string[];
  assembly_types: string[];
  min_layers: number;
  max_layers: number;
  materials: string[];
  surface_finishes: string[];
  min_order_qty: number;
  max_order_qty: number;
  lead_time_days: number;
  monthly_capacity: number;
  smallest_package: string;
  certifications: string[];
  contact_email: string;
  contact_wechat: string;
  verified: boolean;
  verified_at?: string;
  status?: string;
  created_at: string;
}

export interface ListManufacturersResponse {
  manufacturers: Manufacturer[];
  total: number;
}

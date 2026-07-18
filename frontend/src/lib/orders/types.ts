export type OrderStatus =
  | "pending"
  | "confirmed"
  | "in_production"
  | "qc"
  | "shipped"
  | "delivered"
  | "cancelled";

export type PaymentStatus = "unpaid" | "partial" | "paid" | "refunded";
export type QCStatus = "pending" | "passed" | "failed";

export interface Order {
  id: string;
  buyer_id: string;
  supplier_id: string;
  quote_id: string;
  rfq_id: string;
  status: OrderStatus;
  payment_status: PaymentStatus;
  qc_status: QCStatus;
  total_usd: number;
  shipping_address: string;
  shipping_city: string;
  shipping_country: string;
  warranty_until: string;
  created_at: string;
  updated_at: string;
}

export interface OrderEvent {
  id: string;
  order_id: string;
  event_type: string;
  from_status: string;
  to_status: string;
  actor_id: string;
  actor_type: string;
  occurred_at: string;
  location: string;
  notes: string;
}

export interface ListOrdersResponse {
  orders: Order[];
  total: number;
  limit: number;
  offset: number;
}

export interface ListOrderEventsResponse {
  events: OrderEvent[];
}

export interface AppendOrderEventBody {
  to_status: string;
  event_type: string;
  notes?: string;
  location?: string;
}

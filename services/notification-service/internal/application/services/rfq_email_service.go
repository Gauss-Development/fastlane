package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"notification-service/internal/application/dto"
	"notification-service/internal/domain/entities"
	"notification-service/internal/infrastructure/email"
	"notification-service/pkg/logger"
)

// RFQEmailService consumes rfq.* / quote.* events and routes them to email:
// supplier magic-link RFQ invites, buyer confirmations, and buyer
// quote-received pings. Email failures are logged per recipient rather than
// returned, so one bad address never sends the whole event to the DLQ.
type RFQEmailService struct {
	notifications *NotificationService
	sender        *email.Sender
	frontendURL   string
	logger        *logger.Logger
}

func NewRFQEmailService(notifications *NotificationService, sender *email.Sender, frontendURL string, logger *logger.Logger) *RFQEmailService {
	return &RFQEmailService{
		notifications: notifications,
		sender:        sender,
		frontendURL:   strings.TrimRight(frontendURL, "/"),
		logger:        logger,
	}
}

func (s *RFQEmailService) ProcessRFQCreatedEvent(ctx context.Context, body []byte) error {
	var event entities.RFQCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal rfq.created event: %w", err)
	}
	s.logger.Info(fmt.Sprintf("processing rfq.created: %s (%d suppliers)", event.RFQID, len(event.Suppliers)))

	invited := 0
	for _, supplier := range event.Suppliers {
		if supplier.ContactEmail == "" {
			s.logger.Warn(fmt.Sprintf("rfq %s: supplier %s has no contact email, skipping invite", event.RFQID, supplier.SupplierID))
			continue
		}
		if supplier.MagicLinkURL == "" {
			s.logger.Warn(fmt.Sprintf("rfq %s: supplier %s has no magic link, skipping invite", event.RFQID, supplier.SupplierID))
			continue
		}
		subject := fmt.Sprintf("New RFQ %s — %s / 新询盘", event.RFQID, supplier.ProductSKU)
		if err := s.sender.Send(ctx, supplier.ContactEmail, subject, supplierInviteHTML(&event, &supplier)); err != nil {
			s.logger.Error(fmt.Sprintf("rfq %s: supplier invite to %s failed: %v", event.RFQID, supplier.SupplierID, err))
			continue
		}
		invited++
	}

	if event.BuyerEmail != "" {
		subject := fmt.Sprintf("RFQ %s sent to %d supplier(s)", event.RFQID, invited)
		if err := s.sender.Send(ctx, event.BuyerEmail, subject, buyerConfirmationHTML(&event, invited)); err != nil {
			s.logger.Error(fmt.Sprintf("rfq %s: buyer confirmation failed: %v", event.RFQID, err))
		}
	}

	s.storeBuyerNotification(ctx, event.BuyerID, entities.NotificationTypeRFQCreated,
		fmt.Sprintf("RFQ %s published", event.RFQID),
		fmt.Sprintf("Sent to %d supplier(s): %s", invited, event.PartSummary),
		map[string]interface{}{"rfq_id": event.RFQID, "suppliers_invited": invited},
	)
	return nil
}

func (s *RFQEmailService) ProcessQuoteSubmittedEvent(ctx context.Context, body []byte) error {
	var event entities.QuoteSubmittedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal quote.submitted event: %w", err)
	}
	s.logger.Info(fmt.Sprintf("processing quote.submitted: %s for rfq %s", event.QuoteID, event.RFQID))

	if event.BuyerEmail != "" {
		subject := fmt.Sprintf("Quote received for %s — $%.2f/unit", event.RFQID, event.PriceUSD)
		if err := s.sender.Send(ctx, event.BuyerEmail, subject, quoteReceivedHTML(&event, s.rfqURL(event.RFQID))); err != nil {
			s.logger.Error(fmt.Sprintf("quote %s: buyer notification failed: %v", event.QuoteID, err))
		}
	}

	s.storeBuyerNotification(ctx, event.BuyerID, entities.NotificationTypeQuoteSubmitted,
		fmt.Sprintf("Quote received on %s", event.RFQID),
		fmt.Sprintf("%s quoted $%.2f/unit, %d-day lead time", event.SupplierName, event.PriceUSD, event.LeadTimeDays),
		map[string]interface{}{"rfq_id": event.RFQID, "quote_id": event.QuoteID, "supplier_id": event.SupplierID},
	)
	return nil
}

func (s *RFQEmailService) storeBuyerNotification(ctx context.Context, buyerID string, notifType entities.NotificationType, title, message string, data map[string]interface{}) {
	if buyerID == "" {
		return
	}
	if _, err := s.notifications.CreateNotification(ctx, &dto.CreateNotificationRequest{
		UserID:  buyerID,
		Type:    string(notifType),
		Title:   title,
		Message: message,
		Data:    data,
	}); err != nil {
		s.logger.Error("failed to store buyer notification: " + err.Error())
	}
}

func (s *RFQEmailService) rfqURL(rfqID string) string {
	if s.frontendURL == "" {
		return ""
	}
	return s.frontendURL + "/rfqs/" + rfqID
}

// Templates follow the Fiberlane design system: monospace, paper background,
// ink text, international-orange CTA, no imagery. Supplier email is bilingual
// EN/中文 because factory sales reps read it on phones.

const emailShell = `<!DOCTYPE html>
<html><body style="margin:0;padding:0;background-color:#F7F5F0;">
<div style="max-width:560px;margin:0 auto;padding:32px 24px;font-family:'JetBrains Mono','Courier New',monospace;color:#0A0A0A;">
<div style="border-bottom:1px solid #0A0A0A;padding-bottom:12px;margin-bottom:24px;">
<span style="font-weight:700;letter-spacing:0.08em;">FIBERLANE</span>
<span style="float:right;font-size:12px;color:#555;">CN &#9472;&#9472;&#9472;&#9654; US</span>
</div>
%s
<div style="border-top:1px solid #ccc;margin-top:32px;padding-top:12px;font-size:11px;color:#888;">
Fiberlane &#8212; Shenzhen &#9472;&#9472;&#9472;&#9654; San Francisco
</div>
</div></body></html>`

func supplierInviteHTML(event *entities.RFQCreatedEvent, supplier *entities.RFQSupplierInvite) string {
	esc := html.EscapeString
	buyer := event.BuyerCompany
	if buyer == "" {
		buyer = "a US buyer"
	}

	qty := "—"
	if event.Qty > 0 {
		qty = fmt.Sprintf("%d units", event.Qty)
	}
	targetDate := event.TargetDate
	if targetDate == "" {
		targetDate = "—"
	}
	delivery := event.ShippingAddress
	if delivery == "" {
		delivery = "United States"
	}

	rows := [][2]string{
		{"PART REQUESTED / 询盘部件", fmt.Sprintf("%s (%s)", esc(supplier.ProductName), esc(supplier.ProductSKU))},
		{"REQUEST / 需求描述", esc(event.QueryText)},
		{"QUANTITY / 数量", esc(qty)},
		{"DELIVERY TO / 交货地点", esc(delivery)},
		{"TARGET DATE / 目标日期", esc(targetDate)},
	}
	if event.Notes != "" {
		rows = append(rows, [2]string{"NOTES / 备注", esc(event.Notes)})
	}

	var table strings.Builder
	for _, row := range rows {
		table.WriteString(fmt.Sprintf(
			`<tr><td style="padding:6px 12px 6px 0;font-size:11px;color:#555;vertical-align:top;white-space:nowrap;">%s</td><td style="padding:6px 0;font-size:13px;">%s</td></tr>`,
			row[0], row[1]))
	}

	content := fmt.Sprintf(`
<p style="font-size:15px;font-weight:700;margin:0 0 4px;">NEW RFQ FROM %s (United States)</p>
<p style="font-size:13px;color:#555;margin:0 0 16px;">来自 %s（美国）的新询盘</p>
<p style="font-size:12px;color:#555;margin:0 0 20px;">%s</p>
<table style="border-collapse:collapse;width:100%%;">%s</table>
<div style="margin:28px 0;">
<a href="%s" style="display:inline-block;background-color:#D54E20;color:#FFFFFF;text-decoration:none;padding:14px 28px;font-size:14px;font-weight:700;letter-spacing:0.06em;">SUBMIT QUOTE / 提交报价 &#9654;</a>
</div>
<p style="font-size:11px;color:#888;">No login required. This link is unique to your company and this RFQ.<br/>无需登录。此链接仅适用于贵公司和本询盘。</p>`,
		esc(strings.ToUpper(buyer)), esc(buyer), esc(event.RFQID), table.String(), supplier.MagicLinkURL)

	return fmt.Sprintf(emailShell, content)
}

func buyerConfirmationHTML(event *entities.RFQCreatedEvent, invited int) string {
	esc := html.EscapeString
	content := fmt.Sprintf(`
<p style="font-size:15px;font-weight:700;margin:0 0 16px;">RFQ PUBLISHED</p>
<p style="font-size:13px;margin:0 0 8px;"><span style="color:#555;">ID</span> &nbsp;%s</p>
<p style="font-size:13px;margin:0 0 8px;"><span style="color:#555;">REQUEST</span> &nbsp;%s</p>
<p style="font-size:13px;margin:0 0 20px;"><span style="color:#555;">SUPPLIERS INVITED</span> &nbsp;%d</p>
<p style="font-size:13px;color:#555;">You will be emailed as quotes arrive. Typical first response is within 24 hours (suppliers are in CST, UTC+8).</p>`,
		esc(event.RFQID), esc(event.QueryText), invited)
	return fmt.Sprintf(emailShell, content)
}

func quoteReceivedHTML(event *entities.QuoteSubmittedEvent, rfqURL string) string {
	esc := html.EscapeString
	cta := ""
	if rfqURL != "" {
		cta = fmt.Sprintf(`<div style="margin:28px 0;"><a href="%s" style="display:inline-block;background-color:#D54E20;color:#FFFFFF;text-decoration:none;padding:14px 28px;font-size:14px;font-weight:700;letter-spacing:0.06em;">COMPARE QUOTES &#9654;</a></div>`, rfqURL)
	}
	content := fmt.Sprintf(`
<p style="font-size:15px;font-weight:700;margin:0 0 16px;">QUOTE RECEIVED</p>
<p style="font-size:13px;margin:0 0 8px;"><span style="color:#555;">RFQ</span> &nbsp;%s</p>
<p style="font-size:13px;margin:0 0 8px;"><span style="color:#555;">SUPPLIER</span> &nbsp;%s</p>
<p style="font-size:13px;margin:0 0 8px;"><span style="color:#555;">UNIT PRICE</span> &nbsp;$%.2f USD</p>
<p style="font-size:13px;margin:0 0 20px;"><span style="color:#555;">LEAD TIME</span> &nbsp;%d days</p>%s`,
		esc(event.RFQID), esc(event.SupplierName), event.PriceUSD, event.LeadTimeDays, cta)
	return fmt.Sprintf(emailShell, content)
}

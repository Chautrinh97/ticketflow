package http

type checkoutResponseDTO struct {
	PaymentID   string `json:"payment_id"`
	CheckoutURL string `json:"checkout_url"`
}

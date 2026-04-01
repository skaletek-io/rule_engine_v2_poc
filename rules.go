package main

type Rule struct {
	ID         string
	Name       string
	TemplateID string
	Expression string
	Severity   string
	Message    string
}

func seedRules() []Rule {
	return []Rule{

		// ── BANKING ─────────────────────────────────────────────────────────
		{
			ID:         "rule_bank_001",
			Name:       "Large International Wire — New Account",
			TemplateID: "bank_wire_transfer",
			Expression: `payment.amount > 10000 && sender.accountAge < 30`,
			Severity:   "high",
			Message:    "High-value wire from account less than 30 days old",
		},
		{
			ID:         "rule_bank_002",
			Name:       "Transfer to High-Risk Jurisdiction",
			TemplateID: "bank_wire_transfer",
			Expression: `receiver.country == "SC" || receiver.country == "KY" || receiver.country == "VU"`,
			Severity:   "high",
			Message:    "Transfer destination is a high-risk offshore jurisdiction",
		},

		// ── FINTECH ─────────────────────────────────────────────────────────
		{
			ID:         "rule_fin_001",
			Name:       "Unverified User — Large Transaction",
			TemplateID: "fintech_payment",
			Expression: `user.kycVerified == false && transaction.amount > 1000`,
			Severity:   "high",
			Message:    "Large transaction by unverified user",
		},
		{
			ID:         "rule_fin_002",
			Name:       "VPN + Crypto Merchant",
			TemplateID: "fintech_payment",
			Expression: `device.isVPN == true && transaction.merchantMCC == 6051`,
			Severity:   "critical",
			Message:    "Crypto merchant purchase made over VPN",
		},
		{
			ID:         "rule_fin_003",
			Name:       "New Account — High Value",
			TemplateID: "fintech_payment",
			Expression: `user.accountAgeDays < 7 && transaction.amount > 5000`,
			Severity:   "medium",
			Message:    "High-value transaction on account less than 7 days old",
		},

		// ── CRYPTO ──────────────────────────────────────────────────────────
		{
			ID:         "rule_crypto_001",
			Name:       "Blacklisted Wallet Withdrawal",
			TemplateID: "crypto_withdrawal",
			Expression: `wallet.isBlacklisted == true`,
			Severity:   "critical",
			Message:    "Withdrawal attempted from a blacklisted wallet address",
		},
		{
			ID:         "rule_crypto_002",
			Name:       "High Risk Score — Large Withdrawal",
			TemplateID: "crypto_withdrawal",
			Expression: `wallet.riskScore > 80 && withdrawal.amountUSD > 50000`,
			Severity:   "critical",
			Message:    "Large withdrawal from high-risk wallet",
		},
		{
			ID:         "rule_crypto_003",
			Name:       "Unverified User — Any Withdrawal",
			TemplateID: "crypto_withdrawal",
			Expression: `user.kycVerified == false && withdrawal.amountUSD > 500`,
			Severity:   "high",
			Message:    "Withdrawal by unverified user exceeds limit",
		},

		// ── HOTEL ────────────────────────────────────────────────────────────
		{
			ID:         "rule_hotel_001",
			Name:       "Bulk Same-Day Booking — Crypto Payment",
			TemplateID: "hotel_reservation",
			Expression: `reservation.roomsBooked > 5 && reservation.leadTimeDays == 0 && reservation.paymentMethod == "crypto"`,
			Severity:   "high",
			Message:    "Bulk same-day booking paid with crypto",
		},
		{
			ID:         "rule_hotel_002",
			Name:       "High-Value Reservation — Sanctioned Nationality",
			TemplateID: "hotel_reservation",
			Expression: `reservation.totalAmount > 50000 && (guest.nationality == "RU" || guest.nationality == "IR" || guest.nationality == "KP")`,
			Severity:   "critical",
			Message:    "High-value reservation by guest from sanctioned country",
		},

		// ── INSURANCE ────────────────────────────────────────────────────────
		{
			ID:         "rule_ins_001",
			Name:       "Claim on Brand-New Policy",
			TemplateID: "insurance_claim",
			Expression: `claimant.policyAgeDays < 30 && claim.amountRequested > 10000`,
			Severity:   "high",
			Message:    "Large claim filed within 30 days of policy inception",
		},
		{
			ID:         "rule_ins_002",
			Name:       "Repeat Claimant — High Value",
			TemplateID: "insurance_claim",
			Expression: `claimant.previousClaims >= 3 && claim.amountRequested > 20000`,
			Severity:   "medium",
			Message:    "High-value claim from customer with 3 or more prior claims",
		},
		{
			ID:         "rule_ins_003",
			Name:       "Claim Near Coverage Limit",
			TemplateID: "insurance_claim",
			Expression: `claim.amountRequested >= policy.coverageLimit * 0.9`,
			Severity:   "medium",
			Message:    "Claim amount is within 10% of policy coverage limit",
		},

		// ── E-COMMERCE ───────────────────────────────────────────────────────
		{
			ID:         "rule_ecom_001",
			Name:       "New Account — Bulk Electronics Order",
			TemplateID: "ecommerce_order",
			Expression: `buyer.accountAgeDays == 0 && order.category == "electronics" && order.itemCount > 10`,
			Severity:   "high",
			Message:    "Brand-new account placing bulk electronics order",
		},
		{
			ID:         "rule_ecom_002",
			Name:       "Billing and Shipping Mismatch — New Card",
			TemplateID: "ecommerce_order",
			Expression: `shipping.addressMatchesBilling == false && payment.isNewCard == true`,
			Severity:   "medium",
			Message:    "New card used with mismatched billing and shipping address",
		},
		{
			ID:         "rule_ecom_003",
			Name:       "High Value Order — New Account",
			TemplateID: "ecommerce_order",
			Expression: `order.totalAmount > 1000 && buyer.accountAgeDays < 1`,
			Severity:   "high",
			Message:    "High-value order placed by account created today",
		},
	}
}

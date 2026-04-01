package main

import "time"

type Event struct {
	ID         string
	TemplateID string
	OccurredAt time.Time
	Payload    map[string]any
}

func seedEvents() []Event {
	return []Event{

		// ── BANKING ─────────────────────────────────────────────────────────
		{
			ID:         "evt_bank_001",
			TemplateID: "bank_wire_transfer",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"sender": map[string]any{
					"fullName":    "Amara Okafor",
					"nationality": "NG",
					"accountAge":  14, // days
				},
				"receiver": map[string]any{
					"fullName": "Phantom Holdings",
					"country":  "SC", // Seychelles
				},
				"payment": map[string]any{
					"amount":   45000.00,
					"currency": "USD",
				},
				"channel": "online",
			},
		},
		{
			ID:         "evt_bank_002",
			TemplateID: "bank_wire_transfer",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"sender": map[string]any{
					"fullName":    "Emeka Nwosu",
					"nationality": "NG",
					"accountAge":  730,
				},
				"receiver": map[string]any{
					"fullName": "University of Lagos",
					"country":  "NG",
				},
				"payment": map[string]any{
					"amount":   2500.00,
					"currency": "NGN",
				},
				"channel": "mobile",
			},
		},

		// ── FINTECH / PAYMENTS ───────────────────────────────────────────────
		{
			ID:         "evt_fin_001",
			TemplateID: "fintech_payment",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"user": map[string]any{
					"id":             "usr_9921",
					"email":          "spoofed@tempmail.io",
					"accountAgeDays": 1,
					"kycVerified":    false,
				},
				"transaction": map[string]any{
					"amount":       9500.00,
					"currency":     "USD",
					"merchantName": "CryptoSwap Pro",
					"merchantMCC":  6051, // non-financial institutions / crypto
				},
				"device": map[string]any{
					"ip":      "185.220.101.5",
					"country": "RU",
					"isVPN":   true,
				},
			},
		},
		{
			ID:         "evt_fin_002",
			TemplateID: "fintech_payment",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"user": map[string]any{
					"id":             "usr_0042",
					"email":          "jane.doe@gmail.com",
					"accountAgeDays": 400,
					"kycVerified":    true,
				},
				"transaction": map[string]any{
					"amount":       35.00,
					"currency":     "USD",
					"merchantName": "Spotify",
					"merchantMCC":  5735,
				},
				"device": map[string]any{
					"ip":      "102.89.23.1",
					"country": "NG",
					"isVPN":   false,
				},
			},
		},

		// ── CRYPTO ──────────────────────────────────────────────────────────
		{
			ID:         "evt_crypto_001",
			TemplateID: "crypto_withdrawal",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"wallet": map[string]any{
					"address":       "1A1zP1eP5QGefi2DMPTfTL5SLmv7Divf",
					"isBlacklisted": true,
					"riskScore":     92,
				},
				"withdrawal": map[string]any{
					"amountUSD":   120000.00,
					"asset":       "BTC",
					"destination": "external",
				},
				"user": map[string]any{
					"id":          "usr_cr_881",
					"kycVerified": false,
				},
			},
		},
		{
			ID:         "evt_crypto_002",
			TemplateID: "crypto_withdrawal",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"wallet": map[string]any{
					"address":       "3FZbgi29cpjq2GjdwV8eyHuJJnkLtktZc5",
					"isBlacklisted": false,
					"riskScore":     12,
				},
				"withdrawal": map[string]any{
					"amountUSD":   200.00,
					"asset":       "ETH",
					"destination": "external",
				},
				"user": map[string]any{
					"id":          "usr_cr_210",
					"kycVerified": true,
				},
			},
		},

		// ── HOTEL RESERVATION ───────────────────────────────────────────────
		{
			ID:         "evt_hotel_001",
			TemplateID: "hotel_reservation",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"guest": map[string]any{
					"fullName":    "Viktor Petrov",
					"nationality": "RU",
					"email":       "v.petrov@protonmail.com",
				},
				"reservation": map[string]any{
					"roomsBooked":   9,
					"nightlyRate":   850.00,
					"totalNights":   14,
					"totalAmount":   107100.00,
					"paymentMethod": "crypto",
					"leadTimeDays":  0, // booked same day
				},
				"property": map[string]any{
					"country": "AE", // UAE
					"city":    "Dubai",
				},
			},
		},
		{
			ID:         "evt_hotel_002",
			TemplateID: "hotel_reservation",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"guest": map[string]any{
					"fullName":    "Sarah Connor",
					"nationality": "US",
					"email":       "sarah.c@gmail.com",
				},
				"reservation": map[string]any{
					"roomsBooked":   1,
					"nightlyRate":   120.00,
					"totalNights":   3,
					"totalAmount":   360.00,
					"paymentMethod": "card",
					"leadTimeDays":  21,
				},
				"property": map[string]any{
					"country": "FR",
					"city":    "Paris",
				},
			},
		},

		// ── INSURANCE CLAIM ─────────────────────────────────────────────────
		{
			ID:         "evt_ins_001",
			TemplateID: "insurance_claim",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"claimant": map[string]any{
					"fullName":       "Marcus Bell",
					"policyAgeDays":  12, // very new policy
					"previousClaims": 4,
				},
				"claim": map[string]any{
					"type":            "theft",
					"amountRequested": 48000.00,
					"currency":        "USD",
					"incidentCountry": "MX",
				},
				"policy": map[string]any{
					"type":          "personal_property",
					"coverageLimit": 50000.00,
				},
			},
		},
		{
			ID:         "evt_ins_002",
			TemplateID: "insurance_claim",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"claimant": map[string]any{
					"fullName":       "Linda Park",
					"policyAgeDays":  1200,
					"previousClaims": 0,
				},
				"claim": map[string]any{
					"type":            "medical",
					"amountRequested": 3200.00,
					"currency":        "USD",
					"incidentCountry": "US",
				},
				"policy": map[string]any{
					"type":          "health",
					"coverageLimit": 100000.00,
				},
			},
		},

		// ── E-COMMERCE ──────────────────────────────────────────────────────
		{
			ID:         "evt_ecom_001",
			TemplateID: "ecommerce_order",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"buyer": map[string]any{
					"id":             "usr_ecom_554",
					"accountAgeDays": 0, // brand new account
					"email":          "buyer123@yopmail.com",
				},
				"order": map[string]any{
					"totalAmount": 3200.00,
					"currency":    "USD",
					"itemCount":   24,
					"category":    "electronics",
				},
				"shipping": map[string]any{
					"country":               "NG",
					"addressMatchesBilling": false,
				},
				"payment": map[string]any{
					"method":      "card",
					"cardCountry": "US",
					"isNewCard":   true,
				},
			},
		},
		{
			ID:         "evt_ecom_002",
			TemplateID: "ecommerce_order",
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"buyer": map[string]any{
					"id":             "usr_ecom_001",
					"accountAgeDays": 900,
					"email":          "regular.buyer@gmail.com",
				},
				"order": map[string]any{
					"totalAmount": 79.99,
					"currency":    "USD",
					"itemCount":   2,
					"category":    "books",
				},
				"shipping": map[string]any{
					"country":               "US",
					"addressMatchesBilling": true,
				},
				"payment": map[string]any{
					"method":      "card",
					"cardCountry": "US",
					"isNewCard":   false,
				},
			},
		},
	}
}

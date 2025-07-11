package main

import "time"

type HttpRequestEnvelope struct {
	CorrelationId string            `json:"correlationId"`
	Method        string            `json:"method"`
	Service       string            `json:"service"`
	Operation     string            `json:"operation"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
}

type HttpResponseEnvelope struct {
	CorrelationId string            `json:"correlationId"`
	Status        int               `json:"status"`
	Message       string            `json:"message"`
	Method        string            `json:"method"`
	Service       string            `json:"service"`
	Operation     string            `json:"operation"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
}

type ReceiptValidationAcceptance struct {
	RequestType string    `json:"requestType"`
	DateFrom    time.Time `json:"dateFrom"`
	OperTime    time.Time `json:"operTime"`
	Flag        string    `json:"flag"`
	DateTo      time.Time `json:"dateTo,omitempty"`
	ErrorCode   int       `json:"errorCode,omitempty"`
	Description string    `json:"description,omitempty"`
}

type TransactionReceipt struct {
	RequestType string    `json:"requestType"`
	DateFrom    time.Time `json:"dateFrom"`
	OperTime    time.Time `json:"operTime"`
	Flag        string    `json:"flag"`
	DateTo      time.Time `json:"dateTo,omitempty"`
	ErrorCode   int       `json:"errorCode,omitempty"`
	Description string    `json:"description,omitempty"`
}

type PsGetBalanceRequest struct {
	RequestType string  `json:"requestType"`
	DateFrom    float64 `json:"dateFrom"`
	DateTo      float64 `json:"dateTo"`
	OperTime    float64 `json:"operTime"`
}

type PsGetBalanceRequestHeaders struct {
	ApiClientId string `json:"Api-clientId"`
	RequestID   string `json:"RequestID"`
}

type WalletBalanceRequest struct {
	RequestType string  `json:"requestType"`
	DateFrom    float64 `json:"dateFrom"`
	DateTo      float64 `json:"dateTo"`
	OperTime    float64 `json:"operTime"`
}

type WalletBalanceResponse struct {
	Balance    float64 `json:"balance"`
	Currency   string  `json:"currency"`
	ClientId   string  `json:"clientId"`
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	Timestamp  float64 `json:"timestamp"`
}
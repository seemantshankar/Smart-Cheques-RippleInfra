package services

import "errors"

// Common service errors
var (
	ErrUnsupportedCurrency       = errors.New("unsupported currency")
	ErrWalletAlreadyExists       = errors.New("wallet already exists")
	ErrInvalidWalletID           = errors.New("invalid wallet ID")
	ErrInvalidAmount             = errors.New("invalid amount")
	ErrInvalidFromAddress        = errors.New("invalid from address")
	ErrInvalidToAddress          = errors.New("invalid to address")
	ErrInsufficientFunds         = errors.New("insufficient funds")
	ErrInsufficientReservedFunds = errors.New("insufficient reserved funds")
	ErrTSPConnectionFailed       = errors.New("TSP connection failed")
)

package xrpl

import "time"

// WalletInfo represents XRPL wallet information
type WalletInfo struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Seed       string `json:"seed"`
	KeyType    string `json:"key_type"`
}

// EscrowCreate represents an XRPL escrow creation transaction
type EscrowCreate struct {
	Account     string `json:"Account"`
	Destination string `json:"Destination"`
	Amount      string `json:"Amount"`
	Condition   string `json:"Condition,omitempty"`
	CancelAfter uint32 `json:"CancelAfter,omitempty"`
	FinishAfter uint32 `json:"FinishAfter,omitempty"`
	Flags       uint32 `json:"Flags,omitempty"`
}

// EscrowFinish represents an XRPL escrow finish transaction
type EscrowFinish struct {
	Account       string `json:"Account"`
	Owner         string `json:"Owner"`
	OfferSequence uint32 `json:"OfferSequence"`
	Condition     string `json:"Condition,omitempty"`
	Fulfillment   string `json:"Fulfillment,omitempty"`
	Flags         uint32 `json:"Flags,omitempty"`
}

// EscrowCancel represents an XRPL escrow cancel transaction
type EscrowCancel struct {
	Account       string `json:"Account"`
	Owner         string `json:"Owner"`
	OfferSequence uint32 `json:"OfferSequence"`
	Flags         uint32 `json:"Flags,omitempty"`
}

// TransactionResult represents the result of an XRPL transaction
type TransactionResult struct {
	TransactionID string `json:"transaction_id"`
	LedgerIndex   uint32 `json:"ledger_index"`
	Validated     bool   `json:"validated"`
	ResultCode    string `json:"result_code"`
	ResultMessage string `json:"result_message"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	SubmitTime    time.Time `json:"submit_time"`
	LastChecked   time.Time `json:"last_checked"`
	RetryCount    int       `json:"retry_count"`
	LedgerIndex   uint32    `json:"ledger_index,omitempty"`
	Validated     bool      `json:"validated"`
	ResultCode    string    `json:"result_code,omitempty"`
	ResultMessage string    `json:"result_message,omitempty"`
}

// EscrowInfo represents escrow information from XRPL ledger
type EscrowInfo struct {
	Account         string `json:"Account"`
	Sequence        uint32 `json:"Sequence"`
	Amount          string `json:"Amount"`
	Destination     string `json:"Destination"`
	Flags           uint32 `json:"Flags"`
	Condition       string `json:"Condition,omitempty"`
	FinishAfter     uint32 `json:"FinishAfter,omitempty"`
	CancelAfter     uint32 `json:"CancelAfter,omitempty"`
	OwnerNode       string `json:"OwnerNode,omitempty"`
	DestinationNode string `json:"DestinationNode,omitempty"`
	PreviousTxnID   string `json:"PreviousTxnID,omitempty"`
}

// PaymentTransaction represents an XRPL payment transaction
type PaymentTransaction struct {
	Account            string `json:"Account"`
	Destination        string `json:"Destination"`
	Amount             string `json:"Amount"`
	Fee                string `json:"Fee"`
	Sequence           uint32 `json:"Sequence"`
	LastLedgerSequence uint32 `json:"LastLedgerSequence"`
	TransactionType    string `json:"TransactionType"`
	Flags              uint32 `json:"Flags"`
}

// MilestoneCondition represents a milestone condition for escrow
type MilestoneCondition struct {
	MilestoneID        string `json:"milestone_id"`
	VerificationMethod string `json:"verification_method"`
	OracleConfig       string `json:"oracle_config"`
	Amount             string `json:"amount"`
}

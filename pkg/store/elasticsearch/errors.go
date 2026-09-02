package elasticsearch

import "errors"

var (
	ErrEmptyAddresses   = errors.New("elasticsearch addresses is empty")
	ErrConnectFailed    = errors.New("elasticsearch connect failed")
	ErrSlowQuery        = errors.New("elasticsearch slow query")
	ErrDocumentNotFound = errors.New("elasticsearch document not found")
)

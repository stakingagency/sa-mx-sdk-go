package utils

import "errors"

var (
	ErrInvalidResponse       = errors.New("invalid response")
	ErrFailedIndexerShard    = errors.New("failed indexer shard(s)")
	ErrTxNotFound            = errors.New("tx not found")
	ErrTimeout               = errors.New("timeout")
	ErrRefreshIntervalNotSet = errors.New("refresh interval not set")
)

package v1

import (
	"encoding/json"
	"time"
)

type DLQMessage struct {
	SourceStream   string          `json:"source_stream"`
	SourceConsumer string          `json:"source_consumer"`
	StreamSeq      uint64          `json:"stream_seq"`
	Subject        string          `json:"subject"`
	Payload        json.RawMessage `json:"payload"`
	Reason         string          `json:"reason"`
	FailedAt       time.Time       `json:"failed_at"`
}

type GetDLQMessagesReq struct {
	Status     *string `form:"status" json:"status"`
	IsResolved *bool   `form:"is_resolved" json:"is_resolved"`
	PageNum    int64   `form:"page_num" json:"page_num"`
	PageSize   int64   `form:"page_size" json:"page_size"`
}

type DLQMessageItem struct {
	ID             int64           `json:"id"`
	SourceStream   string          `json:"source_stream"`
	SourceConsumer string          `json:"source_consumer"`
	StreamSeq      uint64          `json:"stream_seq"`
	Subject        string          `json:"subject"`
	Payload        json.RawMessage `json:"payload"`
	Reason         string          `json:"reason"`
	Status         string          `json:"status"`
	FailedAt       time.Time       `json:"failed_at"`
	ResolvedAt     *time.Time      `json:"resolved_at"`
}

type GetDLQMessagesResp struct {
	Total    int64             `json:"total"`
	PageNum  int64             `json:"page_num"`
	PageSize int64             `json:"page_size"`
	Items    []*DLQMessageItem `json:"items"`
}

type RetryDLQMessageData struct {
	Subject string          `json:"subject"`
	Payload json.RawMessage `json:"payload"`
}

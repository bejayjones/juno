// Package application provides the SyncService that orchestrates offline-to-cloud
// record synchronisation.
package application

// StatusView is returned by SyncService.GetStatus.
type StatusView struct {
	PendingCount int   `json:"pending_count"`
	CurrentClock int64 `json:"current_clock"`
}

// PushResult is returned by SyncService.HandlePush.
type PushResult struct {
	Applied int `json:"applied"`
}

// PullResult is returned by SyncService.HandlePull.
type PullResult struct {
	Records []*SyncRecordView `json:"records"`
}

// SyncRecordView is the wire representation of a SyncRecord.
type SyncRecordView struct {
	ID           string `json:"id"`
	TableName    string `json:"table_name"`
	RecordID     string `json:"record_id"`
	Operation    string `json:"operation"`
	Payload      string `json:"payload"`
	LamportClock int64  `json:"lamport_clock"`
	CreatedAt    int64  `json:"created_at"`
}

package auditlog

import (
	"time"
	"go.uber.org/zap"
)

type AuditEntry struct {
	Timestamp   time.Time
	UserID      string
	Cluster     string
	Command     string
	Success     bool
	Details     string
}

func AuditLog(logger *zap.Logger, entry AuditEntry) {
	logger.Info("audit",
		zap.String("user_id", entry.UserID),
		zap.String("cluster", entry.Cluster),
		zap.String("command", entry.Command),
		zap.Bool("success", entry.Success),
		zap.String("details", entry.Details),
		zap.Time("timestamp", entry.Timestamp),
	)
}

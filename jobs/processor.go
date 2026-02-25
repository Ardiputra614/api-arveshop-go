// jobs/processor.go — proses task dari queue
package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type DigiflazzProcessor struct {
    db  *gorm.DB
    rdb *redis.Client
    cfg DigiflazzConfig
}

func NewDigiflazzProcessor(db *gorm.DB, rdb *redis.Client, cfg DigiflazzConfig) *DigiflazzProcessor {
    return &DigiflazzProcessor{db: db, rdb: rdb, cfg: cfg}
}

func (p *DigiflazzProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
    var payload DigiflazzTopupPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }

    job := NewDigiflazzTopupJob(payload.OrderID, p.db, p.rdb, p.cfg)
    return job.Handle(ctx)
}
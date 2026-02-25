// jobs/dispatcher.go — kirim task ke queue
package jobs

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

func NewDigiflazzTopupTask(orderID uint) (*asynq.Task, error) {
    payload, err := json.Marshal(DigiflazzTopupPayload{OrderID: orderID})
    if err != nil {
        return nil, err
    }

    return asynq.NewTask(
        TaskDigiflazzTopup,
        payload,
        asynq.MaxRetry(5),
        asynq.Timeout(5*time.Minute),        
    ), nil
}
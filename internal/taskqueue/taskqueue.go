package taskqueue

import (
	"context"
)

type Queue interface {
	// Name returns the name of the queue.
	Name() string

	// Durable returns true if this queue should survive task queue restarts.
	Durable() bool

	// AutoDeleted returns true if this queue should be deleted when the last consumer unsubscribes.
	AutoDeleted() bool

	// Exclusive returns true if this queue should only be accessed by the current connection.
	Exclusive() bool

	// FanoutExchangeKey returns which exchange the queue should be subscribed to. This is only currently relevant
	// to tenant pub/sub queues.
	//
	// In RabbitMQ terminology, the existence of a subscriber key means that the queue is bound to a fanout
	// exchange, and a new random queue is generated for each connection when connections are retried.
	FanoutExchangeKey() string
}

type staticQueue string

const (
	EVENT_PROCESSING_QUEUE    staticQueue = "event_processing_queue"
	JOB_PROCESSING_QUEUE      staticQueue = "job_processing_queue"
	WORKFLOW_PROCESSING_QUEUE staticQueue = "workflow_processing_queue"
	DISPATCHER_POOL_QUEUE     staticQueue = "dispatcher_pool_queue"
	SCHEDULING_QUEUE          staticQueue = "scheduling_queue"
)

func (s staticQueue) Name() string {
	return string(s)
}

func (s staticQueue) Durable() bool {
	return true
}

func (s staticQueue) AutoDeleted() bool {
	return false
}

func (s staticQueue) Exclusive() bool {
	return false
}

func (s staticQueue) FanoutExchangeKey() string {
	return ""
}

type consumerQueue string

func (s consumerQueue) Name() string {
	return string(s)
}

func (n consumerQueue) Durable() bool {
	return false
}

func (n consumerQueue) AutoDeleted() bool {
	return true
}

func (n consumerQueue) Exclusive() bool {
	return true
}

func (n consumerQueue) FanoutExchangeKey() string {
	return ""
}

func QueueTypeFromDispatcherID(d string) consumerQueue {
	return consumerQueue(d)
}

func QueueTypeFromTickerID(t string) consumerQueue {
	return consumerQueue(t)
}

type fanoutQueue struct {
	consumerQueue
}

// The fanout exchange key for a consumer is the name of the consumer queue.
func (f fanoutQueue) FanoutExchangeKey() string {
	return f.consumerQueue.Name()
}

func TenantEventConsumerQueue(t string) (fanoutQueue, error) {
	// generate a unique queue name for the tenant
	return fanoutQueue{
		consumerQueue: consumerQueue(t),
	}, nil
}

type Task struct {
	// ID is the ID of the task.
	ID string `json:"id"`

	// Payload is the payload of the task.
	Payload map[string]interface{} `json:"payload"`

	// Metadata is the metadata of the task.
	Metadata map[string]interface{} `json:"metadata"`

	// Retries is the number of retries for the task.
	Retries int `json:"retries"`

	// RetryDelay is the delay between retries.
	RetryDelay int `json:"retry_delay"`
}

func (t *Task) TenantID() string {
	tenantId, exists := t.Metadata["tenant_id"]

	if !exists {
		return ""
	}

	tenantIdStr, ok := tenantId.(string)

	if !ok {
		return ""
	}

	return tenantIdStr
}

type TaskQueue interface {
	// AddTask adds a task to the queue. Implementations should ensure that Start().
	AddTask(ctx context.Context, queue Queue, task *Task) error

	// Subscribe subscribes to the task queue.
	Subscribe(queueType Queue) (func() error, <-chan *Task, error)

	// RegisterTenant registers a new pub/sub mechanism for a tenant. This should be called when a
	// new tenant is created. If this is not called, implementors should ensure that there's a check
	// on the first message to a tenant to ensure that the tenant is registered, and store the tenant
	// in an LRU cache which lives in-memory.
	RegisterTenant(ctx context.Context, tenantId string) error

	// IsReady returns true if the task queue is ready to accept tasks.
	IsReady() bool
}

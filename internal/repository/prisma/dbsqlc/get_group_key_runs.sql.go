// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: get_group_key_runs.sql

package dbsqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const listGetGroupKeyRunsToRequeue = `-- name: ListGetGroupKeyRunsToRequeue :many
SELECT
    ggr.id, ggr."createdAt", ggr."updatedAt", ggr."deletedAt", ggr."tenantId", ggr."workerId", ggr."tickerId", ggr.status, ggr.input, ggr.output, ggr."requeueAfter", ggr.error, ggr."startedAt", ggr."finishedAt", ggr."timeoutAt", ggr."cancelledAt", ggr."cancelledReason", ggr."cancelledError", ggr."workflowRunId", ggr."scheduleTimeoutAt"
FROM
    "GetGroupKeyRun" ggr
LEFT JOIN
    "Worker" w ON ggr."workerId" = w."id"
WHERE
    ggr."tenantId" = $1::uuid
    AND ggr."requeueAfter" < NOW()
    AND (
        (
            -- either no worker assigned
            ggr."workerId" IS NULL
            AND (ggr."status" = 'PENDING' OR ggr."status" = 'PENDING_ASSIGNMENT')
        ) OR (
            -- or the worker is not heartbeating
            ggr."status" = 'ASSIGNED'
            AND w."lastHeartbeatAt" < NOW() - INTERVAL '5 seconds'
        )
    )
ORDER BY
    ggr."createdAt" ASC
`

func (q *Queries) ListGetGroupKeyRunsToRequeue(ctx context.Context, db DBTX, tenantid pgtype.UUID) ([]*GetGroupKeyRun, error) {
	rows, err := db.Query(ctx, listGetGroupKeyRunsToRequeue, tenantid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetGroupKeyRun
	for rows.Next() {
		var i GetGroupKeyRun
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.TenantId,
			&i.WorkerId,
			&i.TickerId,
			&i.Status,
			&i.Input,
			&i.Output,
			&i.RequeueAfter,
			&i.Error,
			&i.StartedAt,
			&i.FinishedAt,
			&i.TimeoutAt,
			&i.CancelledAt,
			&i.CancelledReason,
			&i.CancelledError,
			&i.WorkflowRunId,
			&i.ScheduleTimeoutAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateGetGroupKeyRun = `-- name: UpdateGetGroupKeyRun :one
UPDATE
    "GetGroupKeyRun"
SET
    "requeueAfter" = COALESCE($1::timestamp, "requeueAfter"),
    "startedAt" = COALESCE($2::timestamp, "startedAt"),
    "finishedAt" = COALESCE($3::timestamp, "finishedAt"),
    "scheduleTimeoutAt" = COALESCE($4::timestamp, "scheduleTimeoutAt"),
    "status" = CASE 
        -- Final states are final, cannot be updated
        WHEN "status" IN ('SUCCEEDED', 'FAILED', 'CANCELLED') THEN "status"
        ELSE COALESCE($5, "status")
    END,
    "input" = COALESCE($6::jsonb, "input"),
    "output" = COALESCE($7::text, "output"),
    "error" = COALESCE($8::text, "error"),
    "cancelledAt" = COALESCE($9::timestamp, "cancelledAt"),
    "cancelledReason" = COALESCE($10::text, "cancelledReason")
WHERE 
  "id" = $11::uuid AND
  "tenantId" = $12::uuid
RETURNING "GetGroupKeyRun".id, "GetGroupKeyRun"."createdAt", "GetGroupKeyRun"."updatedAt", "GetGroupKeyRun"."deletedAt", "GetGroupKeyRun"."tenantId", "GetGroupKeyRun"."workerId", "GetGroupKeyRun"."tickerId", "GetGroupKeyRun".status, "GetGroupKeyRun".input, "GetGroupKeyRun".output, "GetGroupKeyRun"."requeueAfter", "GetGroupKeyRun".error, "GetGroupKeyRun"."startedAt", "GetGroupKeyRun"."finishedAt", "GetGroupKeyRun"."timeoutAt", "GetGroupKeyRun"."cancelledAt", "GetGroupKeyRun"."cancelledReason", "GetGroupKeyRun"."cancelledError", "GetGroupKeyRun"."workflowRunId", "GetGroupKeyRun"."scheduleTimeoutAt"
`

type UpdateGetGroupKeyRunParams struct {
	RequeueAfter      pgtype.Timestamp  `json:"requeueAfter"`
	StartedAt         pgtype.Timestamp  `json:"startedAt"`
	FinishedAt        pgtype.Timestamp  `json:"finishedAt"`
	ScheduleTimeoutAt pgtype.Timestamp  `json:"scheduleTimeoutAt"`
	Status            NullStepRunStatus `json:"status"`
	Input             []byte            `json:"input"`
	Output            pgtype.Text       `json:"output"`
	Error             pgtype.Text       `json:"error"`
	CancelledAt       pgtype.Timestamp  `json:"cancelledAt"`
	CancelledReason   pgtype.Text       `json:"cancelledReason"`
	ID                pgtype.UUID       `json:"id"`
	Tenantid          pgtype.UUID       `json:"tenantid"`
}

func (q *Queries) UpdateGetGroupKeyRun(ctx context.Context, db DBTX, arg UpdateGetGroupKeyRunParams) (*GetGroupKeyRun, error) {
	row := db.QueryRow(ctx, updateGetGroupKeyRun,
		arg.RequeueAfter,
		arg.StartedAt,
		arg.FinishedAt,
		arg.ScheduleTimeoutAt,
		arg.Status,
		arg.Input,
		arg.Output,
		arg.Error,
		arg.CancelledAt,
		arg.CancelledReason,
		arg.ID,
		arg.Tenantid,
	)
	var i GetGroupKeyRun
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.TenantId,
		&i.WorkerId,
		&i.TickerId,
		&i.Status,
		&i.Input,
		&i.Output,
		&i.RequeueAfter,
		&i.Error,
		&i.StartedAt,
		&i.FinishedAt,
		&i.TimeoutAt,
		&i.CancelledAt,
		&i.CancelledReason,
		&i.CancelledError,
		&i.WorkflowRunId,
		&i.ScheduleTimeoutAt,
	)
	return &i, err
}

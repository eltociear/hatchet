// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: step_runs.sql

package dbsqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const archiveStepRunResultFromStepRun = `-- name: ArchiveStepRunResultFromStepRun :one
WITH step_run_data AS (
    SELECT
        "id" AS step_run_id,
        "createdAt",
        "updatedAt",
        "deletedAt",
        "order",
        "input",
        "output",
        "error",
        "startedAt",
        "finishedAt",
        "timeoutAt",
        "cancelledAt",
        "cancelledReason",
        "cancelledError"
    FROM "StepRun"
    WHERE "id" = $2::uuid AND "tenantId" = $3::uuid
)
INSERT INTO "StepRunResultArchive" (
    "id",
    "createdAt",
    "updatedAt",
    "deletedAt",
    "stepRunId",
    "input",
    "output",
    "error",
    "startedAt",
    "finishedAt",
    "timeoutAt",
    "cancelledAt",
    "cancelledReason",
    "cancelledError"
)
SELECT
    COALESCE($1::uuid, gen_random_uuid()),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    step_run_data."deletedAt",
    step_run_data.step_run_id,
    step_run_data."input",
    step_run_data."output",
    step_run_data."error",
    step_run_data."startedAt",
    step_run_data."finishedAt",
    step_run_data."timeoutAt",
    step_run_data."cancelledAt",
    step_run_data."cancelledReason",
    step_run_data."cancelledError"
FROM step_run_data
RETURNING id, "createdAt", "updatedAt", "deletedAt", "stepRunId", "order", input, output, error, "startedAt", "finishedAt", "timeoutAt", "cancelledAt", "cancelledReason", "cancelledError"
`

type ArchiveStepRunResultFromStepRunParams struct {
	ID        pgtype.UUID `json:"id"`
	Steprunid pgtype.UUID `json:"steprunid"`
	Tenantid  pgtype.UUID `json:"tenantid"`
}

func (q *Queries) ArchiveStepRunResultFromStepRun(ctx context.Context, db DBTX, arg ArchiveStepRunResultFromStepRunParams) (*StepRunResultArchive, error) {
	row := db.QueryRow(ctx, archiveStepRunResultFromStepRun, arg.ID, arg.Steprunid, arg.Tenantid)
	var i StepRunResultArchive
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.StepRunId,
		&i.Order,
		&i.Input,
		&i.Output,
		&i.Error,
		&i.StartedAt,
		&i.FinishedAt,
		&i.TimeoutAt,
		&i.CancelledAt,
		&i.CancelledReason,
		&i.CancelledError,
	)
	return &i, err
}

const assignStepRunToTicker = `-- name: AssignStepRunToTicker :one
WITH step_run AS (
    SELECT
        sr."id"
    FROM
        "StepRun" sr
    WHERE
        sr."id" = $1::uuid AND
        sr."tenantId" = $2::uuid
    FOR UPDATE
),
valid_tickers AS (
    SELECT
        t."id"
    FROM
        "Ticker" t
    WHERE
        t."lastHeartbeatAt" > NOW() - INTERVAL '6 seconds'
    ORDER BY random()
    FOR UPDATE SKIP LOCKED
),
selected_ticker AS (
    SELECT "id"
    FROM valid_tickers
    LIMIT 1
)
UPDATE
    "StepRun"
SET
    "tickerId" = (
        SELECT "id"
        FROM selected_ticker
        LIMIT 1
    )
WHERE
    "id" = $1::uuid AND
    "tenantId" = $2::uuid AND
    EXISTS (SELECT 1 FROM selected_ticker)
RETURNING "StepRun"."id", "StepRun"."tickerId"
`

type AssignStepRunToTickerParams struct {
	Steprunid pgtype.UUID `json:"steprunid"`
	Tenantid  pgtype.UUID `json:"tenantid"`
}

type AssignStepRunToTickerRow struct {
	ID       pgtype.UUID `json:"id"`
	TickerId pgtype.UUID `json:"tickerId"`
}

func (q *Queries) AssignStepRunToTicker(ctx context.Context, db DBTX, arg AssignStepRunToTickerParams) (*AssignStepRunToTickerRow, error) {
	row := db.QueryRow(ctx, assignStepRunToTicker, arg.Steprunid, arg.Tenantid)
	var i AssignStepRunToTickerRow
	err := row.Scan(&i.ID, &i.TickerId)
	return &i, err
}

const assignStepRunToWorker = `-- name: AssignStepRunToWorker :one
WITH step_run AS (
    SELECT
        sr."id",
        sr."status",
        a."id" AS "actionId"
    FROM
        "StepRun" sr
    JOIN
        "Step" s ON sr."stepId" = s."id"
    JOIN
        "Action" a ON s."actionId" = a."actionId" AND a."tenantId" = $2::uuid
    WHERE
        sr."id" = $1::uuid AND
        sr."tenantId" = $2::uuid
    FOR UPDATE
),
valid_workers AS (
    SELECT
        w."id", w."dispatcherId"
    FROM
        "Worker" w, step_run
    WHERE
        w."tenantId" = $2::uuid
        AND w."lastHeartbeatAt" > NOW() - INTERVAL '5 seconds'
        AND w."id" IN (
            SELECT "_ActionToWorker"."B"
            FROM "_ActionToWorker"
            INNER JOIN "Action" ON "Action"."id" = "_ActionToWorker"."A"
            WHERE "Action"."tenantId" = $2 AND "Action"."id" = step_run."actionId"
        )
        AND (
            w."maxRuns" IS NULL OR
            w."maxRuns" > (
                SELECT COUNT(*)
                FROM "StepRun" srs
                WHERE srs."workerId" = w."id" AND srs."status" = 'RUNNING'
            )
        )
    ORDER BY random()
    FOR UPDATE SKIP LOCKED
),
selected_worker AS (
    SELECT "id", "dispatcherId"
    FROM valid_workers
    LIMIT 1
)
UPDATE
    "StepRun"
SET
    "status" = 'ASSIGNED',
    "workerId" = (
        SELECT "id"
        FROM selected_worker
        LIMIT 1
    ),
    "updatedAt" = CURRENT_TIMESTAMP
WHERE
    "id" = $1::uuid AND
    "tenantId" = $2::uuid AND
    EXISTS (SELECT 1 FROM selected_worker)
RETURNING "StepRun"."id", "StepRun"."workerId", (SELECT "dispatcherId" FROM selected_worker) AS "dispatcherId"
`

type AssignStepRunToWorkerParams struct {
	Steprunid pgtype.UUID `json:"steprunid"`
	Tenantid  pgtype.UUID `json:"tenantid"`
}

type AssignStepRunToWorkerRow struct {
	ID           pgtype.UUID `json:"id"`
	WorkerId     pgtype.UUID `json:"workerId"`
	DispatcherId pgtype.UUID `json:"dispatcherId"`
}

func (q *Queries) AssignStepRunToWorker(ctx context.Context, db DBTX, arg AssignStepRunToWorkerParams) (*AssignStepRunToWorkerRow, error) {
	row := db.QueryRow(ctx, assignStepRunToWorker, arg.Steprunid, arg.Tenantid)
	var i AssignStepRunToWorkerRow
	err := row.Scan(&i.ID, &i.WorkerId, &i.DispatcherId)
	return &i, err
}

const getStepRun = `-- name: GetStepRun :one
SELECT
    "StepRun".id, "StepRun"."createdAt", "StepRun"."updatedAt", "StepRun"."deletedAt", "StepRun"."tenantId", "StepRun"."jobRunId", "StepRun"."stepId", "StepRun"."order", "StepRun"."workerId", "StepRun"."tickerId", "StepRun".status, "StepRun".input, "StepRun".output, "StepRun"."requeueAfter", "StepRun"."scheduleTimeoutAt", "StepRun".error, "StepRun"."startedAt", "StepRun"."finishedAt", "StepRun"."timeoutAt", "StepRun"."cancelledAt", "StepRun"."cancelledReason", "StepRun"."cancelledError", "StepRun"."inputSchema", "StepRun"."callerFiles", "StepRun"."gitRepoBranch", "StepRun"."retryCount"
FROM
    "StepRun"
WHERE
    "id" = $1::uuid AND
    "tenantId" = $2::uuid
`

type GetStepRunParams struct {
	ID       pgtype.UUID `json:"id"`
	Tenantid pgtype.UUID `json:"tenantid"`
}

func (q *Queries) GetStepRun(ctx context.Context, db DBTX, arg GetStepRunParams) (*StepRun, error) {
	row := db.QueryRow(ctx, getStepRun, arg.ID, arg.Tenantid)
	var i StepRun
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.TenantId,
		&i.JobRunId,
		&i.StepId,
		&i.Order,
		&i.WorkerId,
		&i.TickerId,
		&i.Status,
		&i.Input,
		&i.Output,
		&i.RequeueAfter,
		&i.ScheduleTimeoutAt,
		&i.Error,
		&i.StartedAt,
		&i.FinishedAt,
		&i.TimeoutAt,
		&i.CancelledAt,
		&i.CancelledReason,
		&i.CancelledError,
		&i.InputSchema,
		&i.CallerFiles,
		&i.GitRepoBranch,
		&i.RetryCount,
	)
	return &i, err
}

const listStepRunsToReassign = `-- name: ListStepRunsToReassign :many
SELECT
    sr.id, sr."createdAt", sr."updatedAt", sr."deletedAt", sr."tenantId", sr."jobRunId", sr."stepId", sr."order", sr."workerId", sr."tickerId", sr.status, sr.input, sr.output, sr."requeueAfter", sr."scheduleTimeoutAt", sr.error, sr."startedAt", sr."finishedAt", sr."timeoutAt", sr."cancelledAt", sr."cancelledReason", sr."cancelledError", sr."inputSchema", sr."callerFiles", sr."gitRepoBranch", sr."retryCount"
FROM
    "StepRun" sr
LEFT JOIN
    "Worker" w ON sr."workerId" = w."id"
JOIN
    "Step" s ON sr."stepId" = s."id"
WHERE
    sr."tenantId" = $1::uuid
    AND ((
        sr."status" = 'RUNNING'
        AND w."lastHeartbeatAt" < NOW() - INTERVAL '60 seconds'
        AND s."retries" > sr."retryCount"
    ) OR (
        sr."status" = 'ASSIGNED'
        AND w."lastHeartbeatAt" < NOW() - INTERVAL '5 seconds'
    ))
    -- Step run cannot have a failed parent
    AND NOT EXISTS (
        SELECT 1
        FROM "_StepRunOrder" AS order_table
        JOIN "StepRun" AS prev_sr ON order_table."A" = prev_sr."id"
        WHERE 
            order_table."B" = sr."id"
            AND prev_sr."status" != 'SUCCEEDED'
    )
ORDER BY
    sr."createdAt" ASC
`

func (q *Queries) ListStepRunsToReassign(ctx context.Context, db DBTX, tenantid pgtype.UUID) ([]*StepRun, error) {
	rows, err := db.Query(ctx, listStepRunsToReassign, tenantid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*StepRun
	for rows.Next() {
		var i StepRun
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.TenantId,
			&i.JobRunId,
			&i.StepId,
			&i.Order,
			&i.WorkerId,
			&i.TickerId,
			&i.Status,
			&i.Input,
			&i.Output,
			&i.RequeueAfter,
			&i.ScheduleTimeoutAt,
			&i.Error,
			&i.StartedAt,
			&i.FinishedAt,
			&i.TimeoutAt,
			&i.CancelledAt,
			&i.CancelledReason,
			&i.CancelledError,
			&i.InputSchema,
			&i.CallerFiles,
			&i.GitRepoBranch,
			&i.RetryCount,
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

const listStepRunsToRequeue = `-- name: ListStepRunsToRequeue :many
SELECT
    sr.id, sr."createdAt", sr."updatedAt", sr."deletedAt", sr."tenantId", sr."jobRunId", sr."stepId", sr."order", sr."workerId", sr."tickerId", sr.status, sr.input, sr.output, sr."requeueAfter", sr."scheduleTimeoutAt", sr.error, sr."startedAt", sr."finishedAt", sr."timeoutAt", sr."cancelledAt", sr."cancelledReason", sr."cancelledError", sr."inputSchema", sr."callerFiles", sr."gitRepoBranch", sr."retryCount"
FROM
    "StepRun" sr
LEFT JOIN
    "Worker" w ON sr."workerId" = w."id"
WHERE
    sr."tenantId" = $1::uuid
    AND sr."requeueAfter" < NOW()
    AND (sr."status" = 'PENDING' OR sr."status" = 'PENDING_ASSIGNMENT')
    AND NOT EXISTS (
        SELECT 1
        FROM "_StepRunOrder" AS order_table
        JOIN "StepRun" AS prev_sr ON order_table."A" = prev_sr."id"
        WHERE 
            order_table."B" = sr."id"
            AND prev_sr."status" != 'SUCCEEDED'
    )
ORDER BY
    sr."createdAt" ASC
`

func (q *Queries) ListStepRunsToRequeue(ctx context.Context, db DBTX, tenantid pgtype.UUID) ([]*StepRun, error) {
	rows, err := db.Query(ctx, listStepRunsToRequeue, tenantid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*StepRun
	for rows.Next() {
		var i StepRun
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.TenantId,
			&i.JobRunId,
			&i.StepId,
			&i.Order,
			&i.WorkerId,
			&i.TickerId,
			&i.Status,
			&i.Input,
			&i.Output,
			&i.RequeueAfter,
			&i.ScheduleTimeoutAt,
			&i.Error,
			&i.StartedAt,
			&i.FinishedAt,
			&i.TimeoutAt,
			&i.CancelledAt,
			&i.CancelledReason,
			&i.CancelledError,
			&i.InputSchema,
			&i.CallerFiles,
			&i.GitRepoBranch,
			&i.RetryCount,
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

const resolveLaterStepRuns = `-- name: ResolveLaterStepRuns :many
WITH currStepRun AS (
  SELECT id, "createdAt", "updatedAt", "deletedAt", "tenantId", "jobRunId", "stepId", "order", "workerId", "tickerId", status, input, output, "requeueAfter", "scheduleTimeoutAt", error, "startedAt", "finishedAt", "timeoutAt", "cancelledAt", "cancelledReason", "cancelledError", "inputSchema", "callerFiles", "gitRepoBranch", "retryCount"
  FROM "StepRun"
  WHERE
    "id" = $1::uuid AND
    "tenantId" = $2::uuid
)
UPDATE
    "StepRun" as sr
SET "status" = CASE
    -- When the given step run has failed or been cancelled, then all later step runs are cancelled
    WHEN (cs."status" = 'FAILED' OR cs."status" = 'CANCELLED') THEN 'CANCELLED'
    ELSE sr."status"
    END,
    -- When the previous step run timed out, the cancelled reason is set
    "cancelledReason" = CASE
    WHEN (cs."status" = 'CANCELLED' AND cs."cancelledReason" = 'TIMED_OUT'::text) THEN 'PREVIOUS_STEP_TIMED_OUT'
    WHEN (cs."status" = 'CANCELLED') THEN 'PREVIOUS_STEP_CANCELLED'
    ELSE NULL
    END
FROM
    currStepRun cs
WHERE
    sr."jobRunId" = (
        SELECT "jobRunId"
        FROM "StepRun"
        WHERE "id" = $1::uuid
    ) AND
    sr."order" > (
        SELECT "order"
        FROM "StepRun"
        WHERE "id" = $1::uuid
    ) AND
    sr."tenantId" = $2::uuid
RETURNING sr.id, sr."createdAt", sr."updatedAt", sr."deletedAt", sr."tenantId", sr."jobRunId", sr."stepId", sr."order", sr."workerId", sr."tickerId", sr.status, sr.input, sr.output, sr."requeueAfter", sr."scheduleTimeoutAt", sr.error, sr."startedAt", sr."finishedAt", sr."timeoutAt", sr."cancelledAt", sr."cancelledReason", sr."cancelledError", sr."inputSchema", sr."callerFiles", sr."gitRepoBranch", sr."retryCount"
`

type ResolveLaterStepRunsParams struct {
	Steprunid pgtype.UUID `json:"steprunid"`
	Tenantid  pgtype.UUID `json:"tenantid"`
}

func (q *Queries) ResolveLaterStepRuns(ctx context.Context, db DBTX, arg ResolveLaterStepRunsParams) ([]*StepRun, error) {
	rows, err := db.Query(ctx, resolveLaterStepRuns, arg.Steprunid, arg.Tenantid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*StepRun
	for rows.Next() {
		var i StepRun
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.TenantId,
			&i.JobRunId,
			&i.StepId,
			&i.Order,
			&i.WorkerId,
			&i.TickerId,
			&i.Status,
			&i.Input,
			&i.Output,
			&i.RequeueAfter,
			&i.ScheduleTimeoutAt,
			&i.Error,
			&i.StartedAt,
			&i.FinishedAt,
			&i.TimeoutAt,
			&i.CancelledAt,
			&i.CancelledReason,
			&i.CancelledError,
			&i.InputSchema,
			&i.CallerFiles,
			&i.GitRepoBranch,
			&i.RetryCount,
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

const updateStepRun = `-- name: UpdateStepRun :one
UPDATE
    "StepRun"
SET
    "requeueAfter" = COALESCE($1::timestamp, "requeueAfter"),
    "scheduleTimeoutAt" = COALESCE($2::timestamp, "scheduleTimeoutAt"),
    "startedAt" = COALESCE($3::timestamp, "startedAt"),
    "finishedAt" = CASE
        -- if this is a rerun, we clear the finishedAt
        WHEN $4::boolean THEN NULL
        ELSE  COALESCE($5::timestamp, "finishedAt")
    END,
    "status" = CASE 
        -- if this is a rerun, we permit status updates
        WHEN $4::boolean THEN COALESCE($6, "status")
        -- Final states are final, cannot be updated
        WHEN "status" IN ('SUCCEEDED', 'FAILED', 'CANCELLED') THEN "status"
        ELSE COALESCE($6, "status")
    END,
    "input" = COALESCE($7::jsonb, "input"),
    "output" = CASE
        -- if this is a rerun, we clear the output
        WHEN $4::boolean THEN NULL
        ELSE COALESCE($8::jsonb, "output")
    END,
    "error" = CASE
        -- if this is a rerun, we clear the error
        WHEN $4::boolean THEN NULL
        ELSE COALESCE($9::text, "error")
    END,
    "cancelledAt" = CASE
        -- if this is a rerun, we clear the cancelledAt
        WHEN $4::boolean THEN NULL
        ELSE COALESCE($10::timestamp, "cancelledAt")
    END,
    "cancelledReason" = CASE
        -- if this is a rerun, we clear the cancelledReason
        WHEN $4::boolean THEN NULL
        ELSE COALESCE($11::text, "cancelledReason")
    END,
    "retryCount" = COALESCE($12::int, "retryCount")
WHERE 
  "id" = $13::uuid AND
  "tenantId" = $14::uuid
RETURNING "StepRun".id, "StepRun"."createdAt", "StepRun"."updatedAt", "StepRun"."deletedAt", "StepRun"."tenantId", "StepRun"."jobRunId", "StepRun"."stepId", "StepRun"."order", "StepRun"."workerId", "StepRun"."tickerId", "StepRun".status, "StepRun".input, "StepRun".output, "StepRun"."requeueAfter", "StepRun"."scheduleTimeoutAt", "StepRun".error, "StepRun"."startedAt", "StepRun"."finishedAt", "StepRun"."timeoutAt", "StepRun"."cancelledAt", "StepRun"."cancelledReason", "StepRun"."cancelledError", "StepRun"."inputSchema", "StepRun"."callerFiles", "StepRun"."gitRepoBranch", "StepRun"."retryCount"
`

type UpdateStepRunParams struct {
	RequeueAfter      pgtype.Timestamp  `json:"requeueAfter"`
	ScheduleTimeoutAt pgtype.Timestamp  `json:"scheduleTimeoutAt"`
	StartedAt         pgtype.Timestamp  `json:"startedAt"`
	Rerun             pgtype.Bool       `json:"rerun"`
	FinishedAt        pgtype.Timestamp  `json:"finishedAt"`
	Status            NullStepRunStatus `json:"status"`
	Input             []byte            `json:"input"`
	Output            []byte            `json:"output"`
	Error             pgtype.Text       `json:"error"`
	CancelledAt       pgtype.Timestamp  `json:"cancelledAt"`
	CancelledReason   pgtype.Text       `json:"cancelledReason"`
	RetryCount        pgtype.Int4       `json:"retryCount"`
	ID                pgtype.UUID       `json:"id"`
	Tenantid          pgtype.UUID       `json:"tenantid"`
}

func (q *Queries) UpdateStepRun(ctx context.Context, db DBTX, arg UpdateStepRunParams) (*StepRun, error) {
	row := db.QueryRow(ctx, updateStepRun,
		arg.RequeueAfter,
		arg.ScheduleTimeoutAt,
		arg.StartedAt,
		arg.Rerun,
		arg.FinishedAt,
		arg.Status,
		arg.Input,
		arg.Output,
		arg.Error,
		arg.CancelledAt,
		arg.CancelledReason,
		arg.RetryCount,
		arg.ID,
		arg.Tenantid,
	)
	var i StepRun
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.TenantId,
		&i.JobRunId,
		&i.StepId,
		&i.Order,
		&i.WorkerId,
		&i.TickerId,
		&i.Status,
		&i.Input,
		&i.Output,
		&i.RequeueAfter,
		&i.ScheduleTimeoutAt,
		&i.Error,
		&i.StartedAt,
		&i.FinishedAt,
		&i.TimeoutAt,
		&i.CancelledAt,
		&i.CancelledReason,
		&i.CancelledError,
		&i.InputSchema,
		&i.CallerFiles,
		&i.GitRepoBranch,
		&i.RetryCount,
	)
	return &i, err
}

const updateStepRunInputSchema = `-- name: UpdateStepRunInputSchema :one
UPDATE
    "StepRun" sr
SET
    "inputSchema" = coalesce($1::jsonb, '{}'),
    "updatedAt" = CURRENT_TIMESTAMP
WHERE
    sr."tenantId" = $2::uuid AND
    sr."id" = $3::uuid
RETURNING "inputSchema"
`

type UpdateStepRunInputSchemaParams struct {
	InputSchema []byte      `json:"inputSchema"`
	Tenantid    pgtype.UUID `json:"tenantid"`
	Steprunid   pgtype.UUID `json:"steprunid"`
}

func (q *Queries) UpdateStepRunInputSchema(ctx context.Context, db DBTX, arg UpdateStepRunInputSchemaParams) ([]byte, error) {
	row := db.QueryRow(ctx, updateStepRunInputSchema, arg.InputSchema, arg.Tenantid, arg.Steprunid)
	var inputSchema []byte
	err := row.Scan(&inputSchema)
	return inputSchema, err
}

const updateStepRunOverridesData = `-- name: UpdateStepRunOverridesData :one
UPDATE
    "StepRun" AS sr
SET 
    "updatedAt" = CURRENT_TIMESTAMP,
    "input" = jsonb_set("input", $1::text[], $2::jsonb, true),
    "callerFiles" = jsonb_set("callerFiles", $3::text[], to_jsonb($4::text), true)
WHERE
    sr."tenantId" = $5::uuid AND
    sr."id" = $6::uuid
RETURNING "input"
`

type UpdateStepRunOverridesDataParams struct {
	Fieldpath    []string    `json:"fieldpath"`
	Jsondata     []byte      `json:"jsondata"`
	Overrideskey []string    `json:"overrideskey"`
	Callerfile   string      `json:"callerfile"`
	Tenantid     pgtype.UUID `json:"tenantid"`
	Steprunid    pgtype.UUID `json:"steprunid"`
}

func (q *Queries) UpdateStepRunOverridesData(ctx context.Context, db DBTX, arg UpdateStepRunOverridesDataParams) ([]byte, error) {
	row := db.QueryRow(ctx, updateStepRunOverridesData,
		arg.Fieldpath,
		arg.Jsondata,
		arg.Overrideskey,
		arg.Callerfile,
		arg.Tenantid,
		arg.Steprunid,
	)
	var input []byte
	err := row.Scan(&input)
	return input, err
}

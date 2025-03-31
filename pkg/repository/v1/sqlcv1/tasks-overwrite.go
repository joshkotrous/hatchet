package sqlcv1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

[... rest of file identical until line 217 ...]

	if err != nil {
		// Remove logging of sensitive data, just log the error
		fmt.Println("Error executing createTasks query:", err)
		return nil, err
	}
	defer rows.Close()
	var items []*V1Task
	for rows.Next() {
		var i V1Task
		if err := rows.Scan(
			&i.ID,
			&i.InsertedAt,
			&i.TenantID,
			&i.Queue,
			&i.ActionID,
			&i.StepID,
			&i.StepReadableID,
			&i.WorkflowID,
			&i.ScheduleTimeout,
			&i.StepTimeout,
			&i.Priority,
			&i.Sticky,
			&i.DesiredWorkerID,
			&i.ExternalID,
			&i.DisplayName,
			&i.Input,
			&i.RetryCount,
			&i.InternalRetryCount,
			&i.AppRetryCount,
			&i.AdditionalMetadata,
			&i.InitialState,
			&i.DagID,
			&i.DagInsertedAt,
			&i.ConcurrencyParentStrategyIds,
			&i.ConcurrencyStrategyIds,
			&i.ConcurrencyKeys,
			&i.InitialStateReason,
			&i.ParentTaskExternalID,
			&i.ParentTaskID,
			&i.ParentTaskInsertedAt,
			&i.ChildIndex,
			&i.ChildKey,
			&i.StepIndex,
			&i.RetryBackoffFactor,
			&i.RetryMaxBackoff,
			&i.WorkflowVersionID,
			&i.WorkflowRunID,
		); err != nil {
			// Remove logging of sensitive data, just log the error
			fmt.Println("Error scanning row in createTasks query:", err)
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		// Remove logging of sensitive data, just log the error
		fmt.Println("Error in rows of createTasks query:", err)
		return nil, err
	}
	return items, nil
}

[... rest of file identical until end ...]
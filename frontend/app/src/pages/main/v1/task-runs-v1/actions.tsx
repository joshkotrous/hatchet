import { Button } from '@/components/v1/ui/button';
import {
  DialogTitle,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
} from '@/components/v1/ui/dialog';
import api, {
  queries,
  V1CancelTaskRequest,
  V1ReplayTaskRequest,
  V1TaskStatus,
} from '@/lib/api';
import { useTenant } from '@/lib/atoms';
import { useApiError } from '@/lib/hooks';
import { ArrowPathIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useCallback, useState } from 'react';
import invariant from 'tiny-invariant';
import {
  TimeWindow,
  useColumnFilters,
} from '../workflow-runs-v1/hooks/column-filters';
import { useToolbarFilters } from '../workflow-runs-v1/hooks/toolbar-filters';
import { Combobox } from '@/components/v1/molecules/combobox/combobox';
import { TaskRunColumn } from '../workflow-runs-v1/components/v1/task-runs-columns';
import {
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
  Select,
} from '@/components/v1/ui/select';
import { DateTimePicker } from '@/components/v1/molecules/time-picker/date-time-picker';

export const TASK_RUN_TERMINAL_STATUSES = [
  V1TaskStatus.CANCELLED,
  V1TaskStatus.FAILED,
  V1TaskStatus.COMPLETED,
];

type ActionType = 'cancel' | 'replay';

type BaseTaskRunActionParams =
  | {
      filter?: never;
      externalIds:
        | NonNullable<V1CancelTaskRequest['externalIds']>
        | NonNullable<V1ReplayTaskRequest['externalIds']>;
    }
  | {
      filter:
        | NonNullable<V1CancelTaskRequest['filter']>
        | NonNullable<V1ReplayTaskRequest['filter']>;
      externalIds?: never;
    };

type TaskRunActionsParams =
  | {
      actionType: 'cancel';
      filter?: never;
      externalIds: NonNullable<V1CancelTaskRequest['externalIds']>;
    }
  | {
      actionType: 'cancel';
      filter: NonNullable<V1CancelTaskRequest['filter']>;
      externalIds?: never;
    }
  | {
      actionType: 'replay';
      filter?: never;
      externalIds: NonNullable<V1ReplayTaskRequest['externalIds']>;
    }
  | {
      actionType: 'replay';
      filter: NonNullable<V1ReplayTaskRequest['filter']>;
      externalIds?: never;
    };

export const useTaskRunActions = () => {
  const { tenant } = useTenant();

  invariant(tenant?.metadata.id);

  const { handleApiError } = useApiError({});

  // Helper function to validate task run action parameters
  const validateParams = (params: TaskRunActionsParams): boolean => {
    // Basic parameter validation
    if (!params || typeof params !== 'object') {
      return false;
    }

    // Validate action type
    if (!['cancel', 'replay'].includes(params.actionType)) {
      return false;
    }

    // Exclusive validation for externalIds and filter
    const hasExternalIds = 'externalIds' in params && params.externalIds !== undefined;
    const hasFilter = 'filter' in params && params.filter !== undefined;

    // Either externalIds or filter should be provided, not both or none
    if ((!hasExternalIds && !hasFilter) || (hasExternalIds && hasFilter)) {
      return false;
    }

    // Validate externalIds if provided
    if (hasExternalIds) {
      if (!Array.isArray(params.externalIds) || params.externalIds.length === 0) {
        return false;
      }
      
      // Check if all externalIds are valid strings
      if (!params.externalIds.every(id => typeof id === 'string' && id.trim().length > 0)) {
        return false;
      }
    }

    // Validate filter if provided
    if (hasFilter) {
      // Check if filter is an object
      if (typeof params.filter !== 'object' || params.filter === null) {
        return false;
      }
      
      // Check if filter has valid properties
      const hasValidFilter = Object.entries(params.filter).some(([key, value]) => {
        if (Array.isArray(value)) {
          return value.length > 0 && value.every(item => 
            (typeof item === 'string' && item.trim().length > 0) || 
            (typeof item === 'object' && item !== null)
          );
        }
        return value !== undefined && value !== null;
      });
      
      if (!hasValidFilter) {
        return false;
      }
    }

    return true;
  };

  const { mutate: handleAction } = useMutation({
    mutationKey: ['task-run:action'],
    mutationFn: async (params: TaskRunActionsParams) => {
      const actionType: ActionType = params.actionType;

      // Validate parameters before making API calls
      if (!validateParams(params)) {
        throw new Error('Invalid task run action parameters');
      }

      switch (actionType) {
        case 'cancel':
          return api.v1TaskCancel(tenant.metadata.id, params);
        case 'replay':
          return api.v1TaskReplay(tenant.metadata.id, params);
        default:
          // eslint-disable-next-line no-case-declarations
          const exhaustiveCheck: never = actionType;
          throw new Error(`Unhandled action type: ${exhaustiveCheck}`);
      }
    },
    onError: handleApiError,
  });

  const handleTaskRunAction = useCallback(
    (params: TaskRunActionsParams) => {
      handleAction(params);
    },
    [handleAction],
  );

  return { handleTaskRunAction };
};

[... rest of file remains unchanged ...]
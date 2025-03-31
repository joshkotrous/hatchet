from typing import Any, cast
import json
import logging
from google.protobuf.timestamp_pb2 import Timestamp

from hatchet_sdk.v0.clients.dispatcher.action_listener import (
    Action,
    ActionListener,
    GetActionListenerRequest,
)
from hatchet_sdk.v0.clients.rest.tenacity_utils import tenacity_retry
from hatchet_sdk.v0.connection import new_conn
from hatchet_sdk.v0.contracts.dispatcher_pb2 import (
    STEP_EVENT_TYPE_COMPLETED,
    STEP_EVENT_TYPE_FAILED,
    ActionEventResponse,
    GroupKeyActionEvent,
    GroupKeyActionEventType,
    OverridesData,
    RefreshTimeoutRequest,
    ReleaseSlotRequest,
    StepActionEvent,
    StepActionEventType,
    UpsertWorkerLabelsRequest,
    WorkerLabels,
    WorkerRegisterRequest,
    WorkerRegisterResponse,
)
from hatchet_sdk.v0.contracts.dispatcher_pb2_grpc import DispatcherStub

from ...loader import ClientConfig
from ...metadata import get_metadata

DEFAULT_REGISTER_TIMEOUT = 30
MAX_PAYLOAD_SIZE = 1_000_000  # 1MB character limit for payloads
logger = logging.getLogger(__name__)


def new_dispatcher(config: ClientConfig) -> "DispatcherClient":
    return DispatcherClient(config=config)


def sanitize_and_validate_payload(payload: str) -> str:
    """
    Sanitizes and validates a payload string to prevent injection of malicious data.
    
    Args:
        payload: The payload string to sanitize and validate
        
    Returns:
        The sanitized payload
        
    Raises:
        ValueError: If the payload is invalid or exceeds size limits
    """
    if not payload:
        return payload
    
    # Prevent excessively large payloads
    if len(payload) > MAX_PAYLOAD_SIZE:
        raise ValueError(f"Payload exceeds maximum size limit of {MAX_PAYLOAD_SIZE} characters")
    
    # If payload looks like JSON, validate its structure
    if payload.strip().startswith(("{", "[")):
        try:
            # Parse and re-serialize to ensure well-formed JSON
            parsed = json.loads(payload)
            return json.dumps(parsed)
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in payload: {str(e)}")
    
    # For non-JSON payloads, return as-is
    return payload


class DispatcherClient:
    config: ClientConfig

    def __init__(self, config: ClientConfig):
        conn = new_conn(config)
        self.client = DispatcherStub(conn)  # type: ignore[no-untyped-call]

        aio_conn = new_conn(config, True)
        self.aio_client = DispatcherStub(aio_conn)  # type: ignore[no-untyped-call]
        self.token = config.token
        self.config = config

    async def get_action_listener(
        self, req: GetActionListenerRequest
    ) -> ActionListener:

        # Override labels with the preset labels
        preset_labels = self.config.worker_preset_labels

        for key, value in preset_labels.items():
            req.labels[key] = WorkerLabels(strValue=str(value))

        # Register the worker
        response: WorkerRegisterResponse = await self.aio_client.Register(
            WorkerRegisterRequest(
                workerName=req.worker_name,
                actions=req.actions,
                services=req.services,
                maxRuns=req.max_runs,
                labels=req.labels,
            ),
            timeout=DEFAULT_REGISTER_TIMEOUT,
            metadata=get_metadata(self.token),
        )

        return ActionListener(self.config, response.workerId)

    async def send_step_action_event(
        self, action: Action, event_type: StepActionEventType, payload: str
    ) -> Any:
        try:
            # Sanitize and validate payload before using it
            try:
                validated_payload = sanitize_and_validate_payload(payload)
            except ValueError as ve:
                # If validation fails, log the error and send a failure event
                error_message = f"Invalid payload: {str(ve)}"
                logger.error(error_message)
                
                if event_type != STEP_EVENT_TYPE_FAILED:
                    await self._try_send_step_action_event(
                        action, 
                        STEP_EVENT_TYPE_FAILED,
                        error_message
                    )
                return
                
            return await self._try_send_step_action_event(action, event_type, validated_payload)
        except Exception as e:
            # for step action events, send a failure event when we cannot send the completed event
            if (
                event_type == STEP_EVENT_TYPE_COMPLETED
                or event_type == STEP_EVENT_TYPE_FAILED
            ):
                try:
                    await self._try_send_step_action_event(
                        action,
                        STEP_EVENT_TYPE_FAILED,
                        f"Failed to send finished event: {str(e)}",
                    )
                except Exception as inner_e:
                    logger.error(f"Failed to send failure event: {str(inner_e)}")

            return

    @tenacity_retry
    async def _try_send_step_action_event(
        self, action: Action, event_type: StepActionEventType, payload: str
    ) -> Any:
        eventTimestamp = Timestamp()
        eventTimestamp.GetCurrentTime()

        event = StepActionEvent(
            workerId=action.worker_id,
            jobId=action.job_id,
            jobRunId=action.job_run_id,
            stepId=action.step_id,
            stepRunId=action.step_run_id,
            actionId=action.action_id,
            eventTimestamp=eventTimestamp,
            eventType=event_type,
            eventPayload=payload,
            retryCount=action.retry_count,
        )

        ## TODO: What does this return?
        return await self.aio_client.SendStepActionEvent(
            event,
            metadata=get_metadata(self.token),
        )

    async def send_group_key_action_event(
        self, action: Action, event_type: GroupKeyActionEventType, payload: str
    ) -> Any:
        try:
            # Sanitize and validate payload before using it
            validated_payload = sanitize_and_validate_payload(payload)
        except ValueError as ve:
            error_message = f"Invalid payload: {str(ve)}"
            logger.error(error_message)
            
            # Create event with error message instead of invalid payload
            eventTimestamp = Timestamp()
            eventTimestamp.GetCurrentTime()
            
            event = GroupKeyActionEvent(
                workerId=action.worker_id,
                workflowRunId=action.workflow_run_id,
                getGroupKeyRunId=action.get_group_key_run_id,
                actionId=action.action_id,
                eventTimestamp=eventTimestamp,
                eventType=event_type,
                eventPayload=error_message,
            )
            
            return await self.aio_client.SendGroupKeyActionEvent(
                event,
                metadata=get_metadata(self.token),
            )

        eventTimestamp = Timestamp()
        eventTimestamp.GetCurrentTime()

        event = GroupKeyActionEvent(
            workerId=action.worker_id,
            workflowRunId=action.workflow_run_id,
            getGroupKeyRunId=action.get_group_key_run_id,
            actionId=action.action_id,
            eventTimestamp=eventTimestamp,
            eventType=event_type,
            eventPayload=validated_payload,
        )

        ## TODO: What does this return?
        return await self.aio_client.SendGroupKeyActionEvent(
            event,
            metadata=get_metadata(self.token),
        )

    def put_overrides_data(self, data: OverridesData) -> ActionEventResponse:
        return cast(
            ActionEventResponse,
            self.client.PutOverridesData(
                data,
                metadata=get_metadata(self.token),
            ),
        )

    def release_slot(self, step_run_id: str) -> None:
        self.client.ReleaseSlot(
            ReleaseSlotRequest(stepRunId=step_run_id),
            timeout=DEFAULT_REGISTER_TIMEOUT,
            metadata=get_metadata(self.token),
        )

    def refresh_timeout(self, step_run_id: str, increment_by: str) -> None:
        self.client.RefreshTimeout(
            RefreshTimeoutRequest(
                stepRunId=step_run_id,
                incrementTimeoutBy=increment_by,
            ),
            timeout=DEFAULT_REGISTER_TIMEOUT,
            metadata=get_metadata(self.token),
        )

    def upsert_worker_labels(
        self, worker_id: str | None, labels: dict[str, str | int]
    ) -> None:
        worker_labels = {}

        for key, value in labels.items():
            if isinstance(value, int):
                worker_labels[key] = WorkerLabels(intValue=value)
            else:
                worker_labels[key] = WorkerLabels(strValue=str(value))

        self.client.UpsertWorkerLabels(
            UpsertWorkerLabelsRequest(workerId=worker_id, labels=worker_labels),
            timeout=DEFAULT_REGISTER_TIMEOUT,
            metadata=get_metadata(self.token),
        )

    async def async_upsert_worker_labels(
        self,
        worker_id: str | None,
        labels: dict[str, str | int],
    ) -> None:
        worker_labels = {}

        for key, value in labels.items():
            if isinstance(value, int):
                worker_labels[key] = WorkerLabels(intValue=value)
            else:
                worker_labels[key] = WorkerLabels(strValue=str(value))

        await self.aio_client.UpsertWorkerLabels(
            UpsertWorkerLabelsRequest(workerId=worker_id, labels=worker_labels),
            timeout=DEFAULT_REGISTER_TIMEOUT,
            metadata=get_metadata(self.token),
        )
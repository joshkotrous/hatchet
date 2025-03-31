// eslint-disable-next-line max-classes-per-file
import { EventEmitter, on } from 'events';
import {
  WorkflowRunEvent,
  SubscribeToWorkflowRunsRequest,
  WorkflowRunEventType,
} from '@hatchet/protoc/dispatcher';
import { isAbortError } from 'abort-controller-x';
import sleep from '@hatchet/util/sleep';
import { RunListenerClient } from './child-listener-client';

export class Streamable {
  listener: AsyncIterable<WorkflowRunEvent>;
  id: string;

  responseEmitter = new EventEmitter();

  constructor(listener: AsyncIterable<WorkflowRunEvent>, id: string) {
    this.listener = listener;
    this.id = id;
  }

  async *stream(): AsyncGenerator<WorkflowRunEvent, void, unknown> {
    while (true) {
      const req: WorkflowRunEvent = await new Promise((resolve) => {
        this.responseEmitter.once('response', resolve);
      });
      yield req;
    }
  }
}

export class RunGrpcPooledListener {
  listener: AsyncIterable<WorkflowRunEvent> | undefined;
  requestEmitter = new EventEmitter();
  signal: AbortController = new AbortController();
  client: RunListenerClient;

  subscribers: Record<string, Streamable> = {};
  onFinish: () => void = () => {};

  constructor(client: RunListenerClient, onFinish: () => void) {
    this.client = client;
    this.init();
    this.onFinish = onFinish;
  }

  private async init(retries = 0) {
    const MAX_RETRIES = 20; // Maximum number of total retries
    const MAX_RETRY_INTERVAL = 5000; // 5 seconds in milliseconds
    const BASE_RETRY_INTERVAL = 100; // 0.1 seconds in milliseconds
    
    let retryCount = retries;
    
    // If we've already exceeded the max retries, log a critical error and exit
    if (retryCount > MAX_RETRIES) {
      this.client.logger.error(`Maximum retry attempts (${MAX_RETRIES}) reached. Giving up.`);
      this.handleMaxRetriesExceeded();
      return;
    }
    
    if (retries > 0) {
      const backoffTime = Math.min(BASE_RETRY_INTERVAL * 2 ** (retries - 1), MAX_RETRY_INTERVAL);
      this.client.logger.info(`Retrying in ... ${backoffTime / 1000} seconds (attempt ${retryCount} of ${MAX_RETRIES})`);
      await sleep(backoffTime);
    }

    try {
      this.client.logger.debug('Initializing child-listener');

      this.signal = new AbortController();
      this.listener = this.client.client.subscribeToWorkflowRuns(this.request(), {
        signal: this.signal.signal,
      });

      if (retries > 0) setTimeout(() => this.replayRequests(), 100);

      for await (const event of this.listener) {
        retryCount = 0; // Reset retry count on successful event processing

        const emitter = this.subscribers[event.workflowRunId];
        if (emitter) {
          emitter.responseEmitter.emit('response', event);
          if (event.eventType === WorkflowRunEventType.WORKFLOW_RUN_EVENT_TYPE_FINISHED) {
            delete this.subscribers[event.workflowRunId];
          }
        }
      }

      this.client.logger.debug('Child listener finished');
    } catch (e: any) {
      if (isAbortError(e)) {
        this.client.logger.debug('Child Listener aborted');
        return;
      }
      this.client.logger.error(`Error in child-listener: ${e.message}`);
    } finally {
      // it is possible the server hangs up early,
      // restart the listener if we still have subscribers
      const subscriberCount = Object.keys(this.subscribers).length;
      this.client.logger.debug(
        `Child listener loop exited with ${subscriberCount} subscribers`
      );
      
      if (subscriberCount === 0) {
        this.client.logger.debug('No subscribers remaining, stopping retry attempts');
        return;
      }
      
      // Only retry if we haven't exceeded the max retries
      if (retryCount + 1 <= MAX_RETRIES) {
        this.client.logger.debug(`Restarting child listener retry ${retryCount + 1} of ${MAX_RETRIES}`);
        this.init(retryCount + 1);
      } else {
        this.client.logger.error(`Maximum retry attempts (${MAX_RETRIES}) reached. Giving up.`);
        this.handleMaxRetriesExceeded();
      }
    }
  }
  
  // Helper method to handle the case when max retries are exceeded
  private handleMaxRetriesExceeded() {
    // Clear subscribers to prevent further retries
    this.subscribers = {};
    
    // Call onFinish callback to notify parent components
    this.onFinish();
  }

  subscribe(request: SubscribeToWorkflowRunsRequest) {
    if (!this.listener) throw new Error('listener not initialized');
    
    // Validate workflowRunId
    if (!request.workflowRunId || typeof request.workflowRunId !== 'string' || request.workflowRunId.trim() === '') {
      throw new Error('Invalid workflowRunId: must be a non-empty string');
    }
    
    // Use the validated workflowRunId
    const safeWorkflowRunId = request.workflowRunId.trim();
    
    // Create a new request object with the safe ID
    const safeRequest = { ...request, workflowRunId: safeWorkflowRunId };
    
    this.subscribers[safeWorkflowRunId] = new Streamable(this.listener, safeWorkflowRunId);
    this.requestEmitter.emit('subscribe', safeRequest);
    return this.subscribers[safeWorkflowRunId];
  }

  replayRequests() {
    const subs = Object.values(this.subscribers);
    this.client.logger.debug(`Replaying ${subs.length} requests...`);

    for (const subscriber of subs) {
      this.requestEmitter.emit('subscribe', { workflowRunId: subscriber.id });
    }
  }

  private async *request(): AsyncIterable<SubscribeToWorkflowRunsRequest> {
    for await (const e of on(this.requestEmitter, 'subscribe')) {
      yield e[0];
    }
  }
}
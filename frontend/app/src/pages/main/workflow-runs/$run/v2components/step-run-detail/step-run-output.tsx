import { CodeHighlighter } from '@/components/ui/code-highlighter';
import { StepRun, StepRunStatus, WorkflowRunShape } from '@/lib/api';
import React from 'react';
import StepRunCodeText from './step-run-error';
import LoggingComponent from '@/components/cloud/logging/logs';

const readableReason = (reason?: string): string => {
  return reason ? reason.toLowerCase().split('_').join(' ') : '';
};

type StepRunOutputProps = {
  stepRun: StepRun;
  workflowRun: WorkflowRunShape;
};

const oneLiner = (text: string) => {
  return (
    <div className="my-4">
      <LoggingComponent
        logs={[
          {
            line: text,
          },
        ]}
        onTopReached={() => {}}
        onBottomReached={() => {}}
      />
    </div>
  );
};

const StepRunOutputCancelled = ({ stepRun }: StepRunOutputProps) => {
  let msg = 'Step run was cancelled';

  if (stepRun.cancelledReason) {
    msg = `Step run was cancelled: ${readableReason(stepRun.cancelledReason)}`;
  }

  return oneLiner(msg);
};

const StepRunOutputPending = ({ stepRun }: StepRunOutputProps) => {
  let msg = 'Waiting to start...';

  if (stepRun.parents) {
    msg = `Waiting for parent steps to complete: ${stepRun.parents.join(', ')}`;
  }

  return oneLiner(msg);
};

const StepRunOutputPendingAssignment = () => {
  return oneLiner('Step is waiting to be assigned to a worker');
};

const StepRunOutputAssigned = () => {
  return oneLiner('Step has been assigned and will start shortly');
};

const StepRunOutputRunning = () => {
  return oneLiner('Step is currently running...');
};

const StepRunOutputSucceeded = ({ stepRun }: StepRunOutputProps) => {
  try {
    // Parse the output with a default empty object if it's null/undefined
    const parsedOutput = stepRun?.output ? JSON.parse(stepRun.output) : {};
    
    // Implement a simple size check to prevent large payloads from being rendered
    const outputStr = JSON.stringify(parsedOutput, null, 2);
    const MAX_DISPLAY_SIZE = 500000; // ~500KB limit for display
    
    if (outputStr.length > MAX_DISPLAY_SIZE) {
      return oneLiner(`Output is too large to display safely (${Math.round(outputStr.length/1024)}KB). Please download the raw output if needed.`);
    }
    
    return (
      <CodeHighlighter
        className="my-4 h-[400px] max-h-[400px] overflow-y-auto"
        language="json"
        maxHeight="400px"
        minHeight="400px"
        code={outputStr}
      />
    );
  } catch (error) {
    // Handle parsing errors gracefully
    return oneLiner(`Error parsing step output: ${error instanceof Error ? error.message : 'Invalid JSON'}`);
  }
};

const StepRunOutputFailed = ({ stepRun }: StepRunOutputProps) => {
  if (!stepRun.error) {
    return oneLiner('Step run failed with no error message');
  }

  return (
    <div className="my-4">
      <StepRunCodeText text={stepRun.error} />
    </div>
  );
};

const StepRunOutputCancelling = () => {
  return oneLiner('Step run is being cancelled');
};

const StepRunOutputBackoff = () => {
  return oneLiner('Step run is in a retry backoff state');
};

const OUTPUT_STATE_MAP: Record<StepRunStatus, React.FC<StepRunOutputProps>> = {
  [StepRunStatus.CANCELLED]: StepRunOutputCancelled,
  [StepRunStatus.PENDING]: StepRunOutputPending,
  [StepRunStatus.PENDING_ASSIGNMENT]: StepRunOutputPendingAssignment,
  [StepRunStatus.ASSIGNED]: StepRunOutputAssigned,
  [StepRunStatus.RUNNING]: StepRunOutputRunning,
  [StepRunStatus.SUCCEEDED]: StepRunOutputSucceeded,
  [StepRunStatus.FAILED]: StepRunOutputFailed,
  [StepRunStatus.CANCELLING]: StepRunOutputCancelling,
  [StepRunStatus.BACKOFF]: StepRunOutputBackoff,
};

const StepRunOutput: React.FC<StepRunOutputProps> = (props) => {
  const Component = OUTPUT_STATE_MAP[props.stepRun.status];
  return <Component {...props} />;
};

export default StepRunOutput;
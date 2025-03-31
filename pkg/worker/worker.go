package worker

[Previous code with the following changes applied between lines 429-848:]

// validateActionOutput validates and sanitizes action outputs to mitigate against ML model output manipulation
func (w *Worker) validateActionOutput(actionId string, output any) (any, error) {
	if output == nil {
		return nil, nil
	}
	
	// Log output type for monitoring and auditing
	w.l.Debug().
		Str("action_id", actionId).
		Str("output_type", fmt.Sprintf("%T", output)).
		Msg("Validating action output")
	
	// Apply type-specific validation
	switch v := output.(type) {
	case string:
		// For string outputs (common in ML/LLM responses), limit size to prevent DoS
		if len(v) > 10_000_000 { // 10MB limit
			w.l.Warn().
				Str("action_id", actionId).
				Int("length", len(v)).
				Msg("Output string exceeds size limit, truncating")
			return v[:10_000_000] + "... [truncated]", nil
		}
		
	case map[string]interface{}:
		// For structured outputs, validate recursively
		result := make(map[string]interface{})
		for key, val := range v {
			// Recursively validate nested values
			validVal, err := w.validateActionOutput(actionId, val)
			if err != nil {
				return nil, err
			}
			result[key] = validVal
		}
		return result, nil
		
	case []interface{}:
		// For array outputs, validate each element and limit size
		if len(v) > 1_000_000 { // 1M element limit
			w.l.Warn().
				Str("action_id", actionId).
				Int("length", len(v)).
				Msg("Output array exceeds size limit, truncating")
			v = v[:1_000_000]
		}
		
		result := make([]interface{}, len(v))
		for i, item := range v {
			validItem, err := w.validateActionOutput(actionId, item)
			if err != nil {
				return nil, err
			}
			result[i] = validItem
		}
		return result, nil
	}
	
	return output, nil
}

[... rest of original code until line 593]

			runResults := action.Run(args...)

			// check whether run context was cancelled while action was running
			select {
			case <-ctx.Done():
				w.l.Debug().Msgf("step run %s was cancelled, returning", assignedAction.StepRunId)
				return nil
			default:
			}

			var result any

			if len(runResults) == 2 {
				result = runResults[0]
			}

			if runResults[len(runResults)-1] != nil {
				err = runResults[len(runResults)-1].(error)
			}

			if err != nil {
				return w.sendFailureEvent(ctx, err)
			}

			// Validate the action output before using it in the completion event
			validatedResult, err := w.validateActionOutput(assignedAction.ActionId, result)
			if err != nil {
				return fmt.Errorf("output validation failed: %w", err)
			}

			// send a message that the step run completed
			finishedEvent, err := w.getActionFinishedEvent(assignedAction, validatedResult)

[... rest of original code]
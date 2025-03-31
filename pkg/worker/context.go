package worker

[Previous code content remains exactly the same until line 482]

func (h *hatchetContext) populateStepDataForGroupKeyRun() error {
	if h.stepData != nil {
		return nil
	}

	inputData := map[string]interface{}{}

	err := json.Unmarshal(h.a.ActionPayload, &inputData)

	if err != nil {
		return err
	}

	// Validate the input data
	if err := validateMapData(inputData); err != nil {
		return fmt.Errorf("invalid group key run payload: %w", err)
	}

	h.stepData = &StepRunData{
		Input: inputData,
	}

	return nil
}

func (h *hatchetContext) populateStepData() error {
	if h.stepData != nil {
		return nil
	}

	// Create a temporary variable for validation
	tempStepData := &StepRunData{}

	jsonBytes := h.a.ActionPayload

	if len(jsonBytes) == 0 {
		jsonBytes = []byte("{}")
	}

	err := json.Unmarshal(jsonBytes, tempStepData)

	if err != nil {
		return err
	}

	// Validate the unmarshaled data
	if err := validateStepRunData(tempStepData); err != nil {
		return fmt.Errorf("invalid action payload: %w", err)
	}

	// If validation passes, assign to h.stepData
	h.stepData = tempStepData
	h.stepData.AdditionalMetadata = h.a.AdditionalMetadata

	return nil
}

// validateStepRunData performs validation checks on StepRunData
func validateStepRunData(data *StepRunData) error {
	// Validate TriggeredBy field
	if data.TriggeredBy != "" && 
	   data.TriggeredBy != TriggeredByEvent && 
	   data.TriggeredBy != TriggeredByCron && 
	   data.TriggeredBy != TriggeredBySchedule {
		return fmt.Errorf("invalid triggered_by value: %s", data.TriggeredBy)
	}

	// Validate maps for empty keys
	if data.Input != nil {
		if err := validateMapData(data.Input); err != nil {
			return fmt.Errorf("invalid input data: %w", err)
		}
	}

	if data.Parents != nil {
		for parentName, parentData := range data.Parents {
			if parentName == "" {
				return fmt.Errorf("parent name cannot be empty")
			}
			if err := validateMapData(parentData); err != nil {
				return fmt.Errorf("invalid parent data for %s: %w", parentName, err)
			}
		}
	}

	if data.Triggers != nil {
		for triggerKey, triggerData := range data.Triggers {
			if triggerKey == "" {
				return fmt.Errorf("trigger key cannot be empty")
			}
			if err := validateMapData(triggerData); err != nil {
				return fmt.Errorf("invalid trigger data for %s: %w", triggerKey, err)
			}
		}
	}

	if data.UserData != nil {
		if err := validateMapData(data.UserData); err != nil {
			return fmt.Errorf("invalid user data: %w", err)
		}
	}

	return nil
}

// validateMapData performs basic validation on map data
func validateMapData(data map[string]interface{}) error {
	for key, value := range data {
		if key == "" {
			return fmt.Errorf("map contains an empty key")
		}
		
		// Validate nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			if err := validateMapData(nestedMap); err != nil {
				return err
			}
		}
	}
	
	return nil
}

[Remaining code content remains exactly the same]
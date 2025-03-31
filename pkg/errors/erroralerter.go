package errors

import (
	"context"

	"github.com/hatchet-dev/hatchet/internal/datautils/merge"
)

type Alerter interface {
	SendAlert(ctx context.Context, err error, data map[string]interface{})
}

type NoOpAlerter struct{}

func (s NoOpAlerter) SendAlert(ctx context.Context, err error, data map[string]interface{}) {}

type Wrapped struct {
	a               Alerter
	data            map[string]interface{}
	sensitiveFields []string
}

func NewWrapped(a Alerter) *Wrapped {
	return &Wrapped{
		a: a,
	}
}

func (w *Wrapped) WithData(data map[string]interface{}) {
	w.data = data
}

// WithSensitiveFields specifies fields that should be redacted from alert data
func (w *Wrapped) WithSensitiveFields(fields []string) *Wrapped {
	w.sensitiveFields = fields
	return w
}

// sanitizeData redacts values for sensitive fields at all levels of the data structure
func (w *Wrapped) sanitizeData(data map[string]interface{}) map[string]interface{} {
	if data == nil || len(w.sensitiveFields) == 0 {
		return data
	}

	// Create a new map to avoid modifying the original
	sanitized := make(map[string]interface{})
	for k, v := range data {
		sensitive := false
		for _, field := range w.sensitiveFields {
			if k == field {
				sensitive = true
				break
			}
		}

		if sensitive {
			sanitized[k] = "[REDACTED]"
		} else if nestedMap, ok := v.(map[string]interface{}); ok {
			// Recursively sanitize nested maps
			sanitized[k] = w.sanitizeData(nestedMap)
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

func (w *Wrapped) WrapErr(err error, data map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Merge data first
	mergedData := merge.MergeMaps(w.data, data)
	
	// Sanitize sensitive fields if any are specified
	if len(w.sensitiveFields) > 0 {
		mergedData = w.sanitizeData(mergedData)
	}

	w.a.SendAlert(context.Background(), err, mergedData)
	return err
}
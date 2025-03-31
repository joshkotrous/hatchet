package v2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/v1/sqlcv1"
)

// ExtensionTimeout is the default timeout for extension operations
const ExtensionTimeout = 30 * time.Second

type PostAssignInput struct {
	HasUnassignedStepRuns bool
}

type SnapshotInput struct {
	Workers               map[string]*WorkerCp
	WorkerSlotUtilization map[string]*SlotUtilization
}

type SlotUtilization struct {
	UtilizedSlots    int
	NonUtilizedSlots int
}

type WorkerCp struct {
	WorkerId string
	MaxRuns  int
	Labels   []*sqlcv1.ListManyWorkerLabelsRow
}

type SlotCp struct {
	WorkerId string
	Used     bool
}

type SchedulerExtension interface {
	SetTenants(tenants []*dbsqlc.Tenant)
	ReportSnapshot(tenantId string, input *SnapshotInput)
	PostAssign(tenantId string, input *PostAssignInput)
	Cleanup() error
}

type Extensions struct {
	mu   sync.RWMutex
	exts []SchedulerExtension
}

func (e *Extensions) Add(ext SchedulerExtension) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.exts == nil {
		e.exts = make([]SchedulerExtension, 0)
	}

	e.exts = append(e.exts, ext)
}

func (e *Extensions) ReportSnapshot(tenantId string, input *SnapshotInput) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(len(e.exts))

	for i, ext := range e.exts {
		// Capture loop variables
		idx := i
		extension := ext
		
		go func() {
			defer wg.Done()
			// Recover from panics
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in ReportSnapshot for extension %d: %v\n", idx, r)
				}
			}()
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
			defer cancel()
			
			// Create a channel to signal completion
			done := make(chan struct{})
			
			go func() {
				extension.ReportSnapshot(tenantId, input)
				close(done)
			}()
			
			// Wait for completion or timeout
			select {
			case <-done:
				// Successfully completed
			case <-ctx.Done():
				fmt.Printf("ReportSnapshot timed out for extension %d\n", idx)
			}
		}()
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
}

func (e *Extensions) PostAssign(tenantId string, input *PostAssignInput) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(len(e.exts))

	for i, ext := range e.exts {
		// Capture loop variables
		idx := i
		extension := ext
		
		go func() {
			defer wg.Done()
			// Recover from panics
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in PostAssign for extension %d: %v\n", idx, r)
				}
			}()
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
			defer cancel()
			
			// Create a channel to signal completion
			done := make(chan struct{})
			
			go func() {
				extension.PostAssign(tenantId, input)
				close(done)
			}()
			
			// Wait for completion or timeout
			select {
			case <-done:
				// Successfully completed
			case <-ctx.Done():
				fmt.Printf("PostAssign timed out for extension %d\n", idx)
			}
		}()
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
}

func (e *Extensions) Cleanup() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	eg := errgroup.Group{}

	for i, ext := range e.exts {
		// Capture loop variables
		idx := i
		extension := ext
		
		eg.Go(func() error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
			defer cancel()
			
			// Create a channel for the result
			resultCh := make(chan error, 1)
			
			go func() {
				// Recover from panics
				defer func() {
					if r := recover(); r != nil {
						resultCh <- fmt.Errorf("panic in Cleanup for extension %d: %v", idx, r)
					}
				}()
				
				resultCh <- extension.Cleanup()
			}()
			
			// Wait for result or timeout
			select {
			case err := <-resultCh:
				return err
			case <-ctx.Done():
				return fmt.Errorf("Cleanup timed out for extension %d", idx)
			}
		})
	}

	return eg.Wait()
}

func (e *Extensions) SetTenants(tenants []*dbsqlc.Tenant) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(len(e.exts))

	for i, ext := range e.exts {
		// Capture loop variables
		idx := i
		extension := ext
		
		go func() {
			defer wg.Done()
			// Recover from panics
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in SetTenants for extension %d: %v\n", idx, r)
				}
			}()
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
			defer cancel()
			
			// Create a channel to signal completion
			done := make(chan struct{})
			
			go func() {
				extension.SetTenants(tenants)
				close(done)
			}()
			
			// Wait for completion or timeout
			select {
			case <-done:
				// Successfully completed
			case <-ctx.Done():
				fmt.Printf("SetTenants timed out for extension %d\n", idx)
			}
		}()
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
}
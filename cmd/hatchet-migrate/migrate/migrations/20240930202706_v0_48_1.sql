-- +goose Up
-- Modify "JobRun" table
-- NOTE: This foreign key constraint is being deliberately removed as part of schema changes in v0.48.1.
-- Replaced with application-level validation to maintain data integrity while allowing more flexible relationships.
ALTER TABLE "JobRun" DROP CONSTRAINT "JobRun_jobId_fkey";

-- Modify "JobRunLookupData" table
-- NOTE: This foreign key constraint is being deliberately removed as part of schema changes in v0.48.1.
-- Replaced with application-level validation to maintain data integrity while allowing more flexible relationships.
ALTER TABLE "JobRunLookupData" DROP CONSTRAINT "JobRunLookupData_tenantId_fkey";

-- +goose Down
-- Restore "JobRun" constraint if migration needs to be rolled back
ALTER TABLE "JobRun" ADD CONSTRAINT "JobRun_jobId_fkey" 
    FOREIGN KEY ("jobId") REFERENCES "Job"(id);
    
-- Restore "JobRunLookupData" constraint if migration needs to be rolled back
ALTER TABLE "JobRunLookupData" ADD CONSTRAINT "JobRunLookupData_tenantId_fkey" 
    FOREIGN KEY ("tenantId") REFERENCES "Tenant"(id);
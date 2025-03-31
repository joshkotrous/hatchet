CREATE OR REPLACE FUNCTION get_v1_partitions_before_date(
    targetTableName text,
    targetDate date
) RETURNS TABLE(partition_name text)
    LANGUAGE plpgsql AS
$
BEGIN
    RETURN QUERY
    SELECT
        inhrelid::regclass::text AS partition_name
    FROM
        pg_inherits
    WHERE
        inhparent = targetTableName::regclass
        AND substring(inhrelid::regclass::text, format('%s_(\d{8})', targetTableName)) ~ '^\d{8}'
        AND (substring(inhrelid::regclass::text, format('%s_(\d{8})', targetTableName))::date) < targetDate;
END;
$;

CREATE OR REPLACE FUNCTION create_v1_range_partition(
    targetTableName text,
    targetDate date
) RETURNS integer
    LANGUAGE plpgsql AS
$
DECLARE
    targetDateStr varchar;
    targetDatePlusOneDayStr varchar;
    newTableName varchar;
BEGIN
    -- Validate table name to prevent SQL injection
    IF targetTableName !~ '^[a-zA-Z0-9_]+$' THEN
        RAISE EXCEPTION 'Invalid table name: %', targetTableName;
    END IF;

    SELECT to_char(targetDate, 'YYYYMMDD') INTO targetDateStr;
    SELECT to_char(targetDate + INTERVAL '1 day', 'YYYYMMDD') INTO targetDatePlusOneDayStr;
    SELECT lower(format('%s_%s', targetTableName, targetDateStr)) INTO newTableName;
    -- exit if the table exists
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = newTableName) THEN
        RETURN 0;
    END IF;

    EXECUTE
        format('CREATE TABLE %I (LIKE %I INCLUDING INDEXES)', newTableName, targetTableName);
    EXECUTE
        format('ALTER TABLE %I SET (
            autovacuum_vacuum_scale_factor = ''0.1'',
            autovacuum_analyze_scale_factor=''0.05'',
            autovacuum_vacuum_threshold=''25'',
            autovacuum_analyze_threshold=''25'',
            autovacuum_vacuum_cost_delay=''10'',
            autovacuum_vacuum_cost_limit=''1000''
        )', newTableName);
    EXECUTE
        format('ALTER TABLE %I ATTACH PARTITION %I FOR VALUES FROM (%L) TO (%L)', targetTableName, newTableName, targetDateStr, targetDatePlusOneDayStr);
    RETURN 1;
END;
$;

[Rest of file unchanged...]
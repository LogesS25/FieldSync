DROP TABLE IF EXISTS supervisor_assignments;
DROP TABLE IF EXISTS placements;
DROP TABLE IF EXISTS practicums;
ALTER TABLE users DROP COLUMN IF EXISTS institution_id;
ALTER TABLE users DROP COLUMN IF EXISTS agency_id;
DROP TABLE IF EXISTS agencies;
DROP TABLE IF EXISTS institutions;
DROP TYPE IF EXISTS practicum_status;

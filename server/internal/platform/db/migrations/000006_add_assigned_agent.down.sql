DROP INDEX IF EXISTS idx_rooms_assigned_agent_id;
ALTER TABLE rooms DROP COLUMN IF EXISTS assigned_agent_id;

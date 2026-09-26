ALTER TABLE rooms ADD COLUMN assigned_agent_id UUID REFERENCES agents(id);

CREATE INDEX idx_rooms_assigned_agent_id ON rooms (assigned_agent_id);

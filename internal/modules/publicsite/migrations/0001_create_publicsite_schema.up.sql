CREATE SCHEMA IF NOT EXISTS publicsite;

CREATE TABLE IF NOT EXISTS publicsite.leads (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    company TEXT NOT NULL DEFAULT '',
    size TEXT NOT NULL DEFAULT '',
    need TEXT NOT NULL DEFAULT '',
    utm_source TEXT NOT NULL DEFAULT '',
    consent_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS publicsite.commercial_availabilities (
    id UUID PRIMARY KEY,
    commercial_id UUID NOT NULL,
    weekday INT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    slot_minutes INT NOT NULL DEFAULT 30 CHECK (slot_minutes > 0),
    timezone TEXT NOT NULL DEFAULT 'Europe/Paris',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS publicsite.booking_slots (
    id UUID PRIMARY KEY,
    commercial_id UUID NOT NULL,
    slot_start TIMESTAMPTZ NOT NULL,
    slot_end TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'free',
    external_event_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (commercial_id, slot_start)
);

CREATE TABLE IF NOT EXISTS publicsite.appointments (
    id UUID PRIMARY KEY,
    lead_id UUID NOT NULL REFERENCES publicsite.leads(id),
    commercial_id UUID NOT NULL,
    slot_id UUID NOT NULL REFERENCES publicsite.booking_slots(id),
    channel TEXT NOT NULL DEFAULT 'video',
    status TEXT NOT NULL DEFAULT 'confirmed',
    cancel_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_publicsite_leads_email ON publicsite.leads(email);
CREATE INDEX IF NOT EXISTS idx_publicsite_slots_commercial ON publicsite.booking_slots(commercial_id, slot_start);
CREATE INDEX IF NOT EXISTS idx_publicsite_appointments_token ON publicsite.appointments(cancel_token);

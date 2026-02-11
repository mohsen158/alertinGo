CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    telegram_chat_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_name TEXT NOT NULL,
    check_type TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata TEXT NOT NULL DEFAULT '{}',
    timeout INTEGER NOT NULL DEFAULT 60,
    re_alert_interval INTEGER NOT NULL DEFAULT 300,
    status TEXT NOT NULL DEFAULT 'unknown',
    is_active BOOLEAN NOT NULL DEFAULT false,
    channel_id UUID REFERENCES notification_channels(id) ON DELETE SET NULL,
    server_ip TEXT NOT NULL DEFAULT '',
    server_name TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(monitor_name, check_type)
);

CREATE TABLE alert_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'firing',
    last_alerted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    fired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES notification_channels(id) ON DELETE SET NULL,
    alert_type TEXT NOT NULL,
    message TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT true,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

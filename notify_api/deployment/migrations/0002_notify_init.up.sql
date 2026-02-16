CREATE TABLE IF NOT EXISTS processed_events (
    event_id uuid PRIMARY KEY,
    consumer_name VARCHAR(255) NOT NULL,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
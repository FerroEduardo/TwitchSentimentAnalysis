-- CREATE DATABASE twitchSentimentAnalysis;

CREATE TABLE IF NOT EXISTS results (
    channel VARCHAR(100) NOT NULL,
    "user" VARCHAR(100) NOT NULL,
    "message_id" VARCHAR(64) NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    message VARCHAR(500) NOT NULL,
    sentiment_positive double precision NOT NULL,
    sentiment_neutral double precision NOT NULL,
    sentiment_negative double precision NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_results_channel ON results(channel);

CREATE INDEX IF NOT EXISTS idx_results_timestamp ON results(timestamp);

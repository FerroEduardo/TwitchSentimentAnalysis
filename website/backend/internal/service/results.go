package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ResultsService struct {
	conn   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewResultsService(conn *pgxpool.Pool, logger *zap.SugaredLogger) *ResultsService {
	return &ResultsService{
		conn:   conn,
		logger: logger,
	}
}

type AverageResult struct {
	Channel                  string    `json:"-" db:"channel"`
	Timestamp                time.Time `json:"timestamp" db:"minute_timestamp"`
	AveragePositiveSentiment float64   `json:"avg_sentiment_positive" db:"avg_sentiment_positive"`
	AverageNeutralSentiment  float64   `json:"avg_sentiment_neutral" db:"avg_sentiment_neutral"`
	AverageNegativeSentiment float64   `json:"avg_sentiment_negative" db:"avg_sentiment_negative"`
}

type Result struct {
	Channel           string    `json:"-" db:"channel"`
	User              string    `json:"user" db:"user"`
	Message           string    `json:"message" db:"message"`
	MessageId         string    `json:"message_id" db:"message_id"`
	Timestamp         time.Time `json:"timestamp" db:"timestamp"`
	PositiveSentiment float64   `json:"sentiment_positive" db:"sentiment_positive"`
	NeutralSentiment  float64   `json:"sentiment_neutral" db:"sentiment_neutral"`
	NegativeSentiment float64   `json:"sentiment_negative" db:"sentiment_negative"`
}

func (s *ResultsService) GetLastHourChannelAverageResults(ctx context.Context, moment time.Time) (map[string][]AverageResult, error) {
	start, end := startAndEndOfHour(moment)

	rows, err := s.conn.Query(ctx, `
SELECT 
    DATE_TRUNC('minute', "timestamp") AS "minute_timestamp",
    channel,
    AVG(sentiment_positive) AS avg_sentiment_positive,
    AVG(sentiment_neutral) AS avg_sentiment_neutral,
    AVG(sentiment_negative) AS avg_sentiment_negative
FROM 
    results
WHERE 
    "timestamp" BETWEEN $1 AND $2
GROUP BY
    "minute_timestamp", channel
ORDER BY
    "minute_timestamp" ASC;
`, start, end)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	databaseResults, err := pgx.CollectRows(rows, pgx.RowToStructByName[AverageResult])
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	result := make(map[string][]AverageResult)
	for _, r := range databaseResults {
		result[r.Channel] = append(result[r.Channel], r)
	}

	return result, nil
}

func (s *ResultsService) GetLastResults(ctx context.Context, limit int64, moment time.Time) (map[string][]Result, error) {
	start, end := startAndEndOfHour(moment)

	rows, err := s.conn.Query(ctx, `
SELECT 
    *
FROM 
    results
WHERE 
    "timestamp" BETWEEN $1 AND $2
ORDER BY
    "timestamp" DESC
LIMIT $3;
`, start, end, limit)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	databaseResults, err := pgx.CollectRows(rows, pgx.RowToStructByName[Result])
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	result := make(map[string][]Result)
	for _, r := range databaseResults {
		result[r.Channel] = append(result[r.Channel], r)
	}

	return result, nil
}

func startAndEndOfHour(t time.Time) (time.Time, time.Time) {
	// Start of the hour: Set minutes, seconds, and nanoseconds to zero
	start := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())

	// End of the hour: Add 59 minutes, 59 seconds, and 999999999 nanoseconds to the start
	end := start.Add(59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)

	return start, end
}

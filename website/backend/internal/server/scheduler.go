package server

import (
	"context"
	"encoding/json"
	"time"
	"website/internal/service"

	"go.uber.org/zap"
)

type Event struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type scheduler struct {
	logger         *zap.SugaredLogger
	hub            *broadcastHub
	ticker         *time.Ticker
	resultsService *service.ResultsService
}

func NewScheduler(
	logger *zap.SugaredLogger,
	hub *broadcastHub,
	duration time.Duration,
	resultsService *service.ResultsService,
) *scheduler {
	return &scheduler{
		logger:         logger,
		hub:            hub,
		ticker:         time.NewTicker(duration),
		resultsService: resultsService,
	}
}

func (s *scheduler) Start(ctx context.Context) {
	s.logger.Info("Starting scheduler")

	for {
		select {
		case <-s.ticker.C:
			s.sendLastResultsEvent(ctx)
			s.sendLastMessagesEvent(ctx)
		case <-ctx.Done():
			s.logger.Info("Stoping scheduler")
			s.ticker.Stop()
			return
		}
	}
}

func (s *scheduler) sendLastResultsEvent(ctx context.Context) {
	if len(s.hub.clients) == 0 {
		// logger.Debug("There are no clients to send messages")
		return
	}
	channelResults, err := s.resultsService.GetLastHourChannelAverageResults(ctx, time.Now())
	if err != nil {
		s.logger.Errorf("Failed to get last hour results: %v", err)
		return
	}
	resultsEvent := Event{
		Event: "results",
		Data:  channelResults,
	}
	bts, err := json.Marshal(resultsEvent)
	if err != nil {
		s.logger.Errorf("Failed to encode results to json: %v", err)
		return
	}
	s.hub.broadcast <- bts
}

func (s *scheduler) sendLastMessagesEvent(ctx context.Context) {
	if len(s.hub.clients) == 0 {
		// logger.Debug("There are no clients to send messages")
		return
	}
	messages, err := s.resultsService.GetLastResults(ctx, 100, time.Now())
	if err != nil {
		s.logger.Errorf("Failed to get last messages: %v", err)
		return
	}
	messagesEvent := Event{
		Event: "messages",
		Data:  messages,
	}
	bts, err := json.Marshal(messagesEvent)
	if err != nil {
		s.logger.Errorf("Failed to encode last messages to json: %v", err)
		return
	}
	s.hub.broadcast <- bts
}

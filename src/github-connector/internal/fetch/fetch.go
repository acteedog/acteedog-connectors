package fetch

import (
	"fmt"
	"github-connector/internal/core"
)

// ActivityFetcher defines the structure for fetching activities from GitHub
type ActivityFetcher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewActivityFetcher creates a new ActivityFetcher instance
func NewActivityFetcher(httpClient HTTPClient, cfg map[string]any, targetDate string, logger core.Logger) (*ActivityFetcher, error) {
	config, err := newConfig(cfg, targetDate)
	if err != nil {
		return nil, err
	}

	return &ActivityFetcher{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// FetchActivities fetches and processes activities from GitHub
func (f *ActivityFetcher) FetchActivities() ([]*Activity, error) {
	f.logger.Info("Starting to fetch activities")

	allEvents, err := f.fetchAllEvents()
	if err != nil {
		return nil, err
	}

	f.logger.Info(fmt.Sprintf("Fetched %d events", len(allEvents)))

	filteredEvents := filterEventsByRepository(allEvents, f.config.repositoryPatterns)
	f.logger.Info(fmt.Sprintf("After repository filtering: %d events", len(filteredEvents)))

	activities := []*Activity{}
	for _, event := range filteredEvents {
		activity, err := transformEvent(event)
		if err != nil {
			f.logger.Debug(fmt.Sprintf("Skipping event: %s", err.Error()))
			continue
		}
		if activity != nil {
			activities = append(activities, activity)
		}
	}

	f.logger.Info("Finished fetching activities")

	return activities, nil
}

func (f *ActivityFetcher) fetchAllEvents() ([]map[string]any, error) {
	allEvents := []map[string]any{}

	// GitHub Events API returns max 300 events (3 pages with per_page=100)
	for page := 1; page <= 3; page++ {
		events, err := f.httpClient.FetchActivities(f.config.token, f.config.username, page)
		if err != nil {
			return nil, fmt.Errorf("error fetching activities on page %d: %w", page, err)
		}

		if len(events) == 0 {
			f.logger.Debug(fmt.Sprintf("No more events found at page %d, stopping pagination", page))
			break
		}

		// Filter events by date and check if we should stop
		filteredEvents, shouldStop := filterEventsByDate(events, f.config.startTime, f.config.endTime)
		allEvents = append(allEvents, filteredEvents...)

		f.logger.Debug(fmt.Sprintf("Page %d: %d events fetched, %d filtered", page, len(events), len(filteredEvents)))

		if shouldStop {
			f.logger.Debug("Reached events outside date range, stopping pagination")
			break
		}
	}

	return allEvents, nil
}

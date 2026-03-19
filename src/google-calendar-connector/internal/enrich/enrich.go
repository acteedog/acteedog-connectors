package enrich

import (
	"fmt"
	"google-calendar-connector/internal/core"
	"time"
)

// ContextEnricher enriches context data from Google Calendar API
type ContextEnricher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewContextEnricher creates a new ContextEnricher
func NewContextEnricher(httpClient HTTPClient, contextType string, params any, logger core.Logger) (*ContextEnricher, error) {
	cfg, err := newConfig(contextType, params)
	if err != nil {
		return nil, err
	}

	return &ContextEnricher{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}, nil
}

// EnrichContext enriches the given context with additional data from Google Calendar
func (e *ContextEnricher) EnrichContext(context *core.Context) (*core.Context, error) {
	e.logger.Info(fmt.Sprintf("Enriching context type: %s", e.config.contextType))

	switch e.config.contextType {
	case core.ResourceTypeSource:
		return e.enrichSource(context)
	case core.ResourceTypeCalendar:
		return e.enrichCalendar(context)
	case core.ResourceTypeEvent:
		return e.enrichEvent(context)
	default:
		return nil, fmt.Errorf("unsupported context type: %s", e.config.contextType)
	}
}

func (e *ContextEnricher) enrichSource(context *core.Context) (*core.Context, error) {
	title := "Google Calendar"
	description := "Activity source from Google Calendar"
	url := core.CalendarAPIBase
	context.Title = &title
	context.Description = &description
	context.Url = &url
	return context, nil
}

func (e *ContextEnricher) enrichCalendar(context *core.Context) (*core.Context, error) {
	calendarID := e.config.enrichmentParams.CalendarID
	if calendarID == "" {
		return nil, fmt.Errorf("calendar_id not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching calendar: %s", calendarID))

	resp, err := e.httpClient.FetchCalendarDetail(calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar detail: %w", err)
	}

	return e.applyCalendarEnrichment(context, resp)
}

func (e *ContextEnricher) applyCalendarEnrichment(context *core.Context, cal *CalendarDetailResponse) (*core.Context, error) {
	title := cal.Summary
	description := cal.Description
	context.Title = &title
	context.Description = &description

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["timezone"] = cal.TimeZone
	metadataMap["access_role"] = cal.AccessRole

	context.Metadata = metadataMap

	return context, nil
}

func (e *ContextEnricher) enrichEvent(context *core.Context) (*core.Context, error) {
	calendarID := e.config.enrichmentParams.CalendarID
	eventID := e.config.enrichmentParams.EventID
	if calendarID == "" || eventID == "" {
		return nil, fmt.Errorf("calendar_id and event_id required in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching event: %s/%s", calendarID, eventID))

	resp, err := e.httpClient.FetchEventDetail(calendarID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event detail: %w", err)
	}

	return e.applyEventEnrichment(context, resp)
}

func (e *ContextEnricher) applyEventEnrichment(context *core.Context, evt *core.Event) (*core.Context, error) {
	title := evt.Summary
	description := evt.Description
	url := evt.HtmlLink

	context.Title = &title
	context.Description = &description
	if url != "" {
		context.Url = &url
	}

	if evt.Created != "" {
		if t, err := time.Parse(time.RFC3339, evt.Created); err == nil {
			utc := t.UTC()
			context.CreatedAt = &utc
		}
	}
	if evt.Updated != "" {
		if t, err := time.Parse(time.RFC3339, evt.Updated); err == nil {
			utc := t.UTC()
			context.UpdatedAt = &utc
		}
	}

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["full_description"] = evt.Description
	if evt.Location != "" {
		metadataMap["location"] = evt.Location
	}
	if evt.Creator != nil && evt.Creator.Email != "" {
		metadataMap["creator_email"] = evt.Creator.Email
	}
	if evt.Organizer != nil {
		metadataMap["organizer_email"] = evt.Organizer.Email
	}
	if len(evt.Attendees) > 0 {
		attendees := make([]string, 0, len(evt.Attendees))
		for _, a := range evt.Attendees {
			attendees = append(attendees, a.Email)
		}
		metadataMap["attendees"] = attendees
	}
	if evt.ConferenceData != nil {
		for _, ep := range evt.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				metadataMap["conference_link"] = ep.URI
				break
			}
		}
	}

	context.Metadata = metadataMap

	return context, nil
}

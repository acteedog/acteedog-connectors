package core

import (
	"fmt"
	"time"
)

// Context represents a context object
type Context struct {
	ConnectorId  string
	CreatedAt    *time.Time
	Description  *string
	Id           string
	Metadata     any
	Name         string
	ParentId     string
	ResourceType string
	Title        *string
	UpdatedAt    *time.Time
	Url          *string
}

// ContextGenerator provides factory methods for creating standardized Context objects
type ContextGenerator struct {
	connectorID string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator() *ContextGenerator {
	return &ContextGenerator{
		connectorID: ConnectorID,
	}
}

// CreateSourceContext creates a source context for Google Calendar
func (g *ContextGenerator) CreateSourceContext() *Context {
	id := MakeSourceContextID()
	title := "Google Calendar"
	description := "Activity source from Google Calendar"
	url := CalendarAPIBase
	return &Context{
		Id:           id,
		Name:         id,
		ParentId:     "",
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeSource,
		Title:        &title,
		Description:  &description,
		Url:          &url,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
}

// CreateCalendarContext creates a calendar context
func (g *ContextGenerator) CreateCalendarContext(calendarID, calendarName string) *Context {
	id := MakeCalendarContextID(calendarID)
	parentID := MakeSourceContextID()
	name := fmt.Sprintf("calendar %s", calendarName)
	title := calendarName
	return &Context{
		Id:           id,
		Name:         name,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeCalendar,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"calendar_id": calendarID,
			},
		},
	}
}

// CreateEventContext creates an event context
func (g *ContextGenerator) CreateEventContext(calendarID, eventID, eventTitle string) *Context {
	id := MakeEventContextID(calendarID, eventID)
	parentID := MakeCalendarContextID(calendarID)
	name := fmt.Sprintf("event %s", eventTitle)
	title := eventTitle
	return &Context{
		Id:           id,
		Name:         name,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeEvent,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"calendar_id": calendarID,
				"event_id":    eventID,
			},
		},
	}
}

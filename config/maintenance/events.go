package maintenance

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/TwiN/logr"
)

// MaintenanceEvent represents a one-off scheduled maintenance window for one or more endpoints.
type MaintenanceEvent struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
}

// IsActive returns true if the event is currently in progress.
func (e *MaintenanceEvent) IsActive() bool {
	now := time.Now()
	return now.After(e.Start) && now.Before(e.End)
}

// IsUpcoming returns true if the event starts within the given duration from now.
func (e *MaintenanceEvent) IsUpcoming(within time.Duration) bool {
	now := time.Now()
	return e.Start.After(now) && e.Start.Before(now.Add(within))
}

// EventsStore holds maintenance events loaded from a JSON file, keyed by endpoint key.
// The file is re-read whenever its modification time changes.
type EventsStore struct {
	filePath string
	mu       sync.RWMutex
	events   map[string][]MaintenanceEvent
	lastMod  time.Time
}

var globalEventsStore *EventsStore

// InitEventsStore initialises the package-level store from the given file path.
// Passing an empty path is valid and results in an always-empty store.
func InitEventsStore(filePath string) *EventsStore {
	s := &EventsStore{
		filePath: filePath,
		events:   make(map[string][]MaintenanceEvent),
	}
	if filePath != "" {
		s.reload()
	}
	globalEventsStore = s
	return s
}

// GetEventsStore returns the package-level EventsStore (nil if never initialised).
func GetEventsStore() *EventsStore {
	return globalEventsStore
}

// ReloadIfModified re-reads the file only when its mtime has changed since the last load.
func (s *EventsStore) ReloadIfModified() {
	if s == nil || s.filePath == "" {
		return
	}
	info, err := os.Stat(s.filePath)
	if err != nil {
		return
	}
	if info.ModTime().After(s.lastMod) {
		s.reload()
	}
}

// GetAllEvents returns a snapshot of all events, keyed by endpoint key.
func (s *EventsStore) GetAllEvents() map[string][]MaintenanceEvent {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]MaintenanceEvent, len(s.events))
	for k, v := range s.events {
		out[k] = v
	}
	return out
}

// GetEventsForEndpoint returns all maintenance events registered for the given endpoint key.
func (s *EventsStore) GetEventsForEndpoint(key string) []MaintenanceEvent {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events[key]
}

// IsEndpointUnderMaintenance returns true when the endpoint currently has an active maintenance event.
func (s *EventsStore) IsEndpointUnderMaintenance(key string) bool {
	for _, ev := range s.GetEventsForEndpoint(key) {
		if ev.IsActive() {
			return true
		}
	}
	return false
}

func (s *EventsStore) reload() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		logr.Errorf("[maintenance.EventsStore] Failed to read events file %s: %v", s.filePath, err)
		return
	}
	var events map[string][]MaintenanceEvent
	if err := json.Unmarshal(data, &events); err != nil {
		logr.Errorf("[maintenance.EventsStore] Failed to parse events file %s: %v", s.filePath, err)
		return
	}
	s.mu.Lock()
	s.events = events
	if info, err := os.Stat(s.filePath); err == nil {
		s.lastMod = info.ModTime()
	}
	s.mu.Unlock()
	logr.Infof("[maintenance.EventsStore] Loaded maintenance events from %s (%d endpoints)", s.filePath, len(events))
}

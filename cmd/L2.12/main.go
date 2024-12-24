package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"log"
	"sync"
)

// Event represents a calendar event
type Event struct {
	UserID      int       `json:"user_id"`
	EventID     int       `json:"event_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Location    string    `json:"location"`
}

// ToJSON serializes Event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

var (
	eventsStore = make(map[int]map[int]Event)
	mu          = &sync.RWMutex{}
	eventIDGen  = 1 // Simple incrementing ID generator
)

// Middleware for logging requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Helper function to parse int parameter
func parseInt(params url.Values, key string) (int, error) {
	value := params.Get(key)
	if value == "" {
		return 0, fmt.Errorf("missing %s parameter", key)
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", key, err)
	}
	return i, nil
}

// Helper function to parse date parameter
func parseDate(params url.Values, key string) (time.Time, error) {
	value := params.Get(key)
	if value == "" {
		return time.Time{}, fmt.Errorf("missing %s parameter", key)
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s parameter: %s", key, err)
	}
	return t, nil
}

// Helper function to get the next unique event ID
func getNextEventID() int {
	mu.Lock()
	defer mu.Unlock()
	eventId := eventIDGen
	eventIDGen++
	return eventId
}

// Helper function to get events for a date range
func getEventsForRange(userID int, startTime, endTime time.Time) []Event {
	var events []Event
	mu.RLock()
	defer mu.RUnlock()
	if eventsStore[userID] == nil {
		return events
	}
	for _, event := range eventsStore[userID] {
		if event.StartTime.After(startTime) && event.StartTime.Before(endTime) {
			events = append(events, event)
		}
	}
	return events
}

// Business logic functions
func createEvent(userID, eventID int, event Event) error {
	mu.Lock()
	defer mu.Unlock()
	if eventsStore[userID] == nil {
		eventsStore[userID] = make(map[int]Event)
	}
	if _, exists := eventsStore[userID][eventID]; exists {
		return fmt.Errorf("event ID %d already exists for user %d", eventID, userID)
	}
	eventsStore[userID][eventID] = event
	return nil
}

func updateEvent(userID, eventID int, updates map[string]string) error {
	mu.Lock()
	defer mu.Unlock()
	if eventsStore[userID] == nil {
		return fmt.Errorf("user %d has no events", userID)
	}
	event, exists := eventsStore[userID][eventID]
	if !exists {
		return fmt.Errorf("event ID %d not found for user %d", eventID, userID)
	}
	// Apply updates to the event
	if title, ok := updates["title"]; ok {
		event.Title = title
	}
	if description, ok := updates["description"]; ok {
		event.Description = description
	}
	if location, ok := updates["location"]; ok {
		event.Location = location
	}
	// Update other fields as needed
	eventsStore[userID][eventID] = event
	return nil
}

func deleteEvent(userID, eventID int) error {
	mu.Lock()
	defer mu.Unlock()
	if eventsStore[userID] == nil {
		return fmt.Errorf("user %d has no events", userID)
	}
	_, exists := eventsStore[userID][eventID]
	if !exists {
		return fmt.Errorf("event ID %d not found for user %d", eventID, userID)
	}
	delete(eventsStore[userID], eventID)
	return nil
}

// Handler for POST /create_event
func createEventHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := r.Form

	userID, err := parseInt(params, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	title := params.Get("title")
	if title == "" {
		respondWithError(w, http.StatusBadRequest, "missing title parameter")
		return
	}

	startTime, err := parseDate(params, "start_time")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	endTime, err := parseDate(params, "end_time")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	eventID := getNextEventID()

	event := Event{
		UserID:    userID,
		EventID:   eventID,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
	}

	err = createEvent(userID, eventID, event)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Event created"})
}

// Handler for POST /update_event
func updateEventHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := r.Form

	userID, err := parseInt(params, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	eventID, err := parseInt(params, "event_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	updates := make(map[string]string)
	if title := params.Get("title"); title != "" {
		updates["title"] = title
	}
	if description := params.Get("description"); description != "" {
		updates["description"] = description
	}
	if location := params.Get("location"); location != "" {
		updates["location"] = location
	}

	if len(updates) == 0 {
		respondWithError(w, http.StatusBadRequest, "no updates provided")
		return
	}

	err = updateEvent(userID, eventID, updates)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Event updated"})
}

// Handler for POST /delete_event
func deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := r.Form

	userID, err := parseInt(params, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	eventID, err := parseInt(params, "event_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = deleteEvent(userID, eventID)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Event deleted"})
}

// Handler for GET /events_for_day
func eventsForDayHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	userID, err := parseInt(queryParams, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	date, err := parseDate(queryParams, "date")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	startTime := date
	endTime := date.AddDate(0, 0, 1)

	events := getEventsForRange(userID, startTime, endTime)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"result": events})
}

// Handler for GET /events_for_week
func eventsForWeekHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	userID, err := parseInt(queryParams, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	date, err := parseDate(queryParams, "date")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	startTime := date
	endTime := date.AddDate(0, 0, 7)

	events := getEventsForRange(userID, startTime, endTime)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"result": events})
}

// Handler for GET /events_for_month
func eventsForMonthHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	userID, err := parseInt(queryParams, "user_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	date, err := parseDate(queryParams, "date")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	startTime := date
	// Calculate end of month
	nextMonth := date.AddDate(0, 1, 0)
	endTime := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, date.Location())

	events := getEventsForRange(userID, startTime, endTime)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"result": events})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// respondWithError sends an error JSON response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func main() {
	// Configure port from config (for simplicity, hardcoding here)
	port := ":8080"

	// Register handlers with middleware
	http.HandleFunc("/create_event", createEventHandler)
	http.HandleFunc("/update_event", updateEventHandler)
	http.HandleFunc("/delete_event", deleteEventHandler)
	http.HandleFunc("/events_for_day", eventsForDayHandler)
	http.HandleFunc("/events_for_week", eventsForWeekHandler)
	http.HandleFunc("/events_for_month", eventsForMonthHandler)

	// Apply logging middleware
	http.Handle("/", loggingMiddleware(http.DefaultServeMux))

	log.Printf("Starting server on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

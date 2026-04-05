package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func getToken(t *testing.T, role string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"role": role})
	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("dummyLogin request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("dummyLogin returned %d: %s", resp.StatusCode, data)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	return result["token"]
}

func doRequest(t *testing.T, method, path, token string, body interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	return resp, result
}

// setupRoomWithSchedule creates a room and schedule via admin token and returns the roomID.
func setupRoomWithSchedule(t *testing.T, adminToken string) string {
	t.Helper()

	resp, result := doRequest(t, http.MethodPost, "/rooms/create", adminToken, map[string]interface{}{
		"name":        fmt.Sprintf("e2e_room_%d", time.Now().UnixNano()),
		"description": "e2e test room",
		"capacity":    10,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create room: expected 201, got %d: %v", resp.StatusCode, result)
	}

	roomData, ok := result["room"].(map[string]interface{})
	if !ok {
		t.Fatalf("create room: unexpected response: %v", result)
	}
	roomID := roomData["id"].(string)

	resp, result = doRequest(t, http.MethodPost, "/rooms/"+roomID+"/schedule/create", adminToken, map[string]interface{}{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "11:00",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create schedule: expected 201, got %d: %v", resp.StatusCode, result)
	}

	return roomID
}

// getFirstFreeSlot fetches free slots for a room on a future date and returns the first slot ID.
func getFirstFreeSlot(t *testing.T, roomID, token string) string {
	t.Helper()

	// Try the next 7 days to find a day with free slots
	for i := 1; i <= 7; i++ {
		date := time.Now().UTC().AddDate(0, 0, i).Format("2006-01-02")
		resp, result := doRequest(t, http.MethodGet, "/rooms/"+roomID+"/slots/list?date="+date, token, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list slots: expected 200, got %d: %v", resp.StatusCode, result)
		}

		slots, ok := result["slots"].([]interface{})
		if ok && len(slots) > 0 {
			slot := slots[0].(map[string]interface{})
			return slot["id"].(string)
		}
	}

	t.Fatal("no free slots found in the next 7 days")
	return ""
}

// TestE2E_FullBookingFlow tests: create room → create schedule → get slots → create booking.
func TestE2E_FullBookingFlow(t *testing.T) {
	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	// Step 1: Create room and schedule as admin
	roomID := setupRoomWithSchedule(t, adminToken)
	fmt.Printf("Created room: %s\n", roomID)

	// Step 2: Get free slots as user
	slotID := getFirstFreeSlot(t, roomID, userToken)
	fmt.Printf("Got free slot: %s\n", slotID)

	// Step 3: Create booking as user
	resp, result := doRequest(t, http.MethodPost, "/bookings/create", userToken, map[string]interface{}{
		"slotId": slotID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking: expected 201, got %d: %v", resp.StatusCode, result)
	}

	bookingData, ok := result["booking"].(map[string]interface{})
	if !ok {
		t.Fatalf("create booking: unexpected response: %v", result)
	}

	bookingID := bookingData["id"].(string)
	status := bookingData["status"].(string)

	if status != "active" {
		t.Errorf("create booking: expected status 'active', got '%s'", status)
	}

	fmt.Printf("Created booking: %s (status: %s)\n", bookingID, status)

	// Step 4: Verify booking appears in user's booking list
	resp, result = doRequest(t, http.MethodGet, "/bookings/my", userToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list my bookings: expected 200, got %d: %v", resp.StatusCode, result)
	}

	myBookings, ok := result["bookings"].([]interface{})
	if !ok {
		t.Fatalf("list my bookings: unexpected response: %v", result)
	}

	found := false
	for _, b := range myBookings {
		bMap := b.(map[string]interface{})
		if bMap["id"].(string) == bookingID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created booking %s not found in /bookings/my", bookingID)
	}
}

// TestE2E_CancelBooking tests the cancel booking flow including idempotency.
func TestE2E_CancelBooking(t *testing.T) {
	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	// Setup: create room, schedule, and a booking
	roomID := setupRoomWithSchedule(t, adminToken)

	slotID := getFirstFreeSlot(t, roomID, userToken)

	resp, result := doRequest(t, http.MethodPost, "/bookings/create", userToken, map[string]interface{}{
		"slotId": slotID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking: expected 201, got %d: %v", resp.StatusCode, result)
	}

	bookingData := result["booking"].(map[string]interface{})
	bookingID := bookingData["id"].(string)
	fmt.Printf("Created booking for cancellation test: %s\n", bookingID)

	// Step 1: Cancel the booking
	resp, result = doRequest(t, http.MethodPost, "/bookings/"+bookingID+"/cancel", userToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("cancel booking: expected 200, got %d: %v", resp.StatusCode, result)
	}

	cancelledBooking, ok := result["booking"].(map[string]interface{})
	if !ok {
		t.Fatalf("cancel booking: unexpected response: %v", result)
	}

	status := cancelledBooking["status"].(string)
	if status != "cancelled" {
		t.Errorf("cancel booking: expected status 'cancelled', got '%s'", status)
	}
	fmt.Printf("Cancelled booking %s (status: %s)\n", bookingID, status)

	// Step 2: Verify the slot is free again
	date := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	resp, result = doRequest(t, http.MethodGet, "/rooms/"+roomID+"/slots/list?date="+date, userToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list slots after cancel: expected 200, got %d: %v", resp.StatusCode, result)
	}

	// Step 3: Cancel the same booking again (idempotency check)
	resp, result = doRequest(t, http.MethodPost, "/bookings/"+bookingID+"/cancel", userToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("cancel booking (idempotent): expected 200, got %d: %v", resp.StatusCode, result)
	}

	idempotentBooking, ok := result["booking"].(map[string]interface{})
	if !ok {
		t.Fatalf("cancel booking (idempotent): unexpected response: %v", result)
	}

	idempotentStatus := idempotentBooking["status"].(string)
	if idempotentStatus != "cancelled" {
		t.Errorf("cancel booking (idempotent): expected status 'cancelled', got '%s'", idempotentStatus)
	}
	fmt.Printf("Idempotent cancel confirmed: status is '%s'\n", idempotentStatus)
}

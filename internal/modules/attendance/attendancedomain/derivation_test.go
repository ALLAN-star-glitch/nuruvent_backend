package attendancedomain

import (
	"testing"
	"time"
)

func TestDeriveSessionStatus_FullAttendance(t *testing.T) {
	start := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	session := HydrateSession(
		"s1",
		ExternalRef{Type: "event", ID: "e1"},
		"schedule-1",
		"Workshop",
		start,
		start.Add(60*time.Minute),
		60,
		ProviderZoom,
		"123",
		"https://zoom.us/j/123",
		SessionStatusEnded,
		start, start,
	)

	join := start
	leave := start.Add(55 * time.Minute)
	record := HydrateAttendanceRecord(
		"r1", "a1", "s1",
		join, &leave, 55*60,
		SourceZoomWebhook,
		start, start,
	)

	status := DeriveSessionStatus(
		session,
		[]*AttendanceRecord{record},
		DefaultDerivationPolicy(),
		start.Add(2*time.Hour),
	)
	if status != StatusFull {
		t.Fatalf("expected Full, got %s", status)
	}
}

func TestDeriveSessionStatus_PartialAttendance(t *testing.T) {
	start := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	session := HydrateSession(
		"s1",
		ExternalRef{Type: "event", ID: "e1"},
		"schedule-1",
		"Workshop",
		start,
		start.Add(60*time.Minute),
		60,
		ProviderZoom,
		"123",
		"https://zoom.us/j/123",
		SessionStatusEnded,
		start, start,
	)

	join := start
	leave := start.Add(20 * time.Minute)
	record := HydrateAttendanceRecord(
		"r1", "a1", "s1",
		join, &leave, 20*60,
		SourceZoomWebhook,
		start, start,
	)

	status := DeriveSessionStatus(
		session,
		[]*AttendanceRecord{record},
		DefaultDerivationPolicy(),
		start.Add(2*time.Hour),
	)
	if status != StatusPartial {
		t.Fatalf("expected Partial, got %s", status)
	}
}

func TestDeriveSessionStatus_NoShow(t *testing.T) {
	start := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	session := HydrateSession(
		"s1",
		ExternalRef{Type: "event", ID: "e1"},
		"schedule-1",
		"Workshop",
		start,
		start.Add(60*time.Minute),
		60,
		ProviderZoom,
		"123",
		"https://zoom.us/j/123",
		SessionStatusEnded,
		start, start,
	)

	status := DeriveSessionStatus(
		session,
		nil,
		DefaultDerivationPolicy(),
		start.Add(2*time.Hour),
	)
	if status != StatusNoShow {
		t.Fatalf("expected NoShow, got %s", status)
	}
}

func TestDeriveSessionStatus_Joined(t *testing.T) {
	start := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	session := HydrateSession(
		"s1",
		ExternalRef{Type: "event", ID: "e1"},
		"schedule-1",
		"Workshop",
		start,
		start.Add(60*time.Minute),
		60,
		ProviderZoom,
		"123",
		"https://zoom.us/j/123",
		SessionStatusLive,
		start, start,
	)

	record := HydrateAttendanceRecord(
		"r1", "a1", "s1",
		start, nil, 0,
		SourceJoinLink,
		start, start,
	)

	status := DeriveSessionStatus(
		session,
		[]*AttendanceRecord{record},
		DefaultDerivationPolicy(),
		start.Add(10*time.Minute),
	)
	if status != StatusJoined {
		t.Fatalf("expected Joined, got %s", status)
	}
}

func TestDeriveRollupStatus_SingleSessionFull(t *testing.T) {
	start := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)
	s := HydrateSession(
		"s1",
		ExternalRef{Type: "event", ID: "e1"},
		"schedule-1",
		"Workshop",
		start,
		start.Add(60*time.Minute),
		60,
		ProviderZoom,
		"123",
		"https://zoom.us/j/123",
		SessionStatusEnded,
		start, start,
	)

	status, attended, confirmed := DeriveRollupStatus(
		[]*Session{s},
		map[string]AttendanceStatus{"s1": StatusFull},
		map[string]bool{},
		DefaultDerivationPolicy(),
		start.Add(2*time.Hour),
	)
	if status != StatusFull || attended != 1 || confirmed != 0 {
		t.Fatalf("expected (Full, 1, 0), got (%s, %d, %d)", status, attended, confirmed)
	}
}
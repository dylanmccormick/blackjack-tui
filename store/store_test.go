package store

// import (
// 	"testing"
// 	"time"
// )

// func TestIsYesterday(t *testing.T) {
// 	tim, err := time.Parse(time.RFC3339, "2026-02-04T12:14:36Z")
// 	if err != nil {
// 		t.Fatal("error parsing time")
// 	}
//
// 	if !isToday(tim) {
// 		t.Fatalf("expected isToday to be %v got=%v", isToday(tim), true)
// 	}
//
// 	tim, err = time.Parse(time.RFC3339, "2026-02-03T12:14:36Z")
// 	if err != nil {
// 		t.Fatal("error parsing time")
// 	}
// 	if !isYesterday(tim) {
// 		t.Fatalf("expected isToday to be %v got=%v", isToday(tim), true)
// 	}
// }

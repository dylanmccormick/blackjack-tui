package auth

// import (
// 	"context"
// 	"fmt"
// 	"net/http/httptest"
// 	"testing"
// 	"time"
// )
//
// func TestPost(t *testing.T) {
// 	sm := NewSessionManager()
// 	sm.Run(context.TODO())
//
// 	session := &Session{}
// 	session.SessionId = "1234"
// 	sendDeviceRequest(session)
// 	t.Logf("%#v\n", session)
// 	sm.AddSession(session)
//
// 	ticker := time.NewTicker(10 * time.Second)
//
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/", nil)
//
// 	for range ticker.C {
// 		if HandleAuthCheck(sm, session.SessionId, w, r) {
// 			ticker.Stop()
// 			fmt.Printf("%#v\n", session)
// 		}
// 	}
// }

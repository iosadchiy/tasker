package tasker

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erikstmartin/go-testdb"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
)

func TestHandlers_HandleRead(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		exists   bool
		wantCode int
		wantBody string
	}{
		{
			name:     "found",
			id:       "3a70ec46-8249-45c8-b758-39155bb8294f",
			exists:   true,
			wantCode: 200,
			wantBody: `{"id":"3a70ec46-8249-45c8-b758-39155bb8294f","status":"running","timestamp":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:     "not found",
			id:       "3a70ec46-8249-45c8-b758-39155bb8294f",
			exists:   false,
			wantCode: 404,
			wantBody: `record not found`,
		},
		{
			name:     "bad request",
			id:       "3",
			wantCode: 400,
			wantBody: `invalid UUID length: 1`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns := []string{"id", "status"}
			if tt.exists {
				testdb.StubQuery(
					`SELECT * FROM "tasks"  WHERE "tasks"."id" = ?`,
					testdb.RowsFromCSVString(columns, tt.id+",running"),
				)
			} else {
				testdb.StubQueryError(
					`SELECT * FROM "tasks"  WHERE "tasks"."id" = ?`,
					gorm.ErrRecordNotFound,
				)
			}
			defer testdb.Reset()

			db, _ := gorm.Open("testdb", "")
			svc := Service{
				db:       db,
				strategy: sleep(50),
			}
			h := &Handlers{
				Service:     svc,
				MaxWaitTime: 50 * time.Millisecond,
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/task", nil)
			r = mux.SetURLVars(r, map[string]string{"id": tt.id})
			h.HandleRead(w, r)
			require.Equal(t, tt.wantCode, w.Code)
			require.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

package tasker

import (
	"context"
	"testing"
	"time"

	"github.com/erikstmartin/go-testdb"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
)

func sleep(ms int) func(ctx context.Context, task Task) {
	return func(context.Context, Task) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func TestService_Create(t *testing.T) {
	testdb.StubExec(
		`INSERT  INTO "tasks" ("id","status","updated_at") VALUES (?,?,?)`,
		testdb.NewResult(0, nil, 1, nil),
	)
	testdb.StubExec(
		`UPDATE "tasks" SET "status" = ?, "updated_at" = ?  WHERE "tasks"."id" = ? `,
		testdb.NewResult(0, nil, 1, nil),
	)

	db, _ := gorm.Open("testdb", "")
	svc := Service{
		db:       db,
		strategy: sleep(50),
	}
	ctx := context.Background()
	task, err := svc.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, "created", task.Status)

	time.Sleep(100 * time.Millisecond)
}

func TestService_Read(t *testing.T) {
	db, _ := gorm.Open("testdb", "")
	svc := Service{
		db:       db,
		strategy: sleep(50),
	}
	id := "3a70ec46-8249-45c8-b758-39155bb8294f"

	columns := []string{"id", "status"}
	testdb.StubQuery(
		`SELECT * FROM "tasks"  WHERE "tasks"."id" = ?`,
		testdb.RowsFromCSVString(columns, id+",running"),
	)

	task, err := svc.Read(context.Background(), uuid.MustParse(id))
	require.NoError(t, err)
	require.Equal(t, id, task.ID.String())
	require.Equal(t, "running", task.Status)
}

func TestService_Poll(t *testing.T) {
	db, _ := gorm.Open("testdb", "")
	svc := NewService(db)
	svc.strategy = sleep(50)

	testdb.StubExec(
		`INSERT  INTO "tasks" ("id","status","updated_at") VALUES (?,?,?)`,
		testdb.NewResult(0, nil, 1, nil),
	)
	testdb.StubExec(
		`UPDATE "tasks" SET "status" = ?, "updated_at" = ?  WHERE "tasks"."id" = ? `,
		testdb.NewResult(0, nil, 1, nil),
	)

	ctx := context.Background()
	task, err := svc.Create(ctx)
	require.NoError(t, err)

	columns := []string{"id", "status"}
	testdb.StubQuery(
		`SELECT * FROM "tasks"  WHERE "tasks"."id" = ?`,
		testdb.RowsFromCSVString(columns, task.ID.String()+",finished"),
	)

	task, err = svc.Poll(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, "finished", task.Status)
}
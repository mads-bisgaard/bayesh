package bayesh

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupCoreWithTempDB(t testing.TB) (*Core, string) {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv(BayeshDirEnvVar, tempDir)

	settings, err := Setup(context.Background(), OsFs{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	core, err := NewCore(context.Background(), settings)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}

	return core, settings.Database
}

func TestRecordEvent_NewEvent(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	defer func() {
		if err := core.Close(); err != nil {
			t.Errorf("core.Close() failed: %v", err)
		}
	}()

	ctx := context.Background()
	cwd := "/home/user/project"
	previousCmd := "ls -l"
	currentCmd := "git status"

	err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd)
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	// Verify the data was inserted correctly
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB for verification: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	var eventCounter int
	err = db.QueryRow("SELECT event_counter FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, previousCmd, currentCmd).Scan(&eventCounter)
	if err != nil {
		t.Fatalf("Failed to query for new event: %v", err)
	}

	if eventCounter != 1 {
		t.Errorf("expected event_counter to be 1, got %d", eventCounter)
	}
}

func TestRecordEvent_ExistingEvent(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	defer func() {
		if err := core.Close(); err != nil {
			t.Errorf("core.Close() failed: %v", err)
		}
	}()

	ctx := context.Background()
	cwd := "/home/user/project"
	previousCmd := "ls -l"
	currentCmd := "git status"

	// Record the event for the first time
	if err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd); err != nil {
		t.Fatalf("First RecordEvent failed: %v", err)
	}

	// Verify initial state and get timestamp
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB for verification: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	var initialTimestamp time.Time
	err = db.QueryRow("SELECT last_modified FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, previousCmd, currentCmd).Scan(&initialTimestamp)
	if err != nil {
		t.Fatalf("Failed to query for initial timestamp: %v", err)
	}

	// Wait a moment to ensure the timestamp will be different
	time.Sleep(2 * time.Millisecond)

	// Record the same event again
	if err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd); err != nil {
		t.Fatalf("Second RecordEvent failed: %v", err)
	}

	// Verify the event was updated
	var updatedCounter int
	var updatedTimestamp time.Time
	err = db.QueryRow("SELECT event_counter, last_modified FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, previousCmd, currentCmd).Scan(&updatedCounter, &updatedTimestamp)
	if err != nil {
		t.Fatalf("Failed to query for updated event: %v", err)
	}

	if updatedCounter != 2 {
		t.Errorf("expected event_counter to be 2, got %d", updatedCounter)
	}
	if !updatedTimestamp.After(initialTimestamp) {
		t.Errorf("expected updated timestamp %v to be after initial timestamp %v", updatedTimestamp, initialTimestamp)
	}
}

func TestInferCommands(t *testing.T) {
	core, _ := setupCoreWithTempDB(t)
	ctx := context.Background()
	record := func(cwd, previous, current string, n int) {
		for range n {
			if err := core.RecordEvent(ctx, cwd, previous, current); err != nil {
				t.Fatal(err)
			}
		}
	}
	record("/a", "git status", "git add .", 3)
	record("/a", "git status", "git diff", 1)
	record("/b", "ls", "git diff", 4)
	// P(add|/a)=3/4, P(add|status)=3/4, P(add|/a,status)=3/4 => 3/4
	// P(diff|/a)=1/4, P(diff|status)=1/4, P(diff|/a,status)=1/4 => 1/4
	got, err := core.InferCommands(ctx, "/a", "git status")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"git add .", "git diff"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInferCommandsTemplates(t *testing.T) {
	core, _ := setupCoreWithTempDB(t)
	ctx := context.Background()
	for _, e := range [][2]string{
		{"git checkout x", "vim a.go"},
		{"vim a.go", "git checkout y"},
		{"git checkout y", "vim b.go"},
		{"vim b.go", "git checkout z"},
		{"git checkout z", "vim c.go"},
		{"vim c.go", "git checkout x"},
	} {
		if err := core.RecordEvent(ctx, "/a", e[0], e[1]); err != nil {
			t.Fatal(err)
		}
	}
	// The varying arguments are folded, also in the events recorded before
	// the slots became variable, so the counts of each template add up, and
	// a never seen previous command still matches its template.
	got, err := core.InferCommands(ctx, "/elsewhere", "git checkout new")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"vim <ARG>"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestRecordEventKeepsTemplatesUpToDate checks that the templates and slot
// statistics stored while recording equal those learned from all events at once.
func TestRecordEventKeepsTemplatesUpToDate(t *testing.T) {
	core, _ := setupCoreWithTempDB(t)
	ctx := context.Background()
	rng := rand.New(rand.NewSource(1))
	pick := func(values ...string) string { return values[rng.Intn(len(values))] }
	newCmd := func() string {
		v := fmt.Sprint(rng.Intn(6))
		return pick(
			"cd "+pick("a", "b", "c", "d", v),
			"git checkout "+pick("main", "main", "main", v),
			"git status",
			"cat "+v+" | grep "+pick("x", "y", v),
			"make U="+v+" all",
			`git commit -m "message `+v+`"`,
			"P="+v+" serve",
		)
	}
	previous := ""
	for range 300 {
		current := newCmd()
		if err := core.RecordEvent(ctx, "/"+pick("a", "b"), previous, current); err != nil {
			t.Fatal(err)
		}
		previous = current
	}

	rows, err := core.db.Query("SELECT previous_cmd, current_cmd, previous_template, current_template, event_counter FROM events")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var events []TemplatedRow
	want := newTemplater()
	for rows.Next() {
		var r TemplatedRow
		var n int
		if err := rows.Scan(&r.PreviousCmd, &r.CurrentCmd, &r.PreviousTemplate, &r.CurrentTemplate, &n); err != nil {
			t.Fatal(err)
		}
		events = append(events, r)
		want.add(r.CurrentCmd, n)
	}
	folded := 0
	for _, r := range events {
		if p, c := want.Template(r.PreviousCmd), want.Template(r.CurrentCmd); p != r.PreviousTemplate || c != r.CurrentTemplate {
			t.Errorf("%q -> %q: stored templates %q -> %q, want %q -> %q", r.PreviousCmd, r.CurrentCmd, r.PreviousTemplate, r.CurrentTemplate, p, c)
		}
		if r.CurrentTemplate != r.CurrentCmd {
			folded++
		}
	}
	if folded == 0 {
		t.Error("expected some templates to fold arguments")
	}
	stored := map[string]map[string]int{}
	slots, err := core.db.Query("SELECT slot, value, runs FROM slots")
	if err != nil {
		t.Fatal(err)
	}
	defer slots.Close()
	for slots.Next() {
		var slot, value string
		var n int
		if err := slots.Scan(&slot, &value, &n); err != nil {
			t.Fatal(err)
		}
		if stored[slot] == nil {
			stored[slot] = map[string]int{}
		}
		stored[slot][value] = n
	}
	if len(stored) != len(want.slots) {
		t.Errorf("stored %d slots, want %d", len(stored), len(want.slots))
	}
	for slot, s := range want.slots {
		if !maps.Equal(stored[slot], s.runs) || len(stored[slot]) != s.distinct {
			t.Errorf("slot %q: stored %v, want %+v", slot, stored[slot], s)
		}
	}
}

// TestInferCommandsDuringLargeWrite checks that opening Bayesh is not blocked
// while record-event re-templates many rows. Here the writer's transaction
// outgrows its page cache, which in SQLite's default rollback journal mode
// locks out readers until it commits.
func TestInferCommandsDuringLargeWrite(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	ctx := context.Background()
	if err := core.RecordEvent(ctx, "/a", "ls", "git status"); err != nil {
		t.Fatal(err)
	}

	writerDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer writerDB.Close()
	writer, err := writerDB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	for _, stmt := range []string{"PRAGMA cache_size = 1", "BEGIN IMMEDIATE"} {
		if _, err := writer.ExecContext(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	defer writer.ExecContext(ctx, "ROLLBACK")
	for i := range 2000 {
		if _, err := writer.ExecContext(ctx, "INSERT INTO slots (slot, value, runs) VALUES (?, ?, 1)", fmt.Sprintf("slot%d", i), "value"); err != nil {
			t.Fatal(err)
		}
	}

	// A reader without a busy timeout fails at once instead of waiting
	readerDB, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=0")
	if err != nil {
		t.Fatal(err)
	}
	defer readerDB.Close()
	reader := &Core{Settings: core.Settings, db: readerDB, now: time.Now}
	got, err := reader.InferCommands(ctx, "/a", "ls")
	if err != nil {
		t.Fatalf("InferCommands during a large write: %v", err)
	}
	if want := []string{"git status"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestConcurrentRecordEvents checks that record-events running at the same
// time, as they do when commands finish in several shells, wait for each other
// instead of failing. WAL mode lets readers run alongside a writer, but
// writers still take turns.
func TestConcurrentRecordEvents(t *testing.T) {
	first, dbPath := setupCoreWithTempDB(t)
	ctx := context.Background()
	const writers, events = 8, 20
	errs := make(chan error, writers)
	for w := range writers {
		go func() {
			// One core per writer, like separate record-event processes
			core, err := NewCore(ctx, first.Settings)
			if err != nil {
				errs <- err
				return
			}
			defer core.Close()
			for i := range events {
				if err := core.RecordEvent(ctx, "/a", fmt.Sprintf("cd %d", w), fmt.Sprintf("cd %d", i)); err != nil {
					errs <- err
					return
				}
			}
			errs <- nil
		}()
	}
	for range writers {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var total int
	if err := db.QueryRow("SELECT SUM(" + colEventCounter + ") FROM " + eventsTable).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != writers*events {
		t.Errorf("recorded %d events, want %d", total, writers*events)
	}
}

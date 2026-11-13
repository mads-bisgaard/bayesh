package bayesh

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type OsFs struct{}

func (OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (OsFs) Create(name string) (*os.File, error) {
	return os.Create(name)
}
func (OsFs) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
func (OsFs) Getenv(key string) string {
	return os.Getenv(key)
}
func (OsFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

type Core struct {
	Settings *Settings
	db       *sql.DB
	// now returns the current time, which tests can set
	now func() time.Time
}

func NewCore(ctx context.Context, settings *Settings) (*Core, error) {
	db, err := sql.Open("sqlite3", settings.DataSourceName())
	if err != nil {
		return nil, err
	}

	return &Core{
		Settings: settings,
		db:       db,
		now:      time.Now,
	}, nil
}

// Close should be called to gracefully close the database connection.
func (c *Core) Close() error {
	if c != nil && c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *Core) InferCommands(ctx context.Context, cwd string, previousCmd string) ([]string, error) {
	queries := New(c.db)
	now := c.now()
	previous, err := c.template(ctx, previousCmd)
	if err != nil {
		return nil, err
	}
	byCwd, err := queries.TemplateCounts(ctx, now, &cwd, nil)
	if err != nil {
		return nil, err
	}
	byPrevious, err := queries.TemplateCounts(ctx, now, nil, &previous)
	if err != nil {
		return nil, err
	}
	byBoth, err := queries.TemplateCounts(ctx, now, &cwd, &previous)
	if err != nil {
		return nil, err
	}
	return RankCommands(byCwd, byPrevious, byBoth), nil
}

// template returns the template of cmd given the recorded events.
func (c *Core) template(ctx context.Context, cmd string) (string, error) {
	t := newTemplater()
	if err := loadSlots(ctx, New(c.db), t, cmd); err != nil {
		return "", err
	}
	return t.Template(cmd), nil
}

func (c *Core) RecordEvent(ctx context.Context, cwd string, previousCmd string, currentCmd string) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != sql.ErrTxDone {
			log.Fatalf("Error rolling back transaction: %v", err)
		}
	}()
	if err := recordEvent(ctx, New(c.db).WithTx(tx), c.now(), cwd, previousCmd, currentCmd); err != nil {
		return err
	}
	return tx.Commit()
}

// recordEvent stores the commands as typed, run at time now, together with their templates. When
// the new run changes whether an argument slot is folded, the stored templates
// of the commands with that slot are updated too.
func recordEvent(ctx context.Context, queries *Queries, now time.Time, cwd string, previousCmd string, currentCmd string) error {
	t := newTemplater()
	if err := loadSlots(ctx, queries, t, previousCmd, currentCmd); err != nil {
		return err
	}
	if err := loadThresholdValues(ctx, queries, t, currentCmd); err != nil {
		return err
	}
	changed := t.add(currentCmd, 1)
	for _, a := range t.parse(currentCmd) {
		if a.slot == "" {
			continue
		}
		if err := queries.AddSlotRuns(ctx, a.slot, a.value, 1, now); err != nil {
			return err
		}
	}

	row, err := queries.GetRow(ctx, cwd, previousCmd, currentCmd)
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	row.Cwd, row.PreviousCmd, row.CurrentCmd = cwd, previousCmd, currentCmd
	row.PreviousTemplate, row.CurrentTemplate = t.Template(previousCmd), t.Template(currentCmd)
	row.EventCounter++
	row.LastModified = now
	if exists {
		err = queries.UpdateRow(ctx, row)
	} else {
		err = queries.InsertRow(ctx, row)
	}
	if err != nil || len(changed) == 0 {
		return err
	}

	// Every command with an argument given to one of the changed commands contains its name
	rows, err := queries.RowsMentioning(ctx, changed)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := loadSlots(ctx, queries, t, r.PreviousCmd, r.CurrentCmd); err != nil {
			return err
		}
		previous, current := t.Template(r.PreviousCmd), t.Template(r.CurrentCmd)
		if previous != r.PreviousTemplate || current != r.CurrentTemplate {
			if err := queries.SetTemplates(ctx, r.ID, previous, current); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadSlots loads the statistics of the argument slots of cmds, and the runs of
// their values, which t doesn't have yet.
func loadSlots(ctx context.Context, queries *Queries, t *templater, cmds ...string) error {
	for _, cmd := range cmds {
		for _, a := range t.parse(cmd) {
			if a.slot == "" {
				continue
			}
			s, ok := t.slots[a.slot]
			if !ok {
				distinct, total, err := queries.SlotStats(ctx, a.slot)
				if err != nil {
					return err
				}
				s = &slotStats{distinct: distinct, total: total, runs: map[string]int{}}
				t.slots[a.slot] = s
			}
			if _, ok := s.runs[a.value]; !ok {
				runs, err := queries.SlotRuns(ctx, a.slot, a.value)
				if err != nil {
					return err
				}
				s.runs[a.value] = runs
			}
		}
	}
	return nil
}

// loadThresholdValues loads the values which stop being literal when cmd is
// run once more, so that templater.add can report it. A run raises the runs
// the values of each slot of cmd need to stay literal (see literalRuns), and
// those with runs in between lose it. Most runs raise it by none, so no query
// is made. The slots of cmd must be loaded.
func loadThresholdValues(ctx context.Context, queries *Queries, t *templater, cmd string) error {
	added := map[string]int{}
	for _, a := range t.parse(cmd) {
		if a.slot != "" {
			added[a.slot]++
		}
	}
	for slot, n := range added {
		s := t.slots[slot]
		lo, hi := literalRuns(s.total), literalRuns(s.total+n)
		if lo == hi {
			continue
		}
		values, err := queries.SlotValuesWithRuns(ctx, slot, lo, hi)
		if err != nil {
			return err
		}
		for value, runs := range values {
			s.runs[value] = runs
		}
	}
	return nil
}

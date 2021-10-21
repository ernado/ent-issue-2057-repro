package main

import (
	"context"
	"os"
	"os/signal"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"ent-id-repro/ent"
)

func run(ctx context.Context, lg *zap.Logger) error {
	client, err := ent.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		return xerrors.Errorf("db open: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			lg.Error("db: close", zap.Error(err))
		}
	}()
	if err := client.Schema.Create(ctx,
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
	); err != nil {
		return xerrors.Errorf("db schema: %w", err)
	}

	for i := 0; i < 2; i++ {
id, err := client.User.Create().
	SetID(uuid.New()).
	SetName("foo").OnConflict(
	sql.ConflictColumns("name"),
	sql.ResolveWithNewValues(),
).UpdateNewValues().ID(ctx)
		if err != nil {
			return xerrors.Errorf("create: %w", err)
		}
		lg.Info("Created", zap.String("id", id.String()))
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	if err := run(ctx, lg); err != nil {
		lg.Error("Failed", zap.Error(err))
		os.Exit(2)
	}
}

package app

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
)

func TestMapExists_TmaDemandNotFound(t *testing.T) {
	ok, err := mapExists(tmadomain.ErrDemandNotFound, tmadomain.ErrDemandNotFound)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ok {
		t.Fatal("expected not exists")
	}
}

func TestMapExists_LegacyPgxNoLongerMatchesTma(t *testing.T) {
	ok, err := mapExists(tmadomain.ErrDemandNotFound, pgx.ErrNoRows)
	if err == nil {
		t.Fatal("legacy pgx sentinel must not swallow ErrDemandNotFound")
	}
	if ok {
		t.Fatal("expected not exists on error path")
	}
	if !errors.Is(err, tmadomain.ErrDemandNotFound) {
		t.Fatalf("want ErrDemandNotFound, got %v", err)
	}
}

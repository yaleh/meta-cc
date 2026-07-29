package records

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
)

func TestBuildIndexedContent_UnchangedCorpusHydratesOnlyCandidateSession(t *testing.T) {
	project := filepath.Join(t.TempDir(), "project")
	now := time.Now().UTC().Truncate(time.Second)
	p := &fakeProvider{
		id:        conversation.ProviderCodex,
		available: true,
		sessions: []conversation.Session{
			{ID: "hit", Provider: conversation.ProviderCodex, CWD: project, UpdatedAt: now},
			{ID: "miss", Provider: conversation.ProviderCodex, CWD: project, UpdatedAt: now},
		},
		turnsBySession: map[string][]conversation.Turn{
			"hit": {{
				ID: "turn-hit", UserText: "needle in canonical content", Timestamp: now,
				Items: []conversation.Item{{ID: "item-hit", Kind: conversation.ItemKindUserMessage, Role: "user", Text: "needle in canonical content", Timestamp: now}},
			}},
			"miss": {{
				ID: "turn-miss", UserText: "unrelated content", Timestamp: now,
				Items: []conversation.Item{{ID: "item-miss", Kind: conversation.ItemKindUserMessage, Role: "user", Text: "unrelated content", Timestamp: now}},
			}},
		},
	}
	registry := providerpkg.NewRegistry(p)
	filters := []conversation.ProviderID{conversation.ProviderCodex}

	if _, _, used := BuildIndexedContent(context.Background(), registry, filters, "project", project, "needle"); !used {
		t.Fatal("first query should build and use the index")
	}
	p.loadTurnsCalls = nil

	records, warnings, used := BuildIndexedContent(context.Background(), registry, filters, "project", project, "needle")
	if !used {
		t.Fatalf("unchanged query unexpectedly fell back: %v", warnings)
	}
	if len(p.loadTurnsCalls) != 1 || p.loadTurnsCalls[0] != "hit" {
		t.Fatalf("expected canonical hydration of hit only, got LoadTurns calls %#v", p.loadTurnsCalls)
	}
	if len(records) != 1 || records[0]["session_id"] != "hit" {
		t.Fatalf("expected canonical hit record, got %#v", records)
	}
}

func TestBuildIndexedContent_ShortLiteralFallsBack(t *testing.T) {
	p := &fakeProvider{id: conversation.ProviderCodex, available: true}
	_, warnings, used := BuildIndexedContent(context.Background(), providerpkg.NewRegistry(p), []conversation.ProviderID{conversation.ProviderCodex}, "project", t.TempDir(), "xy")
	if used || len(warnings) != 0 {
		t.Fatalf("short substring must fall back cleanly, used=%v warnings=%v", used, warnings)
	}
}

func TestBuildIndexedContent_UnicodeFoldHydratesCandidate(t *testing.T) {
	project := filepath.Join(t.TempDir(), "project")
	now := time.Now().UTC().Truncate(time.Second)
	p := &fakeProvider{
		id: conversation.ProviderCodex, available: true,
		sessions: []conversation.Session{
			{ID: "kelvin", Provider: conversation.ProviderCodex, CWD: project, UpdatedAt: now},
			{ID: "accent", Provider: conversation.ProviderCodex, CWD: project, UpdatedAt: now},
			{ID: "miss", Provider: conversation.ProviderCodex, CWD: project, UpdatedAt: now},
		},
		turnsBySession: map[string][]conversation.Turn{
			"kelvin": {{ID: "t1", UserText: "KELVIN scale", Items: []conversation.Item{{ID: "i1", Kind: conversation.ItemKindUserMessage, Role: "user", Text: "KELVIN scale"}}}},
			"accent": {{ID: "t2", UserText: "ÉCOLE ÅNGSTRÖM", Items: []conversation.Item{{ID: "i2", Kind: conversation.ItemKindUserMessage, Role: "user", Text: "ÉCOLE ÅNGSTRÖM"}}}},
			"miss":   {{ID: "t3", UserText: "unrelated", Items: []conversation.Item{{ID: "i3", Kind: conversation.ItemKindUserMessage, Role: "user", Text: "unrelated"}}}},
		},
	}
	registry := providerpkg.NewRegistry(p)
	filters := []conversation.ProviderID{conversation.ProviderCodex}

	if _, warnings, used := BuildIndexedContent(context.Background(), registry, filters, "project", project, "kelvin"); !used {
		t.Fatalf("initial Unicode-fold query fell back: %v", warnings)
	}
	p.loadTurnsCalls = nil

	for _, query := range []string{"kelvin", "école", "ångström"} {
		p.loadTurnsCalls = nil
		records, warnings, used := BuildIndexedContent(context.Background(), registry, filters, "project", project, query)
		if !used {
			t.Fatalf("query %q unexpectedly fell back: %v", query, warnings)
		}
		if len(records) != 1 {
			t.Fatalf("query %q got records %#v", query, records)
		}
		if len(p.loadTurnsCalls) != 1 || p.loadTurnsCalls[0] == "miss" {
			t.Fatalf("query %q hydrated %#v, want only its matching session", query, p.loadTurnsCalls)
		}
	}
}

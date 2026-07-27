package ftsindex

import "github.com/yaleh/meta-cc/internal/conversation"

// Hydrate resolves a SearchHit's provenance (TurnID/ItemID) against a
// freshly loaded set of canonical Turns (e.g. via provider.Provider's
// LoadTurns(ctx, hit.ThreadID)) and returns the exact Turn/Item, never a
// cached copy. This is the "exact-record hydration" half of the Contract:
// a search hit's Snippet is only for candidate selection — callers needing
// the authoritative content must hydrate.
func Hydrate(turns []conversation.Turn, hit SearchHit) (conversation.Turn, conversation.Item, bool) {
	for _, turn := range turns {
		if turn.ID != hit.TurnID {
			continue
		}
		for idx, item := range turn.Items {
			if itemID(turn.ID, idx, item) == hit.ItemID {
				return turn, item, true
			}
		}
	}
	return conversation.Turn{}, conversation.Item{}, false
}

package memory

import (
	"context"
	"testing"
)

func TestInMemoryStore_DoesNotLeakInternalPointers(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	if err := s.Add(ctx, &Episode{ID: "e1", SessionID: "s1", Role: "user", Content: "hello", Metadata: map[string]string{"k": "v"}}); err != nil {
		t.Fatal(err)
	}

	got, err := s.Get(ctx, "e1")
	if err != nil {
		t.Fatal(err)
	}
	got.Metadata["k"] = "mutated"
	got.Content = "mutated"

	again, err := s.Get(ctx, "e1")
	if err != nil {
		t.Fatal(err)
	}
	if again.Content != "hello" || again.Metadata["k"] != "v" {
		t.Fatalf("internal episode was mutated through returned pointer: %+v", again)
	}
}

func TestSQLiteStoresInMemoryAreIsolated(t *testing.T) {
	ctx := context.Background()
	s1, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s1.Close()
	s2, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	if err := s1.Add(ctx, &Episode{ID: "only-in-s1", SessionID: "s", Role: "user", Content: "x"}); err != nil {
		t.Fatal(err)
	}
	count, err := s2.Count(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("two :memory: stores share state: s2 count=%d", count)
	}
}

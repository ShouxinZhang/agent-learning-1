package backend_test

import (
	"testing"

	"agent-dou-dizhu/internal/tiandi/rules"
)

func TestRulesCatalogContainsRequiredHands(t *testing.T) {
	catalog := rules.CatalogData()

	required := map[string]bool{
		"single":                false,
		"pair":                  false,
		"triple":                false,
		"triple_with_single":    false,
		"triple_with_pair":      false,
		"four_with_two_pairs":   false,
		"four_with_two_singles": false,
		"straight":              false,
		"serial_pairs":          false,
		"plane_base":            false,
		"plane_with_singles":    false,
		"plane_with_pairs":      false,
		"bomb_four":             false,
		"bomb_five_plus":        false,
		"rocket":                false,
		"pure_laizi_bomb":       false,
		"laizi_substitute_bomb": false,
	}

	for _, section := range catalog.Sections {
		for _, item := range section.Items {
			if _, ok := required[item.Key]; ok {
				required[item.Key] = true
			}
		}
	}

	for key, found := range required {
		if !found {
			t.Fatalf("required hand rule %q not found", key)
		}
	}
}

func TestRulesCatalogBombPriorityOrder(t *testing.T) {
	catalog := rules.CatalogData()
	want := []string{
		"rocket",
		"bomb_five_plus",
		"pure_laizi_bomb",
		"bomb_four",
		"laizi_substitute_bomb",
	}

	if len(catalog.BombPriority) != len(want) {
		t.Fatalf("got %d bomb priorities, want %d", len(catalog.BombPriority), len(want))
	}

	for i, expected := range want {
		if catalog.BombPriority[i].Key != expected {
			t.Fatalf("priority %d: got %s want %s", i, catalog.BombPriority[i].Key, expected)
		}
	}
}

func TestRulesCatalogSequenceHighIsA(t *testing.T) {
	catalog := rules.CatalogData()
	if catalog.SequenceHigh != "A" {
		t.Fatalf("sequence high: got %s want A", catalog.SequenceHigh)
	}
}

func TestRulesCatalogContainsComparisonAndLaiziNotes(t *testing.T) {
	catalog := rules.CatalogData()
	if len(catalog.ComparisonNotes) == 0 {
		t.Fatal("expected comparison notes")
	}
	if len(catalog.LaiziResolutionNotes) == 0 {
		t.Fatal("expected laizi resolution notes")
	}
}

package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// renderMindbreakSource renders the given definition to Go source with the
// shared card renderer, collapsing gofmt's field-alignment whitespace so the
// substring assertions do not depend on column padding.
func renderMindbreakSource(t *testing.T, def *game.CardDef) string {
	t.Helper()
	source, err := (Renderer{}).RenderCardSource(
		&ScryfallCard{Name: def.Name, Layout: "normal"},
		[]*game.CardDef{def},
		nil,
		"testcards",
	)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(source), " ")
}

// TestRenderOpponentCastSpellsAlternativeCost proves the renderer emits
// Mindbreak Trap's per-opponent spells-cast {0} alternative as typed source: the
// condition constant, its count threshold, and the explicit {0} mana cost.
func TestRenderOpponentCastSpellsAlternativeCost(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Mindbreak Render",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
		Types:    []types.Card{types.Instant},
		AlternativeCosts: []cost.Alternative{{
			Label:          "Pay {0}",
			ManaCost:       opt.Val(cost.Mana{cost.O(0)}),
			Condition:      cost.AlternativeConditionOpponentCastSpellsThisTurn,
			ConditionCount: 3,
		}},
	}}
	normalized := renderMindbreakSource(t, def)
	for _, want := range []string{
		"Condition: cost.AlternativeConditionOpponentCastSpellsThisTurn,",
		"ConditionCount: 3,",
		"cost.O(0)",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("rendered source missing %q:\n%s", want, normalized)
		}
	}
}

// TestRenderExileTargetSpells proves the renderer emits the variable-count exile
// effect as typed source: the ExileTargetSpells primitive, the all-target
// stack-objects group reference, and the any-number stack-spell target spec.
func TestRenderExileTargetSpells(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Exile Render",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 0,
				MaxTargets: 99,
				Constraint: "any number of target spells",
				Allow:      game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					StackObjectKinds: []game.StackObjectKind{game.StackSpell},
				},
			}},
			Sequence: []game.Instruction{{
				Primitive: game.ExileTargetSpells{Object: game.AllTargetStackObjectsReference(0)},
			}},
		}.Ability()),
	}}
	normalized := renderMindbreakSource(t, def)
	for _, want := range []string{
		"game.ExileTargetSpells{",
		"game.AllTargetStackObjectsReference(0)",
		"game.TargetAllowStackObject",
		"game.StackSpell",
		"MaxTargets: 99,",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("rendered source missing %q:\n%s", want, normalized)
		}
	}
}

package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestRenderPermanentsOnBattlefieldAlternativeCost proves the renderer emits the
// board-state alternative-cost condition (Blasphemous Edict) as typed source:
// the condition constant, its count threshold, and the counted permanent type,
// pulling in the types import the permanent-type literal needs.
func TestRenderPermanentsOnBattlefieldAlternativeCost(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Board Edict",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
		Types:    []types.Card{types.Sorcery},
		AlternativeCosts: []cost.Alternative{{
			Label:                  "Pay {B}",
			ManaCost:               opt.Val(cost.Mana{cost.B}),
			Condition:              cost.AlternativeConditionPermanentsOnBattlefield,
			ConditionCount:         13,
			ConditionPermanentType: types.Creature,
		}},
	}}
	source, err := (Renderer{}).RenderCardSource(
		&ScryfallCard{Name: "Board Edict", Layout: "normal"},
		[]*game.CardDef{def},
		nil,
		"testcards",
	)
	if err != nil {
		t.Fatal(err)
	}
	// Collapse gofmt's field-alignment whitespace so the assertions do not
	// depend on the exact column padding of the rendered struct literal.
	normalized := strings.Join(strings.Fields(source), " ")
	for _, want := range []string{
		"Condition: cost.AlternativeConditionPermanentsOnBattlefield,",
		"ConditionCount: 13,",
		"ConditionPermanentType: types.Creature,",
		`"github.com/natefinch/council4/mtg/game/types"`,
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("rendered source missing %q:\n%s", want, source)
		}
	}
}

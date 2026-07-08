package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableProwl exercises the four mechanics Prowl, Stoic
// Strategist // Prowl, Pursuit Vehicle combines (issue #2826), each through the
// full parse -> compile -> lower -> render path:
//
//   - Front: "Whenever Prowl attacks, exile up to one other target tapped
//     creature or Vehicle. For as long as that card remains exiled, its owner may
//     play it." lowers to a single ExilePermanentForPlay primitive whose target
//     is the up-to-one, source-excluding, tapped creature-or-Vehicle union, linked
//     under the shared self-exile key.
//   - Front: "Whenever a player plays a card exiled with Prowl, you draw a card
//     and convert Prowl." lowers to a PlaysLinkedExileCard trigger (matching the
//     same link key) whose body draws and transforms.
//   - Back: "Whenever another creature or Vehicle you control enters, ..." lowers
//     its type-or-subtype subject union onto Selection.AnyOf and validates.
//   - Back: "If this is the second time this ability has resolved this turn,
//     convert Prowl." marks the ability CountsResolutionsThisTurn and gates the
//     transform on SourceAbilityResolutionOrdinalThisTurn.
func TestGenerateExecutableProwl(t *testing.T) {
	t.Parallel()
	frontPower, frontToughness := "3", "3"
	backPower, backToughness := "2", "3"
	card := &ScryfallCard{
		Name:     "Prowl, Stoic Strategist // Prowl, Pursuit Vehicle",
		Layout:   "transform",
		TypeLine: "Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle",
		CardFaces: []ScryfallCardFace{
			{
				Name:     "Prowl, Stoic Strategist",
				ManaCost: "{3}{W}",
				TypeLine: "Legendary Artifact Creature — Robot",
				OracleText: "More Than Meets the Eye {2}{W} (You may cast this card converted for {2}{W}.)\n" +
					"Whenever Prowl attacks, exile up to one other target tapped creature or Vehicle. For as long as that card remains exiled, its owner may play it.\n" +
					"Whenever a player plays a card exiled with Prowl, you draw a card and convert Prowl.",
				Power:     &frontPower,
				Toughness: &frontToughness,
			},
			{
				Name:     "Prowl, Pursuit Vehicle",
				TypeLine: "Legendary Artifact — Vehicle",
				OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
					"Whenever another creature or Vehicle you control enters, put a +1/+1 counter on Prowl. If this is the second time this ability has resolved this turn, convert Prowl.",
				Power:     &backPower,
				Toughness: &backToughness,
			},
		},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		// Front attack trigger: exile-for-play with the tapped, source-excluding
		// creature-or-Vehicle union target.
		"Event:  game.EventAttackerDeclared,",
		"Primitive: game.ExilePermanentForPlay{",
		`LinkedKey: game.LinkedKey("exiled-with-source"),`,
		"Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Vehicle\")}}}, Tapped: game.TriTrue, ExcludeSource: true}),",
		// Front plays-exiled trigger: draw + convert on the same link key.
		"Event:                game.EventCardPlayedFromExile,",
		`PlaysLinkedExileCard: game.LinkedKey("exiled-with-source"),`,
		// Back enters trigger: type-or-subtype subject union and the
		// Nth-resolution-this-turn gate.
		"SubjectSelection: game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Vehicle\")}}}}",
		"CountsResolutionsThisTurn: true,",
		"SourceAbilityResolutionOrdinalThisTurn: 2,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

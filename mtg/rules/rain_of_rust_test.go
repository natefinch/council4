package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Rain of Rust printing revision 13a6ecc6-83c8-408b-8600-687cdb383ca9,
// Oracle ID 074e467e-f574-4b44-b5a3-798b592a5a87, exact text:
// "Choose one —
// • Destroy target artifact.
// • Destroy target land.
// Entwine {3}{R} (Choose both if you pay the entwine cost.)"
// Scryfall rulings URI https://api.scryfall.com/cards/13a6ecc6-83c8-408b-8600-687cdb383ca9/rulings
// retrieved 2026-11-20; the response had an empty data list.
// Official Wizards Comprehensive Rules effective 2026-11-06, verified through
// https://magic.wizards.com/en/rules; TXT
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// retained at .cardwork/rain-of-rust/MagicCompRules-20261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 700.2/700.2a/c and 601.2b/c require modal choices and targets only for chosen
// modes; 601.2f-h and 702.42a make the extra cost payable when both are chosen;
// 702.42b and 608.2c resolve the modes in printed order; 608.2b rechecks target
// legality; and 701.8a-c defines destruction and its replacement. Runtime trace:
// `EscalateCost` adds one cost per mode after the first
// (`mtg/rules/payment/options.go`), which is one entwine cost for these two
// modes. Each mode receives its own targets (`mtg/rules/effects.go`), and
// `Context.Destroy` uses the Engine's destruction/replacement/zone-change path
// (`mtg/rules/card_program_batch.go`). Prior entwine evidence:
// `mtg/cards/authored/one_dozen_eyes.go` and `mtg/rules/one_dozen_eyes_test.go`.
func TestRainOfRustEntwineDestroysArtifactAndLand(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("074e467e-f574-4b44-b5a3-798b592a5a87")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Rain of Rust" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Target Artifact", Types: []types.Card{types.Artifact},
	}})
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Target Land", Types: []types.Card{types.Land},
	}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	setMainPhasePriority(g, game.Player1)

	engine := NewEngineWithPrograms(nil, registry.Programs())
	cast := engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(artifact.ObjectID),
		game.PermanentTarget(land.ObjectID),
	}, 0, []int{0, 1}))
	if !cast {
		t.Fatal("could not cast Rain of Rust with both modes and the entwine cost")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining after entwined cast = %d, want 0", got)
	}
	stackObject, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("cast Rain of Rust was not put on the stack")
	}
	if len(stackObject.ChosenModes) != 2 || stackObject.ChosenModes[0] != 0 || stackObject.ChosenModes[1] != 1 {
		t.Fatalf("stack modes = %v, want [0 1]", stackObject.ChosenModes)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(land.CardInstanceID) {
		t.Fatal("entwined Rain of Rust did not destroy both targets")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Rain of Rust did not go to its owner's graveyard")
	}
}

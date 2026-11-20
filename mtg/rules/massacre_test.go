package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Massacre Oracle ID ce1866e9-28de-4bcc-9999-11958e6c13a6, Nemesis #58,
// printing revision f05f5d93-50d1-4aa6-af05-383a6808345b. Current Oracle text,
// verified 2026-11-19 from
// https://api.scryfall.com/cards/f05f5d93-50d1-4aa6-af05-383a6808345b:
// {2}{B}{B} Sorcery. "If an opponent controls a Plains and you control a Swamp,
// you may cast this spell without paying its mana cost. All creatures get
// -2/-2 until end of turn."
// Rulings retrieved 2026-11-19 from
// https://api.scryfall.com/cards/f05f5d93-50d1-4aa6-af05-383a6808345b/rulings:
// empty list.
// The official rules page (checked 2026-11-19) lists Comprehensive Rules
// effective 2026-11-06:
// https://magic.wizards.com/en/rules
// Read §§118.9a-d, 601.2a-h, 611.2a,c and 613.4c in the current TXT:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// The workflow's pinned edition is 2026-09-25; the same relevant clauses were
// read in its official TXT (SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// Both editions require the alternative to replace the printed cost only when
// its conditions hold, and leave the spell's mana cost/value unchanged. The
// -2/-2 effect fixes its set of creatures on resolution, lasts until end of
// turn, and applies in the P/T-modification layer.
func TestAuthoredMassacreScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, authored.Massacre.New())
	swamp := addBasicLandPermanent(g, game.Player1, types.Swamp)
	swamp.Tapped = true
	setMainPhasePriority(g, game.Player1)

	casterCreature := addCreatureWithPT(g, game.Player1, 4, 4)
	opponentCreature := addCreatureWithPT(g, game.Player2, 3, 3)
	act := action.CastSpell(spellID, nil, 0, nil)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Massacre was castable without an opponent-controlled Plains and without mana")
	}

	addBasicLandPermanent(g, game.Player2, types.Plains)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Massacre's free alternative was not offered with the caster's Swamp and an opponent's Plains")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("Massacre could not be cast using its conditional alternative cost")
	}
	if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 0 {
		t.Fatalf("Massacre did not use its free alternative: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	for _, tc := range []struct {
		name      string
		permanent *game.Permanent
		want      int
	}{
		{name: "caster's creature", permanent: casterCreature, want: 2},
		{name: "opponent's creature", permanent: opponentCreature, want: 1},
	} {
		if got := effectivePower(g, tc.permanent); got != tc.want {
			t.Errorf("%s power after Massacre = %d, want %d", tc.name, got, tc.want)
		}
		if got, ok := effectiveToughness(g, tc.permanent); !ok || got != tc.want {
			t.Errorf("%s toughness after Massacre = %d (%t), want %d", tc.name, got, ok, tc.want)
		}
	}
}

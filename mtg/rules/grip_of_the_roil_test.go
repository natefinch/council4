package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: OGW #56 printing 6cf33069-013f-43ab-9a31-e67bb6dc996d,
// Oracle ID c3c87292-2fc8-4759-b7b6-18a8b539b5d1; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Exact
// Oracle text: {2}{U} Instant — “Surge {1}{U} (You may cast this spell for its
// surge cost if you or a teammate has cast another spell this turn.) / Tap
// target creature. It doesn't untap during its controller's next untap step. /
// Draw a card.” Scryfall rulings URI
// https://api.scryfall.com/cards/6cf33069-013f-43ab-9a31-e67bb6dc996d/rulings
// fetched 2026-11-21 and returned six WotC rulings dated 2016-01-22: an already
// tapped creature is a legal target and still stays frozen; the effect tracks
// the creature rather than its controller, so a control change before its old
// controller's next untap makes it skip its new controller's next untap; this
// card uses Surge only as a discount (other Surge cards can have paid-cost
// riders), a teammate's earlier spell qualifies, and an instant's prior spell
// may be resolved, countered, or still on the stack; copies of a Surge instant
// or sorcery count as having had Surge paid; Surge does not change mana cost or
// mana value. Raw response: .cardwork/sources/grip-of-the-roil-rulings.json.
// Official Wizards Comprehensive Rules effective 2026-09-25 (the task's pinned
// edition), https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx;
// verified local copy .cardwork/sources/MagicCompRules_20260925.docx (SHA-256
// 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960). Read
// 118.9a-c, 702.117a, 601.2b/c/f-h, 115.1a, 608.2b, 701.26a-b, 502.3, and
// 121.1: Surge is an optional alternative cost that leaves mana value intact;
// the spell chooses and rechecks its creature target, then taps it, skips its
// controller's next untap for it, and draws the controller's top library card.
func TestGripOfTheRoilScenario(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("c3c87292-2fc8-4759-b7b6-18a8b539b5d1")
	if err != nil {
		t.Fatalf("default-registry lookup for Grip of the Roil: %v", err)
	}
	if def.Name != "Grip of the Roil" {
		t.Fatalf("Oracle ID resolved to %q, want Grip of the Roil", def.Name)
	}

	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})
	setMainPhasePriority(g, game.Player1)

	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grip of the Roil target",
		Types: []types.Card{types.Creature},
	}})
	spellID := addCardToHand(g, game.Player1, def)
	drawID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Grip of the Roil draw",
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) || !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Grip of the Roil was not castable for Surge targeting a creature")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 2 ||
		len(spell.Targets) != 1 || spell.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("Surge cast = %+v, want the Surge alternative, two mana, and the creature target", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped || !target.SkipsNextUntap {
		t.Fatal("Grip of the Roil did not tap and freeze its target")
	}
	if !g.Players[game.Player1].Hand.Contains(drawID) {
		t.Fatal("Grip of the Roil did not draw its controller's card")
	}
}

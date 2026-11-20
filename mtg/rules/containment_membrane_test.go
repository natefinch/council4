package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

const containmentMembraneTestOracleID = "e3bf1a95-4363-4915-8f43-b423714861ec"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Containment Membrane Oracle ID e3bf1a95-4363-4915-8f43-b423714861ec,
// printing revision 49dd3963-a4d7-4992-b6f8-753996390bbf. The pinned Oracle
// text is {2}{U}, “Surge {U} (You may cast this spell for its surge cost if
// you or a teammate has cast another spell this turn.) / Enchant creature /
// Enchanted creature doesn't untap during its controller's untap step.”
// The current Scryfall printing record fetched 2026-11-20 from
// https://api.scryfall.com/cards/49dd3963-a4d7-4992-b6f8-753996390bbf matches
// the pinned text and identity.
// Scryfall rulings
// https://api.scryfall.com/cards/49dd3963-a4d7-4992-b6f8-753996390bbf/rulings
// fetched 2026-11-20 and returned six WotC rulings dated 2016-01-22: the Aura
// may enchant an untapped creature; the restriction matters when it would
// untap and doesn't stop other spells or abilities from untapping it; a prior
// spell may have resolved or been countered; Surge changes neither mana cost
// nor mana value; and some Surge cards have additional effects while others
// only have the alternative cost (Containment Membrane has no additional
// effect).
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 118.9a-d, 601.2b-c/f-h and 702.117a (optional Surge alternative cost,
// its cast condition and unchanged mana value); 303.4a-b/e/g-i, 608.2b and
// 608.3b-c, 702.5a-b (Aura target, attachment and resolution); 502.3,
// 611.3a-b and 701.26a-b (the attached static effect continuously prohibits
// its creature's untap-step untap, not untaps from spells or abilities).
func TestContainmentMembraneSurgeAttachesAndStopsUntap(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(containmentMembraneTestOracleID)
	if err != nil {
		t.Fatalf("default registry lookup for Containment Membrane: %v", err)
	}
	if def.Name != "Containment Membrane" {
		t.Fatalf("Oracle ID lookup returned %q, want Containment Membrane", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})

	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Containment Membrane target",
		Types: []types.Card{types.Creature},
	}})
	target.Tapped = true
	auraID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cast := action.CastSpell(auraID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) ||
		!engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Containment Membrane was not castable for its Surge cost targeting the creature")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.AlternativeCost != 1 || obj.ManaSpentToCast != 1 {
		t.Fatalf("Surge cast = %+v, want the first alternative cost and one blue mana paid", obj)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 0 {
		t.Fatalf("blue mana remaining after Surge cast = %d, want 0", got)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatal("Containment Membrane did not enter the battlefield")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != target.ObjectID {
		t.Fatalf("Containment Membrane attached to %v, want creature %v", aura.AttachedTo, target.ObjectID)
	}

	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw step card"}})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !target.Tapped {
		t.Fatal("enchanted creature untapped during its controller's untap step")
	}
}

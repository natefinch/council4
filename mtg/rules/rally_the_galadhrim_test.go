package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Rally the Galadhrim Oracle ID 84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6,
// printing revision 009a2276-0014-45c2-90b7-064594009780 (LTC 499, released
// 2023-11-03), {2}{G}{U} Sorcery. Oracle text: "Create a token that's a copy
// of target creature you control. Conspire (As you cast this spell, you may tap
// two untapped creatures you control that share a color with it. When you do,
// copy it and you may choose a new target for the copy.)" Scryfall rulings at
// https://api.scryfall.com/cards/009a2276-0014-45c2-90b7-064594009780/rulings
// returned 11 entries on 2026-11-21, including the sharing-color payment,
// independent non-cast copy, optional new-target choice, and token copy values.
// Read the official Wizards Comprehensive Rules effective 2026-11-06:
// https://magic.wizards.com/en/rules and
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// (verified local copy .cardwork/sources/MagicCompRules-20261106.docx).
// Rules 111.1-6 govern token ownership, control, and characteristics; 601.2a-i
// governs casting, target declaration, and cost payment; 702.78a-b defines the
// optional two-creature Conspire cost and its cast trigger; 707.2 and 707.10/
// 10c govern copied values, spell-copy decisions, and the optional new target.
func TestRallyTheGaladhrimCreatesTargetedCreatureCopiesWithConspire(t *testing.T) {
	const oracleID = "84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Rally the Galadhrim by Oracle ID: %v", err)
	}
	if def.Name != "Rally the Galadhrim" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	originalTarget := addCombatPermanent(g, game.Player1, vanillaCreature("Original target", 2, 2))
	copyTarget := addCombatPermanent(g, game.Player1, vanillaCreature("Copy target", 3, 5))
	firstPayment := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	secondPayment := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == copyTarget.ObjectID {
					return []int{option.Index}
				}
			}
			t.Fatal("Conspire did not offer the second controlled creature as a new target")
		}
		return chooseConspireCost(req)
	})
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(spellID, []game.Target{game.PermanentTarget(originalTarget.ObjectID)}, 0, nil), agents, log) {
		t.Fatal("could not cast Rally the Galadhrim targeting a creature you control")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 4 ||
		original.Targets[0].PermanentID != originalTarget.ObjectID ||
		!g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("Rally the Galadhrim did not pay its mana and Conspire costs: %+v", original)
	}
	for _, creature := range []*game.Permanent{firstPayment, secondPayment} {
		permanent, ok := permanentByObjectID(g, creature.ObjectID)
		if !ok || !permanent.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !isConspireTrigger(trigger) {
		t.Fatal("the top stack object is not Rally the Galadhrim's Conspire trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.CastCopy || len(copy.Targets) != 1 ||
		copy.Targets[0].PermanentID != copyTarget.ObjectID ||
		original.Targets[0].PermanentID != originalTarget.ObjectID {
		t.Fatalf("Conspire did not create its independently retargeted spell copy: %+v", copy)
	}

	battlefieldSize := len(g.Battlefield)
	engine.resolveTopOfStack(g, log)
	if len(g.Battlefield) != battlefieldSize+1 {
		t.Fatal("the Conspire copy did not create a creature token")
	}
	copyToken := g.Battlefield[len(g.Battlefield)-1]
	if !copyToken.Token || copyToken.Controller != game.Player1 ||
		copyToken.TokenDef.Name != "Copy target" ||
		copyToken.TokenDef.Power.Val.Value != 3 || copyToken.TokenDef.Toughness.Val.Value != 5 {
		t.Fatalf("Conspire copy created the wrong token: %+v", copyToken)
	}

	engine.resolveTopOfStack(g, log)
	if len(g.Battlefield) != battlefieldSize+2 {
		t.Fatal("the original spell did not create its creature token")
	}
	originalToken := g.Battlefield[len(g.Battlefield)-1]
	if !originalToken.Token || originalToken.Controller != game.Player1 ||
		originalToken.TokenDef.Name != "Original target" ||
		originalToken.TokenDef.Power.Val.Value != 2 || originalToken.TokenDef.Toughness.Val.Value != 2 ||
		!g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatalf("original spell created the wrong token or did not finish in its owner's graveyard: %+v", originalToken)
	}
}

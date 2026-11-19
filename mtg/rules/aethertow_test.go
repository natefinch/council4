package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Aethertow Oracle ID dd0cac88-0a78-408f-8eb0-080f79150180, printing revision
// c0b34ecd-6dd4-49b0-ab49-3c11f70e0c4a (MM3 205): "Put target attacking or
// blocking creature on top of its owner's library. Conspire (As you cast this
// spell, you may tap two untapped creatures you control that share a color with
// it. When you do, copy it and you may choose a new target for the copy.)"
// Scryfall rulings URI
// https://api.scryfall.com/cards/c0b34ecd-6dd4-49b0-ab49-3c11f70e0c4a/rulings
// returned an empty list when retrieved 2026-11-19.
// Official Wizards Comprehensive Rules, effective 2026-11-06, confirmed by
// https://magic.wizards.com/en/rules; TXT:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (retrieved copy .cardwork/source-review/MagicCompRules-20261106.txt).
// Read 105.2-3 for spell colors; 400.3/400.7 for owner libraries and new zone
// objects; 506.4, 508.1a/f, and 509.1a/g for combat status; 601.2b/c/e-i for
// casting, target selection, costs, and cast triggers; 603.2 for triggering;
// 608.2b-c for target legality and ordered resolution; 702.78a-b for Conspire;
// and 707.10/.10c for spell copies and optional new targets. The spell's
// attacking-or-blocking restriction is checked when targeting and on resolution;
// its resolving instruction moves the creature to its owner's library.
func TestAethertowConspirePutsCombatCreaturesOnOwnersLibraries(t *testing.T) {
	const oracleID = "dd0cac88-0a78-408f-8eb0-080f79150180"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Aethertow by Oracle ID: %v", err)
	}
	if def.Name != "Aethertow" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	spellID := addCardToHand(g, game.Player1, def)
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Attacker", 2, 2))
	blocker := addCombatPermanent(g, game.Player2, vanillaCreature("Blocker", 2, 2))
	blocker.Owner = game.Player3
	g.CardInstances[blocker.CardInstanceID].Owner = game.Player3
	firstConspireCreature := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	secondConspireCreature := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	playerOneTop := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player one top"}})
	playerTwoTop := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player two top"}})
	playerThreeTop := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Player three top"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackers := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}))
	if !engine.applyDeclareAttackers(g, game.Player1, attackers) {
		t.Fatal("could not declare Aethertow's target as an attacker")
	}
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	blockers := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, blockers) {
		t.Fatal("could not declare Aethertow's second target as a blocker")
	}
	g.Turn.PriorityPlayer = game.Player1

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == blocker.ObjectID {
					return []int{option.Index}
				}
			}
			t.Fatal("the blocking creature was not offered as a legal new target")
		}
		return chooseConspireCost(req)
	})
	log := &TurnLog{}
	cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(attacker.ObjectID)}, 0, nil)
	if !engine.applyActionWithChoices(g, game.Player1, cast, agents, log) {
		t.Fatal("could not cast Aethertow targeting the attacking creature")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 4 ||
		!g.Players[game.Player1].ManaPool.IsEmpty() || len(original.Targets) != 1 ||
		original.Targets[0].PermanentID != attacker.ObjectID {
		t.Fatalf("Aethertow did not pay its hybrid mana and Conspire costs: %+v", original)
	}
	for _, creature := range []*game.Permanent{firstConspireCreature, secondConspireCreature} {
		if permanent, ok := permanentByObjectID(g, creature.ObjectID); !ok || !permanent.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !isConspireTrigger(trigger) {
		t.Fatal("the top stack object is not Aethertow's Conspire trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.CastCopy || len(copy.Targets) != 1 ||
		copy.Targets[0].PermanentID != blocker.ObjectID {
		t.Fatalf("Aethertow's Conspire copy did not target the blocking creature: %+v", copy)
	}

	engine.resolveTopOfStackWithChoices(g, agents, log)
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("the copied Aethertow left its blocking creature on the battlefield")
	}
	if top, ok := g.Players[game.Player3].Library.Top(); !ok || top != blocker.CardInstanceID {
		t.Fatalf("blocking creature reached player %d's library top as %d, want owner %d card %d",
			game.Player3, top, game.Player3, blocker.CardInstanceID)
	}
	if top, ok := g.Players[game.Player2].Library.Top(); !ok || top != playerTwoTop {
		t.Fatal("blocking creature was put in its controller's library instead of its owner's")
	}

	engine.resolveTopOfStackWithChoices(g, agents, log)
	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("Aethertow left its attacking creature on the battlefield")
	}
	if top, ok := g.Players[game.Player1].Library.Top(); !ok || top != attacker.CardInstanceID ||
		g.Players[game.Player1].Library.Size() != 2 || g.Players[game.Player3].Library.Size() != 2 ||
		!g.Players[game.Player1].Library.Contains(playerOneTop) ||
		!g.Players[game.Player3].Library.Contains(playerThreeTop) {
		t.Fatal("Aethertow did not put its attacking creature on top of its owner's library")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("Aethertow did not finish in its owner's graveyard")
	}
}

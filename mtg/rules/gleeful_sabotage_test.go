package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle revision 7802c7f8-ca2a-4bac-a579-77a8204aa5d4,
// Oracle ID 25fe48be-95c4-4011-8c60-4628fb5ecfcb: destroy target artifact or
// enchantment; Conspire permits tapping two untapped creatures you control that
// share a color with the spell while casting, then copies the spell with an
// optional new target. The pinned printing's Scryfall rulings URL,
// https://api.scryfall.com/cards/7802c7f8-ca2a-4bac-a579-77a8204aa5d4/rulings,
// returned an empty list when retrieved 2026-11-19. Official Comprehensive
// Rules effective November 6, 2026,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 105.2-3 and 302.6 define the color and tapping constraints; 601.2a-i governs
// casting, target choice, and additional-cost payment; 603.2-3 governs the
// cast-trigger; 608.2b and 701.8 govern target legality and destruction;
// 702.78a-b defines Conspire; 707.10 and 707.10c govern the copy and optional
// retargeting. The copy uses the original spell's cast decisions and resolves
// independently against its selected legal target.
func TestGleefulSabotageConspireDestroysArtifactAndEnchantment(t *testing.T) {
	const oracleID = "25fe48be-95c4-4011-8c60-4628fb5ecfcb"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Gleeful Sabotage by Oracle ID: %v", err)
	}
	if def.Name != "Gleeful Sabotage" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	spellID := addCardToHand(g, game.Player1, def)
	artifact := addNamedTypedPermanent(g, game.Player2, "Artifact target", types.Artifact)
	enchantment := addNamedTypedPermanent(g, game.Player2, "Enchantment target", types.Enchantment)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == enchantment.ObjectID {
					return []int{option.Index}
				}
			}
			t.Fatal("the enchantment was not offered as a legal new target")
		}
		return chooseConspireCost(req)
	})
	log := &TurnLog{}
	cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	if !engine.applyActionWithChoices(g, game.Player1, cast, agents, log) {
		t.Fatal("could not cast Gleeful Sabotage")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 2 {
		t.Fatalf("expected the spell and its Conspire cost to be cast and paid: %+v", original)
	}
	for _, creature := range []*game.Permanent{first, second} {
		if p, _ := permanentByObjectID(g, creature.ObjectID); !p.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !isConspireTrigger(trigger) {
		t.Fatal("the top stack object is not Gleeful Sabotage's Conspire trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.CastCopy || len(copy.Targets) != 1 ||
		copy.Targets[0].PermanentID != enchantment.ObjectID {
		t.Fatalf("Conspire copy did not use the chosen enchantment target: %+v", copy)
	}

	engine.resolveTopOfStack(g, log)
	if !g.Players[game.Player2].Graveyard.Contains(enchantment.CardInstanceID) ||
		g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) {
		t.Fatal("the copied spell did not destroy only its enchantment target")
	}
	engine.resolveTopOfStack(g, log)
	if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
		!g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("the original spell did not destroy its artifact target and finish in its owner's graveyard")
	}
}

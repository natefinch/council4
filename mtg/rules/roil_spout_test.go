package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned BFZ #219 printing ee9fdecc-ba16-4aa2-9c20-7c2942ec143e,
// Oracle ID ebc1d343-26f1-4d85-ac3a-610da24c990e, revision ee9fdecc-ba16-4aa2-9c20-7c2942ec143e;
// corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current text verified 2026-11-19 at
// https://api.scryfall.com/cards/ee9fdecc-ba16-4aa2-9c20-7c2942ec143e:
// "{1}{W}{U} Sorcery. Put target creature on top of its owner's library.
// Awaken 4—{4}{W}{U} (If you cast this spell for {4}{W}{U}, also put four
// +1/+1 counters on target land you control and it becomes a 0/0 Elemental
// creature with haste. It's still a land.)" Rulings URI
// https://api.scryfall.com/cards/ee9fdecc-ba16-4aa2-9c20-7c2942ec143e/rulings
// returned six WotC rulings dated 2015-08-25: Awaken is optional and its target
// exists only when that cost is paid; all-illegal targets stop resolution;
// otherwise illegal targets are unaffected; the land retains its other types,
// subtypes, mana abilities, and color.
// The workflow's rules pin is effective 2026-09-25, verified in the official
// document https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf.
// Read 108.3 and 400.7 (owner/library and new zone objects), 115.3 (separate
// target instances may choose the same object), 601.2b-c/f-h (alternative costs
// and branch-dependent targets are chosen while casting), 608.2b-c (target
// legality and instructions resolve in text order), 702.113a-b (Awaken cost,
// target, counters, type, haste, and retained land type), 122.1a (+1/+1 counters
// modify power and toughness), 205.1b and 613.1d/f/g (type, ability, and P/T
// layers), 613.4b (set base P/T), 611.2a (no stated duration lasts to game end),
// and 702.10b (haste bypasses summoning-sickness restrictions). Also read these
// relevant passages in the current official edition effective 2026-11-06,
// verified at https://magic.wizards.com/en/rules and
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf.
func roilSpoutEngine(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("ebc1d343-26f1-4d85-ac3a-610da24c990e")
	if err != nil {
		t.Fatalf("lookup Roil Spout by Oracle ID: %v", err)
	}
	if def.Name != "Roil Spout" {
		t.Fatalf("Oracle ID resolved to %q, want Roil Spout", def.Name)
	}
	return NewEngineWithPrograms(nil, registry.Programs()), def
}

func TestRoilSpoutScenario(t *testing.T) {
	t.Run("Awaken moves the creature to its owner's library and animates a land", func(t *testing.T) {
		engine, def := roilSpoutEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		creature := addCombatPermanent(g, game.Player2, vanillaCreature("Roil Spout target", 2, 2))
		creature.Owner = game.Player3
		g.CardInstances[creature.CardInstanceID].Owner = game.Player3
		creatureBottom := addCardToLibraryNamed(g, game.Player2, "Roil Spout controller's library")
		addCardToLibraryNamed(g, game.Player3, "Roil Spout owner's library")
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		land.SummoningSick = true
		g.Players[game.Player1].ManaPool.Add(mana.C, 4)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{
			game.PermanentTarget(creature.ObjectID),
			game.PermanentTarget(land.ObjectID),
		}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("Awaken cast with {4}{W}{U}, a creature, and a controlled Forest is not legal")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("Awaken cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 6 ||
			len(spell.Targets) != 2 || spell.Targets[0].PermanentID != creature.ObjectID ||
			spell.Targets[1].PermanentID != land.ObjectID {
			t.Fatalf("cast spell = %+v, want Awaken alternative, six mana, creature then land targets", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
			t.Fatal("target creature remained on the battlefield")
		}
		if top, ok := g.Players[game.Player3].Library.Top(); !ok || top != creature.CardInstanceID {
			t.Fatalf("target creature was not put on owner %d's library top", game.Player3)
		}
		if top, ok := g.Players[game.Player2].Library.Top(); !ok || top != creatureBottom {
			t.Fatalf("target creature was put on controller %d's library instead of its owner's", game.Player2)
		}
		if land.Counters.Get(counter.PlusOnePlusOne) != 4 ||
			!permanentHasType(g, land, types.Land) || !permanentHasType(g, land, types.Creature) ||
			!permanentHasSubtype(g, land, types.Forest) || !permanentHasSubtype(g, land, types.Elemental) {
			t.Fatal("Awakened Forest did not gain four counters and become a land Elemental creature")
		}
		if got := effectivePower(g, land); got != 4 {
			t.Fatalf("Awakened Forest power = %d, want 4", got)
		}
		if got := effectiveToughnessMust(t, g, land); got != 4 {
			t.Fatalf("Awakened Forest toughness = %d, want 4", got)
		}
		if !hasKeyword(g, land, game.Haste) {
			t.Fatal("Awakened Forest does not have haste")
		}
	})

	t.Run("mana-cost cast needs only the creature target", func(t *testing.T) {
		engine, def := roilSpoutEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		creature := addCombatPermanent(g, game.Player2, vanillaCreature("Roil Spout target", 2, 2))
		addCardToLibraryNamed(g, game.Player2, "Roil Spout library below target")
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("mana-cost cast should be legal with only its creature target")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("mana-cost cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 0 || spell.ManaSpentToCast != 3 || len(spell.Targets) != 1 {
			t.Fatalf("cast spell = %+v, want mana cost, three mana, one creature target", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if top, ok := g.Players[game.Player2].Library.Top(); !ok || top != creature.CardInstanceID {
			t.Fatal("target creature was not put on top of its owner's library")
		}
		if permanentHasType(g, land, types.Creature) || land.Counters.Get(counter.PlusOnePlusOne) != 0 {
			t.Fatal("mana-cost cast applied the Awaken effect")
		}
	})
}

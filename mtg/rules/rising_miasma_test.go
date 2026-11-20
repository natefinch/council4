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

// Source review: pinned BFZ #122 printing 4f9a8e87-3b8b-4dbf-9c1e-0a3290a33a0b,
// Oracle ID 83fb7dfb-64a4-4320-b544-830c29f24df9, revision
// 4f9a8e87-3b8b-4dbf-9c1e-0a3290a33a0b; corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current text verified 2026-11-20 at
// https://api.scryfall.com/cards/4f9a8e87-3b8b-4dbf-9c1e-0a3290a33a0b:
// "{3}{B} Sorcery. All creatures get -2/-2 until end of turn. Awaken 3—
// {5}{B}{B} (If you cast this spell for {5}{B}{B}, also put three +1/+1
// counters on target land you control and it becomes a 0/0 Elemental creature
// with haste. It's still a land.)" Its rulings URI
// https://api.scryfall.com/cards/4f9a8e87-3b8b-4dbf-9c1e-0a3290a33a0b/rulings
// returned seven WotC rulings dated 2015-08-25. They confirm the global -2/-2
// happens before the Awakened land becomes a creature (so a noncreature land
// isn't shrunk), the Awaken cost and target are optional/branch-dependent, and
// the land retains its other types, subtypes, supertypes, mana abilities and
// color. Raw printing and rulings responses are retained under
// .cardwork/rising-miasma/.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 601.2b-c/f-h (alternative cost and conditional target), 608.2c
// (instructions in order), 611.2a,c (duration and the affected set's
// resolution-time snapshot), 122.1a and 613.4c (counters and P/T modification),
// 205.1b and 613.1d/g (retained land type, type and P/T layers), 702.10b
// (haste), and 702.113a-b (Awaken cost, target, counters and animation). These
// passages require choosing Awaken at casting, shrinking the creatures present
// when the first effect resolves, then adding three counters and animating the
// target land; an ordinary noncreature land is not in the shrunken set.
func risingMiasmaEngine(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("83fb7dfb-64a4-4320-b544-830c29f24df9")
	if err != nil {
		t.Fatalf("lookup Rising Miasma by Oracle ID: %v", err)
	}
	if def.Name != "Rising Miasma" {
		t.Fatalf("Oracle ID resolved to %q, want Rising Miasma", def.Name)
	}
	return NewEngineWithPrograms(nil, registry.Programs()), def
}

func TestRisingMiasmaScenario(t *testing.T) {
	t.Run("Awaken shrinks current creatures before animating the target land", func(t *testing.T) {
		engine, def := risingMiasmaEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		ownCreature := addCombatPermanent(g, game.Player1, vanillaCreature("Rising Miasma ally", 4, 4))
		opponentCreature := addCombatPermanent(g, game.Player2, vanillaCreature("Rising Miasma opponent", 4, 4))
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Players[game.Player1].ManaPool.Add(mana.C, 5)
		g.Players[game.Player1].ManaPool.Add(mana.B, 2)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("Awaken cast with {5}{B}{B} and a controlled Forest is not legal")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("Awaken cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 7 ||
			len(spell.Targets) != 1 || spell.Targets[0].PermanentID != land.ObjectID {
			t.Fatalf("cast spell = %+v, want Awaken alternative, seven mana, and Forest target", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		for i, creature := range []*game.Permanent{ownCreature, opponentCreature} {
			if got := effectivePower(g, creature); got != 2 {
				t.Errorf("creature %d power = %d, want 2", i, got)
			}
			if got := effectiveToughnessMust(t, g, creature); got != 2 {
				t.Errorf("creature %d toughness = %d, want 2", i, got)
			}
		}
		if land.Counters.Get(counter.PlusOnePlusOne) != 3 ||
			!permanentHasType(g, land, types.Land) || !permanentHasType(g, land, types.Creature) ||
			!permanentHasSubtype(g, land, types.Forest) || !permanentHasSubtype(g, land, types.Elemental) {
			t.Fatal("Awakened Forest did not gain three counters and become a land Elemental creature")
		}
		if got := effectivePower(g, land); got != 3 {
			t.Fatalf("Awakened Forest power = %d, want 3 (it must not receive the earlier -2/-2)", got)
		}
		if got := effectiveToughnessMust(t, g, land); got != 3 {
			t.Fatalf("Awakened Forest toughness = %d, want 3 (it must not receive the earlier -2/-2)", got)
		}
		if !hasKeyword(g, land, game.Haste) {
			t.Fatal("Awakened Forest does not have haste")
		}
	})

	t.Run("mana-cost cast shrinks creatures without an Awaken target", func(t *testing.T) {
		engine, def := risingMiasmaEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		ownCreature := addCombatPermanent(g, game.Player1, vanillaCreature("Rising Miasma ally", 4, 4))
		opponentCreature := addCombatPermanent(g, game.Player2, vanillaCreature("Rising Miasma opponent", 4, 4))
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Players[game.Player1].ManaPool.Add(mana.C, 3)
		g.Players[game.Player1].ManaPool.Add(mana.B, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, nil, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("mana-cost cast should be legal without the Awaken target")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("mana-cost cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 0 || spell.ManaSpentToCast != 4 || len(spell.Targets) != 0 {
			t.Fatalf("cast spell = %+v, want mana cost, four mana spent, and no targets", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		for i, creature := range []*game.Permanent{ownCreature, opponentCreature} {
			if got := effectivePower(g, creature); got != 2 {
				t.Errorf("creature %d power = %d, want 2", i, got)
			}
			if got := effectiveToughnessMust(t, g, creature); got != 2 {
				t.Errorf("creature %d toughness = %d, want 2", i, got)
			}
		}
		if permanentHasType(g, land, types.Creature) || land.Counters.Get(counter.PlusOnePlusOne) != 0 {
			t.Fatal("mana-cost cast applied the Awaken effect")
		}
	})

	t.Run("Awakened land creature was already included in the shrink", func(t *testing.T) {
		engine, def := risingMiasmaEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		landCreatureDef := vanillaCreature("Rising Miasma land creature", 4, 4)
		landCreatureDef.CardFace.Types = append(landCreatureDef.CardFace.Types, types.Land)
		landCreatureDef.CardFace.Subtypes = []types.Sub{types.Forest}
		land := addCombatPermanent(g, game.Player1, landCreatureDef)
		g.Players[game.Player1].ManaPool.Add(mana.C, 5)
		g.Players[game.Player1].ManaPool.Add(mana.B, 2)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("Awaken cast with a controlled Forest creature is not legal")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("Awaken cast on the land creature failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if land.Counters.Get(counter.PlusOnePlusOne) != 3 ||
			!permanentHasType(g, land, types.Land) || !permanentHasType(g, land, types.Creature) ||
			!permanentHasSubtype(g, land, types.Forest) || !permanentHasSubtype(g, land, types.Elemental) {
			t.Fatal("Awakened land creature did not retain its land type and gain counters and Elemental type")
		}
		if got := effectivePower(g, land); got != 1 {
			t.Fatalf("Awakened land creature power = %d, want 1 after the earlier -2/-2 and three counters", got)
		}
		if got := effectiveToughnessMust(t, g, land); got != 1 {
			t.Fatalf("Awakened land creature toughness = %d, want 1 after the earlier -2/-2 and three counters", got)
		}
		if !hasKeyword(g, land, game.Haste) {
			t.Fatal("Awakened land creature does not have haste")
		}
	})
}

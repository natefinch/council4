package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Source review: Mizzium Skin (Oracle ID 02d998a6-0afd-4be1-a209-05618fcc43ea, printing
// d9859344-4efc-4b87-a3fb-147e496cee68) says a creature you control gets +0/+1
// and hexproof until end of turn, with overload {1}{U}. Its Scryfall rulings URI
// https://api.scryfall.com/cards/d9859344-4efc-4b87-a3fb-147e496cee68/rulings
// returned four WotC rulings dated 2024-01-12: normal casting has one target;
// overload has none and may affect hexproof/protected permanents; total cost
// starts with the normal or alternative cost, then increases and reductions,
// while mana value stays unchanged; and a free-cast effect cannot choose
// overload.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// read 601.2b/c/f, 608.2b, 611.2a/c, 612.1, 613.1f, 613.4c, 702.11a/b and
// 702.96a/b. These require the overload alternative cost to replace the normal
// cost and text, the overloaded spell to have no targets, the creature set to
// be locked when the effect resolves, +0/+1 and hexproof to apply, and both
// effects to expire at end of turn. Rulings retrieved 2026-11-11.
//
// Source review: Bounty of Might (Oracle ID 6f067bfd-1c59-463f-99e1-61161b43b405, printing
// e9a950d8-e6e2-42d6-83d7-a60c15d2035d) has three separate target-creature
// paragraphs, each granting +3/+3 until end of turn. Its Scryfall rulings URI
// https://api.scryfall.com/cards/e9a950d8-e6e2-42d6-83d7-a60c15d2035d/rulings
// returned one WotC ruling dated 2018-10-05: the same creature may be chosen
// for multiple target instances, including all three, to receive +9/+9.
// The same official rules edition: read 601.2c, 608.2b, 611.2a and 613.4c.
// These require separate target choices (which may repeat across instances),
// each legal target to receive its corresponding +3/+3, and the bonuses to
// expire at end of turn. Rulings retrieved 2026-11-11.
//
// Source review: Long Goodbye (Oracle ID 9d63ba8a-a5cd-4fdb-bba3-95612be0077c, printing
// c3896705-bbd2-4ffb-a590-ee78e0eabdc5) can't be countered, including by ward,
// and destroys target creature or planeswalker with mana value 3 or less. Its
// Its Scryfall rulings URI
// https://api.scryfall.com/cards/c3896705-bbd2-4ffb-a590-ee78e0eabdc5/rulings
// returned one WotC ruling dated 2024-02-02: the target's ward cost may be
// declined, but Long Goodbye still won't be countered. The same official rules
// edition: read 101.2, 202.3, 601.2c, 608.2b, 701.6a, 701.8a and 702.21a. These
// require a "can't" rule to take precedence, target mana value to be calculated
// from its mana cost, the legal target to be checked on resolution, destruction
// to move it to its owner's graveyard, and ward to counter the spell only when
// it is counterable. Rulings retrieved 2026-11-11.

const (
	mizziumSkinOracleID = "02d998a6-0afd-4be1-a209-05618fcc43ea"
	bountyOfMightID     = "6f067bfd-1c59-463f-99e1-61161b43b405"
	longGoodbyeOracleID = "9d63ba8a-a5cd-4fdb-bba3-95612be0077c"
)

func TestMizziumSkinScenario(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(mizziumSkinOracleID)
	if err != nil {
		t.Fatalf("exact Mizzium Skin registry lookup failed: %v", err)
	}
	if def.Name != "Mizzium Skin" {
		t.Fatalf("exact Mizzium Skin registry lookup returned %q", def.Name)
	}

	t.Run("normal target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		target := addCombatPermanent(g, game.Player1, vanillaCreature("Mizzium Skin target", 2, 2))
		opponent := addCombatPermanent(g, game.Player2, vanillaCreature("Mizzium Skin opponent", 2, 2))
		spell := addCardToHand(g, game.Player1, def)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		setMainPhasePriority(g, game.Player1)

		if !engine.applyAction(g, game.Player1, action.CastSpell(
			spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
		)) {
			t.Fatal("Mizzium Skin did not cast targeting a creature its controller controls")
		}
		if g.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("Mizzium Skin did not pay {U}")
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if got, ok := effectiveToughness(g, target); !ok || got != 3 || !hasKeyword(g, target, game.Hexproof) {
			t.Fatalf("target toughness/hexproof = %d, %v; want 3 and hexproof", got, hasKeyword(g, target, game.Hexproof))
		}
		if got := effectivePower(g, target); got != 2 {
			t.Fatalf("target power = %d, want unchanged at 2", got)
		}
		if got, ok := effectiveToughness(g, opponent); !ok || got != 2 || hasKeyword(g, opponent, game.Hexproof) {
			t.Fatalf("opponent toughness/hexproof = %d, %v; want 2 and no hexproof", got, hasKeyword(g, opponent, game.Hexproof))
		}
	})

	t.Run("overload", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		first := addCombatPermanent(g, game.Player1, vanillaCreature("Mizzium Skin first", 2, 2))
		second := addCombatPermanent(g, game.Player1, vanillaCreature("Mizzium Skin second", 3, 3))
		opponent := addCombatPermanent(g, game.Player2, vanillaCreature("Mizzium Skin opponent", 2, 2))
		spell := addCardToHand(g, game.Player1, def)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		setMainPhasePriority(g, game.Player1)

		if !engine.applyAction(g, game.Player1, action.CastOverloadedSpellFaceFromZone(
			spell, zone.Hand, game.FaceFront, nil,
		)) {
			t.Fatal("Mizzium Skin did not cast for its {1}{U} overload cost")
		}
		if g.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("Mizzium Skin overload did not pay {1}{U}")
		}
		stacked, ok := g.Stack.Peek()
		if !ok || !stacked.Overloaded || len(stacked.Targets) != 0 {
			t.Fatalf("overloaded stack object = %+v, want overloaded with no targets", stacked)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		for _, creature := range []*game.Permanent{first, second} {
			got, valid := effectiveToughness(g, creature)
			want := 4
			if creature == first {
				want = 3
			}
			if !valid || got != want || !hasKeyword(g, creature, game.Hexproof) {
				t.Fatalf("overloaded creature toughness/hexproof = %d, %v; want +1 and hexproof", got, hasKeyword(g, creature, game.Hexproof))
			}
			wantPower := 3
			if creature == first {
				wantPower = 2
			}
			if power := effectivePower(g, creature); power != wantPower {
				t.Fatalf("overloaded creature power = %d, want unchanged at %d", power, wantPower)
			}
		}
		if got, ok := effectiveToughness(g, opponent); !ok || got != 2 || hasKeyword(g, opponent, game.Hexproof) {
			t.Fatalf("opponent toughness/hexproof = %d, %v; want 2 and no hexproof", got, hasKeyword(g, opponent, game.Hexproof))
		}
	})
}

func TestBountyOfMightScenario(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(bountyOfMightID)
	if err != nil {
		t.Fatalf("exact Bounty of Might registry lookup failed: %v", err)
	}
	if def.Name != "Bounty of Might" {
		t.Fatalf("exact Bounty of Might registry lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player2, vanillaCreature("Bounty first", 2, 2))
	second := addCombatPermanent(g, game.Player1, vanillaCreature("Bounty second", 3, 3))
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
		game.PermanentTarget(first.ObjectID),
	}, 0, nil)) {
		t.Fatal("Bounty of Might did not cast with its three targets, including a repeated creature")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Bounty of Might did not pay {4}{G}{G}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if power := effectivePower(g, first); power != 8 {
		t.Fatalf("first target power = %d, want 8 after two +3/+3 effects", power)
	}
	if toughness, ok := effectiveToughness(g, first); !ok || toughness != 8 {
		t.Fatalf("first target toughness = %d, %v; want 8", toughness, ok)
	}
	if power := effectivePower(g, second); power != 6 {
		t.Fatalf("second target power = %d, want 6 after one +3/+3 effect", power)
	}
	if toughness, ok := effectiveToughness(g, second); !ok || toughness != 6 {
		t.Fatalf("second target toughness = %d, %v; want 6", toughness, ok)
	}
}

func TestLongGoodbyeScenario(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(longGoodbyeOracleID)
	if err != nil {
		t.Fatalf("exact Long Goodbye registry lookup failed: %v", err)
	}
	if def.Name != "Long Goodbye" {
		t.Fatalf("exact Long Goodbye registry lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Long Goodbye warded creature",
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.G}),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.WardStaticAbility(cost.Mana{cost.O(20)}),
		},
	}}
	target := addCombatPermanent(g, game.Player2, targetDef)
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Long Goodbye did not cast targeting a mana-value-3 creature with ward")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Long Goodbye did not pay {1}{B}")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("target's ward ability did not trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 1 || g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("ward countered Long Goodbye despite its uncounterable text")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if _, stillOnBattlefield := permanentByObjectID(g, target.ObjectID); stillOnBattlefield {
		t.Fatal("Long Goodbye did not destroy its legal target")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("destroyed target did not move to its owner's graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Long Goodbye did not move to its owner's graveyard")
	}

	planeswalkerGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	planeswalkerEngine := NewEngine(nil)
	planeswalkerDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Long Goodbye planeswalker",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
		Types:    []types.Card{types.Planeswalker},
		Loyalty:  opt.Val(3),
	}}
	planeswalker := addCombatPermanent(planeswalkerGame, game.Player2, planeswalkerDef)
	planeswalker.Counters.Add(counter.Loyalty, 3)
	planeswalkerSpell := addCardToHand(planeswalkerGame, game.Player1, def)
	planeswalkerGame.Players[game.Player1].ManaPool.Add(mana.B, 1)
	planeswalkerGame.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(planeswalkerGame, game.Player1)
	if !planeswalkerEngine.applyAction(planeswalkerGame, game.Player1, action.CastSpell(
		planeswalkerSpell, []game.Target{game.PermanentTarget(planeswalker.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Long Goodbye did not cast targeting a mana-value-3 planeswalker")
	}
	planeswalkerEngine.resolveTopOfStack(planeswalkerGame, &TurnLog{})
	if _, stillOnBattlefield := permanentByObjectID(planeswalkerGame, planeswalker.ObjectID); stillOnBattlefield {
		t.Fatal("Long Goodbye did not destroy its planeswalker target")
	}
}

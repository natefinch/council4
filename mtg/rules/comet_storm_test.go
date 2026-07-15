package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// cometStormDef builds the executable shape cardgen lowers Comet Storm to: an
// {X}{R} instant with Multikicker {1} whose spell chooses 1 + kicker any-targets
// (CountEqualsKickerPlusOne) and deals X to each of them (EachTarget).
func cometStormDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Comet Storm",
			ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
			Types:    []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{
					game.KickerKeyword{Cost: cost.Mana{cost.O(1)}, Multi: true},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets:               1,
					MaxTargets:               21,
					Constraint:               "any target",
					Allow:                    game.TargetAllowPermanent | game.TargetAllowPlayer,
					CountEqualsKickerPlusOne: true,
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:     game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
					Recipient:  game.AnyTargetDamageRecipient(0),
					EachTarget: true,
				}}},
			}.Ability()),
		},
	}
}

// toughCreature adds a high-toughness creature so it survives the Comet Storm
// damage in a cast-and-resolve test (state-based actions would otherwise remove
// a lethally damaged target before its marked damage can be asserted).
func toughCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Tough Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 20}),
	}})
}

// TestCometStormCastTargetCountMatchesKicker proves the cast-time
// CountEqualsKickerPlusOne validation: a legal cast must choose exactly one
// target plus one per kicker payment, and every other target count is rejected.
func TestCometStormCastTargetCountMatchesKicker(t *testing.T) {
	// setup builds a fresh game whose Player1 holds Comet Storm with ample mana
	// and Player2 controls one creature, returning the cast targets by role.
	setup := func() (*Engine, *game.Game, game.Target, game.Target, game.Target, func([]game.Target, int) action.Action, func([]game.Target) action.Action) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 20)
		spellID := addCardToHand(g, game.Player1, cometStormDef())
		creature := toughCreature(g, game.Player2)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		unkicked := func(targets []game.Target) action.Action {
			return action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, targets, 1, nil)
		}
		kicked := func(targets []game.Target, kicks int) action.Action {
			return action.CastMultikickedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, targets, 1, nil, kicks)
		}
		return engine, g, game.PlayerTarget(game.Player1), game.PlayerTarget(game.Player2),
			game.PermanentTarget(creature.ObjectID), kicked, unkicked
	}

	cases := []struct {
		name string
		want bool
		make func(p1, p2, perm game.Target, kicked func([]game.Target, int) action.Action, unkicked func([]game.Target) action.Action) action.Action
	}{
		{"unkicked one target", true, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return u([]game.Target{p2})
		}},
		{"unkicked two targets rejected", false, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return u([]game.Target{p1, p2})
		}},
		{"one kick two targets", true, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return k([]game.Target{p2, perm}, 1)
		}},
		{"one kick one target rejected", false, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return k([]game.Target{p2}, 1)
		}},
		{"two kicks three targets", true, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return k([]game.Target{p1, p2, perm}, 2)
		}},
		{"two kicks two targets rejected", false, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return k([]game.Target{p2, perm}, 2)
		}},
		{"one kick duplicate target rejected", false, func(p1, p2, perm game.Target, k func([]game.Target, int) action.Action, u func([]game.Target) action.Action) action.Action {
			return k([]game.Target{p2, p2}, 1)
		}},
	}
	for _, tc := range cases {
		engine, g, p1, p2, perm, kicked, unkicked := setup()
		act := tc.make(p1, p2, perm, kicked, unkicked)
		if got := engine.applyAction(g, game.Player1, act); got != tc.want {
			t.Fatalf("%s: applyAction legal = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCometStormLegalActionsIncludeMultipleKicks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 20)
	spellID := addCardToHand(g, game.Player1, cometStormDef())
	creature := toughCreature(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	want := action.CastMultikickedSpellFaceFromZone(
		spellID,
		zone.Hand,
		game.FaceFront,
		[]game.Target{
			game.PlayerTarget(game.Player1),
			game.PlayerTarget(game.Player2),
			game.PermanentTarget(creature.ObjectID),
		},
		1,
		nil,
		2,
	)
	if !containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatal("legal actions do not include Comet Storm kicked twice with three targets")
	}
}

// TestCometStormUnkickedDealsXToSingleTarget casts an unkicked Comet Storm with
// X = 3 at one target and proves it takes the full X.
func TestCometStormUnkickedDealsXToSingleTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	spellID := addCardToHand(g, game.Player1, cometStormDef())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	startingLife := g.Players[game.Player2].Life

	act := action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront,
		[]game.Target{game.PlayerTarget(game.Player2)}, 3, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(unkicked cast) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != startingLife-3 {
		t.Fatalf("player life = %d, want %d", got, startingLife-3)
	}
}

// TestCometStormKickedDealsXToEveryTarget casts a Comet Storm kicked twice with
// X = 2 at a mix of two players and one creature, and proves every one of the
// 1 + 2 targets independently takes the full X.
func TestCometStormKickedDealsXToEveryTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 20)
	spellID := addCardToHand(g, game.Player1, cometStormDef())
	creature := toughCreature(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	p1Life := g.Players[game.Player1].Life
	p2Life := g.Players[game.Player2].Life

	act := action.CastMultikickedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront,
		[]game.Target{
			game.PlayerTarget(game.Player1),
			game.PlayerTarget(game.Player2),
			game.PermanentTarget(creature.ObjectID),
		}, 2, nil, 2)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(twice-kicked cast) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != p1Life-2 {
		t.Fatalf("player1 life = %d, want %d", got, p1Life-2)
	}
	if got := g.Players[game.Player2].Life; got != p2Life-2 {
		t.Fatalf("player2 life = %d, want %d", got, p2Life-2)
	}
	got, ok := permanentByObjectID(g, creature.ObjectID)
	if !ok {
		t.Fatal("creature target left the battlefield")
	}
	if got.MarkedDamage != 2 {
		t.Fatalf("creature marked damage = %d, want 2", got.MarkedDamage)
	}
}

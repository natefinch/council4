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
	"github.com/natefinch/council4/opt"
)

// Source review: Petrified Wood-Kin, Guildpact #91; Oracle ID
// 3bc9649c-4350-4c6a-8441-dc76a5f833af; pinned printing revision
// a56fbf32-f08f-4882-b8d4-41ccb60cba63 from the pinned corpus
// (SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4).
// Oracle text, three paragraphs:
// “This spell can't be countered.”
// “Bloodthirst X (This creature enters with X +1/+1 counters on it, where X is
// the damage dealt to your opponents this turn.)”
// “Protection from instants”
// Scryfall rulings retrieved 2026-11-22 from
// https://api.scryfall.com/cards/a56fbf32-f08f-4882-b8d4-41ccb60cba63/rulings:
// two WotC rulings dated 2006-02-01. They state that protection from instants
// applies only on the battlefield; instant spells cannot target it there,
// damage from instant sources is prevented, and instant abilities cannot
// target it. In other zones, including the stack, it may be targeted.
// Read the official Wizards Comprehensive Rules effective 2026-11-06 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 601.2a-i (casting and paying costs), 113.6g (the can't-be-countered ability
// functions on the stack), 120.1-3 (damage and its results), 614.12 (self
// replacement for battlefield entry), 702.16a-f (protection from card types,
// targeting and damage prevention), and 702.54a-b (Bloodthirst X counts total
// damage dealt to the entering permanent's controller's opponents this turn).
func TestPetrifiedWoodKinCastsUncounterablyWithBloodthirstAndProtectionFromInstants(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("3bc9649c-4350-4c6a-8441-dc76a5f833af")
	if err != nil {
		t.Fatalf("exact Petrified Wood-Kin lookup failed: %v", err)
	}
	if def.Name != "Petrified Wood-Kin" {
		t.Fatalf("exact lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := &TurnLog{}

	damageSourceTwo := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	markPlayerCombatDamage(g, damageSourceTwo, game.Player2, 2, log)
	damageSourceThree := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	markPlayerCombatDamage(g, damageSourceThree, game.Player3, 3, log)

	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Petrified Wood-Kin did not cast for {6}{G}")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.SourceID != cardID {
		t.Fatalf("top of stack = %+v, want Petrified Wood-Kin", spell)
	}
	if spell.ManaSpentToCast != 7 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("cast did not pay {6}{G}: stack=%+v, remaining mana=%d",
			spell, g.Players[game.Player1].ManaPool.Total())
	}

	counterProbe := &game.CardDef{CardFace: game.CardFace{
		Name:     "Counter Probe",
		ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Text: "Counter target spell.",
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					StackObjectKinds: []game.StackObjectKind{game.StackSpell},
				},
			}},
			Sequence: []game.Instruction{{
				Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
			}},
		}.Ability()),
	}}
	counterID := addCardToHand(g, game.Player2, counterProbe)
	g.Players[game.Player2].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player2
	counterAction := action.CastSpell(counterID, []game.Target{game.StackObjectTarget(spell.ID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player2), counterAction) {
		t.Fatal("instant counter spell could not target Petrified Wood-Kin on the stack")
	}
	if !engine.applyAction(g, game.Player2, counterAction) {
		t.Fatal("counter probe failed to cast")
	}
	engine.resolveTopOfStack(g, log)
	if _, ok := stackObjectByID(g, spell.ID); !ok {
		t.Fatal("Petrified Wood-Kin was countered")
	}

	engine.resolveTopOfStack(g, log)
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatal("Petrified Wood-Kin did not enter the battlefield")
	}
	if got, want := permanent.Counters.Get(counter.PlusOnePlusOne), 5; got != want {
		t.Fatalf("Bloodthirst X counters = %d, want %d for 5 damage dealt to opponents this turn", got, want)
	}

	ordinaryCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	instantID := addCardToHand(g, game.Player2, targetCreatureInstant())
	setMainPhasePriority(g, game.Player2)
	if !engine.canCastSpell(g, game.Player2, instantID,
		[]game.Target{game.PermanentTarget(ordinaryCreature.ObjectID)}, 0, nil) {
		t.Fatal("instant probe could not target an ordinary creature")
	}
	if engine.canCastSpell(g, game.Player2, instantID,
		[]game.Target{game.PermanentTarget(permanent.ObjectID)}, 0, nil) {
		t.Fatal("instant probe could target Petrified Wood-Kin with protection from instants")
	}
}

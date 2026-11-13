package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3829 (Selection.ToughnessLessThanSource) is proved
// with a test-local shape of the blocked card Thunderkin Awakener (oracle
// c0171343-9c0f-4001-8f4e-a5af9256c4d2, printing
// 289b649b-c93a-44ad-bdd6-866f8b2f1bc2, {1}{R} Creature — Elemental Shaman 1/2)
// Oracle "Haste / Whenever this creature attacks, choose target Elemental
// creature card in your graveyard with toughness less than this creature's
// toughness. Return that card to the battlefield tapped and attacking. Sacrifice
// it at the beginning of the next end step." Its five 2019-07-12 Scryfall
// rulings: the target must have lesser toughness both as it attacks and as the
// trigger resolves; if Thunderkin leaves before resolution its toughness as it
// last existed on the battlefield decides legality; the controller chooses what
// the returned creature attacks; it never "attacked"; a returned creature that
// already left is not sacrificed. CR 2026-11-06 (MagicCompRules 20261106.txt):
// 113.7a, 508.4, 603.3d, 608.2b (target recheck uses the departed source's last
// known information), 608.2h. Its "tapped and attacking" entry is
// PutOnBattlefield.EntryAttacking (#3835), proved in
// put_on_battlefield_attacking_test.go.

func thunderkinShapeDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Thunderkin Shape",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental, types.Shaman},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{thunderkinShapeTargetSpec()},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:         game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							EntryTapped:    true,
							EntryAttacking: true,
							PublishLinked:  game.LinkedKey("thunderkin-returned"),
						},
						LocalProducts: game.LocalProducts{Links: []game.LinkedKey{"thunderkin-returned"}},
					},
					{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
						Timing:         game.DelayedAtBeginningOfNextEndStep,
						CapturedObject: opt.Val(game.LinkedObjectReference("thunderkin-returned")),
						Content: game.Mode{Sequence: []game.Instruction{{
							Primitive: game.Sacrifice{Object: game.CapturedObjectReference()},
						}}}.Ability(),
					}}},
				},
			}.Ability(),
		}},
	}}
}

func thunderkinShapeTargetSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			RequiredTypes:           []types.Card{types.Creature},
			SubtypesAny:             []types.Sub{types.Elemental},
			Controller:              game.ControllerYou,
			ToughnessLessThanSource: true,
		}),
	}
}

func toughnessTestCreature(name string, sub types.Sub, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{sub},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

func thunderkinCandidateCards(g *game.Game, source *game.Permanent) []id.ID {
	spec := thunderkinShapeTargetSpec()
	var cards []id.ID
	for _, target := range targetCandidatesForSpec(g, game.Player1, nil, source.ObjectID, game.Event{}, &spec) {
		cards = append(cards, target.CardID)
	}
	return cards
}

// declareThunderkinAttack declares source attacking Player2 and puts its attack
// trigger on the stack, returning the trigger.
func declareThunderkinAttack(t *testing.T, g *game.Game, engine *Engine, source *game.Permanent) *game.StackObject {
	t.Helper()
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}) {
		t.Fatal("Thunderkin shape could not attack")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attack created no trigger with a legal target")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want Thunderkin shape trigger", trigger)
	}
	return trigger
}

func requireCardOnlyTarget(t *testing.T, trigger *game.StackObject, cardID id.ID) {
	t.Helper()
	if len(trigger.Targets) != 1 || trigger.Targets[0].Kind != game.TargetCard || trigger.Targets[0].CardID != cardID {
		t.Fatalf("trigger targets = %+v, want card %d", trigger.Targets, cardID)
	}
}

// A 1/2 source may target only an Elemental creature card its controller owns
// with printed toughness less than 2; the returned card is sacrificed at the
// beginning of the next end step.
func TestThunderkinShapeTargetsLesserToughnessElementalAndSacrificesItAtEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	lesser := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Lesser Elemental", types.Elemental, 3, 1))
	addCardToGraveyard(g, game.Player1, toughnessTestCreature("Equal Elemental", types.Elemental, 0, 2))
	addCardToGraveyard(g, game.Player1, toughnessTestCreature("Lesser Human", types.Human, 1, 1))
	addCardToGraveyard(g, game.Player2, toughnessTestCreature("Opposing Elemental", types.Elemental, 1, 1))

	if got := thunderkinCandidateCards(g, source); len(got) != 1 || got[0] != lesser {
		t.Fatalf("target candidates = %v, want only the lesser-toughness Elemental %d", got, lesser)
	}
	trigger := declareThunderkinAttack(t, g, engine, source)
	requireCardOnlyTarget(t, trigger, lesser)
	engine.resolveTopOfStack(g, &TurnLog{})

	returned, ok := reanimatedPermanent(g, lesser)
	if !ok || returned.Controller != game.Player1 || !returned.Tapped {
		t.Fatalf("returned permanent = %+v (present %v), want tapped under Player1", returned, ok)
	}
	if attack, attacking := attackerDeclarationFor(g, returned.ObjectID); !attacking || attack.Target.Player != game.Player2 {
		t.Fatalf("returned attack = %+v (attacking %v), want attacking Player2", attack, attacking)
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, returned.ObjectID); ok {
		t.Fatal("returned card was not sacrificed at the beginning of the next end step")
	}
	if z, ok := cardZone(g, lesser); !ok || z != zone.Graveyard {
		t.Fatalf("sacrificed card zone = %v (%v), want graveyard", z, ok)
	}
}

// The source bound reads the source's current toughness while it is on the
// battlefield, so counters widen or narrow the legal targets.
func TestToughnessLessThanSourceReadsCurrentSourceToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	lesser := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Lesser Elemental", types.Elemental, 1, 1))
	equal := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Equal Elemental", types.Elemental, 1, 2))

	source.Counters.Add(counter.PlusOnePlusOne, 1)
	if got := thunderkinCandidateCards(g, source); len(got) != 2 ||
		!containsID(got, lesser) || !containsID(got, equal) {
		t.Fatalf("2/3 source candidates = %v, want both %d and %d", got, lesser, equal)
	}
	source.Counters.Remove(counter.PlusOnePlusOne, 1)
	source.Counters.Add(counter.MinusOneMinusOne, 1)
	if got := thunderkinCandidateCards(g, source); len(got) != 0 {
		t.Fatalf("0/1 source candidates = %v, want none", got)
	}
}

// Ruling 2019-07-12 / CR 608.2b, 608.2h: when the source leaves before its
// trigger resolves, target legality uses its last known battlefield toughness
// (3 with a +1/+1 counter), not the printed toughness of the card it became.
func TestToughnessLessThanSourceUsesLastKnownToughnessAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	source.Counters.Add(counter.PlusOnePlusOne, 1)
	target := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Equal Printed Elemental", types.Elemental, 2, 2))

	trigger := declareThunderkinAttack(t, g, engine, source)
	requireCardOnlyTarget(t, trigger, target)
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("could not remove the source before resolution")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned, ok := reanimatedPermanent(g, target)
	if !ok || !returned.Tapped {
		t.Fatalf("target was not returned using the departed source's last known toughness: %+v (%v)", returned, ok)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %+v, want the next-end-step sacrifice", g.DelayedTriggers)
	}
}

// Ruling 2019-07-12 / CR 608.2b: the target must still have lesser toughness as
// the trigger resolves, so shrinking the source makes it illegal.
func TestToughnessLessThanSourceRechecksTargetAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	target := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Lesser Elemental", types.Elemental, 1, 1))

	trigger := declareThunderkinAttack(t, g, engine, source)
	requireCardOnlyTarget(t, trigger, target)
	source.Counters.Add(counter.MinusOneMinusOne, 1)
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := reanimatedPermanent(g, target); ok {
		t.Fatal("target returned although its toughness was no longer less than the source's")
	}
	if z, ok := cardZone(g, target); !ok || z != zone.Graveyard {
		t.Fatalf("illegal target zone = %v (%v), want graveyard", z, ok)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %+v, want none after the target became illegal", g.DelayedTriggers)
	}
}

// A permanent subject is judged by its current toughness, a Selection carrying
// only this bound still reaches the matcher, and a source never seen on the
// battlefield yields no bound.
func TestToughnessLessThanSourcePermanentSubjectAndUnknownSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, toughnessTestCreature("Source", types.Elemental, 1, 3))
	creature := addCombatPermanent(g, game.Player2, toughnessTestCreature("Other", types.Human, 1, 2))
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			ToughnessLessThanSource: true,
		}),
	}
	target := game.PermanentTarget(creature.ObjectID)
	if !targetMatchesSpec(g, game.Player1, source.ObjectID, game.Event{}, &spec, target) {
		t.Fatal("2-toughness permanent did not match a 3-toughness source")
	}
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	if targetMatchesSpec(g, game.Player1, source.ObjectID, game.Event{}, &spec, target) {
		t.Fatal("permanent with current toughness 3 matched a 3-toughness source")
	}
	creature.Counters.Remove(counter.PlusOnePlusOne, 1)
	if targetMatchesSpec(g, game.Player1, g.IDGen.Next(), game.Event{}, &spec, target) {
		t.Fatal("unknown source produced a toughness bound")
	}
}

func containsID(ids []id.ID, want id.ID) bool {
	for _, got := range ids {
		if got == want {
			return true
		}
	}
	return false
}

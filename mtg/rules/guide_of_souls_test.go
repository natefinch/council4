package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// guidePaymentAgent answers the optional energy payment of Guide of Souls'
// attack trigger with accept or decline, and always picks the first offered
// target for the reflexive "when you do" trigger (the lone attacking creature).
type guidePaymentAgent struct {
	pay bool
}

func (guidePaymentAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a guidePaymentAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		if a.pay {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceTarget:
		return []int{0}
	default:
		return request.DefaultSelection
	}
}

// guideOfSoulsAttackAbility mirrors the lowered second triggered ability of
// Guide of Souls: on attack, you may pay {E}{E}{E}; when you do, a reflexive
// trigger puts two +1/+1 counters and a flying counter on target attacking
// creature and permanently makes it an Angel. The reflexive trigger's target is
// chosen only when it goes on the stack, after the payment resolves, and it is
// gated on the payment's published result.
func guideOfSoulsAttackAbility() *game.TriggeredAbility {
	return &game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Pay{Payment: game.ResolutionPayment{
						Prompt: "Pay {E}{E}{E}?",
						AdditionalCosts: []cost.Additional{{
							Kind:   cost.AdditionalEnergy,
							Text:   "pay {E}{E}{E}",
							Amount: 3,
						}},
					}},
					PublishResult: game.ResultKey("controller-paid"),
				},
				{
					Primitive: game.CreateReflexiveTrigger{Trigger: game.ReflexiveTriggerDef{
						Content: game.Mode{
							Targets: []game.TargetSpec{{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target attacking creature",
								Allow:      game.TargetAllowPermanent,
								Selection: opt.Val(game.Selection{
									RequiredTypesAny: []types.Card{types.Creature},
									CombatState:      game.CombatStateAttacking,
								}),
							}},
							Sequence: []game.Instruction{
								{Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								}},
								{Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Flying,
								}},
								{Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{{
										Layer:       game.LayerType,
										AddSubtypes: []types.Sub{types.Angel},
									}},
									Duration: game.DurationPermanent,
								}},
							},
						}.Ability(),
					}},
					ResultGate: opt.Val(game.InstructionResultGate{
						Key:       "controller-paid",
						Succeeded: game.TriTrue,
					}),
				},
			},
		}.Ability(),
	}
}

// setupGuideOfSoulsAttack builds a game where Player1 controls Guide of Souls and
// a lone attacking creature (the reflexive target), gives Player1 the requested
// energy, declares the attacker, and pushes the attack triggered ability.
func setupGuideOfSoulsAttack(t *testing.T, energy int) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, creatureDef("Guide of Souls"))
	attacker := addCombatPermanent(g, game.Player1, creatureDef("Attacking Bear"))
	g.Players[game.Player1].EnergyCounters = energy
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	ability := guideOfSoulsAttackAbility()
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: ability,
	})
	return g, attacker
}

// TestGuideOfSoulsAttackPaymentBuffsTarget proves the accept path: paying
// {E}{E}{E} spends exactly three energy, and the reflexive trigger — whose target
// is chosen only after the payment resolves — places two +1/+1 counters and a
// flying counter on the attacking creature and permanently makes it an Angel.
func TestGuideOfSoulsAttackPaymentBuffsTarget(t *testing.T) {
	g, attacker := setupGuideOfSoulsAttack(t, 3)
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: guidePaymentAgent{pay: true}}

	resolveStackWithTriggers(engine, g, agents)

	if got := g.Players[game.Player1].EnergyCounters; got != 0 {
		t.Fatalf("energy = %d, want 0 after paying {E}{E}{E}", got)
	}
	if got := attacker.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
	if got := attacker.Counters.Get(counter.Flying); got != 1 {
		t.Fatalf("flying counters = %d, want 1", got)
	}
	if !permanentHasSubtype(g, attacker, types.Angel) {
		t.Fatal("attacker did not become an Angel")
	}
	if !permanentHasType(g, attacker, types.Creature) {
		t.Fatal("attacker lost its creature type, want Angel added in addition")
	}
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %d, want 0 (drained)", len(g.PendingReflexiveTriggers))
	}
}

// TestGuideOfSoulsAttackPaymentDeclinedDoesNothing proves the decline path:
// refusing the optional payment spends no energy, queues no reflexive trigger,
// and leaves the attacking creature with no counters and its original types.
func TestGuideOfSoulsAttackPaymentDeclinedDoesNothing(t *testing.T) {
	g, attacker := setupGuideOfSoulsAttack(t, 3)
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: guidePaymentAgent{pay: false}}

	resolveStackWithTriggers(engine, g, agents)

	if got := g.Players[game.Player1].EnergyCounters; got != 3 {
		t.Fatalf("energy = %d, want 3 unchanged after declining", got)
	}
	if got := attacker.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters = %d, want 0 after declining", got)
	}
	if attacker.Counters.Get(counter.Flying) != 0 {
		t.Fatal("flying counter placed despite declining the payment")
	}
	if permanentHasSubtype(g, attacker, types.Angel) {
		t.Fatal("attacker became an Angel despite declining the payment")
	}
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %d, want 0 (never queued)", len(g.PendingReflexiveTriggers))
	}
}

// TestGuideOfSoulsAttackNoEnergyCannotPay proves that a controller without enough
// energy cannot pay the optional cost, so no energy is spent and the reflexive
// buff never happens even when the controller wants to pay.
func TestGuideOfSoulsAttackNoEnergyCannotPay(t *testing.T) {
	g, attacker := setupGuideOfSoulsAttack(t, 2)
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: guidePaymentAgent{pay: true}}

	resolveStackWithTriggers(engine, g, agents)

	if got := g.Players[game.Player1].EnergyCounters; got != 2 {
		t.Fatalf("energy = %d, want 2 unchanged (cannot pay {E}{E}{E} with 2)", got)
	}
	if attacker.Counters.Get(counter.PlusOnePlusOne) != 0 || attacker.Counters.Get(counter.Flying) != 0 {
		t.Fatal("counters placed despite being unable to pay")
	}
	if permanentHasSubtype(g, attacker, types.Angel) {
		t.Fatal("attacker became an Angel despite being unable to pay")
	}
}

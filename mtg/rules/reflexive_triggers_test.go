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

// reflexiveEnablingAgent answers the optional enabling "you may" with enable and
// always picks the first offered target for the reflexive trigger. Picking the
// first target is sufficient for the Eden-style proof because the only legal
// targets are cards the enabling action just added to the graveyard, so any
// legal choice is a card that did not exist as a target before the enabling
// action resolved.
type reflexiveEnablingAgent struct {
	enable bool
}

func (reflexiveEnablingAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a reflexiveEnablingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		if a.enable {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceTarget:
		return []int{0}
	default:
		return request.DefaultSelection
	}
}

// edenLikeReflexiveAbility mirrors Eden, Seat of the Sanctum's lowered ability:
// mill two cards (mandatory), then optionally sacrifice this permanent, and when
// you do, return a target creature card from your graveyard to your hand. The
// return is a reflexive triggered ability (game.CreateReflexiveTrigger) gated on
// the sacrifice's published result so its target is chosen when the reflexive
// trigger is put on the stack — after the mill has resolved.
func edenLikeReflexiveAbility() game.ActivatedAbility {
	return game.ActivatedAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Mill{
						Amount: game.Fixed(2),
						Player: game.ControllerReference(),
					},
				},
				{
					Primitive:     game.Sacrifice{Object: game.SourceCardPermanentReference()},
					Optional:      true,
					PublishResult: game.ResultKey("if-you-do"),
				},
				{
					Primitive: game.CreateReflexiveTrigger{
						Trigger: game.ReflexiveTriggerDef{
							Content: game.Mode{
								Targets: []game.TargetSpec{{
									MinTargets: 1,
									MaxTargets: 1,
									Constraint: "target creature card from your graveyard",
									Allow:      game.TargetAllowCard,
									TargetZone: zone.Graveyard,
									Selection: opt.Val(game.Selection{
										RequiredTypes: []types.Card{types.Creature},
										Controller:    game.ControllerYou,
									}),
								}},
								Sequence: []game.Instruction{{
									Primitive: game.MoveCard{
										Card:        game.CardReference{Kind: game.CardReferenceTarget},
										FromZone:    zone.Graveyard,
										Destination: zone.Hand,
									},
								}},
							}.Ability(),
						},
					},
					ResultGate: opt.Val(game.InstructionResultGate{
						Key:       "if-you-do",
						Succeeded: game.TriTrue,
					}),
				},
			},
		}.Ability(),
	}
}

func pushEdenLikeAbility(g *game.Game, source *game.Permanent) {
	ability := edenLikeReflexiveAbility()
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackActivatedAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      source.Controller,
		InlineActivated: &ability,
	})
}

// TestReflexiveTriggerTargetsCardMilledByEnablingResolution proves the Eden fix:
// the reflexive "when you do, return a target creature card from your graveyard"
// chooses its target AFTER the enabling resolution has milled cards, so a card
// milled moments earlier is a legal target. Before resolution the graveyard has
// no creature card at all, so the only way a creature can return to hand is if
// the reflexive trigger's target was chosen from the post-mill graveyard.
func TestReflexiveTriggerTargetsCardMilledByEnablingResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Eden Analogue"))
	milledA := addCardToLibrary(g, game.Player1, creatureDef("Milled Bear"))
	milledB := addCardToLibrary(g, game.Player1, creatureDef("Milled Ox"))

	if g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("graveyard not empty before resolution: %d", g.Players[game.Player1].Graveyard.Size())
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: reflexiveEnablingAgent{enable: true}}
	pushEdenLikeAbility(g, source)
	resolveStackWithTriggers(engine, g, agents)

	if onBattlefieldByCard(g, source.CardInstanceID) {
		t.Fatal("source permanent was not sacrificed by the enabling action")
	}
	returned := 0
	for _, cardID := range []id.ID{milledA, milledB} {
		if g.Players[game.Player1].Hand.Contains(cardID) {
			returned++
		}
	}
	if returned != 1 {
		t.Fatalf("milled creatures returned to hand = %d, want exactly 1 (the reflexive target)", returned)
	}
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %d, want 0 (drained)", len(g.PendingReflexiveTriggers))
	}
}

// TestReflexiveTriggerDoesNotFireWhenEnablingActionDeclined proves the reflexive
// trigger is gated on the enabling action: declining the optional sacrifice must
// neither sacrifice the source nor queue or resolve the reflexive return.
func TestReflexiveTriggerDoesNotFireWhenEnablingActionDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Eden Analogue"))
	milledA := addCardToLibrary(g, game.Player1, creatureDef("Milled Bear"))
	milledB := addCardToLibrary(g, game.Player1, creatureDef("Milled Ox"))

	agents := [game.NumPlayers]PlayerAgent{game.Player1: reflexiveEnablingAgent{enable: false}}
	pushEdenLikeAbility(g, source)
	resolveStackWithTriggers(engine, g, agents)

	if !onBattlefieldByCard(g, source.CardInstanceID) {
		t.Fatal("source permanent was sacrificed despite declining the optional action")
	}
	for _, cardID := range []id.ID{milledA, milledB} {
		if g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatal("reflexive return fired even though the enabling action was declined")
		}
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("milled creature left the graveyard without the reflexive trigger firing")
		}
	}
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %d, want 0 (never queued)", len(g.PendingReflexiveTriggers))
	}
}

// TestReflexiveTriggerResolvesEventPermanentReference proves the reflexive
// trigger carries the enabling ability's triggering event: a reflexive body that
// references the triggering event's permanent ("that creature", modeled as
// EventPermanentReference) resolves against that event even though the reflexive
// trigger is a new stack object. This mirrors cards like Sparktongue Dragon and
// Heart Piercer Manticore whose reflexive effect acts on the event permanent.
func TestReflexiveTriggerResolvesEventPermanentReference(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Enabler"))
	witness := addCombatPermanent(g, game.Player1, creatureDef("Event Witness"))

	ability := &game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive:     game.Sacrifice{Object: game.SourceCardPermanentReference()},
					Optional:      true,
					PublishResult: game.ResultKey("if-you-do"),
				},
				{
					Primitive: game.CreateReflexiveTrigger{
						Trigger: game.ReflexiveTriggerDef{
							Content: game.Mode{
								Sequence: []game.Instruction{{
									Primitive: game.AddCounter{
										Amount:      game.Fixed(1),
										Object:      game.EventPermanentReference(),
										CounterKind: counter.PlusOnePlusOne,
									},
								}},
							}.Ability(),
						},
					},
					ResultGate: opt.Val(game.InstructionResultGate{
						Key:       "if-you-do",
						Succeeded: game.TriTrue,
					}),
				},
			},
		}.Ability(),
	}

	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		InlineTrigger:   ability,
		TriggerEvent:    game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: witness.ObjectID},
		HasTriggerEvent: true,
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: reflexiveEnablingAgent{enable: true}}
	resolveStackWithTriggers(engine, g, agents)

	if onBattlefieldByCard(g, source.CardInstanceID) {
		t.Fatal("source permanent was not sacrificed by the enabling action")
	}
	if got := witness.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("event permanent +1/+1 counters = %d, want 1 (reflexive body resolved against the triggering event)", got)
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.Permanent {
	return addCombatPermanent(g, controller, triggeredCreature(pattern, instructions, targets))
}

func addOptionalTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.Permanent {
	card := triggeredCreature(pattern, instructions, targets)
	card.TriggeredAbilities[0].Optional = true
	return addCombatPermanent(g, controller, card)
}

func addTriggeredPermanentWithCondition(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, lifeAtLeast int, instructions []game.Instruction) *game.Permanent {
	permanent := addTriggeredPermanent(g, controller, pattern, instructions, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningCondition = opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: lifeAtLeast}},
	})
	return permanent
}

func addSelfEnterInterveningTrigger(g *game.Game, condition *game.TriggerCondition) *game.Permanent {
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, nil, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	condition.Type = game.TriggerWhen
	condition.Pattern = game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}
	card.Def.TriggeredAbilities[0].Trigger = *condition
	return permanent
}

func addSelfDiesCounterAbsenceTrigger(g *game.Game, kind counter.Kind) *game.Permanent {
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningIfEventPermanentHadNoCounterKind = opt.Val(kind)
	return permanent
}

func addSelfDiesCounterPresenceTrigger(g *game.Game, kind counter.Kind) *game.Permanent {
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningIfEventPermanentHadCounterKind = opt.Val(kind)
	return permanent
}

func selfDiesEventCardDefinition(primitive game.Primitive) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Returning Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(),
		}},
	}}
}

func selfDiesAdventureDefinition() *game.CardDef {
	def := selfDiesEventCardDefinition(game.GrantCastPermission{
		Card:     game.CardReference{Kind: game.CardReferenceEvent},
		FromZone: zone.Graveyard,
		Face:     game.FaceAlternate,
		Duration: game.DurationUntilEndOfYourNextTurn,
	})
	def.Layout = game.LayoutAdventure
	def.Alternate = opt.Val(game.CardFace{
		Name:         "Returning Adventure",
		Types:        []types.Card{types.Sorcery},
		Subtypes:     []types.Sub{types.Adventure},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}}.Ability()),
	})
	return def
}

func addCounterTransferTriggerSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Counter Transfer Source",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{Event: game.EventZoneChanged, Controller: game.TriggerControllerYou, RequirePermanentTypes: []types.Card{types.Artifact}, MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard}, InterveningIf: "it had counters on it", InterveningIfEventPermanentHadCounters: true},
				Content: game.Mode{
					Targets: []game.TargetSpec{
						{MinTargets: 0, MaxTargets: 1, Constraint: "artifact or creature you control"},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.MoveCounters{
								Object: game.TargetPermanentReference(0),
								Source: game.CounterSourceSpec{
									Kind: game.CounterSourceEventPermanent,
								},
								AllKinds: true,
							},
						},
					},
				}.Ability(),
			},
		}},
	})
}

func TestTriggerPatternRequireNonToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		RequireNonToken:       true,
	}
	source := addTriggeredPermanent(g, game.Player1, pattern, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	token, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Token", Types: []types.Card{types.Creature}}})
	if !ok {
		t.Fatal("createTokenPermanent failed")
	}
	card := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nontoken", Types: []types.Card{types.Creature}}})

	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: token.ObjectID,
		TokenName:   "Token",
		TokenDef:    token.TokenDef,
	}) {
		t.Fatal("non-token trigger matched token event")
	}
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: card.ObjectID,
		CardID:      card.CardInstanceID,
	}) {
		t.Fatal("non-token trigger did not match nontoken event")
	}
}

func TestPhasedOutPermanentDoesNotDetectTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventCardDrawn,
		Controller: game.TriggerControllerYou,
	}, nil, nil)
	event := game.Event{
		Kind:       game.EventCardDrawn,
		Controller: game.Player1,
		Player:     game.Player1,
		Amount:     1,
	}

	source.PhasedOut = true
	if pending := engine.detectTriggeredAbilities(g, []game.Event{event}); len(pending) != 0 {
		t.Fatalf("phased-out source detected %d triggers, want 0", len(pending))
	}
	source.PhasedOut = false
	if pending := engine.detectTriggeredAbilities(g, []game.Event{event}); len(pending) != 1 {
		t.Fatalf("phased-in source detected %d triggers, want 1", len(pending))
	}
}

func triggeredCreature(pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Triggered Creature",
		Types:     []types.Card{types.Creature},
		ManaCost:  greenCost(),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: *pattern},
				Content: game.Mode{
					Targets:  targets,
					Sequence: instructions,
				}.Ability(),
			},
		}},
	}
}

type choiceOnlyAgent struct {
	choices [][]int
	next    int
}

func (*choiceOnlyAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	return action.Pass()
}

func (a *choiceOnlyAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if a.next >= len(a.choices) {
		return nil
	}
	choice := append([]int(nil), a.choices[a.next]...)
	a.next++
	return choice
}

// TestUnionTokenTriggerFiresOnCreateAndSacrifice exercises the event-union
// trigger pattern ("Whenever you create or sacrifice a token"): the same
// triggered ability must fire on both the token-created event and the
// token-sacrificed event.
func TestUnionTokenTriggerFiresOnCreateAndSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventTokenCreated,
		UnionEvent:       game.EventPermanentSacrificed,
		Player:           game.TriggerPlayerYou,
		SubjectSelection: game.Selection{TokenOnly: true},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn One"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Two"}})

	token, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Treasure",
		Types: []types.Card{types.Artifact},
	}})
	if !ok {
		t.Fatal("token was not created")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token-created event did not fire the union trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after create = %d, want 1", got)
	}

	if !sacrificePermanent(g, token) {
		t.Fatal("token was not sacrificed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token-sacrificed event did not fire the union trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size after sacrifice = %d, want 2", got)
	}
}

// TestUnionEnterAttackTriggerFiresOnEnterAndAttack exercises the event-union
// trigger pattern "Whenever this creature enters or attacks": the same
// self-scoped triggered ability must fire on both the enters-the-battlefield
// event and the attack event, and must ignore another permanent's attack.
func TestUnionEnterAttackTriggerFiresOnEnterAndAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:      game.EventPermanentEnteredBattlefield,
		UnionEvent: game.EventAttackerDeclared,
		Source:     game.TriggerSourceSelf,
	}
	source := addTriggeredPermanent(g, game.Player1, pattern, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("enter-or-attack union did not match the enter event")
	}
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("enter-or-attack union did not match the attack event")
	}

	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other",
		Types: []types.Card{types.Creature},
	}})
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: other.ObjectID,
	}) {
		t.Fatal("self-scoped union matched another permanent's attack event")
	}
}

package rules

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func minscBooTimelessHeroesDef() *game.CardDef {
	boo := minscBooTokenDef()
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Minsc & Boo, Timeless Heroes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:         []color.Color{color.Red, color.Green},
			CanBeCommander: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Planeswalker},
			Subtypes:       []types.Sub{types.Minsc},
			Loyalty:        opt.Val(3),
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(boo),
								},
							},
						},
					}.Ability(),
				},
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(boo),
								},
							},
						},
					}.Ability(),
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				{
					LoyaltyCost: 1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature with trample or haste",
								Allow:      game.TargetAllowPermanent,
								Selection: opt.Val(game.Selection{
									AnyOf: []game.Selection{
										{Keyword: game.Trample},
										{Keyword: game.Haste},
									},
									RequiredTypesAny: []types.Card{types.Creature},
								}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(3),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
				{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:               game.Fixed(1),
									Player:               game.ControllerReference(),
									Selection:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
									PublishLinked:        game.LinkedKey("sacrificed-creature"),
									PublishObjectBinding: true,
								},
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "any target",
													Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.Damage{
														Amount: game.Dynamic(game.DynamicAmount{
															Kind:       game.DynamicAmountObjectPower,
															Multiplier: 1,
															Object:     game.LinkedObjectReference("sacrificed-creature"),
														}),
														Recipient:    game.AnyTargetDamageRecipient(0),
														DamageSource: opt.Val(game.SourcePermanentReference()),
													},
												},
												{
													Primitive: game.Draw{
														Amount: game.Dynamic(game.DynamicAmount{
															Kind:       game.DynamicAmountObjectPower,
															Multiplier: 1,
															Object:     game.LinkedObjectReference("sacrificed-creature"),
														}),
														Player: game.ControllerReference(),
													},
													Condition: opt.Val(game.EffectCondition{
														Object: game.LinkedObjectReference("sacrificed-creature"),
														Condition: opt.Val(game.Condition{
															Object: opt.Val(game.LinkedObjectReference("sacrificed-creature")),
															ObjectMatches: opt.Val(game.Selection{
																SubtypesAny: []types.Sub{types.Hamster},
															}),
														}),
													}),
												},
											},
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
				},
			},
		},
	}
}

func minscBooTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Boo",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Hamster},
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.HasteStaticBody,
		},
	}}
}

type minscReflexiveAgent struct {
	sacrificeName        string
	targetAfterSacrifice *bool
}

func (minscReflexiveAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a minscReflexiveAgent) ChooseChoice(observation PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceResolution:
		for i := range request.Options {
			if request.Options[i].Card.Exists &&
				request.Options[i].Card.Val.Name == a.sacrificeName {
				return []int{i}
			}
		}
	case game.ChoiceTarget:
		if a.targetAfterSacrifice != nil {
			*a.targetAfterSacrifice = true
			for _, permanent := range observation.Battlefield() {
				if permanent.Name == a.sacrificeName {
					*a.targetAfterSacrifice = false
				}
			}
		}
		for i := range request.Options {
			for _, target := range request.Options[i].Targets {
				if target.Kind == game.TargetPlayer && target.PlayerID == game.Player2 {
					return []int{i}
				}
			}
		}
	default:
	}
	return request.DefaultSelection
}

func minscReflexiveAbility() game.ActivatedAbility {
	def := minscBooTimelessHeroesDef()
	return game.ActivatedAbility{Content: def.LoyaltyAbilities[1].Content}
}

func pushMinscReflexiveAbility(g *game.Game, source *game.Permanent) {
	ability := minscReflexiveAbility()
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackActivatedAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		InlineActivated: &ability,
	})
}

func minscHamsterDef(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Hamster},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func TestMinscBooOptionalTokenTriggers(t *testing.T) {
	for _, triggerIndex := range []int{0, 1} {
		for _, accept := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%s", triggerIndex, map[bool]string{false: "decline", true: "accept"}[accept]), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				def := minscBooTimelessHeroesDef()
				source := addCombatPermanent(g, game.Player1, def)
				ability := def.TriggeredAbilities[triggerIndex]
				g.Stack.Push(&game.StackObject{
					ID:            g.IDGen.Next(),
					Kind:          game.StackTriggeredAbility,
					SourceID:      source.ObjectID,
					SourceCardID:  source.CardInstanceID,
					Controller:    game.Player1,
					InlineTrigger: &ability,
					AbilityIndex:  triggerIndex,
				})
				choice := 0
				if accept {
					choice = 1
				}
				agents := [game.NumPlayers]PlayerAgent{
					game.Player1: &choiceOnlyAgent{choices: [][]int{{choice}}},
				}

				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

				booCount := 0
				for _, permanent := range g.Battlefield {
					if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Boo" {
						booCount++
					}
				}
				want := 0
				if accept {
					want = 1
				}
				if booCount != want {
					t.Fatalf("Boo tokens = %d, want %d", booCount, want)
				}
			})
		}
	}
}

func TestMinscBooPlusOneUsesLiveTrampleOrHasteUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	minsc := addCombatPermanent(g, game.Player1, minscBooTimelessHeroesDef())
	minsc.Counters.Add(counter.Loyalty, 3)
	tramplerDef := creatureDef("Trampler")
	tramplerDef.StaticAbilities = []game.StaticAbility{game.TrampleStaticBody}
	trampler := addCombatPermanent(g, game.Player1, tramplerDef)
	hastyDef := creatureDef("Hasty")
	hastyDef.StaticAbilities = []game.StaticAbility{game.HasteStaticBody}
	hasty := addCombatPermanent(g, game.Player2, hastyDef)
	ordinary := addCombatPermanent(g, game.Player1, creatureDef("Ordinary"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	actions := engine.legalActions(g, game.Player1)
	noTarget := action.ActivateAbility(minsc.ObjectID, 0, nil, 0)
	trampleTarget := action.ActivateAbility(minsc.ObjectID, 0, []game.Target{game.PermanentTarget(trampler.ObjectID)}, 0)
	hasteTarget := action.ActivateAbility(minsc.ObjectID, 0, []game.Target{game.PermanentTarget(hasty.ObjectID)}, 0)
	ordinaryTarget := action.ActivateAbility(minsc.ObjectID, 0, []game.Target{game.PermanentTarget(ordinary.ObjectID)}, 0)
	if !containsAction(actions, noTarget) ||
		!containsAction(actions, trampleTarget) ||
		!containsAction(actions, hasteTarget) {
		t.Fatalf("legal actions do not include the no-target, trample, and haste choices: %+v", actions)
	}
	if containsAction(actions, ordinaryTarget) {
		t.Fatal("ordinary creature without trample or haste was a legal target")
	}
	if !engine.applyAction(g, game.Player1, trampleTarget) {
		t.Fatal("failed to activate +1 targeting creature with trample")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := trampler.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
}

func TestMinscReflexiveSacrificeUsesLKIAndLateTargeting(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Minsc Source"))
	boo := addTokenCreaturePermanent(g, game.Player1, "Boo")
	boo.TokenDef = minscHamsterDef("Boo", 4)
	for i := range 4 {
		addCardToLibrary(g, game.Player1, creatureDef("Card "+strings.Repeat("x", i+1)))
	}
	late := false
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: minscReflexiveAgent{sacrificeName: "Boo", targetAfterSacrifice: &late},
	}

	pushMinscReflexiveAbility(g, source)
	movePermanentToZone(g, source, zone.Graveyard)
	resolveStackWithTriggers(engine, g, agents)

	if !late {
		t.Fatal("reflexive target was not chosen after the sacrificed token left")
	}
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("player 2 life = %d, want 36", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Hand.Size() != 4 {
		t.Fatalf("cards drawn = %d, want 4", g.Players[game.Player1].Hand.Size())
	}
}

func TestMinscReflexiveNegativePowerUsesZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Minsc Source"))
	hamster := addCombatPermanent(g, game.Player1, minscHamsterDef("Negative Hamster", -2))
	_ = hamster
	addCardToLibrary(g, game.Player1, creatureDef("Undrawn Card"))
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: minscReflexiveAgent{sacrificeName: "Negative Hamster"},
	}

	pushMinscReflexiveAbility(g, source)
	resolveStackWithTriggers(engine, g, agents)

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("player 2 life = %d, want 40", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("cards drawn = %d, want 0", g.Players[game.Player1].Hand.Size())
	}
}

func TestMinscReflexiveSurvivesSacrificeZoneReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, landDef("Minsc Source"))
	hamster := addCombatPermanent(g, game.Player1, minscHamsterDef("Exiled Hamster", 3))
	for i := range 3 {
		addCardToLibrary(g, game.Player1, creatureDef("Draw "+strings.Repeat("x", i+1)))
	}
	replacement := game.ReplacementEffect{
		Description:   "exile instead",
		MatchEvent:    game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		MatchToZone:   true,
		ToZone:        zone.Graveyard,
		ReplaceToZone: zone.Exile,
	}
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
		Replacement: &replacement,
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: minscReflexiveAgent{sacrificeName: "Exiled Hamster"},
	}

	pushMinscReflexiveAbility(g, source)
	resolveStackWithTriggers(engine, g, agents)

	if !g.Players[game.Player1].Exile.Contains(hamster.CardInstanceID) {
		t.Fatal("sacrificed Hamster did not use the exile replacement")
	}
	if g.Players[game.Player2].Life != 37 || g.Players[game.Player1].Hand.Size() != 3 {
		t.Fatalf("life=%d hand=%d, want 37 and 3", g.Players[game.Player2].Life, g.Players[game.Player1].Hand.Size())
	}
}

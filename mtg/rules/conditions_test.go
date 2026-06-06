package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestConditionControllerControlsPermanentFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "types.Snow Mountain",
		Supertypes: []types.Super{types.Basic, types.Snow},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Mountain}},
	})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 7}),
		Toughness: opt.Val(game.PT{Value: 7})},
	})

	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Land},
			Supertypes: []types.Super{types.Basic},
			SubtypesAny: []types.Sub{
				types.Swamp,
				types.Mountain,
			},
			MinCount: 1,
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not match controlled basic Mountain")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition matched another player's Mountain")
	}

	powerCondition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types: []types.Card{types.Creature},
			Power: opt.Val(compare.Int{
				Op:    compare.GreaterOrEqual,
				Value: 7,
			}),
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, powerCondition) {
		t.Fatal("condition did not match controlled creature with power >= 7")
	}
	powerCondition.Val.Negate = true
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, powerCondition) {
		t.Fatal("negated condition matched")
	}
}

func TestConditionControllerControlsPermanentFilterCanExcludeSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5})},
	})
	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:         []types.Card{types.Creature},
			ExcludeSource: true,
			Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
		},
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: source}, condition) {
		t.Fatal("condition matched source as another creature")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Other",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4})},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, source: source}, condition) {
		t.Fatal("condition did not match another large creature")
	}
}

func TestConditionTargetEnteredThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "New Creature",
		Types: []types.Card{types.Creature}},
	})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	}
	condition := opt.Val(game.Condition{TargetEnteredThisTurn: opt.Val(0)})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched before enter event")
	}
	emitEvent(g, game.GameEvent{Kind: game.EventPermanentEnteredBattlefield, PermanentID: creature.ObjectID})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition did not match target that entered this turn")
	}
	g.Turn.TurnNumber++
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched target that entered on a previous turn")
	}
}

func TestConditionCastFromZoneRequiresNonCopyStackObject(t *testing.T) {
	condition := opt.Val(game.Condition{CastFromZone: opt.Val(zone.Graveyard)})
	obj := &game.StackObject{
		Kind:       game.StackSpell,
		Controller: game.Player1,
		SourceZone: zone.Graveyard,
	}
	if !conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition did not match spell cast from graveyard")
	}
	obj.Copy = true
	if conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched copied stack object")
	}
	obj.Copy = false
	obj.SourceZone = zone.Hand
	if conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched spell cast from hand")
	}
}

func TestConditionControllerControlsTotalPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Small", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 3})}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Medium", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 5})}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})

	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Creature},
			TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not match total creature power >= 8")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition matched another player's creatures")
	}
}

func TestConditionEventPermanentNameUniqueAmongControlledAndGraveyardCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unique := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	condition := opt.Val(game.Condition{EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true})
	ctx := conditionContext{
		controller: game.Player1,
		event: &game.GameEvent{
			PermanentID: unique.ObjectID,
		},
	}
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition counted the event permanent as another matching creature")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition matched with another controlled creature of the same name")
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unique = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	ctx.event = &game.GameEvent{PermanentID: unique.ObjectID}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition matched with a same-name creature card in graveyard")
	}
}

func TestActivationConditionRestrictsExplicitAndAutoMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	verge := addCombatPermanent(g, game.Player1, conditionalRedManaLand())
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Red Spell",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			Text: "Do nothing.",
		}}},
	})

	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(verge.ObjectID, 0, nil, 0)) {
		t.Fatal("conditional red mana ability was legal without Swamp or Mountain")
	}
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("auto-payment treated conditional red mana as available without Swamp or Mountain")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain}},
	})
	if !containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(verge.ObjectID, 0, nil, 0)) {
		t.Fatal("conditional red mana ability was not legal with Mountain")
	}
	if !containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("auto-payment did not use conditional red mana with Mountain")
	}
}

func TestInterveningConditionChecksControllerPermanentPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBugenhagenLikePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired without controlled creature with power >= 7")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 7})},
	})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Again"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not fire with controlled creature with power >= 7")
	}
	g.Battlefield = g.Battlefield[:1]
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want intervening-if recheck to skip draw", got)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestStaticConditionGraveyardAbilityGrantsHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types: []types.Card{types.Creature}},
	})
	other := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Other Creature",
		Types: []types.Card{types.Creature}},
	})
	angerID := addCardToGraveyard(g, game.Player1, angerLikeCard())
	angerCard := g.CardInstances[angerID]

	if hasKeyword(g, creature, game.Haste) {
		t.Fatal("Anger-like graveyard ability granted haste without Mountain")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain}},
	})
	if !hasKeyword(g, creature, game.Haste) {
		t.Fatal("Anger-like graveyard ability did not grant haste with Mountain")
	}
	if hasKeyword(g, other, game.Haste) {
		t.Fatal("Anger-like graveyard ability granted haste to opponent's creature")
	}

	g.Players[game.Player1].Graveyard.Remove(angerID)
	battlefieldAnger := addCombatPermanent(g, game.Player1, angerCard.Def)
	if hasKeyword(g, creature, game.Haste) {
		t.Fatal("graveyard-only Anger-like static ability functioned from battlefield")
	}
	if !hasKeyword(g, battlefieldAnger, game.Haste) {
		t.Fatal("battlefield Anger-like source did not have its own haste keyword")
	}
}

func TestConditionalEntersTappedCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, cinderLikeLand())
	engine := NewEngine(nil)
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land without basics failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; !got.Tapped {
		t.Fatalf("land = %+v, want tapped without two basic lands", got)
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest}},
	})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Island}},
	})
	cardID = addCardToHand(g, game.Player1, cinderLikeLand())
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land with basics failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; got.Tapped {
		t.Fatalf("land = %+v, want untapped with two basic lands", got)
	}
}

func setSorcerySpeedTurn(g *game.Game, playerID game.PlayerID) {
	g.Turn.ActivePlayer = playerID
	g.Turn.PriorityPlayer = playerID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
}

func conditionalRedManaLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Conditional Red Land",
		Types: []types.Card{types.Land},
		Abilities: []game.AbilityDef{{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {R}. Activate only if you control a Swamp or a Mountain.",
			IsManaAbility: true,
			ActivationCondition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
				},
			}),
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
			Effects: []game.Effect{{
				Type:        game.EffectAddMana,
				Amount:      1,
				ManaColor:   mana.R,
				TargetIndex: game.TargetIndexController,
			}},
		}}},
	}
}

func addBugenhagenLikePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Bugenhagen-like",
		Types: []types.Card{types.Creature},
		Abilities: []game.AbilityDef{{
			Kind: game.TriggeredAbility,
			Text: "At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event: game.EventCardDrawn,
				},
				InterveningIf: "if you control a creature with power 7 or greater",
				InterveningCondition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						Types: []types.Card{types.Creature},
						Power: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 7,
						}),
					},
				}),
			}),
			Effects: []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController}},
		}}},
	})
}

func angerLikeCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Anger-like",
		Types: []types.Card{types.Creature},
		Abilities: []game.AbilityDef{
			{
				Kind:             game.StaticAbility,
				Text:             "Haste",
				KeywordAbilities: game.SimpleKeywords(game.Haste),
			},
			{
				Kind:           game.StaticAbility,
				Text:           "As long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
				ZoneOfFunction: zone.Graveyard,
				Condition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						SubtypesAny: []types.Sub{types.Mountain},
					},
				}),
				Effects: []game.Effect{{
					Type: game.EffectApplyContinuous,
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:       game.LayerAbility,
						Selector:    game.EffectSelectorCreaturesYouControl,
						AddKeywords: []game.Keyword{game.Haste},
					}},
				}},
			},
		}},
	}
}

func cinderLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cinder-like",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbilityDef{
			game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
				Negate: true,
				ControllerControls: game.PermanentFilter{
					Types:      []types.Card{types.Land},
					Supertypes: []types.Super{types.Basic},
					MinCount:   2,
				},
			}),
		}},
	}
}

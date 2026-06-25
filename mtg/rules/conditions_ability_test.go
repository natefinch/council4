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

func TestControllerControlsNamedCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	// Oracle text spells the gate with a hyphen ("Urza's Power-Plant") while the
	// printed card name uses a space; the predicate must reconcile both.
	condition := opt.Val(game.Condition{ControllerControlsNamed: []string{"Urza's Power-Plant", "Urza's Tower"}})

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Urza's Mine"}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-named condition passed while controlling neither named permanent")
	}

	plant := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Urza's Power Plant"}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-named condition passed while controlling only the Power Plant")
	}

	tower := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Urza's Tower"}})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-named condition failed while controlling both named permanents")
	}

	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("controls-named condition passed for a player controlling neither permanent")
	}

	tower.Controller = game.Player2
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-named condition passed after the Tower changed controllers")
	}
	tower.Controller = game.Player1

	plant.PhasedOut = true
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-named condition passed while the Power Plant is phased out")
	}
}

// tronManaAbility builds the Urza tron land mana ability: "{T}: Add {C}. If you
// control an Urza's Power-Plant and an Urza's Tower, add {C}{C} instead." The
// base {C} production is gated on NOT meeting the named-permanent condition and
// the {C}{C} bonus on meeting it, so exactly one production resolves.
func tronManaAbility() game.ManaAbility {
	named := game.Condition{ControllerControlsNamed: []string{"Urza's Power-Plant", "Urza's Tower"}}
	notNamed := named
	notNamed.Negate = true
	baseGate := opt.Val(game.EffectCondition{Condition: opt.Val(notNamed)})
	bonusGate := opt.Val(game.EffectCondition{Condition: opt.Val(named)})
	return game.ManaAbility{
		Text:            "{T}: Add {C}. If you control an Urza's Power-Plant and an Urza's Tower, add {C}{C} instead.",
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}, Condition: baseGate},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}, Condition: bonusGate},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}, Condition: bonusGate},
		}}.Ability(),
	}
}

// TestTronManaAbilityAddsBaseThenBonus proves the lowered tron mana ability
// adds the base {C} when its named-permanent condition fails and the {C}{C}
// bonus instead when the controller also controls an Urza's Power Plant and an
// Urza's Tower.
func TestTronManaAbilityAddsBaseThenBonus(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := tronManaAbility()
	mine := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Urza's Mine", Types: []types.Card{types.Land}}},
		&body,
	)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mine.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(tron mana ability, base) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 1 {
		t.Fatalf("colorless mana = %d, want 1 (base production only)", got)
	}

	mine.Tapped = false
	g.Players[game.Player1].ManaPool = mana.Pool{}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Urza's Power Plant", Types: []types.Card{types.Land}}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Urza's Tower", Types: []types.Card{types.Land}}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mine.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(tron mana ability, bonus) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2 (bonus production)", got)
	}
}

func TestControllerControlsCommanderCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	condition := opt.Val(game.Condition{ControllerControlsCommander: true})

	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-commander condition passed with no commander on the battlefield")
	}

	commander := addCommanderPermanent(g, game.Player1)
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-commander condition failed while controlling commander")
	}

	commander.PhasedOut = true
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("controls-commander condition passed while commander is phased out")
	}
	commander.PhasedOut = false

	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("controls-commander condition passed for a player who does not control the commander")
	}
}

func TestControllerGainedLifeThisTurnCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	condition := opt.Val(game.Condition{ControllerGainedLifeThisTurnAtLeast: 3})

	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("gained-life condition passed with no life gained this turn")
	}

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 2})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("gained-life condition passed at 2 life gained, want threshold 3")
	}

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 1})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("gained-life condition failed at 3 life gained this turn")
	}

	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("gained-life condition passed for a player who gained no life this turn")
	}
}

func TestActivationConditionRestrictsAutoMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	verge := addCombatPermanent(g, game.Player1, conditionalRedManaLand())
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Red Spell",
		ManaCost:     opt.Val(cost.Mana{cost.R}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("auto-payment treated conditional red mana as available without Swamp or Mountain")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain}},
	})
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(verge.ObjectID, 0, nil, 0)) {
		t.Fatal("payment-only conditional mana ability was exposed as a standalone action")
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

	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
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
	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
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

func TestActivationConditionChecksControlledCreaturesTotalPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setSorcerySpeedTurn(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Formidable Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           "{G}: Draw a card. Activate only if creatures you control have total power 8 or greater.",
			ManaCost:       greenCost(),
			ZoneOfFunction: zone.Battlefield,
			ActivationCondition: opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
					TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
				}),
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	// The source alone has only power 2, below the total-power threshold.
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal with controlled total power below 8")
	}

	// An opponent's large creature must not satisfy the controller predicate.
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent Giant",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 9})},
	})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal counting an opponent's creature power")
	}

	// Adding controlled creatures to reach total power 8 makes it legal.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Big Ally",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 6})},
	})
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability not legal with controlled total power 8")
	}
}

func TestActivationConditionChecksEventHistoryAttackedThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setSorcerySpeedTurn(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Raiding Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           "{G}: Draw a card. Activate only if you attacked this turn.",
			ManaCost:       greenCost(),
			ZoneOfFunction: zone.Battlefield,
			ActivationCondition: opt.Val(game.Condition{
				EventHistory: opt.Val(game.EventHistoryCondition{
					Pattern: game.TriggerPattern{
						Event:      game.EventAttackerDeclared,
						Controller: game.TriggerControllerYou,
					},
					Window: game.EventHistoryCurrentTurn,
				}),
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	// No attack has happened this turn, so the ability is illegal.
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal before the controller attacked this turn")
	}

	// An opponent's attack must not satisfy the controller-relative predicate.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal when only an opponent attacked this turn")
	}

	// Once the controller attacks this turn, the ability becomes legal.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability not legal after the controller attacked this turn")
	}
}

func TestActivationConditionChecksCreatedTokenThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setSorcerySpeedTurn(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Token Idol",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           "{G}: Draw a card. Activate only if you created a token this turn.",
			ManaCost:       greenCost(),
			ZoneOfFunction: zone.Battlefield,
			ActivationCondition: opt.Val(game.Condition{
				ControllerCreatedTokenThisTurn: true,
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	tokenDef := &game.CardDef{CardFace: game.CardFace{Name: "Treasure", Types: []types.Card{types.Artifact}}}

	// No token has been created this turn, so the ability is illegal.
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal before the controller created a token this turn")
	}

	// An opponent creating a token must not satisfy the controller predicate.
	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player2, TokenDef: tokenDef})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal when only an opponent created a token this turn")
	}

	// A nontoken permanent entering under the controller must not satisfy it.
	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player1})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability legal when only a nontoken permanent entered this turn")
	}

	// Once the controller creates a token this turn, the ability becomes legal.
	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player1, TokenDef: tokenDef})
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ability not legal after the controller created a token this turn")
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

func TestConditionalEntersTappedLegendaryCreatureCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, minasTirithLikeLand())
	engine := NewEngine(nil)
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land without a legendary creature failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; !got.Tapped {
		t.Fatalf("land = %+v, want tapped without a legendary creature", got)
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	// A non-legendary creature must not satisfy the condition.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Grizzly Bears",
		Types: []types.Card{types.Creature}}})
	cardID = addCardToHand(g, game.Player1, minasTirithLikeLand())
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land with only a non-legendary creature failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; !got.Tapped {
		t.Fatalf("land = %+v, want tapped with only a non-legendary creature", got)
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Llanowar Visionary",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature}}})
	cardID = addCardToHand(g, game.Player1, minasTirithLikeLand())
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land with a legendary creature failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; got.Tapped {
		t.Fatalf("land = %+v, want untapped with a legendary creature", got)
	}
}

func minasTirithLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Minas-Tirith-like",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedIfReplacement("Minas Tirith enters tapped unless you control a legendary creature.", &game.Condition{
				Negate: true,
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Supertypes:    []types.Super{types.Legendary},
					},
				}),
			}),
		}},
	}
}

func TestLifeAndOpponentConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}

	g.Players[game.Player1].Life = 10
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 10}}})) {
		t.Fatal("controller life-at-least condition failed at threshold")
	}
	g.Players[game.Player1].Life = 9
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 10}}})) {
		t.Fatal("controller life-at-least condition passed below threshold")
	}

	g.Players[game.Player2].Life = 13
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyPlayerLifeAtMost: 13})) {
		t.Fatal("any-player life condition failed at threshold")
	}
	g.Players[game.Player1].Life = 14
	g.Players[game.Player2].Life = 14
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyPlayerLifeAtMost: 13})) {
		t.Fatal("any-player life condition passed with all players above threshold")
	}

	g.Players[game.Player3].Eliminated = true
	g.Players[game.Player4].Eliminated = true
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentCountAtLeast: 2})) {
		t.Fatal("opponent-count condition included eliminated players")
	}
	g.Players[game.Player3].Eliminated = false
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentCountAtLeast: 2})) {
		t.Fatal("opponent-count condition failed with two alive opponents")
	}
}

func TestControllerLifeRelativeConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}

	// StartingLife defaults to 40; "at least 10 life more than your starting
	// life total" needs current life >= 50.
	aboveStarting := opt.Val(game.Condition{ControllerLifeAtLeastAboveStarting: 10})
	g.Players[game.Player1].Life = 50
	if !conditionSatisfied(g, ctx, aboveStarting) {
		t.Fatal("life-above-starting condition failed at threshold")
	}
	g.Players[game.Player1].Life = 49
	if conditionSatisfied(g, ctx, aboveStarting) {
		t.Fatal("life-above-starting condition passed below threshold")
	}

	atMost := opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 5}}})
	g.Players[game.Player1].Life = 5
	if !conditionSatisfied(g, ctx, atMost) {
		t.Fatal("life-at-most condition failed at threshold")
	}
	g.Players[game.Player1].Life = 6
	if conditionSatisfied(g, ctx, atMost) {
		t.Fatal("life-at-most condition passed above threshold")
	}

	// "0 or less life" must remain an active predicate, not be treated as empty.
	atMostZero := opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 0}}})
	g.Players[game.Player1].Life = 0
	if !conditionSatisfied(g, ctx, atMostZero) {
		t.Fatal("life-at-most-zero condition failed at zero life")
	}
	g.Players[game.Player1].Life = 1
	if conditionSatisfied(g, ctx, atMostZero) {
		t.Fatal("life-at-most-zero condition passed above zero life")
	}
}

func TestNegativeConditionThresholdsFailClosed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	conditions := []game.Condition{
		{Negate: true, Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: -1}}},
		{Negate: true, Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: -1}}},
		{Negate: true, ControllerLifeAtLeastAboveStarting: -1},
		{Negate: true, AnyPlayerLifeAtMost: -1},
		{Negate: true, OpponentCountAtLeast: -1},
	}
	for _, condition := range conditions {
		if conditionSatisfied(g, ctx, opt.Val(condition)) {
			t.Fatalf("conditionSatisfied(%+v) = true, want false", condition)
		}
		countConditions := []game.Condition{
			{Negate: true, ControlsMatching: opt.Val(game.SelectionCount{MinCount: -1})},
			{Negate: true, AnyOpponentControls: opt.Val(game.SelectionCount{MinCount: -1})},
			{Negate: true, OpponentsControl: opt.Val(game.SelectionCount{MinCount: -1})},
		}
		for _, condition := range countConditions {
			if conditionSatisfied(g, ctx, opt.Val(condition)) {
				t.Fatalf("conditionSatisfied(%+v) = true, want false", condition)
			}
		}
	}
}

func TestControllerLibrarySizeAndLifeExactlyConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}

	// "you have N or more cards in your library" (Battle of Wits).
	librarySize := opt.Val(game.Condition{ControllerLibrarySizeAtLeast: 2})
	if conditionSatisfied(g, ctx, librarySize) {
		t.Fatal("library-size condition passed with an empty library")
	}
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card A", Types: []types.Card{types.Land}}})
	if conditionSatisfied(g, ctx, librarySize) {
		t.Fatal("library-size condition passed below threshold")
	}
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card B", Types: []types.Card{types.Land}}})
	if !conditionSatisfied(g, ctx, librarySize) {
		t.Fatal("library-size condition failed at threshold")
	}

	// "you have exactly N life" (Near-Death Experience).
	lifeExactly := opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.Equal, Value: 1}}})
	g.Players[game.Player1].Life = 1
	if !conditionSatisfied(g, ctx, lifeExactly) {
		t.Fatal("life-exactly condition failed at exact life")
	}
	g.Players[game.Player1].Life = 2
	if conditionSatisfied(g, ctx, lifeExactly) {
		t.Fatal("life-exactly condition passed above exact life")
	}
	g.Players[game.Player1].Life = 0
	if conditionSatisfied(g, ctx, lifeExactly) {
		t.Fatal("life-exactly condition passed below exact life")
	}
}

func TestOpponentPermanentCountConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	count := game.SelectionCount{
		Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
		MinCount:  2,
	}
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player3, types.Island)

	if conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyOpponentControls: opt.Val(count)})) {
		t.Fatal("any-opponent condition combined permanents across opponents")
	}
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentsControl: opt.Val(count)})) {
		t.Fatal("collective opponents condition did not combine permanents")
	}
	addBasicLandPermanent(g, game.Player2, types.Mountain)
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyOpponentControls: opt.Val(count)})) {
		t.Fatal("any-opponent condition failed for one opponent with enough permanents")
	}
}

func TestLifeConditionalEntersTappedReplacement(t *testing.T) {
	for _, test := range []struct {
		name       string
		life       int
		wantTapped bool
	}{
		{name: "below threshold", life: 9, wantTapped: true},
		{name: "at threshold", life: 10, wantTapped: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = test.life
			setSorcerySpeedTurn(g, game.Player1)
			cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:  "Life Land",
				Types: []types.Card{types.Land},
				ReplacementAbilities: []game.ReplacementAbility{
					game.EntersTappedIfReplacement("This land enters tapped unless you have 10 or more life.", &game.Condition{
						Negate:     true,
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 10}},
					}),
				},
			}})
			if !NewEngine(nil).applyPlayLand(g, game.Player1, cardID) {
				t.Fatal("applyPlayLand() = false")
			}
			if got := g.Battlefield[len(g.Battlefield)-1].Tapped; got != test.wantTapped {
				t.Fatalf("Tapped = %v, want %v", got, test.wantTapped)
			}
		})
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
		ManaAbilities: []game.ManaAbility{{
			Text: "{T}: Add {R}. Activate only if you control a Swamp or a Mountain.",
			ActivationCondition: opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{
						SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
					},
				}),
			}),
			AdditionalCosts: cost.Tap,
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.AddMana{
					Amount:    game.Fixed(1),
					ManaColor: mana.R,
				}}},
			}.Ability(),
		}}},
	}
}

func addBugenhagenLikePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Bugenhagen-like",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event: game.EventCardDrawn,
				},
				InterveningIf: "if you control a creature with power 7 or greater",
				InterveningCondition: opt.Val(game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 7,
							}),
						},
					}),
				}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}}},
	})
}

func angerLikeCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Anger-like",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
			{
				ZoneOfFunction: zone.Graveyard,
				Condition: opt.Val(game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{
							SubtypesAny: []types.Sub{types.Mountain},
						},
					}),
				}),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer: game.LayerAbility,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
					AddKeywords: []game.Keyword{game.Haste},
				}},
			},
		}},
	}
}

func cinderLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cinder-like",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
				Negate: true,
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Supertypes:    []types.Super{types.Basic},
					},
					MinCount: 2,
				}),
			}),
		}},
	}
}

// TestAggregateControllerLifeComparators verifies that the unified
// AggregateComparison representation evaluates every controller-life comparator
// (at least / at most / exactly) through the single aggregate path, that an
// empty Aggregates slice disables the predicate, and that multiple aggregate
// operands are ANDed together.
func TestAggregateControllerLifeComparators(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}

	life := func(n int) {
		g.Players[game.Player1].Life = n
	}
	cond := func(ops ...game.AggregateComparison) opt.V[game.Condition] {
		return opt.Val(game.Condition{Aggregates: ops})
	}
	atLeast := game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 10}
	atMost := game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 20}
	exactly := game.AggregateComparison{Aggregate: game.AggregateControllerLife, Op: compare.Equal, Value: 15}

	life(10)
	if !conditionSatisfied(g, ctx, cond(atLeast)) {
		t.Fatal("at-least failed at threshold")
	}
	life(9)
	if conditionSatisfied(g, ctx, cond(atLeast)) {
		t.Fatal("at-least passed below threshold")
	}
	life(20)
	if !conditionSatisfied(g, ctx, cond(atMost)) {
		t.Fatal("at-most failed at threshold")
	}
	life(21)
	if conditionSatisfied(g, ctx, cond(atMost)) {
		t.Fatal("at-most passed above threshold")
	}
	life(15)
	if !conditionSatisfied(g, ctx, cond(exactly)) {
		t.Fatal("exactly failed at exact life")
	}
	life(16)
	if conditionSatisfied(g, ctx, cond(exactly)) {
		t.Fatal("exactly passed off exact life")
	}

	// Empty Aggregates slice is treated as no predicate.
	emptyCond := game.Condition{}
	if !emptyCond.Empty() {
		t.Fatal("zero-value condition is not Empty")
	}
	withAgg := game.Condition{Aggregates: []game.AggregateComparison{atLeast}}
	if withAgg.Empty() {
		t.Fatal("condition with an aggregate reported Empty")
	}

	// Multiple aggregate operands are ANDed: 10 <= life <= 20.
	band := cond(atLeast, atMost)
	life(15)
	if !conditionSatisfied(g, ctx, band) {
		t.Fatal("AND band failed inside range")
	}
	life(9)
	if conditionSatisfied(g, ctx, band) {
		t.Fatal("AND band passed below range")
	}
	life(21)
	if conditionSatisfied(g, ctx, band) {
		t.Fatal("AND band passed above range")
	}
}

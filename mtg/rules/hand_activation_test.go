package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestHandActivatedAbilityDiscardsSourceAndUsesStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, handActivatedTestCard())
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	activation := action.ActivateAbility(cardID, 0, nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("legal actions do not include the hand activation")
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("applyAction() = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("mana cost did not tap the available Forest")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) ||
		!g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("activated card was not discarded from hand")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility ||
		obj.SourceCardID != cardID || obj.SourceZone != zone.Hand {
		t.Fatalf("stack object = %+v, want hand-sourced activated ability", obj)
	}
}

func TestHandActivatedAbilityUsesUnreducedCostWithNoMatchingPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, handActivatedReducedTestCard())
	g.Turn.PriorityPlayer = game.Player1
	activation := action.ActivateAbility(cardID, 0, nil, 0)

	addBasicLandPermanent(g, game.Player1, types.Forest)
	if actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("activation with cost {1}{G} was legal with only one mana and no legendary creatures")
	}

	addBasicLandPermanent(g, game.Player1, types.Forest)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("activation with cost {1}{G} was not legal with two mana and no legendary creatures")
	}
}

func TestHandActivatedAbilityReductionFloorsGenericAndPreservesColoredMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, handActivatedReducedTestCard())
	addLegendaryCreaturePermanent(g, game.Player1)
	addLegendaryCreaturePermanent(g, game.Player1)
	g.Turn.PriorityPlayer = game.Player1
	activation := action.ActivateAbility(cardID, 0, nil, 0)

	if actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("multiple reductions removed the colored {G} requirement")
	}
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("multiple reductions did not floor the generic cost at zero")
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("applyAction() = false, want reduced activation to succeed")
	}
	if !forest.Tapped {
		t.Fatal("reduced activation did not pay its colored {G} cost")
	}
}

func TestHandActivatedAbilityCannotActivateOutsideHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, handActivatedTestCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	g.Turn.PriorityPlayer = game.Player1
	activation := action.ActivateAbility(cardID, 0, nil, 0)

	if actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("hand-only activated ability was legal from the graveyard")
	}
	if engine.applyAction(g, game.Player1, activation) {
		t.Fatal("hand-only activated ability applied from the graveyard")
	}
}

func TestBattlefieldActivatedAbilityUsesSourceCostReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addReducedBattlefieldActivator(g, game.Player1)
	addLegendaryCreaturePermanent(g, game.Player1)
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.PriorityPlayer = game.Player1
	activation := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("battlefield activation did not apply its source-local cost reduction")
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("applyAction() = false, want reduced battlefield activation to succeed")
	}
	if !source.Tapped {
		t.Fatal("reduced battlefield activation did not pay its tap cost")
	}
}

func handActivatedTestCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Hand Activator",
			Types: []types.Card{types.Creature},
			ActivatedAbilities: []game.ActivatedAbility{{
				ManaCost:       opt.Val(cost.Mana{cost.G}),
				ZoneOfFunction: zone.Hand,
				AdditionalCosts: []cost.Additional{{
					Kind:   cost.AdditionalDiscard,
					Text:   "Discard this card",
					Amount: 1,
					Source: zone.Hand,
				}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Draw{
						Amount: game.Fixed(1),
						Player: game.ControllerReference(),
					},
				}}}.Ability(),
			}},
		},
	}
}

func handActivatedReducedTestCard() *game.CardDef {
	def := handActivatedTestCard()
	ability := &def.ActivatedAbilities[0]
	ability.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.G})
	ability.CostModifiers = []game.CostModifier{{
		Kind:               game.CostModifierAbility,
		PerObjectReduction: 1,
		CountSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Supertypes:    []types.Super{types.Legendary},
			Controller:    game.ControllerYou,
		},
	}}
	return def
}

func addLegendaryCreaturePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCreaturePermanent(g, controller)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("legendary creature card instance not found")
	}
	card.Def.Supertypes = []types.Super{types.Legendary}
	return permanent
}

func addReducedBattlefieldActivator(g *game.Game, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Reduced Activator",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{{
				ManaCost: opt.Val(cost.Mana{cost.O(5)}),
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalTap,
				}},
				CostModifiers: []game.CostModifier{{
					Kind:               game.CostModifierAbility,
					PerObjectReduction: 1,
					CountSelection: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Supertypes:    []types.Super{types.Legendary},
						Controller:    game.ControllerYou,
					},
				}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Draw{
						Amount: game.Fixed(1),
						Player: game.ControllerReference(),
					},
				}}}.Ability(),
			}},
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

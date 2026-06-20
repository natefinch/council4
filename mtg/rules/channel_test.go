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

func TestChannelActivationReducesGenericOnlyDiscardsSelfAndBouncesToOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	channelID := addCardToHand(g, game.Player1, otawaraTestCard())
	otherCardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Other Card"}})
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	for range 3 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:       "Legendary Creature",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
		}})
	}
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Artifact",
		Types: []types.Card{types.Artifact},
	}})
	land := addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(channelID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("reduced Channel activation was not legal with one Island")
	}
	landAct := action.ActivateAbility(channelID, 0, []game.Target{game.PermanentTarget(land.ObjectID)}, 0)
	if containsAction(engine.legalActions(g, game.Player1), landAct) {
		t.Fatal("Channel activation illegally targeted a land")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(Channel) = false, want true")
	}
	if !island.Tapped {
		t.Fatal("Channel did not preserve and pay the blue requirement")
	}
	if g.Players[game.Player1].Hand.Contains(channelID) || !g.Players[game.Player1].Graveyard.Contains(channelID) {
		t.Fatal("Channel source was not discarded from hand")
	}
	if !g.Players[game.Player1].Hand.Contains(otherCardID) {
		t.Fatal("Channel discarded another card instead of itself")
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("target moved before Channel resolved")
	}

	engine.resolveTopOfStack(g, nil)

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target remained on battlefield after Channel resolved")
	}
	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("target did not return to its owner's hand")
	}
}

func TestChannelReductionFloorsAtZeroAndNeverReducesColoredMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	channelID := addCardToHand(g, game.Player1, otawaraTestCard())
	for range 5 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:       "Legendary Creature",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
		}})
	}
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Creature",
		Types: []types.Card{types.Creature},
	}})
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(channelID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Channel was legal without blue mana; reduction touched colored requirement")
	}
	addBasicLandPermanent(g, game.Player1, types.Island)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Channel was not legal after generic reduction floored at zero with blue available")
	}
}

func TestBattlefieldActivationUsesSourceCostReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Reducing Permanent",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
			CostModifiers: []game.CostModifier{{
				Kind:               game.CostModifierAbility,
				PerObjectReduction: 1,
				CountSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Supertypes:    []types.Super{types.Legendary},
					Controller:    game.ControllerYou,
				},
			}},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
					},
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Bounce{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability(),
		}},
	}})
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	for range 3 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:       "Legendary Creature",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
		}})
	}
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Creature",
		Types: []types.Card{types.Creature},
	}})
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("battlefield activation did not use its source-scoped reduction")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(reduced battlefield ability) = false, want true")
	}
	if !island.Tapped {
		t.Fatal("reduced battlefield activation did not preserve and pay the blue requirement")
	}
}

func otawaraTestCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Otawara, Soaring City",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
			AdditionalCosts: []cost.Additional{{
				Kind:       cost.AdditionalDiscard,
				Amount:     1,
				Source:     zone.Hand,
				SourceSelf: true,
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
			ZoneOfFunction: zone.Hand,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{PermanentTypes: []types.Card{
						types.Artifact, types.Creature, types.Enchantment, types.Planeswalker,
					}},
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Bounce{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability(),
		}},
	}}
}

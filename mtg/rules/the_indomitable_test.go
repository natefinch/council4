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

func indomitableTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "The Indomitable",
		ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Artifact},
		Subtypes:   []types.Sub{types.Vehicle},
		Power:      opt.Val(game.PT{Value: 6}),
		Toughness:  opt.Val(game.PT{Value: 6}),
		StaticAbilities: []game.StaticAbility{{
			ZoneOfFunction: zone.Graveyard,
			Condition: opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{
						SubtypesAny: []types.Sub{types.Pirate, types.Vehicle},
						Tapped:      game.TriTrue,
					},
					MinCount: 3,
				}),
			}),
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Graveyard,
				AffectedSource: true,
			}},
		}},
	}}
}

func addTappedSubtypePermanent(g *game.Game, controller game.PlayerID, name string, cardTypes []types.Card, subtypes ...types.Sub) *game.Permanent {
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    cardTypes,
		Subtypes: subtypes,
	}})
	permanent.Tapped = true
	return permanent
}

func indomitableCastAction(cardID game.ObjectID) action.Action {
	return action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
}

func TestTheIndomitableThresholdAndSubtypeUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, indomitableTestDef())
	addTappedSubtypePermanent(g, game.Player1, "Pirate Vehicle", []types.Card{types.Artifact, types.Creature}, types.Pirate, types.Vehicle)
	addTappedSubtypePermanent(g, game.Player1, "Pirate", []types.Card{types.Creature}, types.Pirate)

	if canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("two qualifying permanents met the threshold; a Pirate Vehicle was double-counted")
	}
	addTappedSubtypePermanent(g, game.Player1, "Vehicle", []types.Card{types.Artifact}, types.Vehicle)
	if !canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("three mixed tapped Pirates and Vehicles did not meet the threshold")
	}
	addTappedSubtypePermanent(g, game.Player1, "Another Pirate", []types.Card{types.Creature}, types.Pirate)
	if !canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("four qualifying permanents did not remain above the threshold")
	}
}

func TestTheIndomitablePermissionTracksLivePermanentState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, indomitableTestDef())
	pirate := addTappedSubtypePermanent(g, game.Player1, "Pirate", []types.Card{types.Creature}, types.Pirate)
	vehicle := addTappedSubtypePermanent(g, game.Player1, "Vehicle", []types.Card{types.Artifact}, types.Vehicle)
	changing := addTappedSubtypePermanent(g, game.Player1, "Changing Permanent", []types.Card{types.Artifact})
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Island)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	act := indomitableCastAction(cardID)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untyped third permanent enabled the graveyard cast")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: changing.ObjectID,
		Layer:            game.LayerType,
		AddSubtypes:      []types.Sub{types.Vehicle},
	})
	if !permanentHasSubtype(g, changing, types.Vehicle) {
		t.Fatal("live type effect did not make the permanent a Vehicle")
	}
	condition := indomitableTestDef().StaticAbilities[0].Condition
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("live Vehicle type change did not satisfy the controls condition")
	}
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("live Vehicle type change did not enable the graveyard cast")
	}
	g.ContinuousEffects = nil
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("ending the Vehicle type change did not disable the graveyard cast")
	}

	g.CardInstances[changing.CardInstanceID].Def.Subtypes = []types.Sub{types.Vehicle}
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("three printed qualifying permanents did not enable the graveyard cast")
	}
	pirate.Tapped = false
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untapping a Pirate did not disable the graveyard cast")
	}
	pirate.Tapped = true
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("retapping a Pirate did not re-enable the graveyard cast")
	}

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: vehicle.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
	})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("an opponent-controlled Vehicle counted for the owner")
	}
	g.ContinuousEffects = nil
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("restoring Vehicle control did not re-enable the graveyard cast")
	}
}

func TestTheIndomitableUsesGraveyardOwnerAndScopesEachCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToGraveyard(g, game.Player2, indomitableTestDef())
	second := addCardToGraveyard(g, game.Player2, indomitableTestDef())
	for i := range 3 {
		addTappedSubtypePermanent(g, game.Player1, "Player 1 Pirate", []types.Card{types.Creature}, types.Pirate)
		if i < 2 {
			addTappedSubtypePermanent(g, game.Player2, "Player 2 Vehicle", []types.Card{types.Artifact}, types.Vehicle)
		}
	}

	if canCastFromZoneByRuleEffect(g, game.Player1, first, zone.Graveyard, game.FaceFront) {
		t.Fatal("nonowner could cast an opponent's graveyard card")
	}
	if canCastFromZoneByRuleEffect(g, game.Player2, first, zone.Graveyard, game.FaceFront) {
		t.Fatal("owner's condition incorrectly counted the opponent's permanents")
	}
	addTappedSubtypePermanent(g, game.Player2, "Player 2 Pirate", []types.Card{types.Creature}, types.Pirate)
	for _, cardID := range []game.ObjectID{first, second} {
		if !canCastFromZoneByRuleEffect(g, game.Player2, cardID, zone.Graveyard, game.FaceFront) {
			t.Fatalf("owner could not cast physical copy %v", cardID)
		}
	}

	g.Players[game.Player2].Graveyard.Remove(first)
	g.Players[game.Player2].Exile.Add(first)
	if canCastFromZoneByRuleEffect(g, game.Player2, first, zone.Graveyard, game.FaceFront) {
		t.Fatal("card retained its graveyard permission after leaving the graveyard")
	}
	if !canCastFromZoneByRuleEffect(g, game.Player2, second, zone.Graveyard, game.FaceFront) {
		t.Fatal("moving one physical copy removed the other copy's self-scoped permission")
	}
}

func TestTheIndomitableGraveyardCastUsesNormalTimingAndCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, indomitableTestDef())
	for range 3 {
		addTappedSubtypePermanent(g, game.Player1, "Pirate", []types.Card{types.Creature}, types.Pirate)
	}
	var islands []*game.Permanent
	for range 3 {
		islands = append(islands, addBasicLandPermanent(g, game.Player1, types.Island))
	}
	act := indomitableCastAction(cardID)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("three mana paid The Indomitable's {2}{U}{U} cost")
	}
	islands = append(islands, addBasicLandPermanent(g, game.Player1, types.Island))
	g.Turn.ActivePlayer = game.Player2
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard permission bypassed normal artifact timing")
	}
	g.Turn.ActivePlayer = game.Player1

	g.CommanderIDs[cardID] = true
	g.Players[game.Player1].CommanderCastCount = 2
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("payable graveyard cast was not proposed")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("payable graveyard cast failed")
	}
	for i, island := range islands {
		if !island.Tapped {
			t.Fatalf("Island %d was not tapped to pay the normal mana cost", i)
		}
	}
	if g.Players[game.Player1].CommanderCastCount != 2 {
		t.Fatal("graveyard cast incorrectly applied or incremented commander tax")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.SourceZone != zone.Graveyard || obj.Flashback {
		t.Fatalf("stack object = %+v, want ordinary graveyard cast", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved card remained in the graveyard")
	}
	if permanentByCardID(g, cardID) == nil {
		t.Fatal("resolved Vehicle did not enter the battlefield")
	}
}

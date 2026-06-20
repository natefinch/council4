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

func TestBoseijuHandActivationTargetLegality(t *testing.T) {
	tests := []struct {
		name       string
		controller game.PlayerID
		def        *game.CardDef
		wantLegal  bool
	}{
		{"opponent artifact", game.Player2, permanentDef("Artifact", nil, types.Artifact), true},
		{"opponent enchantment", game.Player2, permanentDef("Enchantment", nil, types.Enchantment), true},
		{"opponent nonbasic land", game.Player2, permanentDef("Nonbasic Land", nil, types.Land), true},
		{"opponent basic artifact land", game.Player2, permanentDef("Basic Artifact Land", []types.Super{types.Basic}, types.Artifact, types.Land), true},
		{"opponent basic land", game.Player2, permanentDef("Basic Land", []types.Super{types.Basic}, types.Land), false},
		{"opponent creature", game.Player2, permanentDef("Creature", nil, types.Creature), false},
		{"your artifact", game.Player1, permanentDef("Your Artifact", nil, types.Artifact), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, boseijuRuntimeTestCard())
			addLegendaryCreaturePermanent(g, game.Player1)
			addBasicLandPermanent(g, game.Player1, types.Forest)
			target := addCombatPermanent(g, tt.controller, tt.def)
			g.Turn.PriorityPlayer = game.Player1

			activation := action.ActivateAbility(cardID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
			if got := actionsContain(engine.legalActions(g, game.Player1), activation); got != tt.wantLegal {
				t.Fatalf("activation legality = %t, want %t", got, tt.wantLegal)
			}
			if !tt.wantLegal && engine.applyAction(g, game.Player1, activation) {
				t.Fatal("illegal Boseiju target was accepted")
			}
		})
	}
}

func TestBoseijuRemovalThenAffectedControllerSearch(t *testing.T) {
	for _, tt := range []struct {
		name        string
		agent       optionalSearchAgent
		wantFetched bool
	}{
		{"accepts", optionalSearchAgent{accept: true, wanted: "Typed Land"}, true},
		{"declines", optionalSearchAgent{accept: false, wanted: "Typed Land"}, false},
		{"fails to find", optionalSearchAgent{accept: true, wanted: "Not Present"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, boseijuRuntimeTestCard())
			addLegendaryCreaturePermanent(g, game.Player1)
			addBasicLandPermanent(g, game.Player1, types.Forest)
			target := addCombatPermanent(g, game.Player2, permanentDef("Target Artifact", nil, types.Artifact))
			typedLandID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name:     "Typed Land",
				Types:    []types.Card{types.Land},
				Subtypes: []types.Sub{types.Forest},
			}})
			g.Turn.PriorityPlayer = game.Player1
			activation := action.ActivateAbility(cardID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
			if !engine.applyAction(g, game.Player1, activation) {
				t.Fatal("Boseiju activation failed")
			}

			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: optionalSearchAgent{accept: !tt.agent.accept, wanted: "Typed Land"},
				game.Player2: tt.agent,
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("target artifact was not destroyed")
			}
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
				t.Fatal("destroyed target did not enter its owner's graveyard")
			}
			if g.Players[game.Player2].Library.Contains(typedLandID) == tt.wantFetched {
				t.Fatalf("typed land library membership does not match fetched=%t", tt.wantFetched)
			}
			var fetched *game.Permanent
			for _, permanent := range g.Battlefield {
				if permanent.CardInstanceID == typedLandID {
					fetched = permanent
					break
				}
			}
			if tt.wantFetched {
				if fetched == nil {
					t.Fatal("land with a basic land type was not fetched")
				}
				if fetched.Controller != game.Player2 || fetched.Tapped {
					t.Fatalf("fetched land = %+v, want untapped under affected player's control", fetched)
				}
			} else if fetched != nil {
				t.Fatal("declined or failed-to-find search still fetched a land")
			}
		})
	}
}

func boseijuRuntimeTestCard() *game.CardDef {
	searcher := game.ObjectControllerReference(game.TargetPermanentReference(0))
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Boseiju, Who Endures",
		Types:      []types.Card{types.Land},
		Supertypes: []types.Super{types.Legendary},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.G}),
			ZoneOfFunction: zone.Hand,
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalDiscard,
				Text:   "Discard this card",
				Amount: 1,
				Source: zone.Hand,
			}},
			CostModifiers: []game.CostModifier{{
				Kind:               game.CostModifierAbility,
				PerObjectReduction: 1,
				CountSelection: &game.Selection{
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
					Selection: opt.Val(game.Selection{
						Controller: game.ControllerOpponent,
						AnyOf: []game.Selection{
							{RequiredTypes: []types.Card{types.Artifact}},
							{RequiredTypes: []types.Card{types.Enchantment}},
							{RequiredTypes: []types.Card{types.Land}, ExcludedSupertype: types.Basic},
						},
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
					{
						Optional:      true,
						OptionalActor: opt.Val(searcher),
						Primitive: game.Search{
							Player: searcher,
							Amount: game.Fixed(1),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								CardType:    opt.Val(types.Land),
								SubtypesAny: []types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain, types.Forest},
							},
						},
					},
				},
			}.Ability(),
		}},
	}}
}

func permanentDef(name string, supertypes []types.Super, cardTypes ...types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Types:      cardTypes,
		Supertypes: supertypes,
	}}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCanoptekWraithCombatDamagePaymentChoiceAndSearch(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	wraith := addCombatPermanent(g, game.Player1, canoptekWraithTestDef())
	forest := basicNamedLand("Forest")
	addCombatPermanent(g, game.Player1, forest)
	addCombatPermanent(g, game.Player2, basicNamedLand("Island"))
	addCardToLibrary(g, game.Player1, forest)
	addCardToLibrary(g, game.Player1, forest)
	addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)

	dealPlayerDamage(g, wraith.CardInstanceID, wraith.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}, {0, 1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, wraith.ObjectID); ok {
		t.Fatal("paid source was not sacrificed")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana = %d, want 0 after payment", got)
	}
	forests := 0
	tappedForests := 0
	for _, permanent := range g.Battlefield {
		if permanentEffectiveName(g, permanent) != "Forest" || effectiveController(g, permanent) != game.Player1 {
			continue
		}
		forests++
		if permanent.Tapped {
			tappedForests++
		}
	}
	if forests != 3 || tappedForests != 2 {
		t.Fatalf("controlled Forests = %d (%d tapped), want 3 (2 tapped)", forests, tappedForests)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want unmatched Island remaining", got)
	}
}

func TestResolutionSourceSacrificeRejectsAbsentChangedControlAndInsufficientMana(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		mana    int
		prepare func(*game.Game, *game.Permanent)
	}{
		{name: "insufficient mana", mana: 2},
		{name: "changed control", mana: 3, prepare: func(_ *game.Game, source *game.Permanent) {
			source.Controller = game.Player2
		}},
		{name: "source absent", mana: 3, prepare: func(g *game.Game, source *game.Permanent) {
			movePermanentToZone(g, source, zone.Graveyard)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, canoptekWraithTestDef())
			other := addCombatCreaturePermanent(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.mana)
			if tc.prepare != nil {
				tc.prepare(g, source)
			}
			obj := &game.StackObject{
				Controller:   game.Player1,
				SourceID:     source.ObjectID,
				SourceCardID: source.CardInstanceID,
			}
			accepted, succeeded := engine.resolveResolutionPaymentValue(
				g,
				obj,
				&game.ResolutionPayment{
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
				},
				[game.NumPlayers]PlayerAgent{},
				&TurnLog{},
			)
			if accepted || succeeded {
				t.Fatalf("payment = (%v, %v), want unavailable", accepted, succeeded)
			}
			if _, ok := permanentByObjectID(g, other.ObjectID); !ok {
				t.Fatal("another creature was substituted for the source sacrifice")
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != tc.mana {
				t.Fatalf("mana = %d, want unchanged %d", got, tc.mana)
			}
		})
	}
}

func TestResolutionSourceSacrificeSupportsTokenSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token, ok := createTokenPermanent(g, game.Player1, canoptekWraithTestDef())
	if !ok {
		t.Fatal("creating token source failed")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	obj := &game.StackObject{Controller: game.Player1, SourceID: token.ObjectID}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	accepted, succeeded := engine.resolveResolutionPaymentValue(
		g,
		obj,
		&game.ResolutionPayment{
			ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
		},
		agents,
		&TurnLog{},
	)
	if !accepted || !succeeded {
		t.Fatalf("payment = (%v, %v), want token source payment", accepted, succeeded)
	}
	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("token source remained on the battlefield")
	}
}

func TestPermanentChoiceCapturesControlledTokenEffectiveName(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token, ok := createTokenPermanent(g, game.Player1, basicNamedLand("Forest"))
	if !ok {
		t.Fatal("creating land token failed")
	}
	addCombatPermanent(g, game.Player2, basicNamedLand("Island"))
	player := game.ControllerReference()
	_, values := resolutionChoiceOptions(g, &game.StackObject{Controller: game.Player1}, game.Player1, &game.ResolutionChoice{
		Kind:            game.ResolutionChoicePermanent,
		PlayerReference: &player,
		Selection:       &game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
	})
	if len(values) != 1 {
		t.Fatalf("values = %#v, want only the controlled land", values)
	}
	result := values[0]
	if result.PermanentID != token.ObjectID || result.CardID != 0 || result.CardName != "Forest" {
		t.Fatalf("result = %#v, want token identity and effective name", result)
	}
}

func canoptekWraithTestDef() *game.CardDef {
	player := game.ControllerReference()
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Canoptek Wraith",
		Types: []types.Card{types.Artifact, types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:               game.EventDamageDealt,
					Source:              game.TriggerSourceSelf,
					Subject:             game.TriggerSubjectDamageSource,
					RequireCombatDamage: true,
					DamageRecipient:     game.DamageRecipientPlayer,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.Pay{Payment: game.ResolutionPayment{
						ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
						AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
					}},
					PublishResult: "controller-paid",
				},
				{
					Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:            game.ResolutionChoicePermanent,
							PlayerReference: &player,
							Selection:       &game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
						},
						PublishChoice: game.ResolutionChosenPermanentChoiceKey,
					},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "controller-paid", Succeeded: game.TriTrue}),
				},
				{
					Primitive: game.Search{
						Player: game.ControllerReference(),
						Spec: game.SearchSpec{
							SourceZone:     zone.Library,
							Destination:    zone.Battlefield,
							Filter:         game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
							NameFromChoice: game.ResolutionChosenPermanentChoiceKey,
							EntersTapped:   true,
						},
						Amount: game.Fixed(2),
					},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "controller-paid", Succeeded: game.TriTrue}),
				},
			}}.Ability(),
		}},
	}}
}

func basicNamedLand(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
	}}
}

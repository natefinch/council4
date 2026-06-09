package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestSplitBothFacesAreLegalFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, splitCard())
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFace(cardID, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want CastSpellFace(front)", legal)
	}
	if !actionsContain(legal, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want CastSpellFace(alternate)", legal)
	}
}

func TestSplitFrontFaceCastAndResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, splitCard())
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceFront, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(front) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceFront {
		t.Fatalf("stack object = %+v, want front face", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("split front face did not go to graveyard")
	}
	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("split front face did not resolve its draw effect")
	}
}

func TestSplitAlternateFaceCastAndResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, splitCard())
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceAlternate {
		t.Fatalf("stack object = %+v, want alternate face", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("life = %d, want 42 from alternate half", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("split alternate face did not go to graveyard")
	}
}

func TestSplitOnlyFrontFaceLegalFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, splitCard())
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Graveyard,
	})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want graveyard CastSpellFace(front)", legal)
	}
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceAlternate, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, did not want graveyard CastSpellFace(alternate)", legal)
	}
}

func splitCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Split First",
			Types:    []types.Card{types.Sorcery},
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability()),
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name:     "Split Second",
			Types:    []types.Card{types.Sorcery},
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
			Colors:   []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
				}},
			}.Ability()),
		}),
	}
}

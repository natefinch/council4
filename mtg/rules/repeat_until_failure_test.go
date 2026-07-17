package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	testRepeatPutResult      game.ResultKey = "test-repeat-put"
	testRepeatContinueResult game.ResultKey = "test-repeat-continue"
)

type repeatLandChoiceAgent struct {
	mayAnswers []bool
	landNames  []string
	mayIndex   int
	landIndex  int
}

func (*repeatLandChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *repeatLandChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		accept := a.mayIndex < len(a.mayAnswers) && a.mayAnswers[a.mayIndex]
		a.mayIndex++
		if accept {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceResolution:
		if a.landIndex >= len(a.landNames) {
			return request.DefaultSelection
		}
		name := a.landNames[a.landIndex]
		a.landIndex++
		for _, option := range request.Options {
			if option.Label == name {
				return []int{option.Index}
			}
		}
	default:
		return request.DefaultSelection
	}
	return request.DefaultSelection
}

func repeatPutLandAndDraw() game.RepeatProcess {
	return game.RepeatProcess{
		Body: game.Mode{Sequence: []game.Instruction{
			{
				Primitive: game.PutFromHandChoice(
					game.ControllerReference(),
					game.Selection{RequiredTypes: []types.Card{types.Land}},
					game.Fixed(1),
					true,
					false,
					false,
				),
				Optional:      true,
				PublishResult: testRepeatPutResult,
			},
			{
				Primitive: game.Draw{
					Player: game.ControllerReference(),
					Amount: game.Fixed(1),
				},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       testRepeatPutResult,
					Succeeded: game.TriTrue,
				}),
				PublishResult: testRepeatContinueResult,
			},
		}}.Ability(),
		ContinueResult: testRepeatContinueResult,
	}
}

func resolveRepeatPutLand(
	g *game.Game,
	obj *game.StackObject,
	agent *repeatLandChoiceAgent,
) {
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	engine.resolveInstructionWithChoices(
		g,
		obj,
		&game.Instruction{Primitive: repeatPutLandAndDraw()},
		agents,
		&TurnLog{},
	)
}

func TestRepeatUntilFailureDrawsLandAndUsesCurrentHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Source", Types: []types.Card{types.Creature},
	}})
	obj := triggeredObjFor(source)
	movePermanentToZone(g, source, zone.Graveyard)

	first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land},
	}})
	nonland := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bolt", Types: []types.Card{types.Instant},
	}})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land},
	}})

	agent := &repeatLandChoiceAgent{
		mayAnswers: []bool{true, true},
		landNames:  []string{"Forest", "Island"},
	}
	resolveRepeatPutLand(g, obj, agent)

	for _, cardID := range []id.ID{first, drawn} {
		permanent, ok := reanimatedPermanent(g, cardID)
		if !ok {
			t.Fatalf("land %d did not enter the battlefield", cardID)
		}
		if !permanent.Tapped {
			t.Fatalf("land %d entered untapped", cardID)
		}
	}
	if !g.Players[game.Player1].Hand.Contains(nonland) {
		t.Fatal("nonland card was exposed or removed from hand")
	}
	if !g.FailedDraws[game.Player1] {
		t.Fatal("empty-library draw did not fail and stop the process")
	}
	if agent.mayIndex != 2 {
		t.Fatalf("optional prompts = %d, want 2", agent.mayIndex)
	}
	entered, draws := 0, 0
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield &&
			(event.CardID == first || event.CardID == drawn) {
			entered++
		}
		if event.Kind == game.EventCardDrawn && event.CardID == drawn {
			draws++
		}
	}
	if entered != 2 || draws != 1 {
		t.Fatalf("land-entry events = %d, draw events = %d; want 2 and 1", entered, draws)
	}
}

func TestRepeatUntilFailureStopsWithNoLandInHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Source", Types: []types.Card{types.Creature},
	}})
	nonland := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bolt", Types: []types.Card{types.Instant},
	}})
	addCardToLibraryNamed(g, game.Player1, "Undrawn")
	agent := &repeatLandChoiceAgent{mayAnswers: []bool{true}}

	resolveRepeatPutLand(g, triggeredObjFor(source), agent)

	if !g.Players[game.Player1].Hand.Contains(nonland) ||
		g.Players[game.Player1].Hand.Size() != 1 ||
		agent.mayIndex != 1 {
		t.Fatal("no-land iteration moved or drew a card, or did not terminate")
	}
}

func TestRepeatUntilFailureCanDeclineAnyIteration(t *testing.T) {
	tests := []struct {
		name      string
		answers   []bool
		wantLands int
		wantDraws int
	}{
		{name: "initially", answers: []bool{false}, wantLands: 0, wantDraws: 0},
		{name: "after success", answers: []bool{true, false}, wantLands: 1, wantDraws: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Source", Types: []types.Card{types.Creature},
			}})
			for _, name := range []string{"Forest", "Mountain"} {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: name, Types: []types.Card{types.Land},
				}})
			}
			addCardToLibraryNamed(g, game.Player1, "Draw")
			addCardToLibraryNamed(g, game.Player1, "Draw 2")

			resolveRepeatPutLand(g, triggeredObjFor(source), &repeatLandChoiceAgent{
				mayAnswers: tt.answers,
				landNames:  []string{"Mountain"},
			})

			lands := 0
			for _, permanent := range g.Battlefield {
				if permanentHasType(g, permanent, types.Land) {
					lands++
				}
			}
			if lands != tt.wantLands {
				t.Fatalf("lands on battlefield = %d, want %d", lands, tt.wantLands)
			}
			draws := 0
			for _, event := range g.Events {
				if event.Kind == game.EventCardDrawn && event.Player == game.Player1 {
					draws++
				}
			}
			if draws != tt.wantDraws {
				t.Fatalf("draw events = %d, want %d", draws, tt.wantDraws)
			}
		})
	}
}

func TestRepeatUntilFailureStopsWhenLandMoveIsReplaced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Source", Types: []types.Card{types.Creature},
	}})
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land},
	}})
	addCardToLibraryNamed(g, game.Player1, "Undrawn")
	resolveInstruction(NewEngine(nil), g, triggeredObjFor(source), game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:   "put into graveyard instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Hand,
			MatchToZone:   true,
			ToZone:        zone.Battlefield,
			ReplaceToZone: zone.Graveyard,
		},
	}, nil)

	resolveRepeatPutLand(g, triggeredObjFor(source), &repeatLandChoiceAgent{
		mayAnswers: []bool{true},
		landNames:  []string{"Forest"},
	})

	if !g.Players[game.Player1].Graveyard.Contains(land) {
		t.Fatal("replacement did not redirect the land")
	}
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("a card was drawn after the land failed to enter")
	}
}

func TestRepeatUntilFailureUsesTriggerControllerAfterControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Source", Types: []types.Card{types.Creature},
	}})
	obj := triggeredObjFor(source)
	source.Controller = game.Player2
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land},
	}})
	addCardToLibraryNamed(g, game.Player1, "Draw")

	resolveRepeatPutLand(g, obj, &repeatLandChoiceAgent{
		mayAnswers: []bool{true, false},
		landNames:  []string{"Forest"},
	})

	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("trigger controller hand size = %d, want one drawn card", g.Players[game.Player1].Hand.Size())
	}
	if g.Players[game.Player2].Hand.Size() != 0 {
		t.Fatal("current source controller incorrectly made the choices")
	}
}

func TestRepeatUntilFailureStopsForEliminatedController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Source", Types: []types.Card{types.Creature},
	}})
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land},
	}})
	g.Players[game.Player1].Eliminated = true
	agent := &repeatLandChoiceAgent{mayAnswers: []bool{true}, landNames: []string{"Forest"}}

	resolveRepeatPutLand(g, triggeredObjFor(source), agent)

	if !g.Players[game.Player1].Hand.Contains(land) || agent.mayIndex != 0 {
		t.Fatal("process resolved for an eliminated controller")
	}
}

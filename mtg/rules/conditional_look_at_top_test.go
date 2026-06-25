package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conditionalLookAtTopRevealSequence is the runtime shape cardgen emits for the
// "look at the top card of your library; if it's a land card, you may reveal it
// and put it into your hand; if you don't, you may put it into your graveyard"
// building block (Sarinth Steelseeker, Traveling Botanist, Territory Culler).
func conditionalLookAtTopRevealSequence(cardType types.Card) []game.Instruction {
	looked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"}
	return []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: "looked",
			},
		},
		{
			Primitive: game.ConditionalDestinationPlace{
				Card:     looked,
				FromZone: zone.Library,
				CardCondition: opt.Val(game.CardSelection{
					Card:      looked,
					Selection: game.Selection{RequiredTypesAny: []types.Card{cardType}},
				}),
				Then:         zone.Hand,
				ThenReveal:   true,
				Else:         zone.Graveyard,
				ElseOptional: true,
			},
		},
	}
}

// TestConditionalLookAtTopMatchingGoesToHandNotGraveyard proves that when the
// looked-at card matches the gate type and the controller accepts the reveal,
// the card moves to hand and the graveyard fallback (gated on the reveal not
// happening) does not fire.
func TestConditionalLookAtTopMatchingGoesToHandNotGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Forest",
		Types: []types.Card{types.Land},
	}})
	addInstructionSpellToStack(g, conditionalLookAtTopRevealSequence(types.Land))
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(topID) {
		t.Fatal("matching land was not moved to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(topID) {
		t.Fatal("matching land that went to hand also went to graveyard")
	}
	if g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("matching land remained in library")
	}
}

// TestConditionalLookAtTopNonMatchingOffersGraveyard proves that when the
// looked-at card does not match the gate type, the reveal publishes a
// not-accepted result and the graveyard fallback fires so the controller may
// bin the card. It validates the new TriFalse-gated graveyard instruction.
func TestConditionalLookAtTopNonMatchingOffersGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Relic",
		Types: []types.Card{types.Artifact},
	}})
	addInstructionSpellToStack(g, conditionalLookAtTopRevealSequence(types.Land))
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(topID) {
		t.Fatal("non-matching card was moved to hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(topID) {
		t.Fatal("non-matching card was not offered to the graveyard fallback")
	}
	if g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("non-matching card binned to graveyard remained in library")
	}
}

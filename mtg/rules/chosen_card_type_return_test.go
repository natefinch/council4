package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestMassReturnFromGraveyardMatchesResolutionChosenCardType(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addCfzGraveyardCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Artifact",
		Types: []types.Card{types.Artifact},
	}})
	creature := addCfzGraveyardCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Creature",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
	}
	const key = game.ChoiceKey("chosen-permanent-type")
	rememberResolutionChoice(obj, string(key), game.ResolutionChoiceResult{
		Kind:     game.ResolutionChoiceCardType,
		CardType: types.Artifact,
	})
	resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	result := handleMassReturnFromGraveyard(resolver, game.MassReturnFromGraveyard{
		Player:      game.ControllerReference(),
		Selection:   game.Selection{ChosenCardTypeFrom: key},
		Destination: zone.Hand,
	})
	if !result.succeeded ||
		!g.Players[game.Player1].Hand.Contains(artifact) ||
		!g.Players[game.Player1].Graveyard.Contains(creature) {
		t.Fatalf("result = %#v, artifact in hand = %v, creature in graveyard = %v",
			result,
			g.Players[game.Player1].Hand.Contains(artifact),
			g.Players[game.Player1].Graveyard.Contains(creature))
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerCouncilOfFourOrdinalPlayerTurnTriggers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "The Council of Four",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Noble",
		ManaCost:   "{3}{W}{U}",
		Power:      new("0"),
		Toughness:  new("8"),
		OracleText: "Whenever a player draws their second card during their turn, you draw a card.\nWhenever a player casts their second spell during their turn, you create a 2/2 white Knight creature token.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	draw := face.TriggeredAbilities[0].Trigger.Pattern
	if draw.Event != game.EventCardDrawn ||
		draw.Player != game.TriggerPlayerAny ||
		draw.PlayerEventOrdinalThisTurn != 2 ||
		draw.CastDuringTurn != game.TriggerTurnEventPlayer {
		t.Fatalf("draw pattern = %#v", draw)
	}
	cast := face.TriggeredAbilities[1].Trigger.Pattern
	if cast.Event != game.EventSpellCast ||
		cast.Controller != game.TriggerControllerAny ||
		cast.PlayerEventOrdinalThisTurn != 2 ||
		cast.CastDuringTurn != game.TriggerTurnEventPlayer {
		t.Fatalf("cast pattern = %#v", cast)
	}
}

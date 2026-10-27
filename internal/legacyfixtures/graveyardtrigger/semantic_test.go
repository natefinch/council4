// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules"
)

type passAndAcceptAgent struct{}

func (passAndAcceptAgent) ChooseAction(rules.PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (passAndAcceptAgent) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	return request.DefaultSelection
}

func TestRenderedGraveyardSourceActuallyTriggers(t *testing.T) {
	source := RTGraveyardSource()
	ordinary := RTDefaultSource()
	recurring := []*game.CardDef{source, SqueeGoblinNabob(), FlamewakePhoenix(), LightningPhoenix(), SilversmoteGhoul()}
	for _, def := range append(recurring, ordinary) {
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("generated definition invalid: %+v", issues)
		}
	}
	for _, def := range recurring {
		found := false
		for _, ability := range def.TriggeredAbilities {
			found = found || ability.ZoneOfFunction == zone.Graveyard
		}
		if !found {
			t.Fatalf("generated graveyard metadata lost for %s", def.Name)
		}
	}
	if ordinary.TriggeredAbilities[0].ZoneOfFunction != zone.None {
		t.Fatal("generated default-zone metadata changed")
	}
	filler := &game.CardDef{CardFace: game.CardFace{Name: "Filler"}}
	deck := make([]*game.CardDef, 12)
	for i := range deck {
		deck[i] = filler
	}
	engine := rules.NewEngine(nil)
	g := engine.NewGoldfishGame(game.PlayerConfig{Deck: deck})
	card := g.IDGen.Next()
	g.CardInstances[card] = &game.CardInstance{ID: card, Def: source, Owner: game.Player1, ZoneVersion: 7}
	g.Players[game.Player1].Graveyard.Add(card)
	other := g.IDGen.Next()
	g.CardInstances[other] = &game.CardInstance{ID: other, Def: ordinary, Owner: game.Player1}
	g.Players[game.Player1].Graveyard.Add(other)
	engine.RunGoldfish(g, passAndAcceptAgent{}, 1)
	if !g.Players[game.Player1].Hand.Contains(card) ||
		!g.Players[game.Player1].Graveyard.Contains(other) || g.Players[game.Player1].Life != 41 {
		t.Fatal("generated source failed runtime return, or a default-zone card triggered from graveyard")
	}
	captured := 0
	for _, event := range g.Events {
		for _, trigger := range event.TriggeredAbilities {
			if trigger.SourceCardID != card {
				continue
			}
			captured++
			if trigger.SourceZone != zone.Graveyard || trigger.SourceZoneVersion != 7 ||
				trigger.Controller != game.Player1 {
				t.Fatal("generated source lost event-time incarnation/owner snapshot")
			}
		}
	}
	if captured != 1 {
		t.Fatalf("captured %d source triggers, want exactly one", captured)
	}
}

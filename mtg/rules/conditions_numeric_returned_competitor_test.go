package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledSpellConditionDoesNotAcquireReturnedPermanent(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Spell Not Returned Creature", Layout: "normal",
		TypeLine: "Creature", ManaCost: "{9}", Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever you cast an instant or sorcery spell, return target creature card from your graveyard to the battlefield. You gain 1 life. If that spell has mana value 5 or greater, you gain 2 life."})
	for _, eventValue := range []int{4, 5, 6} {
		for _, returnedValue := range []int{2, 8} {
			t.Run(fmt.Sprintf("spell=%d/returned=%d", eventValue, returnedValue), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player1, def)
				cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Returned Competitor", Types: []types.Card{types.Creature},
					ManaCost: opt.Val(cost.Mana{cost.O(returnedValue)}),
					Power:    opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
				}})
				spell := spellWithManaValue(g, game.Player1, cost.Mana{cost.O(eventValue)})
				obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
					SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
					HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventSpellCast, StackObjectID: spell.ID,
						Controller: game.Player1, ManaValue: stackObjectKnownManaValue(g, spell)},
					Targets: []game.Target{currentCardTarget(t, g, cardID)},
				}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				want := 41
				if eventValue >= 5 {
					want += 2
				}
				if g.Players[game.Player1].Life != want {
					t.Fatalf("life=%d, want %d; compared returned/source permanent instead of spell", g.Players[game.Player1].Life, want)
				}
				if g.Players[game.Player1].Graveyard.Contains(cardID) {
					t.Fatal("return producer did not execute")
				}
			})
		}
	}
}

package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledBlinkConditionAndConsequenceShareNewObject(t *testing.T) {
	t.Parallel()
	body := "Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it."
	for _, modal := range []bool{false, true} {
		for _, matches := range []bool{false, true} {
			t.Run(fmt.Sprintf("modal=%t/matches=%t", modal, matches), func(t *testing.T) {
				t.Parallel()
				text := body
				if modal {
					text = "Choose one —\n• " + body + "\n• Draw a card."
				}
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Published Untap", Layout: "normal", TypeLine: "Instant", OracleText: text,
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				targetDef := &game.CardDef{CardFace: game.CardFace{Name: "Subject", Types: []types.Card{types.Creature}}}
				if matches {
					targetDef.Subtypes = []types.Sub{types.Elf}
				}
				old := addCombatPermanent(g, game.Player1, targetDef)
				decoy := addCombatPermanent(g, game.Player1, targetDef)
				decoy.Tapped = true
				sourceID := addCardInstance(g, game.Player1, def)
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
					SourceID: sourceID, SourceCardID: sourceID, ChosenModes: []int{0}, TargetCounts: []int{1},
					Targets: []game.Target{game.PermanentTarget(old.ObjectID)},
				}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				var returned *game.Permanent
				for _, permanent := range g.Battlefield {
					if permanent.CardInstanceID == old.CardInstanceID {
						returned = permanent
					}
				}
				if returned == nil || returned.ObjectID == old.ObjectID || returned.Tapped == matches {
					t.Fatalf("new permanent=%+v, want tapped=%t", returned, !matches)
				}
				if !decoy.Tapped {
					t.Fatal("returned-subject consequence untapped a different permanent")
				}
			})
		}
	}
}

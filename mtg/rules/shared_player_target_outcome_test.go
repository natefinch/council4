package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledSharedPlayerRecipientTargetOutcomeVerbs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		action              string
		life, hand, library int
	}{
		{"they gain 2 life.", 42, 0, 3},
		{"they lose 2 life.", 38, 0, 3},
		{"they draw two cards.", 40, 2, 1},
		{"they mill two cards.", 40, 0, 1},
	} {
		for _, prefix := range []string{"", "Target player draws a card. ", "Exile target creature card from a graveyard. "} {
			t.Run(prefix+tt.action, func(t *testing.T) {
				t.Parallel()
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Target Outcome Recipient", Layout: "normal", TypeLine: "Instant",
					OracleText: prefix + "Target opponent sacrifices a creature of their choice. If that player does, " + tt.action,
				})
				for _, protected := range []bool{false, true} {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					for player := range g.Players {
						for range 3 {
							addCardToLibrary(g, game.PlayerID(player), evidenceCard("Outcome library", 1))
						}
					}
					creature := addCombatCreaturePermanent(g, game.Player2)
					creature.Controller = game.Player3
					if protected {
						cantSacrificeControlNotOwnEnchantment(g, game.Player3)
					}
					obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
					switch prefix {
					case "Target player draws a card. ":
						obj.Targets = append(obj.Targets, game.PlayerTarget(game.Player4))
						obj.TargetCounts = append(obj.TargetCounts, 1)
					case "Exile target creature card from a graveyard. ":
						prior := addCardToGraveyard(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
							Name: "Earlier card target", Types: []types.Card{types.Creature},
						}})
						obj.Targets = append(obj.Targets, game.CardTarget(prior))
						obj.TargetCounts = append(obj.TargetCounts, 1)
					default:
					}
					obj.Targets = append(obj.Targets, game.PlayerTarget(game.Player3))
					obj.TargetCounts = append(obj.TargetCounts, 1)
					result, available := NewEngine(nil).resolveAbilityContentReceipt(g, obj, def.SpellAbility.Val,
						[game.NumPlayers]PlayerAgent{}, &TurnLog{}, "if-you-do")
					if !available || result.Succeeded == protected {
						t.Fatalf("protected=%v: actual sacrifice result=%+v available=%v", protected, result, available)
					}
					for player, state := range g.Players {
						life, hand, library := 40, 0, 3
						if game.PlayerID(player) == game.Player3 && !protected {
							life, hand, library = tt.life, tt.hand, tt.library
						}
						if game.PlayerID(player) == game.Player4 && prefix == "Target player draws a card. " {
							hand, library = 1, 2
						}
						if state.Life != life || state.Hand.Size() != hand || state.Library.Size() != library {
							t.Fatalf("protected=%v player %d: life/hand/library=%d/%d/%d, want %d/%d/%d",
								protected, player, state.Life, state.Hand.Size(), state.Library.Size(), life, hand, library)
						}
					}
				}
			})
		}
	}
}

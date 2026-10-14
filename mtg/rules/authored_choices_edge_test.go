package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAuthoredGildedDrakeNoTargetOrUnavailableExchange(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"no target", "source left", "same controller", "foreign owner"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			def := authored.GildedDrake.New()
			drake := addCombatPermanent(g, game.Player1, def)
			target := addCombatCreaturePermanent(g, game.Player2)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
				SourceID: drake.ObjectID, SourceCardID: drake.CardInstanceID, Controller: game.Player1,
				InlineTrigger: &def.TriggeredAbilities[0],
				Targets:       []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
			}
			switch scenario {
			case "no target":
				obj.Targets, obj.TargetCounts = nil, []int{0}
			case "source left":
				movePermanentToZone(g, drake, zone.Exile)
			case "same controller":
				drake.Controller = game.Player2
			case "foreign owner":
				drake.Owner = game.Player3
				card, _ := g.GetCardInstance(drake.CardInstanceID)
				card.Owner = game.Player3
			default:
				t.Fatal("unknown fixture")
			}
			g.Stack.Push(obj)
			engine.resolveTopOfStack(g, &TurnLog{})
			switch scenario {
			case "no target":
				if !g.Players[game.Player1].Graveyard.Contains(drake.CardInstanceID) {
					t.Fatal("zero-target trigger did not sacrifice Drake")
				}
			case "source left":
				if effectiveController(g, target) != game.Player2 {
					t.Fatal("half an exchange occurred after the source left")
				}
			case "same controller":
				if effectiveController(g, target) != game.Player2 || permanentByCardID(g, drake.CardInstanceID) == nil {
					t.Fatal("same-controller exchange succeeded or the ability sacrificed another player's Drake")
				}
			case "foreign owner":
				if effectiveController(g, target) != game.Player1 || effectiveController(g, drake) != game.Player2 ||
					drake.Owner != game.Player3 {
					t.Fatal("control exchange confused ownership with control")
				}
			default:
				t.Fatal("unknown fixture")
			}
		})
	}
}

func TestAuthoredGiftsUngivenZeroThroughFourFinds(t *testing.T) {
	t.Parallel()
	for count := range 5 {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			for index := range count {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + index))}})
			}
			addImplementationSpellToStack(g, game.Player1, authored.GiftsUngiven.New(), []game.Target{game.PlayerTarget(game.Player2)})
			engine.resolveTopOfStack(g, &TurnLog{})
			wantHand := max(0, count-2)
			// Gifts itself also enters its owner's graveyard.
			if g.Players[game.Player1].Hand.Size() != wantHand ||
				g.Players[game.Player1].Graveyard.Size() != min(count, 2)+1 {
				t.Fatal("fewer finds did not put every found card in the required partition")
			}
		})
	}
}

func TestAuthoredGiftsUngivenFailToFindAndIllegalOpponent(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"decline", "duplicate names", "eliminated opponent"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			for range 3 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Same"}})
			}
			addImplementationSpellToStack(g, game.Player1, authored.GiftsUngiven.New(), []game.Target{game.PlayerTarget(game.Player2)})
			agents := [game.NumPlayers]PlayerAgent{}
			switch scenario {
			case "decline":
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{}}}
			case "duplicate names":
			case "eliminated opponent":
				g.Players[game.Player2].Eliminated = true
			default:
				t.Fatal("unknown fixture")
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)
			wantLibrary := 3
			if scenario == "duplicate names" {
				wantLibrary = 2
			}
			if g.Players[game.Player1].Library.Size() != wantLibrary || g.Players[game.Player1].Hand.Size() != 0 {
				t.Fatal("failed/limited search invented cards or selected duplicate names")
			}
			if scenario == "eliminated opponent" && len(log.Choices) != 0 {
				t.Fatal("all-illegal Gifts target still searched or chose a partition")
			}
		})
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Independent review against f8978af4da2765535365a9c0c5825688018349a8:
// exact printings and empty rulings responses listed in buyback_value_test.go
// were opened independently; corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Independently read verified CR 2026-09-25, SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca:
// 118.1-3/7-8, 121, 307, 400.7, 601.2, 608.2, 701.8, 702.27a,
// 707.10, 614.1/4-11, 616.1. These probes isolate copied controller-relative
// instructions and replaced empty draws, rather than just rerunning author tests.

func TestBuybackValueIndependentWhispersCopiedController(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	one := addCardToLibrary(g, game.Player1, vanillaCreature("Original top", 2, 2))
	two := addCardToLibrary(g, game.Player2, vanillaCreature("Copier top", 2, 2))
	tc := buybackValueCase(t, authored.WhispersOfTheMuse.OracleID)
	source, original := castProductionBuyback(t, g, engine, registry, tc,
		game.Player1, game.Target{}, true, tc.manaToPayWithBuyback())
	reiterate := buybackValueCase(t, authored.Reiterate.OracleID)
	castProductionBuyback(t, g, engine, registry, reiterate, game.Player2,
		game.StackObjectTarget(original.ID), false, reiterate.manaToPay)
	engine.resolveTopOfStack(g, &TurnLog{})
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.Controller != game.Player2 || !copy.BuybackPaid {
		t.Fatal("real Reiterate did not create the paid Whispers copy for Player Two")
	}
	for name, branch := range map[string]*game.Game{"live": g, "clone": g.Clone()} {
		engine.resolveTopOfStack(branch, &TurnLog{})
		if !branch.Players[game.Player2].Hand.Contains(two) ||
			!branch.Players[game.Player1].Library.Contains(one) ||
			branch.Players[game.Player1].Hand.Contains(source) ||
			branch.Players[game.Player1].Graveyard.Contains(source) ||
			branch.Stack.Size() != 1 {
			t.Fatalf("%s: copy drew for the original controller or moved its physical source", name)
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		if !branch.Players[game.Player1].Hand.Contains(one) ||
			!branch.Players[game.Player1].Hand.Contains(source) ||
			branch.Players[game.Player2].Hand.Contains(source) ||
			countEventsOfKind(branch.Events, game.EventCardDrawn) != 2 {
			t.Fatalf("%s: original did not independently draw and buy back", name)
		}
	}
}

func TestBuybackValueIndependentBurialCopiedGraveyard(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	for _, retarget := range []bool{false, true} {
		t.Run(map[bool]string{false: "keep now-illegal target", true: "retarget positive control"}[retarget], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			one := addCardToGraveyard(g, game.Player1, vanillaCreature("Original creature", 2, 2))
			two := addCardToGraveyard(g, game.Player2, vanillaCreature("Copier creature", 2, 2))
			tc := buybackValueCase(t, authored.DisturbedBurial.OracleID)
			source, original := castProductionBuyback(t, g, engine, registry, tc,
				game.Player1, currentCardTarget(t, g, one), true, tc.manaToPayWithBuyback())
			reiterate := buybackValueCase(t, authored.Reiterate.OracleID)
			castProductionBuyback(t, g, engine, registry, reiterate, game.Player2,
				game.StackObjectTarget(original.ID), false, reiterate.manaToPay)
			choices := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceTarget {
					if retarget {
						return chooseAuthoredTargets(t, request, currentCardTarget(t, g, two))
					}
					for _, option := range request.Options {
						if option.Label == "Keep current targets" {
							return []int{option.Index}
						}
					}
					t.Fatal("copy must be permitted to retain an illegal target")
				}
				return request.DefaultSelection
			}}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: choices}, &TurnLog{})
			copy, ok := g.Stack.Peek()
			want := one
			if retarget {
				want = two
			}
			if !ok || !copy.Copy || copy.Controller != game.Player2 ||
				len(copy.Targets) != 1 || copy.Targets[0].CardID != want {
				t.Fatal("real copied Burial did not retain the chosen target")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Graveyard.Contains(one) ||
				g.Players[game.Player2].Hand.Contains(two) != retarget ||
				g.Players[game.Player2].Graveyard.Contains(two) == retarget ||
				g.Players[game.Player1].Hand.Contains(source) || g.Stack.Size() != 1 {
				t.Fatal("copy did not recheck 'your graveyard' relative to its own controller")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(one) ||
				!g.Players[game.Player1].Hand.Contains(source) {
				t.Fatal("copy's target decision changed the original Burial")
			}
		})
	}
}

func TestBuybackValueIndependentWhispersReplacedEmptyDraw(t *testing.T) {
	t.Parallel()
	for _, accept := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline loses positive control", true: "replace empty draw"}[accept], func(t *testing.T) {
			g, engine, registry := abundanceGame(t)
			if g.Players[game.Player1].Library.Size() != 0 {
				t.Fatal("fixture must start with an empty library")
			}
			tc := buybackValueCase(t, authored.WhispersOfTheMuse.OracleID)
			source, _ := castProductionBuyback(t, g, engine, registry, tc,
				game.Player1, game.Target{}, true, tc.manaToPayWithBuyback())
			choices := &abundanceChoices{accepts: []bool{accept}, kinds: []int{1}}
			policies := [game.NumPlayers]PlayerAgent{game.Player1: choices}
			engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(source) ||
				g.Players[game.Player1].Eliminated ||
				g.FailedDraws[game.Player1] == accept ||
				countEventsOfKind(g.Events, game.EventCardDrawn) != 0 {
				t.Fatal("empty draw replacement, successful buyback, or pre-priority loss timing is wrong")
			}
			result := engine.Simulator().ResolvePriority(g, simPassPolicies())
			if result.Players[game.Player1].Eliminated == accept {
				t.Fatal("next real priority checkpoint did not distinguish replaced from failed draw")
			}
		})
	}
}

func TestBuybackValueIndependentPulseOwnerAfterControlChange(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Stolen artifact", Types: []types.Card{types.Artifact},
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerControl,
		NewController: opt.Val(game.Player3), Duration: game.DurationPermanent,
	})
	if effectiveController(g, target) != game.Player3 {
		t.Fatal("fixture did not transfer artifact control while retaining its owner")
	}
	tc := buybackValueCase(t, authored.ShatteringPulse.OracleID)
	source, _ := castProductionBuyback(t, g, engine, registry, tc,
		game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, target.CardInstanceID) != nil ||
		!g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
		g.Players[game.Player3].Graveyard.Contains(target.CardInstanceID) ||
		!g.Players[game.Player1].Hand.Contains(source) {
		t.Fatal("Pulse destroyed into the controller's graveyard instead of its owner's, or failed buyback")
	}
}

package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestControllerObservationResultAndChoicesCompose(t *testing.T) {
	const observation = "You may look at the top card of target opponent's library. "
	const consequence = "If you do, you may put that card on the bottom of that player's library."
	for _, prefix := range []string{"", "Target player gains 1 life. ", "Exile target card from your graveyard. ", "Counter target spell. "} {
		for _, rider := range []string{"", "You gain 1 life. "} {
			t.Run(prefix+rider, func(t *testing.T) {
				card := &ScryfallCard{Name: "Controller Observation", Layout: "normal", TypeLine: "Sorcery",
					OracleText: prefix + observation + rider + consequence}
				defs, diagnostics, err := CompileCardDefs(card)
				if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
					t.Fatalf("compile: %v %v", err, diagnostics)
				}
				mode := defs[0].SpellAbility.Val.Modes[0]
				index := 0
				if prefix != "" {
					index = 1
				}
				look := captureTestPrimitive[game.LookAtLibraryTop](t, mode.Sequence[index].Primitive)
				move := mode.Sequence[len(mode.Sequence)-1]
				if look.Player != game.TargetPlayerReference(index) ||
					!mode.Sequence[index].Optional || mode.Sequence[index].OptionalActor.Exists ||
					mode.Sequence[index].OptionalActorGroup.Exists || !move.Optional || move.OptionalActor.Exists ||
					!move.ResultGate.Exists || move.ResultGate.Val.Key != mode.Sequence[index].PublishResult ||
					look.PublishLinked == "" {
					t.Fatalf("look=%+v optional=%v actor=%+v group=%+v publisher=%q moveOptional=%v moveActor=%+v gate=%+v",
						look, mode.Sequence[index].Optional, mode.Sequence[index].OptionalActor,
						mode.Sequence[index].OptionalActorGroup, mode.Sequence[index].PublishResult,
						move.Optional, move.OptionalActor, move.ResultGate)
				}
				assertCardPaths(t, card, fmt.Sprintf("Targets[%d].Allow = game.TargetAllowPlayer", index),
					"Primitive.(game.MoveCard).Card.Kind = game.CardReferenceLinked", "DestinationBottom = true")
				if rider != "" && mode.Sequence[index+1].ResultGate.Exists {
					t.Fatal("independent controller rider acquired observation gate")
				}
			})
		}
	}
}

func TestPrecognitionPinnedObservationComposition(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Precognition", Layout: "normal", TypeLine: "Enchantment", ManaCost: "{4}{U}",
		OracleText: "At the beginning of your upkeep, you may look at the top card of target opponent's library. If you do, you may put that card on the bottom of that player's library.",
	}, "Primitive.(game.LookAtLibraryTop).Player.kind = game.PlayerReferenceTargetPlayer",
		"Primitive.(game.MoveCard).Card.Kind = game.CardReferenceLinked", "DestinationBottom = true",
		"Sequence[0].Optional = true", "Sequence[1].Optional = true")
}

func TestControllerObservationPreservesModeAndOwnershipBoundaries(t *testing.T) {
	const body = "You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library."
	for _, text := range []string{
		"Choose one \u2014\n\u2022 " + body + "\n\u2022 You gain 2 life.",
		"If you control an artifact, " + body,
	} {
		assertCardPaths(t, &ScryfallCard{Name: "Observation Scope", Layout: "normal", TypeLine: "Sorcery", OracleText: text},
			"Primitive.(game.LookAtLibraryTop)", "Primitive.(game.MoveCard).Card.Kind = game.CardReferenceLinked",
			"Sequence[0].PublishResult", "Sequence[2].ResultGate.Val.Key")
	}
	for _, text := range []string{
		"Repeat the following process two times. " + body,
		"You may look at the top two cards of target opponent's library. If you do, you may put that card on the bottom of that player's library.",
		"Target player may look at the top card of their library. If they do, you gain 1 life.",
		"You may look at the top card of target opponent's library. Reveal the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Unproved Observation Scope", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
	}
}

func TestControllerObservationCannotBorrowForeignActorOutcome(t *testing.T) {
	for _, actor := range []string{"they do", "that player does"} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Foreign Observer", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "You may look at the top card of target opponent's library. If " + actor + ", you gain 1 life."})
	}
	for _, text := range []string{
		"Target player may draw a card. If that player does, you gain 1 life.",
		"Target player may discard a card. If that player does, you gain 1 life.",
		"Target player gains 1 life. That player may sacrifice a creature. If that player does, you gain 1 life.",
	} {
		compilation, diagnostics := compileTestOracle(text, parser.Context{InstantOrSorcery: true}, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		body := compilation.Abilities[0].Content
		found := false
		for i, effect := range body.Effects {
			if body.ControllerObservesTargetLibraryTop(effect) {
				t.Fatalf("player action became an observation: %s", text)
			}
			if !effect.Optional {
				continue
			}
			found = true
			sequence := []game.Instruction{{Optional: true}}
			if !bindSequenceOptionalActor(body, i, sequence, map[shared.Span]int{body.Targets[0].Span: 0}) ||
				!sequence[0].OptionalActor.Exists || sequence[0].OptionalActor.Val != game.TargetPlayerReference(0) {
				t.Fatal("genuine player action lost the affected player's choice")
			}
		}
		if !found {
			t.Fatal("missing player optional action")
		}
	}
}

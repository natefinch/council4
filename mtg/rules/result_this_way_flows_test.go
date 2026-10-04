package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestResultThisWayElseBranchSemantics(t *testing.T) {
	for _, otherwise := range []bool{false, true} {
		clause, nonmatchingDraws := "If you don't", 0
		if otherwise {
			clause, nonmatchingDraws = "Otherwise", 1
		}
		t.Run(clause, func(t *testing.T) {
			defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
				Name: "Else Probe", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "You may discard a card. If a land card is discarded this way, draw two cards. " + clause + ", draw a card.",
			})
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("compile = %v, %v", diagnostics, err)
			}
			for _, test := range []struct {
				name    string
				card    types.Card
				decline bool
				draws   int
			}{
				{"matching", types.Land, false, 2},
				{"nonmatching successful discard", types.Creature, false, nonmatchingDraws},
				{"failed discard", "", false, 1},
				{"declined discard", types.Land, true, 1},
			} {
				t.Run(test.name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					if test.card != "" {
						addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{test.card}}})
					}
					for range 3 {
						addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
					}
					agents := [game.NumPlayers]PlayerAgent{}
					if test.decline {
						agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}}}
					}
					NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
						defs[0].SpellAbility.Val, agents, &TurnLog{})
					if got := 3 - g.Players[game.Player1].Library.Size(); got != test.draws {
						t.Fatalf("drawn = %d, want %d", got, test.draws)
					}
				})
			}
		})
	}
}

func TestResultThisWayColoredPermanentSelection(t *testing.T) {
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Colored Result Probe", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may sacrifice a creature. If a creature with one or more colors was sacrificed this way, you gain 3 life.",
	})
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("compile = %v, %v", diagnostics, err)
	}
	for _, colored := range []bool{false, true} {
		t.Run(map[bool]string{false: "colorless", true: "colored"}[colored], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}}
			if colored {
				def.Colors = []color.Color{color.Red}
			}
			addCombatPermanent(g, game.Player1, def)
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
				defs[0].SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 40
			if colored {
				want = 43
			}
			if got := g.Players[game.Player1].Life; got != want {
				t.Fatalf("life = %d, want %d", got, want)
			}
		})
	}
}

func TestResultObjectGroupEnvelopesPublishActualMembers(t *testing.T) {
	for _, envelope := range []string{"each player", "optional group", "tempting offer"} {
		t.Run(envelope, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			addCardToHand(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			instruction := &game.Instruction{
				Primitive:     game.Discard{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(1)},
				PublishResult: "discard",
			}
			if envelope == "each player" {
				instruction.ForEachPlayerGroup = opt.Val(game.OpponentsReference())
			} else {
				instruction.Optional = true
				instruction.OptionalActorGroup = opt.Val(game.OpponentsReference())
				instruction.TemptingOffer = envelope == "tempting offer"
			}
			gate := game.InstructionResultGate{Key: "discard", Succeeded: game.TriTrue,
				ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}})}
			if err := game.ValidateInstructionSequence([]game.Instruction{
				*instruction, {Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)}, ResultGate: opt.Val(gate)},
			}); err != nil {
				t.Fatal(err)
			}
			obj := &game.StackObject{Controller: game.Player1}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !instructionResultGateSatisfied(g, obj, gate) {
				t.Fatal("group envelope did not publish the actual discarded land")
			}
			engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if instructionResultGateSatisfied(g, obj, gate) || len(obj.ResolutionResultObjects["discard"]) != 0 {
				t.Fatal("failed repeated group publication retained old result objects")
			}
		})
	}
}

func TestNegatedResultObjectGateRequiresPublication(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if instructionResultGateSatisfied(g, &game.StackObject{}, game.InstructionResultGate{
		Key: "missing", Succeeded: game.TriTrue, ObjectSelection: opt.Val(game.Selection{}), Negate: true,
	}) {
		t.Fatal("negation admitted a missing result publication")
	}
}

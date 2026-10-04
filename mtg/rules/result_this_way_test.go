package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestResultThisWayDiscardSelection(t *testing.T) {
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Lord Windgrace", Layout: "normal",
		TypeLine: "Legendary Planeswalker — Windgrace", ManaCost: "{3}{B}{R}{G}",
		Loyalty:    new("4"),
		OracleText: "+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile = %v, %v, %v", defs, diagnostics, err)
	}
	content := defs[0].LoyaltyAbilities[0].Content
	for _, test := range []struct {
		name      string
		cardType  types.Card
		wantDraws int
	}{
		{"land", types.Land, 2},
		{"nonland", types.Creature, 1},
		{"empty hand", "", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if test.cardType != "" {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Discarded", Types: []types.Card{test.cardType},
				}})
			}
			for range 3 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			}
			obj := &game.StackObject{Controller: game.Player1}
			log := &TurnLog{}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, log)
			if got := g.Players[game.Player1].Hand.Size(); got != test.wantDraws {
				t.Fatalf("drawn cards = %d, want %d (interposed draw is unconditional)", got, test.wantDraws)
			}
		})
	}
}

func TestResultThisWayProducerSelections(t *testing.T) {
	for _, test := range []struct {
		name, text                 string
		from                       zone.Type
		matchType, otherType       types.Card
		matchSubtype, otherSubtype types.Sub
	}{
		{"discard", "You may discard a card. If a land card is discarded this way, you gain 3 life.", zone.Hand, types.Land, types.Creature, "", ""},
		{"mill", "You may mill a card. If a creature card was milled this way, you gain 3 life.", zone.Library, types.Creature, types.Land, "", ""},
		{"destroy", "You may destroy target artifact or land. If an artifact was destroyed this way, you gain 3 life.", zone.Battlefield, types.Artifact, types.Land, "", ""},
		{"sacrifice", "You may sacrifice a creature. If a Goblin was sacrificed this way, you gain 3 life.", zone.Battlefield, types.Creature, types.Creature, "Goblin", "Elf"},
		{"exile", "You may exile target creature. If a Pirate was exiled this way, you gain 3 life.", zone.Battlefield, types.Creature, types.Creature, "Pirate", "Elf"},
		{"exile card", "You may exile a card from a graveyard. If a creature card was exiled this way, you gain 3 life.", zone.Graveyard, types.Creature, types.Land, "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
				Name: "Result Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: test.text,
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile = %v, %v, %v", defs, diagnostics, err)
			}
			content := defs[0].SpellAbility.Val
			for _, branch := range []struct {
				name                      string
				matching, absent, decline bool
				wantLife                  int
			}{
				{"matching", true, false, false, 43},
				{"nonmatching", false, false, false, 40},
				{"no actual result", true, true, false, 40},
				{"declined", true, false, true, 40},
			} {
				t.Run(branch.name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					obj := &game.StackObject{Controller: game.Player1}
					cardType, subtype := test.otherType, test.otherSubtype
					if branch.matching {
						cardType, subtype = test.matchType, test.matchSubtype
					}
					def := &game.CardDef{CardFace: game.CardFace{Name: "Candidate", Types: []types.Card{cardType}}}
					if subtype != "" {
						def.Subtypes = []types.Sub{subtype}
					}
					if !branch.absent {
						switch test.from {
						case zone.Hand:
							addCardToHand(g, game.Player1, def)
						case zone.Library:
							addCardToLibrary(g, game.Player1, def)
						case zone.Graveyard:
							addCardToGraveyard(g, game.Player1, def)
						case zone.Battlefield:
							permanent := addCombatPermanent(g, game.Player1, def)
							obj.Targets = []game.Target{game.PermanentTarget(permanent.ObjectID)}
						default:
							t.Fatalf("unsupported test source zone %v", test.from)
						}
					}
					agents := [game.NumPlayers]PlayerAgent{}
					if branch.decline {
						agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}}}
					}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
					if got := g.Players[game.Player1].Life; got != branch.wantLife {
						t.Fatalf("life = %d, want %d", got, branch.wantLife)
					}
					result := obj.ResolutionResults["if-you-do"]
					if result.Accepted != !branch.decline {
						t.Fatalf("accepted = %t, want %t", result.Accepted, !branch.decline)
					}
					if !branch.absent && !branch.decline && !result.Succeeded {
						t.Fatal("noun mismatch must not change the producer's success")
					}
					if branch.absent || branch.decline {
						if len(obj.ResolutionResultObjects["if-you-do"]) != 0 {
							t.Fatal("failed or declined producer published actual result members")
						}
					}
					NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
						Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "if-you-do", Succeeded: game.TriTrue}),
					}, agents, &TurnLog{})
					wantLife := branch.wantLife
					if result.Succeeded {
						wantLife++
					}
					if got := g.Players[game.Player1].Life; got != wantLife {
						t.Fatalf("unfiltered result consumer life = %d, want %d", got, wantLife)
					}
				})
			}
		})
	}
}

func TestResultObjectSelectionUsesActualGroupMembers(t *testing.T) {
	for _, matching := range []bool{false, true} {
		t.Run(map[bool]string{false: "no matching member", true: "later matching member"}[matching], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			secondType := types.Creature
			if matching {
				secondType = types.Land
			}
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{secondType}}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First", Types: []types.Card{types.Creature}}})
			addInstructionSpellToStack(g, []game.Instruction{
				{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(2), Destination: zone.Graveyard}, PublishResult: "mill"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "mill", Succeeded: game.TriTrue,
						ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}})})},
			})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			want := 40
			if matching {
				want = 43
			}
			if got := g.Players[game.Player1].Life; got != want {
				t.Fatalf("life = %d, want %d", got, want)
			}
		})
	}
}

package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestResultThisWayBlinkKeepsDepartureCharacteristics(t *testing.T) {
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Siren's Ruse", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile target creature you control, then return that card to the battlefield under its owner's control. If a Pirate was exiled this way, draw a card.",
	})
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("compile = %v, %v", diagnostics, err)
	}
	for _, pirate := range []bool{false, true} {
		t.Run(map[bool]string{false: "non-Pirate", true: "temporary Pirate"}[pirate], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{"Elf"},
			}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}
			engine := NewEngine(nil)
			agents := [game.NumPlayers]PlayerAgent{}
			log := &TurnLog{}
			if pirate {
				engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ApplyContinuous{
					Object:            opt.Val(game.TargetPermanentReference(0)),
					ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerType, AddSubtypes: []types.Sub{"Pirate"}}},
				}}, agents, log)
				if !permanentHasSubtype(g, permanent, types.Pirate) {
					t.Fatal("setup did not make the target a Pirate")
				}
			}
			sequence := defs[0].SpellAbility.Val.Modes[0].Sequence
			// Resolve the spell's instructions inside its own local-product frame,
			// observing and mutating state between them during that lifetime.
			restore := enterLocalProductFrame(g, obj, sequence)
			resolver := newEffectResolver(engine, g, obj, agents, log)
			resolver.resolveInstruction(&sequence[0])
			saved := obj.ResolutionResultObjects["if-you-do"]
			if len(saved) != 1 || slices.Contains(saved[0].Subtypes, types.Pirate) != pirate {
				t.Fatalf("departure snapshots = %#v", saved)
			}
			resolver.resolveInstruction(&sequence[1])
			if len(g.Battlefield) != 1 || g.Battlefield[0].ObjectID == permanent.ObjectID {
				t.Fatal("linked return did not produce a new permanent")
			}
			if permanentHasSubtype(g, g.Battlefield[0], types.Pirate) {
				t.Fatal("temporary Pirate subtype survived blink")
			}
			// Subsequent LKI updates cannot change the result's frozen values.
			lki := g.LastKnownInformation[permanent.ObjectID]
			lki.Subtypes[0] = types.Pirate
			if pirate {
				lki.Subtypes[0] = "Elf"
			}
			g.LastKnownInformation[permanent.ObjectID] = lki
			resolver.resolveInstruction(&sequence[2])
			want := 0
			if pirate {
				want = 1
			}
			if got := g.Players[game.Player1].Hand.Size(); got != want {
				t.Fatalf("drawn = %d, want %d", got, want)
			}
			restore()
			assertResultKeysCleaned(t, obj, "if-you-do")
		})
	}
}

func TestResultObjectDestroyPreventionAndInvalidTarget(t *testing.T) {
	for _, test := range []struct {
		name            string
		keyword         game.Keyword
		shield, invalid bool
		wantLife        int
	}{
		{"destroyed", game.KeywordNone, false, false, 43},
		{"indestructible", game.Indestructible, false, false, 40},
		{"shield counter", game.KeywordNone, true, false, 40},
		{"invalid target", game.KeywordNone, false, true, 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatCreaturePermanent(g, game.Player1, test.keyword)
			if test.shield {
				target.Counters.Add(counter.Shield, 1)
			}
			if test.invalid {
				movePermanentToZone(g, target, zone.Exile)
			}
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			instructions := []game.Instruction{
				{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}, PublishResult: "destroy"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "destroy", Succeeded: game.TriTrue,
						ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})})},
			}
			engine := NewEngine(nil)
			for i := range instructions {
				engine.resolveInstructionWithChoices(g, obj, &instructions[i], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			}
			if got := g.Players[game.Player1].Life; got != test.wantLife {
				t.Fatalf("life = %d, want %d", got, test.wantLife)
			}
			if test.wantLife == 40 && len(obj.ResolutionResultObjects["destroy"]) != 0 {
				t.Fatal("attempted or stale target was published as an actual result")
			}
		})
	}
}

func TestResultObjectExileRequiresActualDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commander: &game.CardDef{CardFace: game.CardFace{
			Name: "Pirate Commander", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Pirate},
		}}},
	})
	cardID := g.Players[game.Player1].CommanderInstanceID
	g.Players[game.Player1].CommandZone.Remove(cardID)
	card, _ := g.GetCardInstance(cardID)
	permanent, ok := createCardPermanentFaceWithOptions(NewEngine(nil), g, card, game.Player1, zone.Command, game.FaceFront, nil, permanentCreationOptions{}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok {
		t.Fatal("commander setup failed")
	}
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile}, PublishResult: "exile",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player1].CommandZone.Contains(cardID) {
		t.Fatal("setup did not redirect commander to command zone")
	}
	if len(obj.ResolutionResultObjects["exile"]) != 0 {
		t.Fatal("redirected exile published an exiled Pirate")
	}
	if !obj.ResolutionResults["exile"].Succeeded {
		t.Fatal("result filtering changed generic move success")
	}
}

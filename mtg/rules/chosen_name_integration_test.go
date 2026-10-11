package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledChosenNameSuffixAndMultipleRealCopies(t *testing.T) {
	for _, tc := range chosenNameRuntimeCompositions[2:] {
		t.Run(tc.text, func(t *testing.T) {
			def := chosenNameRuntimeDef(t, tc.text)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
			addCombatPermanent(g, game.Player1, basicNamedLand("Island"))
			for range tc.searches {
				addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
				addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
			}
			source := addCardToLibrary(g, game.Player4, def)
			g.Players[game.Player4].Library.Remove(source)
			original := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
				SourceID: source, SourceCardID: source, Controller: game.Player1}
			g.Stack.Push(original)
			engine := NewEngine(nil)
			addEffectSpellToStack(g, game.Player1,
				game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
				[]game.Target{game.StackObjectTarget(original.ID)})
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			stackCopy, ok := g.Stack.Peek()
			if !ok || !stackCopy.Copy || stackCopy.ID == original.ID {
				t.Fatal("no real stack copy")
			}
			cloned := g.Clone()
			for _, name := range []string{"Island", "Forest"} {
				agent := &chosenNameAgent{name: name, count: 1}
				engine.resolveTopOfStackWithChoices(cloned,
					[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
				if len(agent.requests) != tc.searches+1 {
					t.Fatal("copy/original failed to choose its own name for each consumer")
				}
			}
			if cloned.Players[game.Player1].Library.Size() != 0 ||
				g.Players[game.Player1].Library.Size() != 2*tc.searches ||
				g.Stack.Size() != 2 || len(stackCopy.ResolutionChoices) != 0 || len(original.ResolutionChoices) != 0 {
				t.Fatal("public copied composition borrowed a choice or mutated the original clone")
			}
		})
	}
}

func TestCompiledChosenNameSuffixAndMultipleCaptureAcrossReentry(t *testing.T) {
	for _, tc := range chosenNameRuntimeCompositions[2:] {
		t.Run(tc.text, func(t *testing.T) {
			def := chosenNameRuntimeDef(t, tc.text)
			sequence := def.SpellAbility.Val.Modes[0].Sequence
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			land := addCombatPermanent(g, game.Player2, basicNamedLand("Forest"))
			land.Controller = game.Player1
			for range 2 * tc.searches {
				addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			}
			island := addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			g.Stack.Push(obj)
			engine := NewEngine(nil)
			agent := &chosenNameAgent{name: "Forest", count: 2}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			engine.resolveInstructionSequence(g, obj, sequence[:1], agents, &TurnLog{})
			choice, ok := linkedResolutionChoice(obj, string(game.ResolutionChosenPermanentChoiceKey))
			if !ok || choice.PermanentID != land.ObjectID || choice.CardName != "Forest" {
				t.Fatal("original public choice did not capture the controlled permanent")
			}
			cloned := g.Clone()
			clonedObject, ok := cloned.Stack.Peek()
			if !ok {
				t.Fatal("clone lost resolving object")
			}
			clonedLand, ok := permanentByObjectID(cloned, land.ObjectID)
			if !ok {
				t.Fatal("clone lost chosen permanent")
			}
			movePermanentToZone(cloned, clonedLand, zone.Graveyard)
			card := cloned.CardInstances[land.CardInstanceID]
			card.Def = basicNamedLand("Island")
			reentered, entered := newEffectResolver(engine, cloned, clonedObject, agents, &TurnLog{}).putResolvedCardOnBattlefieldValue(
				card, zone.Graveyard, game.Player3, nil, permanentCreationOptions{})
			if !entered || reentered.ObjectID == land.ObjectID ||
				reentered.CardInstanceID != land.CardInstanceID || reentered.Owner != game.Player2 {
				t.Fatal("chosen card did not reenter with a new object and its original physical owner")
			}
			engine.resolveInstructionSequence(cloned, clonedObject, sequence[1:], agents, &TurnLog{})
			if cloned.Players[game.Player1].Library.Size() != 1 ||
				!cloned.Players[game.Player1].Library.Contains(island) ||
				g.Players[game.Player1].Library.Size() != 2*tc.searches+1 {
				t.Fatal("public consumers substituted current name/controller or mutated the original game")
			}
			finalChoice, ok := linkedResolutionChoice(clonedObject, string(game.ResolutionChosenPermanentChoiceKey))
			if !ok || finalChoice != choice {
				t.Fatal("a later consumer replaced the exact original captured choice")
			}
		})
	}
}

func TestCompiledChosenNameCanoptekCopyPaysAndGatesEveryConsequence(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Chosen payment copy", Layout: "normal", TypeLine: "Artifact Creature",
		OracleText: "When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then " + chosenNameRuntimeSearch,
		Power:      new("2"), Toughness: new("1"),
	})
	sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatal("payment duplicated or consequence lost during projection")
	}
	if _, ok := sequence[0].Primitive.(game.Pay); !ok || sequence[0].PublishResult == "" {
		t.Fatal("missing original payment publisher")
	}
	for _, instruction := range sequence[1:] {
		if !instruction.ResultGate.Exists || instruction.ResultGate.Val.Key != sequence[0].PublishResult ||
			instruction.ResultGate.Val.Succeeded != game.TriTrue {
			t.Fatal("projected consequence lost the actual payment gate")
		}
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
	for range 2 {
		addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	}
	island := addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	engine := NewEngine(nil)
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("compiled trigger was not queued")
	}
	original, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("missing original trigger")
	}
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	stackCopy, ok := g.Stack.Peek()
	if !ok || !stackCopy.Copy {
		t.Fatal("no actual copied trigger")
	}
	for _, state := range []*game.Game{g, g.Clone()} {
		log := &TurnLog{}
		engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}, {0, 1}}},
		}, log)
		if state.Players[game.Player1].Library.Size() != 1 || !state.Players[game.Player1].Library.Contains(island) {
			t.Fatal("paid copied trigger lost the exact chosen-name search")
		}
		if _, present := permanentByObjectID(state, source.ObjectID); present {
			t.Fatal("copy failed to sacrifice the exact original source")
		}
		afterCopy := len(log.Choices)
		engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}, {0}}},
		}, log)
		if len(log.Choices) != afterCopy || state.Players[game.Player1].Library.Size() != 1 ||
			state.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("source-unavailable original reused copied payment or choice state")
		}
	}
}

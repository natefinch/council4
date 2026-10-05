package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func captureCopyToken(t *testing.T, g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
	permanent := addCombatPermanent(g, controller, def)
	permanent.Token = true
	permanent.TokenDef = def
	return permanent
}

func TestFixedPhaseEachTokenCopyPublishesActualBatch(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t,
		"For each token you control, create a token that's a copy of that permanent. "+
			"You gain 1 life. Exile them at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	first := captureCopyToken(t, g, game.Player1, "First")
	second := captureCopyToken(t, g, game.Player1, "Second")
	decoy := captureCopyToken(t, g, game.Player2, "Decoy")
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want one", len(g.DelayedTriggers))
	}
	if captured := g.DelayedTriggers[0].CapturedObjectIDs; len(captured) != 4 {
		t.Fatalf("captured copies = %v, want four replacement-modified copies", captured)
	}
	for _, objectID := range g.DelayedTriggers[0].CapturedObjectIDs {
		if objectID == first.ObjectID || objectID == second.ObjectID || objectID == decoy.ObjectID {
			t.Fatal("publication included an original or unrelated token")
		}
	}
	sequence[0].Optional = true
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
	if len(g.DelayedTriggers) != 2 || len(g.DelayedTriggers[1].CapturedObjectIDs) != 0 {
		t.Fatal("declined copy producer reused a previous batch")
	}
	sequence[0].Optional = false
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 3 || len(g.DelayedTriggers[2].CapturedObjectIDs) != 12 {
		t.Fatal("repeated resolution did not capture all fresh copies of the six current tokens")
	}
	for _, old := range g.DelayedTriggers[0].CapturedObjectIDs {
		for _, fresh := range g.DelayedTriggers[2].CapturedObjectIDs {
			if old == fresh {
				t.Fatal("later publication included an old product instead of its new copy")
			}
		}
	}
	clone := g.Clone()
	engine.runEndingPhase(clone, [game.NumPlayers]PlayerAgent{})
	for _, trigger := range g.DelayedTriggers {
		for _, objectID := range trigger.CapturedObjectIDs {
			if _, live := permanentByObjectID(clone, objectID); live {
				t.Fatal("an actual copied token escaped delayed disposal")
			}
			if _, live := permanentByObjectID(g, objectID); !live {
				t.Fatal("clone disposal changed the original game")
			}
		}
	}
	for _, original := range []*game.Permanent{first, second, decoy} {
		if _, live := permanentByObjectID(clone, original.ObjectID); !live {
			t.Fatal("delayed disposal removed an original or unrelated token")
		}
	}
}

func TestFixedPhaseSpecializedCopyPublishers(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		source game.TokenCopySource
	}{
		{"trigger-batch", game.TokenCopySourceChosenFromTriggerBatch},
		{"populate", game.TokenCopySourceChosenControlledCreatureToken},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			first := captureCopyToken(t, g, game.Player1, "First")
			second := captureCopyToken(t, g, game.Player1, "Second")
			captureCopyToken(t, g, game.Player2, "Decoy")
			addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
			obj := &game.StackObject{
				Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: twilightDivinerTrigger(), HasTriggerEvent: true,
				TriggerEvent: game.Event{
					Kind: game.EventPermanentEnteredBattlefield, PermanentID: first.ObjectID, Controller: game.Player1,
				},
			}
			sequence := []game.Instruction{
				{Primitive: game.CreateToken{
					Amount: game.Fixed(1), PublishLinked: "copy-products",
					Source: game.TokenCopyOf(game.TokenCopySpec{Source: test.source}),
				}},
				{Primitive: game.GainLife{Amount: game.Fixed(1)}},
				{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
					Timing:              game.DelayedAtBeginningOfNextEndStep,
					CapturedObjectGroup: opt.Val(game.LinkedObjectReference("copy-products")),
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.MovePermanent{Group: game.CapturedObjectsGroup(), Destination: zone.Exile},
					}}}.Ability(),
				}}},
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: fixedSelectionAgent{selection: []int{0}}}
			engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
			if len(g.DelayedTriggers) != 1 {
				t.Fatalf("delayed triggers = %d, want one", len(g.DelayedTriggers))
			}
			if captured := g.DelayedTriggers[0].CapturedObjectIDs; len(captured) != 2 {
				t.Fatalf("captured copies = %v, want two replacement-modified copies", captured)
			}
			captured := append([]game.ObjectID(nil), g.DelayedTriggers[0].CapturedObjectIDs...)
			sequence[0].Condition = opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{ControllerHandEmpty: true, Negate: true}),
			})
			engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
			if len(g.DelayedTriggers) != 2 || len(g.DelayedTriggers[1].CapturedObjectIDs) != 0 {
				t.Fatal("false-gated specialized copy producer reused a previous batch")
			}
			engine.runEndingPhase(g, agents)
			for _, objectID := range captured {
				if _, live := permanentByObjectID(g, objectID); live {
					t.Fatal("a specialized copy escaped delayed disposal")
				}
			}
			first.Token, second.Token = false, false
			obj.HasTriggerEvent = false
			sequence[0].Condition = opt.V[game.EffectCondition]{}
			engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
			if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].CapturedObjectIDs) != 0 {
				t.Fatal("unavailable copy subject reused a stale publication")
			}
		})
	}
}

func TestFixedPhaseEachTokenCopyCapturesReplacementOutputs(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name        string
		replacement game.TokenCreationReplacementSpec
		wantNames   map[string]int
	}{
		{
			name:        "partial-prevention",
			replacement: game.TokenCreationReplacementSpec{Multiplier: 1, Addend: -1, Types: []types.Card{types.Artifact}},
			wantNames:   map[string]int{"Second": 1},
		},
		{
			name:        "zero-output",
			replacement: game.TokenCreationReplacementSpec{Multiplier: 1, Addend: -1, Types: []types.Card{types.Creature}},
			wantNames:   map[string]int{},
		},
		{
			name: "identity-substitution",
			replacement: game.TokenCreationReplacementSpec{
				Multiplier: 1, Types: []types.Card{types.Artifact},
				ReplaceDef: &game.CardDef{CardFace: game.CardFace{Name: "Substitute", Types: []types.Card{types.Artifact}}},
			},
			wantNames: map[string]int{"Substitute": 1, "Second": 1},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t,
				"For each token you control, create a token that's a copy of that permanent. "+
					"You gain 1 life. Exile them at the beginning of the next end step.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			first := captureCopyToken(t, g, game.Player1, "First")
			first.TokenDef.Types = append(first.TokenDef.Types, types.Artifact)
			captureCopyToken(t, g, game.Player1, "Second")
			spec := test.replacement
			spec.Filter = game.TriggerControllerYou
			replacement := &game.CardDef{CardFace: game.CardFace{
				Name: "Creation Replacement", Types: []types.Card{types.Enchantment},
				ReplacementAbilities: []game.ReplacementAbility{
					game.TokenCreationReplacementFiltered("Replace artifact copies", &spec),
				},
			}}
			if issues := game.ValidateCardDef(replacement); len(issues) != 0 {
				t.Fatalf("replacement fixture: %v", issues)
			}
			addReplacementPermanent(t, g, game.Player1, replacement)
			obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
			key := linkedObjectSourceKey(g, obj, "sequence-effect-0-product")
			rememberLinkedObject(g, key, permanentLinkedObjectRef(first))
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if len(g.DelayedTriggers) != 1 {
				t.Fatalf("delayed triggers = %d, want one", len(g.DelayedTriggers))
			}
			names := make(map[string]int)
			for _, objectID := range g.DelayedTriggers[0].CapturedObjectIDs {
				permanent, ok := permanentByObjectID(g, objectID)
				if !ok || permanent.ObjectID == first.ObjectID {
					t.Fatal("replacement output reused the seeded stale publication")
				}
				names[permanent.TokenDef.Name]++
			}
			if len(names) != len(test.wantNames) {
				t.Fatalf("actual published replacement names = %v, want %v", names, test.wantNames)
			}
			for name, count := range test.wantNames {
				if names[name] != count {
					t.Fatalf("actual published replacement names = %v, want %v", names, test.wantNames)
				}
			}
			captured := append([]game.ObjectID(nil), g.DelayedTriggers[0].CapturedObjectIDs...)
			engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
			for _, objectID := range captured {
				if _, live := permanentByObjectID(g, objectID); live {
					t.Fatal("an actual replacement output escaped delayed disposal")
				}
			}
			if _, live := permanentByObjectID(g, first.ObjectID); !live {
				t.Fatal("delayed disposal removed the original token")
			}
		})
	}
}

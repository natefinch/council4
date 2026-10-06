package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestConditionalRepeatInvocationReceiptOwnership(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		modes                   []int
		legacy                  bool
		readOuter               bool
		skip                    bool
		bonus                   bool
		nested                  bool
		unconditional           bool
		wantLife                int
		wantDraws               int
		expectedValidationError string
	}{
		{name: "unselected publisher", modes: []int{1}, wantLife: 22},
		{name: "read only inherits outer", modes: []int{1}, readOuter: true, wantLife: 22,
			expectedValidationError: `instruction[1]: mode 1: instruction[0]: ResultGate references key "continue" not yet published`},
		{name: "legacy unselected publisher", modes: []int{1}, legacy: true, readOuter: true, wantLife: 22,
			expectedValidationError: `instruction[1]: mode 1: instruction[0]: ResultGate references key "continue" not yet published`},
		{name: "selected publisher", modes: []int{0}, wantLife: 21, wantDraws: 2},
		{name: "legacy selected publisher", modes: []int{0}, legacy: true, wantLife: 21, wantDraws: 2},
		{name: "skipped publisher", modes: []int{0}, skip: true, wantLife: 21},
		{name: "publisher then reader", modes: []int{0, 1}, wantLife: 24, wantDraws: 2},
		{name: "reader then publisher", modes: []int{1, 0}, wantLife: 24, wantDraws: 2},
		{name: "bonus selected modes", modes: []int{0, 1}, bonus: true, wantLife: 24, wantDraws: 2},
		{name: "nested bounded repeat", modes: []int{1}, nested: true, wantLife: 23},
		{name: "unconditional repeat", modes: []int{1}, unconditional: true, wantLife: 23},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			for range 2 {
				addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 2, 3))
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, ChosenModes: tc.modes,
				ResolutionResults:       map[string]game.InstructionResolutionResult{"continue": {Succeeded: true, Amount: 99}},
				ResolvedAmounts:         map[string]int{"continue": 99},
				ResolvedExcessDamage:    map[string]int{"continue": 17},
				ResolutionResultObjects: map[string][]game.ObjectSnapshot{"continue": {{}}},
			}
			publisher := game.Instruction{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				PublishResult: "continue", LocalProducts: game.LocalProducts{Results: []game.ResultKey{"continue"}}}
			if tc.legacy {
				publisher.LocalProducts = game.LocalProducts{}
			}
			if tc.skip {
				publisher.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.SourcePermanentReference()),
					ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})})
			}
			reader := game.Instruction{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}
			if tc.readOuter {
				reader.ResultGate = opt.Val(game.InstructionResultGate{Key: "continue", Succeeded: game.TriTrue})
			}
			body := game.AbilityContent{MinModes: 1, MaxModes: 2, Modes: []game.Mode{
				{Sequence: []game.Instruction{publisher}},
				{Sequence: []game.Instruction{reader}},
			}}
			if tc.bonus {
				body.MaxModes = 1
				body.ModeChoiceBonus = game.ModeChoiceBonus{Condition: game.ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1}
			}
			repeat := game.RepeatProcess{Body: body, ContinueResult: "continue"}
			if tc.unconditional {
				repeat.ContinueResult = ""
				repeat.Times = game.Fixed(2)
			}
			var process game.Primitive = repeat
			if tc.nested {
				process = game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{
					Sequence: []game.Instruction{{Primitive: repeat}},
				}.Ability()}
			}
			sequence := []game.Instruction{
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
					PublishResult: "continue", LocalProducts: game.LocalProducts{Results: []game.ResultKey{"continue"}}},
				{Primitive: process},
			}
			err := game.ValidateInstructionSequence(sequence)
			if tc.expectedValidationError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedValidationError) {
					t.Fatalf("validation error=%v, want %q", err, tc.expectedValidationError)
				}
				return
			}
			if err != nil {
				t.Fatalf("invalid repeat fixture: %v", err)
			}
			receipt, available := NewEngine(nil).resolveInstructionSequenceReceipt(g, obj, sequence,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{}, "continue")
			if g.Players[game.Player1].Life != tc.wantLife || g.Players[game.Player1].Hand.Size() != tc.wantDraws {
				t.Errorf("life=%d draws=%d, want %d/%d", g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size(), tc.wantLife, tc.wantDraws)
			}
			if !available || !receipt.Succeeded || receipt.Amount != 1 {
				t.Errorf("repeat replaced enclosing invocation receipt: %#v available=%v", receipt, available)
			}
			if outer := obj.ResolutionResults["continue"]; !outer.Succeeded || outer.Amount != 99 ||
				obj.ResolvedAmounts["continue"] != 99 || obj.ResolvedExcessDamage["continue"] != 17 ||
				len(obj.ResolutionResultObjects["continue"]) != 1 {
				t.Errorf("enclosing products were not restored: receipt=%#v amounts=%v excess=%v objects=%v",
					outer, obj.ResolvedAmounts, obj.ResolvedExcessDamage, obj.ResolutionResultObjects)
			}
		})
	}
	t.Run("unselected mode returns unavailable", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		obj := &game.StackObject{Controller: game.Player1, ChosenModes: []int{1},
			ResolutionResults: map[string]game.InstructionResolutionResult{"continue": {Succeeded: true, Amount: 99}}}
		body := game.AbilityContent{MinModes: 1, MaxModes: 1, Modes: []game.Mode{
			{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				PublishResult: "continue", LocalProducts: game.LocalProducts{Results: []game.ResultKey{"continue"}}}}},
			{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}},
		}}
		receipt, available := NewEngine(nil).resolveAbilityContentReceipt(g, obj, body,
			[game.NumPlayers]PlayerAgent{}, &TurnLog{}, "continue")
		if available || receipt.Succeeded || receipt.Amount != 0 {
			t.Errorf("unproduced receipt inherited outer success: %#v available=%v", receipt, available)
		}
		if outer := obj.ResolutionResults["continue"]; !outer.Succeeded || outer.Amount != 99 {
			t.Fatal("querying an unproduced receipt destroyed the outer receipt")
		}
	})
}

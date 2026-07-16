package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	springheartPaidResult   = game.ResultKey("springheart-landfall-paid")
	springheartCopiedResult = game.ResultKey("springheart-landfall-copied")
)

// springheartInsectTokenDef builds the fixed 1/1 green Insect creature token the
// landfall body creates whenever no copy token is made.
func springheartInsectTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Insect",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// springheartAttachedToControlledCreature is the "this permanent is attached to
// a creature you control" gate: the source Aura's attached permanent must be a
// creature its controller controls. It resolves the attached permanent through
// SourceAttachedPermanentReference, so it fails closed when the source is an
// unattached creature (Springheart cast as a creature or fallen off its host).
func springheartAttachedToControlledCreature() game.Condition {
	return game.Condition{
		Text:   "this permanent is attached to a creature you control",
		Object: opt.Val(game.SourceAttachedPermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}
}

// lowerLandfallPayCopyTokenTrigger composes Springheart Nantuko's recognized
// landfall body from existing primitives. The controller may pay {1}{G} only
// while the source is attached to a creature they control (the payment gate);
// paying publishes a success result that arms a true reflexive trigger, whose
// later resolution copies the still-attached creature or, when no copy is made,
// creates a fixed 1/1 green Insect. Every path that creates no copy token
// creates one Insect instead: the reflexive fallback (paid but detached before
// payoff), the declined-payment branch (attached but unpaid), and the
// unattached branch (no payment offered). Evaluating the copy and its
// attachment gate at reflexive resolution keeps attach, detach, and control
// changes between payment and payoff correct, and the copy reads the referenced
// creature's copiable values with the token owned and controlled by the
// ability's controller.
func lowerLandfallPayCopyTokenTrigger(ability compiler.CompiledAbility) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		summary, detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger kind")
	}

	attached := springheartAttachedToControlledCreature()
	notAttached := springheartAttachedToControlledCreature()
	notAttached.Text = "this permanent is not attached to a creature you control"
	notAttached.Negate = true

	insect := func() game.Instruction {
		return game.Instruction{Primitive: game.CreateToken{
			Amount: game.Fixed(1),
			Source: game.TokenDef(springheartInsectTokenDef()),
		}}
	}

	// Reflexive payoff: copy the still-attached creature, else a fixed Insect.
	copyInsectFallback := insect()
	copyInsectFallback.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       springheartCopiedResult,
		Succeeded: game.TriFalse,
	})
	reflexiveContent := game.Mode{Sequence: []game.Instruction{
		{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source: game.TokenCopySourceObject,
					Object: game.SourceAttachedPermanentReference(),
				}),
			},
			PublishResult: springheartCopiedResult,
		},
		copyInsectFallback,
	}}.Ability()

	pay := game.Instruction{
		Primitive: game.Pay{Payment: game.ResolutionPayment{
			Prompt:   "Pay {1}{G}?",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		}},
		Condition:     opt.Val(game.EffectCondition{Text: attached.Text, Condition: opt.Val(attached)}),
		PublishResult: springheartPaidResult,
	}
	reflexive := game.Instruction{
		Primitive: game.CreateReflexiveTrigger{Trigger: game.ReflexiveTriggerDef{Content: reflexiveContent}},
		ResultGate: opt.Val(game.InstructionResultGate{
			Key:       springheartPaidResult,
			Succeeded: game.TriTrue,
		}),
	}
	declinedInsect := insect()
	declinedInsect.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       springheartPaidResult,
		Succeeded: game.TriFalse,
	})
	unattachedInsect := insect()
	unattachedInsect.Condition = opt.Val(game.EffectCondition{Text: notAttached.Text, Condition: opt.Val(notAttached)})

	return game.TriggeredAbility{
		Text:    ability.Text,
		Trigger: game.TriggerCondition{Type: triggerType, Pattern: pattern},
		Content: game.Mode{Text: ability.Text, Sequence: []game.Instruction{
			pay,
			reflexive,
			declinedInsect,
			unattachedInsect,
		}}.Ability(),
	}, nil
}

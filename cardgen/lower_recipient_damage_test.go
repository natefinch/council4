package cardgen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestRecipientDamageFramePolicies(t *testing.T) {
	t.Parallel()
	for _, family := range []struct {
		name      string
		text      string
		lower     func(contentCtx) (game.AbilityContent, *shared.Diagnostic)
		recipient game.DamageRecipient
		detail    string
	}{
		{"controller", "{0}: This creature deals 2 damage to you.", lowerControllerDamageSpell,
			game.PlayerDamageRecipient(game.ControllerReference()), "exact fixed or X damage to you"},
		{"event player", "Whenever an opponent draws a card, this creature deals 2 damage to that player.", lowerEventPlayerDamageSpell,
			game.PlayerDamageRecipient(game.EventPlayerReference()), "exact fixed, X, or source-power damage to that player"},
		{"event related", "Whenever this creature blocks or becomes blocked by a creature, this creature deals 2 damage to that creature.", lowerEventRelatedPermanentDamageSpell,
			game.ObjectDamageRecipient(game.EventRelatedPermanentReference()), "exact fixed, X, or source-power damage to that creature"},
	} {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(family.text, parser.Context{CardName: "Frame Creature"}, compiler.Context{})
			if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
				t.Fatalf("compile: abilities=%d diagnostics=%+v", len(compilation.Abilities), diagnostics)
			}
			ability := compilation.Abilities[0]
			fresh := func() contentCtx {
				content := ability.Content
				content.Effects = append([]compiler.CompiledEffect(nil), content.Effects...)
				content.References = append([]compiler.CompiledReference(nil), content.References...)
				return contentCtx{content: content, span: content.Span, enclosingKind: ability.Kind}
			}
			for _, amount := range []struct {
				name string
				in   compiler.CompiledAmount
				want game.Quantity
			}{
				{"fixed", compiler.CompiledAmount{Known: true, Value: 2}, game.Fixed(2)},
				{"X", compiler.CompiledAmount{VariableX: true}, game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})},
			} {
				t.Run(amount.name, func(t *testing.T) {
					t.Parallel()
					ctx := fresh()
					ctx.content.Effects[0].Amount = amount.in
					content, diagnostic := family.lower(ctx)
					want := game.Mode{Sequence: []game.Instruction{{Primitive: game.Damage{
						Amount: amount.want, Recipient: family.recipient,
						DamageSource: primaryDamageSource(ability.Content.References),
					}}}}.Ability()
					if diagnostic != nil || !reflect.DeepEqual(content, want) {
						t.Fatalf("content=%+v diagnostic=%+v want=%+v", content, diagnostic, want)
					}
				})
			}
			for _, invalid := range []struct {
				name   string
				mutate func(*compiler.AbilityContent)
			}{
				{"inexact", func(c *compiler.AbilityContent) { c.Effects[0].Exact = false }},
				{"negated", func(c *compiler.AbilityContent) { c.Effects[0].Negated = true }},
				{"target", func(c *compiler.AbilityContent) { c.Targets = []compiler.CompiledTarget{{}} }},
				{"recipient group", func(c *compiler.AbilityContent) {
					c.Effects[0].DamageRecipient.GroupSelectors = []compiler.CompiledSelector{{}}
				}},
				{"condition", func(c *compiler.AbilityContent) { c.Conditions = []compiler.CompiledCondition{{}} }},
				{"keyword", func(c *compiler.AbilityContent) { c.Keywords = []compiler.CompiledKeyword{{}} }},
				{"mode", func(c *compiler.AbilityContent) { c.Modes = []compiler.CompiledMode{{}} }},
				{"zero", func(c *compiler.AbilityContent) { c.Effects[0].Amount = compiler.CompiledAmount{Known: true} }},
				{"negative", func(c *compiler.AbilityContent) {
					c.Effects[0].Amount = compiler.CompiledAmount{Known: true, Value: -1}
				}},
				{"unknown amount", func(c *compiler.AbilityContent) { c.Effects[0].Amount = compiler.CompiledAmount{} }},
				{"missing references", func(c *compiler.AbilityContent) { c.References = nil }},
				{"default source", func(c *compiler.AbilityContent) { c.References[0] = compiler.CompiledReference{} }},
				{"foreign source", func(c *compiler.AbilityContent) { c.References[0].Binding = compiler.ReferenceBindingTarget }},
			} {
				t.Run(invalid.name, func(t *testing.T) {
					t.Parallel()
					ctx := fresh()
					invalid.mutate(&ctx.content)
					content, diagnostic := family.lower(ctx)
					if diagnostic == nil || diagnostic.Summary != "unsupported damage spell" ||
						diagnostic.Detail != "the executable source backend supports only "+family.detail ||
						!reflect.DeepEqual(content, game.AbilityContent{}) {
						t.Fatalf("content=%+v diagnostic=%+v", content, diagnostic)
					}
				})
			}
			if family.name != "controller" {
				for _, binding := range []compiler.ReferenceBinding{
					compiler.ReferenceBindingUnsupported, compiler.ReferenceBindingTarget,
				} {
					t.Run(fmt.Sprintf("unbound recipient %v", binding), func(t *testing.T) {
						t.Parallel()
						ctx := fresh()
						ctx.content.References[len(ctx.content.References)-1].Binding = binding
						if _, diagnostic := family.lower(ctx); diagnostic == nil {
							t.Fatal("accepted a recipient without the required event binding")
						}
					})
				}
			}
			for _, invalid := range []struct {
				name   string
				mutate func(*compiler.AbilityContent)
			}{
				{"divided dispatch", func(c *compiler.AbilityContent) { c.Effects[0].Divided = true }},
				{"recipient dispatch", func(c *compiler.AbilityContent) {
					c.Effects[0].DamageRecipient.Reference = parser.DamageRecipientReferenceNone
				}},
				{"kind dispatch", func(c *compiler.AbilityContent) { c.Effects[0].Kind = compiler.EffectDraw }},
				{"count dispatch", func(c *compiler.AbilityContent) { c.Effects = append(c.Effects, c.Effects[0]) }},
			} {
				t.Run(invalid.name, func(t *testing.T) {
					t.Parallel()
					ctx := fresh()
					invalid.mutate(&ctx.content)
					assertPanics(t, invalid.name, func() { family.lower(ctx) })
				})
			}
		})
	}
}

func TestRecipientDamageCompiledCapabilities(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		text string
		want string
	}{
		{"controller fixed", "{0}: This creature deals 2 damage to you.", "Amount.fixed = 2"},
		{"controller X", "{X}: This creature deals X damage to you.", "Amount.dynamic.Kind = game.DynamicAmountX"},
		{"event fixed", "Whenever an opponent draws a card, this creature deals 2 damage to that player.", "Recipient.player.kind = game.PlayerReferenceEventPlayer"},
		{"event X", "Whenever an opponent draws a card, this creature deals X damage to that player.", "Amount.dynamic.Kind = game.DynamicAmountX"},
		{"event power", "Whenever an opponent casts a noncreature spell, this creature deals damage equal to its power to that player.", "Amount.dynamic.Kind = game.DynamicAmountObjectPower"},
		{"event hand", "Whenever an opponent attacks you, this creature deals damage to that player equal to the number of cards in their hand.", "Amount.dynamic.Kind = game.DynamicAmountCountCardsInZone"},
		{"related fixed", "Whenever this creature blocks or becomes blocked by a creature, this creature deals 2 damage to that creature.", "Recipient.object.kind = game.ObjectReferenceEventRelatedPermanent"},
		{"related X", "Whenever this creature blocks or becomes blocked by a creature, this creature deals X damage to that creature.", "Amount.dynamic.Kind = game.DynamicAmountX"},
		{"related power", "Whenever this creature blocks or becomes blocked by a creature, this creature deals damage equal to its power to that creature.", "Amount.dynamic.Kind = game.DynamicAmountObjectPower"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Frame Creature", Layout: "normal", TypeLine: "Creature",
				Power: new("5"), Toughness: new("8"), OracleText: tc.text}
			assertCardPaths(t, card, tc.want, "DamageSource.Val.kind = game.ObjectReferenceSourcePermanent")
			assertCardPathsAbsent(t, card, ".Targets[")
		})
	}
	for _, text := range []string{
		"{0}: This creature deals damage equal to its power to you.",
		"{0}: This creature deals 2 damage to that player.",
		"{0}: This creature deals 2 damage to that creature.",
		"{0}: This creature deals 2 damage to you. The damage can't be prevented.",
		"Whenever an opponent draws a card, this creature deals 2 damage to that player. The damage can't be prevented.",
		"Whenever this creature blocks or becomes blocked by a creature, this creature deals 2 damage to that creature. The damage can't be prevented.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Frame Refusal", Layout: "normal", TypeLine: "Creature", OracleText: text})
		})
	}
}

func TestRecipientDamageIndependentClauseCompositions(t *testing.T) {
	t.Parallel()
	for _, family := range []struct {
		preamble string
		damage   string
	}{
		{"{0}: ", "This creature deals 2 damage to you."},
		{"Whenever an opponent draws a card, ", "this creature deals 2 damage to that player."},
		{"Whenever this creature blocks or becomes blocked by a creature, ", "this creature deals 2 damage to that creature."},
	} {
		for _, rider := range []string{"Draw a card.", "You gain 1 life.", "Scry 1."} {
			for _, before := range []bool{false, true} {
				text := family.preamble + family.damage + " " + rider
				if before {
					text = family.preamble + rider + " " + family.damage
				}
				t.Run(text, func(t *testing.T) {
					t.Parallel()
					face := lowerSingleFace(t, &ScryfallCard{Name: "Frame Composition", Layout: "normal",
						TypeLine: "Creature", OracleText: text, Power: new("5"), Toughness: new("8")})
					content := game.AbilityContent{}
					if len(face.ActivatedAbilities) != 0 {
						content = face.ActivatedAbilities[0].Content
					} else {
						content = face.TriggeredAbilities[0].Content
					}
					sequence := content.Modes[0].Sequence
					index := 0
					if before {
						index = 1
					}
					if len(sequence) != 2 || sequence[index].Primitive.Kind() != game.PrimitiveDamage ||
						sequence[1-index].Primitive.Kind() == game.PrimitiveDamage {
						t.Fatalf("independent clause order changed: %+v", sequence)
					}
				})
			}
		}
	}
}

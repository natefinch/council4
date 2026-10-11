package cardgen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSharedPlayerRecipientDomains(t *testing.T) {
	t.Parallel()
	permanent := targetCreatureFixture()
	player := compiler.CompiledTarget{
		Exact: true, Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Selector: compiler.CompiledSelector{Kind: compiler.SelectorOpponent},
	}
	stack := permanent
	stack.Selector.Kind = compiler.SelectorSpell
	card := player
	card.Selector = compiler.CompiledSelector{
		Kind: compiler.SelectorCard, Zone: zone.Graveyard, Controller: compiler.ControllerOpponent,
	}
	targetRef := compiler.CompiledReference{Kind: compiler.ReferencePronoun, Binding: compiler.ReferenceBindingTarget, Occurrence: 3}
	thatPlayer := targetRef
	thatPlayer.Kind = compiler.ReferenceThatPlayer
	eventCompilation, diagnostics := compileTestOracle("Whenever an opponent loses life, that player mills two cards.", parser.Context{}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	eventRef := eventCompilation.Abilities[0].Content.References[0]
	for _, tt := range []struct {
		name       string
		context    parser.EffectContextKind
		targets    []compiler.CompiledTarget
		references []compiler.CompiledReference
		player     game.PlayerReference
		group      game.PlayerGroupReference
	}{
		{"controller", parser.EffectContextController, nil, nil, game.ControllerReference(), game.PlayerGroupReference{}},
		{"defending", parser.EffectContextDefendingPlayer, nil, nil, game.DefendingPlayerReference(), game.PlayerGroupReference{}},
		{"opponents", parser.EffectContextEachOpponent, nil, nil, game.PlayerReference{}, game.OpponentsReference()},
		{"all players", parser.EffectContextEachPlayer, nil, nil, game.PlayerReference{}, game.AllPlayersReference()},
		{"event player", parser.EffectContextReferencedPlayer, nil, []compiler.CompiledReference{eventRef}, game.EventPlayerReference(), game.PlayerGroupReference{}},
		{"permanent controller", parser.EffectContextReferencedObjectController, []compiler.CompiledTarget{permanent}, []compiler.CompiledReference{targetRef}, game.ObjectControllerReference(game.TargetPermanentReference(3)), game.PlayerGroupReference{}},
		{"permanent owner", parser.EffectContextReferencedObjectOwner, []compiler.CompiledTarget{permanent}, []compiler.CompiledReference{targetRef}, game.ObjectOwnerReference(game.TargetPermanentReference(3)), game.PlayerGroupReference{}},
		{"stack controller", parser.EffectContextReferencedObjectController, []compiler.CompiledTarget{stack}, []compiler.CompiledReference{targetRef}, game.ObjectControllerReference(game.TargetStackObjectReference(3)), game.PlayerGroupReference{}},
		{"prior player", parser.EffectContextReferencedPlayer, []compiler.CompiledTarget{player}, []compiler.CompiledReference{thatPlayer}, game.TargetPlayerReference(3), game.PlayerGroupReference{}},
		{"prior graveyard owner", parser.EffectContextReferencedPlayer, []compiler.CompiledTarget{card}, []compiler.CompiledReference{thatPlayer}, game.ObjectOwnerReference(game.TargetCardReference(3)), game.PlayerGroupReference{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := contentCtx{content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{Context: tt.context}},
				Targets: tt.targets,
			}}
			got, ok := resolvePlayerRecipient(ctx, tt.references)
			if !ok || got.player != tt.player || got.group != tt.group || len(got.targets) != 0 {
				t.Fatalf("recipient=%+v ok=%v, want player=%v group=%v with no new targets", got, ok, tt.player, tt.group)
			}
			ctx.text = "not Oracle"
			ctx.content.Effects[0].Text = "not Oracle"
			opaque, ok := resolvePlayerRecipient(ctx, tt.references)
			if !ok || !reflect.DeepEqual(got, opaque) {
				t.Fatal("recipient resolution depends on retained text")
			}
		})
	}
}

func TestSharedPlayerRecipientCardDefCompositions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text, path, primitive, recipient string
		index                                  int
	}{
		{"life second target", "Tap target creature. Target opponent loses 2 life.", "SpellAbility.Val", "LoseLife", "Player.targetIndex = 1", 1},
		{"draw second target", "Tap target creature. Target opponent draws two cards.", "SpellAbility.Val", "Draw", "Player.targetIndex = 1", 1},
		{"mill second target", "Tap target creature. Target player mills two cards.", "SpellAbility.Val", "MoveTopOfLibrary", "Player.targetIndex = 1", 1},
		{"owner after independent action", "Tap target artifact. Destroy target creature. Scry 1. Its owner gains 2 life.", "SpellAbility.Val", "GainLife", "Player.object.Val.targetIndex = 1", 3},
		{"controller after independent action", "Tap target artifact. Destroy target creature. Scry 1. Its controller draws a card.", "SpellAbility.Val", "Draw", "Player.object.Val.targetIndex = 1", 3},
		{"mill after independent action", "Tap target artifact. Destroy target creature. Scry 1. Its controller mills two cards.", "SpellAbility.Val", "MoveTopOfLibrary", "Player.object.Val.targetIndex = 1", 3},
		{"stack controller", "Counter target spell. Its controller draws a card.", "SpellAbility.Val", "Draw", "Player.object.Val.kind = game.ObjectReferenceTargetStackObject", 1},
		{"opponents life", "Each opponent loses 2 life.", "SpellAbility.Val", "LoseLife", "PlayerGroup.Kind = game.PlayerGroupReferenceOpponents", 0},
		{"all draw", "Each player draws a card.", "SpellAbility.Val", "Draw", "PlayerGroup.Kind = game.PlayerGroupReferenceAllPlayers", 0},
		{"opponents mill", "Each opponent mills two cards.", "SpellAbility.Val", "MoveTopOfLibrary", "PlayerGroup.Kind = game.PlayerGroupReferenceOpponents", 0},
		{"conditional draw", "If you control a creature, draw a card. You gain 2 life.", "SpellAbility.Val", "Draw", "Player.kind = game.PlayerReferenceController", 0},
		{"optional draw and life", "You may draw a card and gain 2 life. You lose 1 life.", "SpellAbility.Val", "GainLife", "Player.kind = game.PlayerReferenceController", 1},
		{"event player draw", "Whenever an opponent casts a spell, they draw a card.", "TriggeredAbilities[0].Content", "Draw", "Player.kind = game.PlayerReferenceEventPlayer", 0},
		{"event player life", "Whenever an opponent casts a spell, they lose 2 life.", "TriggeredAbilities[0].Content", "LoseLife", "Player.kind = game.PlayerReferenceEventPlayer", 0},
		{"event player mill", "Whenever an opponent loses life, that player mills that many cards.", "TriggeredAbilities[0].Content", "MoveTopOfLibrary", "Player.kind = game.PlayerReferenceEventPlayer", 0},
		{"departed event controller", "Whenever a creature dies, that creature's controller loses 2 life.", "TriggeredAbilities[0].Content", "LoseLife", "Player.object.Val.kind = game.ObjectReferenceEventPermanent", 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			typeLine := "Instant"
			if tt.path != "SpellAbility.Val" {
				typeLine = "Enchantment"
			}
			card := &ScryfallCard{Name: "Recipient Test", Layout: "normal", TypeLine: typeLine, OracleText: tt.text}
			path := fmt.Sprintf("%s.Modes[0].Sequence[%d].Primitive.(game.%s).", tt.path, tt.index, tt.primitive)
			assertCardPaths(t, card, path+tt.recipient)
			assertCardPathsAbsent(t, card, path+"Player.kind = game.PlayerReferenceNone")
		})
	}
}

func TestSharedPlayerRecipientOptionalAndModalCardDefs(t *testing.T) {
	t.Parallel()
	optional := &ScryfallCard{Name: "Optional Recipients", Layout: "normal", TypeLine: "Instant",
		OracleText: "You may draw a card and gain 2 life. You lose 1 life."}
	assertCardPaths(t, optional,
		"Sequence[0].Optional = true",
		"Sequence[0].PublishOptionalDecision = \"optional-clause-1\"",
		"Sequence[1].OptionalDecisionGate = \"optional-clause-1\"")
	assertCardPathsAbsent(t, optional, "Sequence[2].Optional", "Sequence[2].OptionalDecisionGate")
	modal := &ScryfallCard{Name: "Modal Recipients", Layout: "normal", TypeLine: "Instant",
		OracleText: "Choose one —\n\u2022 Target player draws a card.\n\u2022 Each opponent loses 2 life."}
	assertCardPaths(t, modal,
		"Modes[0].Sequence[0].Primitive.(game.Draw).Player.kind = game.PlayerReferenceTargetPlayer",
		"Modes[1].Sequence[0].Primitive.(game.LoseLife).PlayerGroup.Kind = game.PlayerGroupReferenceOpponents")
	assertCardPathsAbsent(t, modal, "Modes[1].Targets")
}

func TestSharedPlayerRecipientUnconsumedOperands(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"You gain 2 life.", "Draw a card.", "Mill two cards."} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(text, parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			for _, edit := range []struct {
				name string
				edit func(*compiler.AbilityContent)
			}{
				{"missing actor", func(c *compiler.AbilityContent) { c.Effects[0].Context = parser.EffectContextUnknown }},
				{"foreign event reference", func(c *compiler.AbilityContent) {
					c.References = []compiler.CompiledReference{{Binding: compiler.ReferenceBindingEventPermanent}}
				}},
				{"foreign event player binding", func(c *compiler.AbilityContent) {
					c.Effects[0].Context = parser.EffectContextEventPlayer
					c.References = []compiler.CompiledReference{{Kind: compiler.ReferencePronoun,
						Pronoun: compiler.ReferencePronounThey, Binding: compiler.ReferenceBindingSource}}
				}},
				{"extra references", func(c *compiler.AbilityContent) {
					c.Effects[0].Context = parser.EffectContextEventPlayer
					c.References = []compiler.CompiledReference{
						{Binding: compiler.ReferenceBindingEventPlayer}, {Binding: compiler.ReferenceBindingEventPlayer},
					}
				}},
				{"unconsumed keyword", func(c *compiler.AbilityContent) { c.Keywords = []compiler.CompiledKeyword{{}} }},
				{"unconsumed condition", func(c *compiler.AbilityContent) { c.Conditions = []compiler.CompiledCondition{{}} }},
				{"unconsumed mode", func(c *compiler.AbilityContent) { c.Modes = []compiler.CompiledMode{{}} }},
				{"unknown amount", func(c *compiler.AbilityContent) {
					c.Effects = append([]compiler.CompiledEffect(nil), c.Effects...)
					c.Effects[0].Amount = compiler.CompiledAmount{}
				}},
			} {
				t.Run(edit.name, func(t *testing.T) {
					mutated := content
					mutated.Effects = append([]compiler.CompiledEffect(nil), content.Effects...)
					edit.edit(&mutated)
					ctx := contentCtx{content: mutated}
					var diagnostic *shared.Diagnostic
					switch mutated.Effects[0].Kind {
					case compiler.EffectGain:
						_, diagnostic = lowerFixedLifeSpell(ctx, "gain",
							func(q game.Quantity, p game.PlayerReference) game.Primitive {
								return game.GainLife{Amount: q, Player: p}
							},
							func(q game.Quantity, p game.PlayerGroupReference) game.Primitive {
								return game.GainLife{Amount: q, PlayerGroup: p}
							})
					case compiler.EffectDraw:
						_, diagnostic = lowerFixedDrawSpell(ctx, nil)
					case compiler.EffectMill:
						_, diagnostic = lowerFixedCardCountPlayerSpell(ctx, nil, "mill", "mills", true,
							func(q game.Quantity, p game.PlayerReference) game.Primitive {
								return game.MoveTopOfLibrary{Amount: q, Player: p, Destination: zone.Graveyard}
							},
							func(q game.Quantity, p game.PlayerGroupReference) game.Primitive {
								return game.MoveTopOfLibrary{Amount: q, PlayerGroup: p, Destination: zone.Graveyard}
							})
					default:
						t.Fatalf("unexpected effect kind: %v", mutated.Effects[0].Kind)
					}
					if diagnostic == nil {
						t.Fatal("an unconsumed operand was accepted")
					}
				})
			}
		})
	}
}

func TestSharedPlayerRecipientUnsupportedCardParameters(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Each opponent loses half their life, rounded up.",
		"Each opponent mills half their library, rounded up.",
		"Draw zero cards.",
		"Target player mills cards equal to your life total.",
		"Destroy target creature. Its owner draws a card.",
		"Destroy target creature. Its owner mills two cards.",
		"Draw a card and do something mysterious.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Recipient Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
		})
	}
}

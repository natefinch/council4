package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const questingBeastOracleText = "Vigilance, deathtouch, haste\n" +
	"Questing Beast can't be blocked by creatures with power 2 or less.\n" +
	"Combat damage that would be dealt by creatures you control can't be prevented.\n" +
	"Whenever Questing Beast deals combat damage to an opponent, it deals that much damage to target planeswalker that player controls."

func TestLowerQuestingBeast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Questing Beast",
		Layout:     "normal",
		ManaCost:   "{2}{G}{G}",
		TypeLine:   "Legendary Creature — Beast",
		OracleText: questingBeastOracleText,
		Power:      new("4"),
		Toughness:  new("4"),
	})

	var blocker, unpreventable bool
	var keywords []game.Keyword
	for _, static := range face.StaticAbilities {
		switch static.VarName {
		case "game.VigilanceStaticBody":
			keywords = append(keywords, game.Vigilance)
		case "game.DeathtouchStaticBody":
			keywords = append(keywords, game.Deathtouch)
		case "game.HasteStaticBody":
			keywords = append(keywords, game.Haste)
		default:
		}

		for _, effect := range static.Body.ContinuousEffects {
			keywords = append(keywords, effect.AddKeywords...)
		}
		for _, effect := range static.Body.RuleEffects {
			switch effect.Kind {
			case game.RuleEffectCantBeBlockedByCreaturesWith:
				blocker = effect.BlockerRestriction.Kind == game.BlockerRestrictionPowerLessOrEqual &&
					effect.BlockerRestriction.Power == 2
			case game.RuleEffectCombatDamageCantBePrevented:
				unpreventable = effect.AffectedSelection.Controller == game.ControllerYou &&
					(slices.Contains(effect.AffectedSelection.RequiredTypesAny, types.Creature) ||
						slices.Contains(effect.AffectedSelection.RequiredTypes, types.Creature))
			default:
			}
		}

	}
	for _, keyword := range []game.Keyword{game.Vigilance, game.Deathtouch, game.Haste} {
		if !slices.Contains(keywords, keyword) {
			t.Errorf("static keywords = %v, missing %v", keywords, keyword)
		}
	}
	if !blocker {
		t.Error("missing power-2-or-less blocker prohibition")
	}
	if !unpreventable {
		t.Error("missing controller-relative combat-damage prevention prohibition")
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	pattern := ability.Trigger.Pattern
	if pattern.Event != game.EventDamageDealt ||
		pattern.Source != game.TriggerSourceSelf ||
		pattern.Subject != game.TriggerSubjectDamageSource ||
		pattern.Player != game.TriggerPlayerOpponent ||
		pattern.DamageRecipient != game.DamageRecipientPlayer ||
		!pattern.RequireCombatDamage {
		t.Fatalf("trigger pattern = %#v", pattern)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Targets) != 1 ||
		!mode.Targets[0].Selection.Exists ||
		!mode.Targets[0].Selection.Val.ControlledByEventPlayer ||
		!slices.Contains(mode.Targets[0].Selection.Val.RequiredTypesAny, types.Planeswalker) {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %#v, want Damage", mode.Sequence[0].Primitive)
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
		t.Fatalf("damage amount = %#v, want event damage", damage.Amount)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("damage source = %#v, want triggering permanent", damage.DamageSource)
	}
}

func TestLowerCombatDamagePreventionProhibitionFromTypedSemantics(t *testing.T) {
	t.Parallel()
	ability := compiler.CompiledAbility{
		Kind: compiler.AbilityStatic,
		Static: &compiler.CompiledStaticSemantics{Declarations: []compiler.StaticDeclaration{{
			Kind: compiler.StaticDeclarationCombatDamagePreventionProhibition,
			CombatDamagePreventionProhibition: &compiler.StaticCombatDamagePreventionProhibitionDeclaration{
				Source: compiler.CompiledSelector{
					Kind:       compiler.SelectorCreature,
					Controller: compiler.ControllerYou,
				},
			},
		}}},
	}
	lowered, handled, diagnostic := lowerStaticDeclarations(ability, &parser.Ability{})
	if !handled || diagnostic != nil {
		t.Fatalf("handled = %v diagnostic = %#v", handled, diagnostic)
	}
	if len(lowered.staticAbilities) != 1 ||
		len(lowered.staticAbilities[0].Body.RuleEffects) != 1 ||
		lowered.staticAbilities[0].Body.RuleEffects[0].Kind != game.RuleEffectCombatDamageCantBePrevented {
		t.Fatalf("lowered = %#v", lowered)
	}
}

func TestGenerateExecutableQuestingBeast(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Questing Beast",
		Layout:     "normal",
		ManaCost:   "{2}{G}{G}",
		TypeLine:   "Legendary Creature — Beast",
		OracleText: questingBeastOracleText,
		Power:      new("4"),
		Toughness:  new("4"),
	}, "q")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.RuleEffectCantBeBlockedByCreaturesWith",
		"game.RuleEffectCombatDamageCantBePrevented",
		"ControlledByEventPlayer: true",
		"game.DynamicAmountEventDamage",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

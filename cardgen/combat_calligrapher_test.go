package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

const combatCalligrapherOracle = "Flying\n" +
	"Inklings can't attack you or planeswalkers you control.\n" +
	"Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent."

func TestLowerCombatCalligrapherMechanics(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Calligrapher",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warlock",
		OracleText: combatCalligrapherOracle,
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.StaticAbilities) != 2 || len(face.TriggeredAbilities) != 1 {
		t.Fatalf("face = %#v", face)
	}
	var restriction *game.RuleEffect
	for i := range face.StaticAbilities {
		for j := range face.StaticAbilities[i].Body.RuleEffects {
			effect := &face.StaticAbilities[i].Body.RuleEffects[j]
			if effect.Kind == game.RuleEffectCantAttack {
				restriction = effect
			}
		}
	}
	if restriction == nil ||
		restriction.AffectedController != game.ControllerAny ||
		restriction.DefendingPlayer != game.PlayerYou ||
		len(restriction.AffectedSelection.SubtypesAny) != 1 ||
		restriction.AffectedSelection.SubtypesAny[0] != types.Inkling {
		t.Fatalf("restriction: controller=%v defender=%v types=%v selection=%#v",
			restriction.AffectedController, restriction.DefendingPlayer,
			restriction.PermanentTypes, restriction.AffectedSelection)
	}

	ability := face.TriggeredAbilities[0]
	pattern := ability.Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Controller != game.TriggerControllerAny ||
		pattern.Player != game.TriggerPlayerOpponent ||
		pattern.AttackRecipient != game.AttackRecipientPlayer ||
		!pattern.OneOrMore ||
		!pattern.OneOrMorePerAttackTarget {
		t.Fatalf("trigger pattern = %#v", pattern)
	}
	create, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %#v, want CreateToken", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if !create.Recipient.Exists ||
		create.Recipient.Val.Kind() != game.PlayerReferenceEventPlayer ||
		!create.EntryTapped ||
		create.EntryAttacking ||
		!create.EntryAttackingDefender.Exists ||
		create.EntryAttackingDefender.Val.Kind() != game.PlayerReferenceDefendingPlayer {
		t.Fatalf("create token = %#v", create)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok ||
		def.Power.Val.Value != 2 ||
		def.Toughness.Val.Value != 1 ||
		len(def.Colors) != 2 ||
		!slices.Contains(def.Colors, color.White) ||
		!slices.Contains(def.Colors, color.Black) ||
		len(def.Subtypes) != 1 ||
		def.Subtypes[0] != types.Inkling ||
		len(def.StaticAbilities) != 1 ||
		len(def.StaticAbilities[0].KeywordAbilities) != 1 ||
		game.KeywordAbilityKind(def.StaticAbilities[0].KeywordAbilities[0]) != game.Flying {
		t.Fatalf("token definition = %#v", def)
	}
}

func TestRenderCombatCalligrapherUsesTypedCorrelation(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Calligrapher",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warlock",
		OracleText: combatCalligrapherOracle,
		Power:      new("3"),
		Toughness:  new("3"),
	}, "c")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("err = %v, diagnostics = %#v", err, diagnostics)
	}
	for _, want := range []string{
		"Kind:              game.RuleEffectCantAttack,",
		"DefendingPlayer:   game.PlayerYou,",
		`SubtypesAny: []types.Sub{types.Sub("Inkling")}`,
		"OneOrMorePerAttackTarget: true,",
		"opt.Val(game.EventPlayerReference())",
		"opt.Val(game.DefendingPlayerReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

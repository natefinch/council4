package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerEllivereOfTheWildCourt(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ellivere of the Wild Court",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Knight",
		OracleText: "Whenever Ellivere enters or attacks, create a Virtuous Role token attached to another target creature you control. " +
			"(If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 for each enchantment you control.)\n" +
			"Whenever an enchanted creature you control deals combat damage to a player, draw a card.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}

	roleTrigger := face.TriggeredAbilities[0]
	roleMode := roleTrigger.Content.Modes[0]
	if roleTrigger.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield ||
		roleTrigger.Trigger.Pattern.UnionEvent != game.EventAttackerDeclared ||
		roleTrigger.Trigger.Pattern.Source != game.TriggerSourceSelf ||
		len(roleMode.Targets) != 1 ||
		!roleMode.Targets[0].Selection.Exists ||
		!roleMode.Targets[0].Selection.Val.ExcludeSource ||
		roleMode.Targets[0].Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("Role trigger = %#v", roleTrigger)
	}
	create, ok := roleMode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || !create.EntryAttachedTo.Exists {
		t.Fatalf("Role create primitive = %#v", roleTrigger.Content)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok || token.Name != "Virtuous Role" ||
		len(token.Types) != 1 || token.Types[0] != types.Enchantment ||
		len(token.Subtypes) != 2 || token.Subtypes[0] != types.Aura || token.Subtypes[1] != types.Role ||
		len(token.StaticAbilities) != 2 {
		t.Fatalf("Virtuous Role token = %#v", token)
	}
	buff := token.StaticAbilities[1].ContinuousEffects[0]
	if buff.Group.Domain() != game.GroupDomainAttachedObject ||
		!buff.PowerDeltaDynamic.Exists ||
		buff.PowerDeltaDynamic.Val.Group.Selection().Controller != game.ControllerYou {
		t.Fatalf("Virtuous Role buff = %#v", buff)
	}

	drawTrigger := face.TriggeredAbilities[1].Trigger.Pattern
	if drawTrigger.Event != game.EventDamageDealt ||
		drawTrigger.Subject != game.TriggerSubjectDamageSource ||
		drawTrigger.Controller != game.TriggerControllerYou ||
		!drawTrigger.RequireCombatDamage ||
		drawTrigger.DamageRecipient != game.DamageRecipientPlayer ||
		!drawTrigger.DamageSourceWasEnchanted ||
		drawTrigger.OneOrMore {
		t.Fatalf("draw trigger = %#v", drawTrigger)
	}
}

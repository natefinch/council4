package cardgen

import "testing"

// TestCompileTheSpearOfBashenga covers the "attacks the
// monarch" attack trigger paired with a "that player controls" destroy target.
// The equipped creature attacking the monarch triggers on EventAttackerDeclared
// with the defending player constrained to the monarch (Player:
// TriggerPlayerMonarch, AttackRecipient: AttackRecipientPlayer), and the destroy
// target is the tapped nonland permanent controlled by that attacked player, so
// the "that player controls" clause lowers to ControlledByDefendingPlayer rather
// than the attacker-relative ControlledByEventPlayer.
func TestCompileTheSpearOfBashenga(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "The Spear of Bashenga",
		Layout:   "normal",
		ManaCost: "{4}{W}",
		TypeLine: "Legendary Artifact — Equipment",
		OracleText: "When The Spear of Bashenga enters, if there is no monarch, you become the monarch.\n" +
			"Equipped creature gets +2/+2 and has vigilance.\n" +
			"Whenever equipped creature attacks the monarch, destroy target tapped nonland permanent that player controls.\n" +
			"Equip {2}",
	}
	assertCardPaths(t, card,
		"TriggeredAbilities[1].Trigger.Pattern.Event = game.EventAttackerDeclared",
		"TriggeredAbilities[1].Trigger.Pattern.Source = game.TriggerSourceAttachedPermanent",
		"TriggeredAbilities[1].Trigger.Pattern.Player = game.TriggerPlayerMonarch",
		"TriggeredAbilities[1].Trigger.Pattern.AttackRecipient = game.AttackRecipientPlayer",
		"TriggeredAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.Destroy).Object.kind = game.ObjectReferenceTargetPermanent",
		"TriggeredAbilities[1].Content.Modes[0].Targets[0].Selection.Val.ExcludedTypes[0] = types.Land",
		"TriggeredAbilities[1].Content.Modes[0].Targets[0].Selection.Val.Tapped = game.TriTrue",
		"TriggeredAbilities[1].Content.Modes[0].Targets[0].Selection.Val.ControlledByDefendingPlayer = true",
	)
	assertCardPathsAbsent(t, card, "TriggeredAbilities[1].Content.Modes[0].Targets[0].Selection.Val.ControlledByEventPlayer")
}

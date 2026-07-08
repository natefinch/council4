package cardgen

import "testing"

// TestLowerTchakaCommanderControlActivationCondition proves T'Chaka, Venerable
// King's graveyard become-monarch ability lowers with the
// "Activate only if you control your commander." restriction as a
// ControllerControlsCommander activation condition. Previously the
// commander-control predicate was recognized only for static "as long as" gates
// and intervening trigger conditions, so its use as an activation restriction
// failed closed as an unsupported activation condition.
func TestLowerTchakaCommanderControlActivationCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "T'Chaka, Venerable King",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Noble Hero",
		ManaCost:   "{G}{W}",
		OracleText: "When T'Chaka enters, mill three cards, then you may put an artifact or land card from among the milled cards into your hand.\n{3}, Exile this card from your graveyard: You become the monarch. Activate only if you control your commander.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ActivationCondition.Exists {
		t.Fatal("activation condition missing, want ControllerControlsCommander")
	}
	if !ability.ActivationCondition.Val.ControllerControlsCommander {
		t.Fatalf("activation condition = %#v, want ControllerControlsCommander", ability.ActivationCondition.Val)
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1 (ETB mill)", len(face.TriggeredAbilities))
	}
}

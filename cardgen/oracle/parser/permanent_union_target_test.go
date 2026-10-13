package parser

import "testing"

func TestPermanentUnionTargetGrammar(t *testing.T) {
	for _, text := range []string{
		"target creature or Vehicle",
		"target Vehicle or creature",
		"target creature or Vehicle you control",
		"target creature or Vehicle an opponent controls",
		"target blue creature or artifact",
		"target creature or tapped Vehicle",
		"target artifact creature or Vehicle",
		"target creature with flying or Vehicle",
		"target creature with mana value 3 or less or Vehicle with mana value 3 or less",
	} {
		t.Run(text, func(t *testing.T) {
			target := singleTarget(t, "Return "+text+" to its owner's hand.")
			if !target.Exact || target.Selection.Kind != SelectionPermanent ||
				len(target.Selection.Alternatives) != 2 {
				t.Fatalf("want fully consumed permanent union: %#v", target)
			}
		})
	}
	target := singleTarget(t, "Tap target blue creature or artifact.")
	if len(target.Selection.ColorsAny) != 0 ||
		len(target.Selection.Alternatives[0].ColorsAny) != 1 ||
		target.Selection.Alternatives[0].ColorsAny[0] != ColorBlue ||
		len(target.Selection.Alternatives[1].ColorsAny) != 0 {
		t.Fatalf("blue must qualify only creature: %#v", target.Selection)
	}
	target = singleTarget(t, "Tap target creature or Vehicle you control.")
	if target.Selection.Controller != SelectionControllerYou ||
		target.Selection.Alternatives[1].Controller != SelectionControllerAny {
		t.Fatalf("whole-union controller was not retained: %#v", target.Selection)
	}
}

func TestPermanentUnionTargetControls(t *testing.T) {
	for _, text := range []string{
		"target creature or artifact",
		"target artifact creature",
		"target Wolf or Werewolf",
	} {
		target := singleTarget(t, "Tap "+text+".")
		if !target.Exact || len(target.Selection.Alternatives) != 0 {
			t.Fatalf("existing target changed: %#v", target)
		}
	}
	target := singleTarget(t, "Tap target artifact creature.")
	if !target.Selection.ConjunctiveTypes ||
		len(target.Selection.RequiredTypesAny) != 2 ||
		target.Selection.RequiredTypesAny[0] != CardTypeArtifact ||
		target.Selection.RequiredTypesAny[1] != CardTypeCreature {
		t.Fatalf("conjunctive types changed: %#v", target.Selection)
	}
}

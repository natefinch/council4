package game

import "testing"

func TestUmbraArmorStaticKeywordBody(t *testing.T) {
	body, ok := KeywordStaticBody(UmbraArmor)
	if !ok {
		t.Fatal("KeywordStaticBody(UmbraArmor) = false")
	}
	if !BodyHasKeyword(&body, UmbraArmor) {
		t.Fatal("UmbraArmorStaticBody does not carry UmbraArmor")
	}
	if body.Text != "Umbra armor" {
		t.Fatalf("body text = %q, want Umbra armor", body.Text)
	}
}

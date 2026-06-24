package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

func TestBodyText(t *testing.T) {
	cases := []struct {
		body Ability
		want string
	}{
		{&ActivatedAbility{Text: "activated"}, "activated"},
		{&ManaAbility{Text: "mana"}, "mana"},
		{&LoyaltyAbility{Text: "loyalty"}, "loyalty"},
		{&TriggeredAbility{Text: "triggered"}, "triggered"},
		{&ChapterAbility{Text: "chapter"}, "chapter"},
		{&ReplacementAbility{Text: "replacement"}, "replacement"},
		{&StaticAbility{Text: "static"}, "static"},
		{&AbilityContent{}, ""},
	}
	for _, c := range cases {
		if got := BodyText(c.body); got != c.want {
			t.Errorf("BodyText(%T) = %q, want %q", c.body, got, c.want)
		}
	}
}

func TestBodyTextNilBody(t *testing.T) {
	if got := BodyText((*ActivatedAbility)(nil)); got != "" {
		t.Fatalf("BodyText(nil) = %q, want empty", got)
	}
}

func TestBodyAdditionalCosts(t *testing.T) {
	activated := &ActivatedAbility{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice}}}
	if got := BodyAdditionalCosts(activated); len(got) != 1 || got[0].Kind != cost.AdditionalSacrifice {
		t.Fatalf("BodyAdditionalCosts(activated) = %+v, want one sacrifice cost", got)
	}
	mana := &ManaAbility{AdditionalCosts: cost.Tap}
	if got := BodyAdditionalCosts(mana); len(got) != 1 || got[0].Kind != cost.AdditionalTap {
		t.Fatalf("BodyAdditionalCosts(mana) = %+v, want tap cost", got)
	}
	if got := BodyAdditionalCosts(&TriggeredAbility{}); got != nil {
		t.Fatalf("BodyAdditionalCosts(triggered) = %+v, want nil", got)
	}
}

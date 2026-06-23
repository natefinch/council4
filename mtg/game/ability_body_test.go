package game

import "testing"

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

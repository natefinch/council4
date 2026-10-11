package parser

import "testing"

func TestLeadingCreateIteratorConsumption(t *testing.T) {
	for _, tc := range []struct {
		source string
		exact  bool
	}{
		{"For each opponent, create a 1/1 white Soldier creature token.", true},
		{"For each creature you control, create a 1/1 white Soldier creature token.", true},
		{"For each creature you control that's a Wolf or a Werewolf, create a 2/2 green Wolf creature token.", true},
		{"For each creature you control that is a Wolf or a Werewolf, create a 2/2 green Wolf creature token.", true},
		{"For each creature you control that's a Wolf and a Werewolf, create a 2/2 green Wolf creature token.", false},
		{"For each creature you control that's a Wolf or a blue Werewolf, create a 2/2 green Wolf creature token.", false},
		{"Whenever this creature attacks, for each token you control, create a 1/1 white Rabbit creature token.", true},
		{"For each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"Then for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"Then, for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"And then for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"You may for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"Then you may for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
		{"Then, you may for each creature card exiled this way, create a 2/2 black Zombie creature token.", false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			doc, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			effect := doc.Abilities[0].Sentences[0].Effects[0]
			if effect.Exact != tc.exact {
				t.Fatalf("exact=%v context=%v amount=%+v tokens=%q",
					effect.Exact, effect.Context, effect.Amount, joinedEffectText(effect.Tokens))
			}
		})
	}
}

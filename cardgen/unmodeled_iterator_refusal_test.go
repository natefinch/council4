package cardgen

import "testing"

func TestUnmodeledIteratorDomainsRemainWholeCardRefusals(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"For each Equipment attached to this creature, create a 1/1 white Soldier creature token.",
		"For each nontoken creature you controlled that died this turn, create a 2/2 black Zombie creature token.",
		"Exile X target creature cards from your graveyard. Then for each creature card exiled this way, create a 2/2 black Zombie creature token.",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, targetCountTestCard(body), "unsupported")
		})
	}
}

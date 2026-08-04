# Card-Support Backlog

Every eligible Scryfall corpus card is evaluated with two signals and routed to the layer that blocks it:

- **Parser signal** (parser-only): `cardgen.ParseCardFaces` + `parser.DocumentCoverage` — is the card parser-complete, and which uncovered components remain?
- **Lowering signal** (full compile): compilecards' canonical report — did the card generate, and if not, which distinct diagnostic summaries blocked lowering? compilecards is the authority; an independent per-card recompile reconciles against it.

This produces two ranked, actionable queues. Regenerate with `mage cardBacklog`.

## Headline

- Eligible cards: 32508
- Supported (generated): 18184
- Parser-complete: 22316
- **Lowering backlog** (parser-complete, not generated): 4896
- **Parser backlog** (not parser-complete, not generated): 9428

Partition check: 18184 supported + 4896 lowering-backlog + 9428 parser-backlog = 32508 eligible. ✓

764 generated cards are not parser-complete. The lowerer fully generates them, but the parser-coverage harness does not span all their must-cover tokens (the residue tracked in `parser-coverage.md`). They are counted as **supported**, not routed to either backlog queue:

- Veil of Summer
- Harald, King of Skemfar
- Stormchaser's Talent
- Puppeteer Clique
- Lantern of Revealing
- Deny the Divine
- Dawnbringer Cleric
- Gix's Caress
- Bedlam Reveler
- Cogwork Assembler
- Chain of Plasma
- Benefaction of Rhonas
- Lore Weaver
- Junk Golem
- Thirst for Identity
- Waterspout Djinn
- Finest Hour
- Pongify
- Defiant Stand
- Rapid Hybridization
- Phantasmal Forces
- Bothersome Quasit
- Ashiok's Forerunner
- Summon: Magus Sisters
- Chorus of Might
- Colossus of the Blood Age
- Chaos Warp
- Assert Authority
- Silvar, Devourer of the Free
- Rune of Protection: Blue
- Phantom Nantuko
- Requisition Raid
- Katara, the Fearless
- Rotting Giant
- Rhino's Rampage
- Timin, Youthful Geist
- Zulaport Enforcer
- Yenna, Redtooth Regent
- Jaheira, Friend of the Forest
- Pillage
- Chain of Acid
- Lively Dirge
- Kamber, the Plunderer
- Student of Warfare
- Niambi, Faithful Healer
- Coralhelm Commander
- Titanic Ultimatum
- Triumphant Reckoning
- Teferi's Protection
- Faerie Impostor

### Reconciliation guard

Generated membership is read from compilecards' canonical report. An independent per-card recompile cross-checks it; the run fails if they diverge.

- Authoritative generated (compilecards report): 18184
- Independent per-card recompile generated: 18184
- Divergences: 0 — the two pipelines agree. ✓

## Lowering queue

Parser-complete cards that do not yet lower, bucketed by distinct lowering diagnostic summary and ranked by affected-card count. Parsing is already done for these cards, so they are the lowest-risk backlog: this is `unsupported-reasons.md` restricted to the parser-complete subset.

| Rank | Reason | Affected (parser-complete) cards | Sole blockers | Example cards |
| --- | --- | --- | --- | --- |
| 1 | unsupported ordered effect sequence | 1681 | 1135 | Wasp, Shrinking Savior; Strength of Night; Mind Extraction; Coalition Relic; Fear of Falling |
| 2 | unsupported optional effect | 506 | 11 | Dazzling Sphinx; Mindclaw Shaman; Remembrance; Park Bleater; Moonring Mirror |
| 3 | unsupported static declaration operation | 276 | 232 | Food Fight; Magma Sliver; Sedge Sliver; Wingrattle Scarecrow; Pompous Gadabout |
| 4 | unsupported static ability | 267 | 184 | Nissa, Worldsoul Speaker; Static Orb; Stenn, Paranoid Partisan; Beluna Grandsquall // Seek Thrills; Avatar of Growth |
| 5 | unsupported counter placement | 204 | 119 | Toluz, Clever Conductor; Sword-Swallowing Seraph; Greater Werewolf; Ent-Draught Basin; Park Bleater |
| 6 | unsupported damage spell | 187 | 158 | Armed Response; Combo Attack; Huatli, Dinosaur Knight; Kill! Maim! Burn!; Call Forth the Tempest |
| 7 | unsupported activation cost | 181 | 128 | Thunderherd Migration; Krovikan Sorcerer; Etchings of the Chosen; Tourach's Gate; City of Shadows |
| 8 | unsupported return spell | 161 | 134 | Dragon Fangs; Dragon Scales; Venser's Diffusion; Scapegoat; Kazandu Stomper |
| 9 | unsupported static declaration group | 160 | 123 | Sedge Sliver; Freewind Equenaut; Rune of Sustenance; Indomitable Might; Cast Through Time |
| 10 | unsupported destroy spell | 128 | 111 | Coils of the Medusa; Unliving Psychopath; Bounty Agent; Rampaging War Mammoth; Feline Sovereign |
| 11 | unsupported search effect | 119 | 78 | Remembrance; Kasmina, Enigma Sage; Avatar of Growth; Increasing Ambition; Quest for the Holy Relic |
| 12 | unsupported token creation | 113 | 86 | Witch's Mark; Goblin Gathering; Sorin, Grim Nemesis; Nesting Dragon; Kibo, Uktabi Prince |
| 13 | unsupported activation ability word | 108 | 100 | Half-Elf Monk; Sagu Pummeler; Blazing Bomb; Champion of Dusan; Red Death, Shipwrecker |
| 14 | unsupported ability word | 96 | 84 | Bloodthorn Flail; The Dalek Emperor; Solar Tide; Terror Tide; Ensnared by the Mara |
| 15 | unsupported enters-tapped replacement | 90 | 55 | Stenn, Paranoid Partisan; Choco-Comet; Nevermore; True-Name Nemesis; Jailbreak |
| 16 | unsupported exile spell | 88 | 63 | Ravnica at War; Toluz, Clever Conductor; Consuming Sinkhole; Sengir Autocrat; Ulamog, the Ceaseless Hunger |
| 17 | unsupported power/toughness spell | 87 | 67 | Murk Dwellers; Park Bleater; Shaper Parasite; Battle Frenzy; Blood Age General |
| 18 | unsupported activation references | 75 | 63 | Planebound Accomplice; Puresight Merrow; Titans' Nest; Spurnmage Advocate; Pulsemage Advocate |
| 19 | unsupported gain-control spell | 66 | 53 | Slave of Bolas; Legacy's Allure; The Super Hero Civil War; Dominating Vampire; Skyfire Kirin |
| 20 | unsupported enters-with-counters replacement | 64 | 47 | Flycatcher Giraffid; Malefic Scythe; Callous Sell-Sword // Burn Together; Bone Devourer; Faerie Squadron |
| 21 | unsupported type line | 61 | 60 | Playable Delusionary Hydra; Notorious Sliver War; City's Blessing // Elemental; Demonic Tourist Laser; Night Brushwagg Ringmaster |
| 22 | unsupported cast effect | 58 | 30 | Oracle of Bones; Founding the Third Path; Forger's Foundry; Spell Queller; Xantid Swarm |
| 23 | unsupported life spell | 56 | 51 | Guiltfeeder; Wall of Reverence; Revered Unicorn; Atarka's Command; Netherborn Phalanx |
| 24 | unsupported temporary keyword spell | 56 | 48 | Order of the Golden Cricket; Pale Wayfarer; Violent Urge; Outmuscle; Gravity Negator |
| 25 | unsupported ability content | 55 | 46 | Vihaan, Goldwaker; Heated Debate; Renegade Doppelganger; Shifting Loyalties; Symmetry Sage |
| 26 | unsupported static declaration condition | 47 | 37 | Desperate Castaways; Nadaar, Selfless Paladin; Veldt; Gloom Stalker; Hazy Homunculus |
| 27 | unsupported draw spell | 42 | 34 | Theft of Dreams; Fatigue; Gregor, Shrewd Magistrate; Nessian Boar; Thought Sponge |
| 28 | unsupported library placement | 40 | 33 | Misinformation; Chittering Rats; Murderous Rider // Swift End; Landscaper Colos; God-Eternal Bontu |
| 29 | unsupported sacrifice spell | 38 | 32 | Yukora, the Prisoner; Demonic Taskmaster; Burning Sands; Papalymo Totolymo; Defiler of Souls |
| 30 | unsupported mixed keyword ability | 38 | 31 | Chief Engineer; Sky Tether; Radiant Destiny; Mystic Decree; Wicker Picker |
| 31 | unsupported attach effect | 35 | 30 | Crown of the Ages; Ronin Warclub; Illusory Gains; Beatrix, Loyal General; Prison Term |
| 32 | unsupported mana effect | 34 | 28 | Dictate of Karametra; Interplanar Beacon; Veldt; Market Festival; Skycloud Egg |
| 33 | unsupported activation condition | 30 | 29 | Ebon Praetor; Inner-Flame Igniter; Everflame Eidolon; Roadside Reliquary; Arch of Orazca |
| 34 | unsupported counter spell | 29 | 24 | Spell Blast; Drown in the Loch; Unyaro Griffin; Hisoka's Defiance; Frontline Medic |
| 35 | unsupported shuffle effect | 28 | 26 | Dwell on the Past; Madblind Mountain; Perpetual Timepiece; Renewing Touch; Piper's Melody |
| 36 | unsupported parameterized keyword | 28 | 23 | Goblin Barrage; Vexing Scuttler; Ulamog's Dreadsire; Garruk's Harbinger; Sporeweb Weaver |
| 37 | unsupported tap spell | 28 | 22 | Torrent Elemental; Gridlock; Dawnglare Invoker; Tectonic Instability; Arena of the Ancients |
| 38 | unsupported mana symbol | 23 | 21 | Pit of Offerings; Blinkmoth Urn; Elemental Resonance; Inner Fire; Songs of the Damned |
| 39 | unsupported activation timing | 21 | 18 | Vivi Ornitier; Hall of Oracles; Tomb Tyrant; In the Trenches; Desert |
| 40 | unsupported card layout | 20 | 20 | Nezumi Graverobber // Nighteyes the Desecrator; Faithful Squire // Kaiso, Memory of Loyalty; Jushi Apprentice // Tomoya the Revealer; Cunning Bandit // Azamuki, Treachery Incarnate; Nezumi Shortfang // Stabwhisker the Odious |
| 41 | unsupported delayed effect | 19 | 16 | Silent Assassin; Wicker Warcrawler; Marchesa, the Black Rose; Rienne, Angel of Rebirth; Ghoulish Impetus |
| 42 | unsupported keyword or ability grant | 19 | 13 | Furystoke Giant; Huatli, Poet of Unity // Roar of the Fifth People; Mist Dragon; Urza's Saga; Quicksmith Spy |
| 43 | unsupported optional replacement effect | 18 | 16 | Parallel Thoughts; Mocking Doppelganger; The Mimeoplasm; Jinnie Fay, Jetmir's Second; Arsenal Thresher |
| 44 | unsupported keyword or ability loss | 15 | 11 | Cephalid Snitch; Scarwood Hag; Final Act; Mist Dragon; Torpid Moloch |
| 45 | unsupported triggered ability | 15 | 11 | Jeering Instigator; Siege Dragon; Spectral Force; Kiyomaro, First to Stand; Tephraderm |
| 46 | unsupported Oracle construct | 15 | 0 | Tetsuo, Imperial Champion; Demolition Stomper; Glacierwood Siege; Vision, Synthezoid Avenger; Reverence |
| 47 | unsupported permanent zone-change trigger | 14 | 14 | Reluctant Dounguard; Scrapshooter; Starforged Sword; Ichorplate Golem; Wretched Camel |
| 48 | unsupported untap spell | 14 | 10 | Magus of the Candelabra; Early Harvest; Reality Spasm; The Thirteenth Doctor; Urtet, Remnant of Memnarch |
| 49 | unsupported phase/step trigger phrase effect | 13 | 13 | Umaro, Raging Yeti; Quiet Disrepair; Sylvan Scavenging; Mister Hyde, Monster Within; Ferocification |
| 50 | unsupported group power/toughness spell | 13 | 10 | Bloodline Culling; Thran Weaponry; Rabble-Rouser; Mercadia's Downfall; Firebird, Blazing Ranger |
| 51 | unsupported emblem ability | 13 | 6 | Zariel, Archduke of Avernus; Tezzeret, Cruel Captain; Chandra, Torch of Defiance; Kaya the Inexorable; Koth, Fire of Resistance |
| 52 | unsupported can't-block effect | 12 | 12 | Blinding Flare; Temur Charm; Manacles of Decay; Goma Fada Vanguard; Mournwillow |
| 53 | unsupported discard spell | 12 | 12 | Tormented Thoughts; Warped Devotion; Zhang Liao, Hero of Hefei; Cabal Conditioning; Jagged Poppet |
| 54 | unsupported multiple spell abilities | 12 | 12 | Orcish Medicine; Agony Warp; Force Away; Incinerating Blast; Bounty of Might |
| 55 | unsupported alternative spell cost | 12 | 9 | Nethergoyf; Conflagrate; Nourishing Shoal; Sickening Shoal; Spinning Darkness |
| 56 | unsupported can't-be-blocked effect | 11 | 10 | Gingerbrute; Speed, Young Avenger; Runed Arch; Leitmotif Composer; Secret Tunnel |
| 57 | unsupported overload effect | 11 | 5 | Mizzium Skin; Corporeal Projection; Mind Rake; Break the Ice; Weapon Surge |
| 58 | unsupported manifest spell | 10 | 8 | Orcish Spy; They Came from the Pipes; Smoke Teller; Omarthis, Ghostfire Initiate; Etrata, Deadly Fugitive |
| 59 | validation failed: invalid-ability-body | 9 | 9 | Sachi, Daughter of Seshiro; Gift of Paradise; Forgotten Monument; Find the Path; New Horizons |
| 60 | unsupported enters-as-copy replacement | 8 | 8 | Mercurial Pretender; Pirated Copy; Masterwork of Ingenuity; Callidus Assassin; Gigantoplasm |

## Parser queue

Cards that are not parser-complete (and do not lower), bucketed by owning component family and normalized uncovered-span cluster, ranked by occurrence. This is the grammar-recognition backlog.

| Rank | Component | Cluster | Count | Example cards |
| --- | --- | --- | --- | --- |
| 1 | condition | if able | 105 | Impetuous Devils; Nacatl Hunt-Pride; Culling Mark; Legion Warboss; Magitek Scythe |
| 2 | effect | it's still a land. | 70 | Hall of Storm Giants; Restless Vinestalk; Vastwood Animist; Restless Spire; Embodiment of Fury |
| 3 | effect | level N | 58 | Builder's Talent; Leader's Talent; Sorcerer Class; Fortune Teller's Talent; Cool but Rude |
| 4 | effect | you may choose new targets for the copy. | 53 | Melek, Izzet Paragon; Lithoform Engine; Echoes of Eternity; Fire Lord Azula; Najal, the Storm Runner |
| 5 | effect | it can't be regenerated. | 51 | Polymorph; Phage the Untouchable; Wooden Stake; Shivan Emissary; Phyrexian Reaper |
| 6 | other | partner | 48 | Ghost of Ramirez DePietro; Sophina, Spearsage Deserter; Silas Renn, Seeker Adept; Krark, the Thumbless; Vial Smasher the Fierce |
| 7 | condition | if you search your library this way | 42 | Vraska's Scorn; Claim Jumper; Fang-Druid Summoner; Grand Master of Flowers; Invasion of Ikoria // Zilortha, Apex of Ikoria |
| 8 | effect | daybound (if a player casts no spells during their own turn, it becomes night next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 9 | effect | nightbound (if a player casts at least two spells during their own turn, it becomes day next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 10 | condition | if a player cast two or more spells last turn | 33 | Ulrich of the Krallenhorde // Ulrich, Uncontested Alpha; Lambholt Elder // Silverpelt Werewolf; Instigator Gang // Wildblood Pack; Daybreak Ranger // Nightfall Predator; Hinterland Logger // Timber Shredder |
| 11 | effect | any player may activate this ability. | 33 | Flailing Manticore; Vintara Elephant; Xantcha, Sleeper Agent; Casey Jones, Asphalt Hooligan; Deadly Designs |
| 12 | condition | if it's a creature card | 30 | Search for Survivors; Hauntwoods Shrieker; Sapling of Colfenor; Domri Rade; Llanowar Empath |
| 13 | condition | if it's a land card | 30 | Unexpected Results; Countryside Crusher; Skyclave Aerialist // Skyclave Invader; Nissa, Vastwood Seer // Nissa, Sage Animist; Thrasios, Triton Hero |
| 14 | effect | take an extra turn after this one. | 29 | Twice Upon a Time // Unlikely Meeting; Temporal Extortion; The Legend of Kuruk // Avatar Kuruk; Alchemist's Gambit; Chance for Glory |
| 15 | condition | if you win | 28 | Research the Deep; Woodland Guidance; Titan's Revenge; Captivating Glance; Sentry Oak |
| 16 | effect | they can't be regenerated. | 28 | Wave of Terror; Tsabo's Decree; Kirtar's Wrath; Reign of Terror; Spreading Plague |
| 17 | effect | aftermath (cast this spell only from your graveyard. then exile it.) | 27 | Heaven // Earth; Struggle // Survive; Claim // Fame; Farm // Market; Appeal // Authority |
| 18 | effect | clash with an opponent. | 25 | Research the Deep; Woodland Guidance; Titan's Revenge; Captivating Glance; Pollen Lullaby |
| 19 | effect | station (tap another creature you control: put charge counters equal to its power on this spacecraft. station only as a sorcery. it's an artifact creature at N +.) | 25 | Wedgelight Rammer; Rescue Skiff; Fell Gravship; Wurmwall Sweeper; Lumen-Class Frigate |
| 20 | effect | you may exert this creature as it attacks. | 25 | Nef-Crop Entangler; Clockwork Droid; Vizier of the True; Watchful Naga; Resolute Survivors |
| 21 | effect | choose one. | 24 | Wail of the Forgotten; Go Nuts!; See Double; Soul Transfer; Prophetic Titan |
| 22 | effect | soulbond (you may pair this creature with another unpaired creature when either enters. they remain paired for as long as you control both of them.) | 24 | Donna Noble; Tandem Lookout; Doom Weaver; Stonewright; Spectral Gateguards |
| 23 | effect | exploit (when this creature enters, you may sacrifice a creature.) | 23 | Fell Stinger; Rakshasa Gravecaller; Infernal Captor; Diver Skaab; Sidisi, Undead Vizier |
| 24 | effect | you may choose the same mode more than once. | 22 | Eldrazi Confluence; Doomsday Confluence; Verdant Confluence; Unite the Coalition; Obscura Confluence |
| 25 | other | choose a background | 22 | Halsin, Emerald Archdruid; Karlach, Fury of Avernus; Shadowheart, Dark Justiciar; Durnan of the Yawning Portal; Viconia, Drow Apostate |
| 26 | other | doctor's companion | 21 | Nyssa of Traken; Donna Noble; Bill Potts; Susan Foreman; Rose Tyler |
| 27 | condition | if you lose the flip | 20 | Goblin Bomb; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Risky Move; Mogg Assassin |
| 28 | effect | learn. | 20 | Professor of Symbology; Field Trip; Sparring Regimen; Igneous Inspiration; Enthusiastic Study |
| 29 | effect | choose target creature. | 19 | Arcbond; Fatal Fissure; Spark of Creativity; Sewers of Estark; Erratic Mutation |
| 30 | effect | open an attraction. | 19 | Complaints Clerk; Discourtesy Clerk; Myra the Magnificent; Soul Swindler; Coming Attraction |
| 31 | trigger | when this creature exploits a creature | 19 | Fell Stinger; Rakshasa Gravecaller; Infernal Captor; Diver Skaab; Sidisi's Faithful |
| 32 | condition | if you can't | 18 | Frankenstein's Monster; Ravenous Demon // Archdemon of Greed; Rust Elemental; Infernal Denizen; Out of the Tombs |
| 33 | effect | you choose a nonland card from it. | 18 | Grief; Memory Theft; Cerebral Confiscation; Check for Traps; Down for Repairs |
| 34 | condition | if you cast a spell this way | 17 | Eye of Duskmantle; Brainstealer Dragon; Noctis, Prince of Lucis; Xander's Pact; Outrageous Robbery |
| 35 | effect | buyback{3}(you may pay an additional{3}as you cast this spell. if you do, put this card into your hand as it resolves.) | 17 | Spell Burst; Fanning the Flames; Invulnerability; Seething Anger; Reiterate |
| 36 | effect | choose a color. | 17 | Skrelv, Defector Mite; Radiant Lotus; Addle; Glory; Sungold Sentinel |
| 37 | effect | flip a coin. | 17 | Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Risky Move; Plasma Caster; Mogg Assassin |
| 38 | effect | teamwork N (as an additional cost to cast this spell, you may tap any number of creatures you control with total power N or more.) | 17 | Go Nuts!; Crossover Collaboration; Repulsor Blast; We Say Thee Nay!; Earth's Mightiest Heroes |
| 39 | effect | x can't be 0. | 17 | Lair of the Hydra; Helm of Obedience; Katara, Water Tribe's Hope; Aladdin's Lamp; Benalish Commander |
| 40 | condition | as long as this creature is paired with another creature | 16 | Spectral Gateguards; Nearheath Pilgrim; Wingcrafter; Hanweir Lancer; Trusted Forcemage |
| 41 | condition | if a player does | 16 | Temporal Extortion; Carrion Rats; Phantasmagorian; Brain Gorgers; Shivan Wumpus |
| 42 | effect | N +\| flying | 16 | Rescue Skiff; Wurmwall Sweeper; Inspirit, Flagship Vessel; Uthros Scanship; Exploration Broodship |
| 43 | effect | backup N (when this creature enters, put a +1/+1 counter on target creature. if that's another creature, it gains the following ability until end of turn.) | 16 | Saiba Cryptomancer; Chomping Kavu; Consuming Aetherborn; Serpent-Blade Assailant; Mirror-Style Master |
| 44 | effect | until end of turn, you don't lose this mana as steps and phases end. | 16 | Colossal Plow; Tanuki Transplanter; Rousing Refrain; Tundra Fumarole; Kessig Naturalist // Lord of the Ulvenwald |
| 45 | trigger | whenever you fully unlock a room | 16 | Optimistic Scavenger; Fear of Infinity; Scrabbling Skullcrab; Dashing Bloodsucker; Erratic Apparition |
| 46 | condition | if you win the flip | 15 | Goblin Bomb; Ral, Monsoon Mage // Ral, Leyline Prodigy; Plasma Caster; Mogg Assassin; Impulsive Maneuvers |
| 47 | effect | choose a card name. | 15 | Tunnel Vision; The Clone Saga; Unmoored Ego; Cheering Fanatic; Liar's Pendulum |
| 48 | effect | choose an opponent. | 15 | Triarch Stalker; Infernal Offering; Benevolent Offering; Slithermuse; Sylvan Offering |
| 49 | effect | cipher (then you may exile this spell card encoded on a creature you control. whenever that creature deals combat damage to a player, its controller may cast a copy of the encoded card without paying its mana cost.) | 15 | Paranoid Delusions; Stolen Identity; Trait Doctoring; Writ of Return; Last Thoughts |
| 50 | effect | job select (when this equipment enters, create a 1/1 colorless hero creature token, then attach this to it.) | 15 | Monk's Fist; Paladin's Arms; Red Mage's Rapier; Bard's Bow; Thief's Knife |
| 51 | other | {1}— | 15 | Unfortunate Accident; Rush of Dread; One Last Job; Smuggler's Surprise; Shifting Grift |
| 52 | condition | if that spell would be put into a graveyard | 14 | Ogre Battlecaster; Impulsivity; Toshiro Umezawa; Halo Forager; Deluxe Dragster |
| 53 | condition | if that spell would be put into your graveyard | 14 | Mavinda, Students' Advocate; Sword of Once and Future; Power Pack; Vohar, Vodalian Desecrator; Dreadhorde Arcanist |
| 54 | condition | if this card is in your graveyard | 14 | Ghastly Remains; Bridge from Below; Master of Death; Ichorid; Jocasta, Automaton Avenger |
| 55 | condition | if this spell was cast using teamwork | 14 | Go Nuts!; Crossover Collaboration; Repulsor Blast; Earth's Mightiest Heroes; Murdock's Crusade |
| 56 | effect | hidden agenda (start the game with this conspiracy face down in the command zone and secretly choose a card name. you may turn this conspiracy face up any time and reveal that name.) | 14 | Muzzio's Preparations; Secret Summoning; Unexpected Potential; Echoing Boon; Brago's Favor |
| 57 | effect | you may pay{1}. | 14 | Ancestral Katana; Dutiful Replicator; Nadir Kraken; Oloro, Ageless Ascetic; Valentin, Dean of the Vein // Lisette, Dean of the Root |
| 58 | condition | if a nonland permanent left the battlefield this turn or a spell was warped this turn | 13 | Decode Transmissions; Temporal Intervention; Hylderblade; Insatiable Skittermaw; Roving Actuator |
| 59 | effect | after this phase, there is an additional combat phase. | 13 | Karlach, Fury of Avernus; Scourge of the Throne; Najeela, the Blade-Blossom; Bumi, Unleashed; Illusionist's Gambit |
| 60 | effect | choose target creature you control. | 13 | Heroic Sacrifice; Mandate of Abaddon; Loki, Lord of Misrule; Hall of Mirrors; Eidolon of Astral Winds |


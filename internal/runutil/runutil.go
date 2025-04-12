// Package runutil provides utilities for generating and managing run names
package runutil

import (
	"math/rand"
	"time"
)

// randomSource is a dedicated random source for generating run names
var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// List of adjectives for generating run names
var adjectives = []string{
	"amber", "ancient", "autumn", "bold", "brave", "bright", "calm", "clever",
	"coastal", "cosmic", "crimson", "crystal", "curious", "daring", "deep",
	"distant", "eager", "elegant", "emerald", "enchanted", "endless", "energetic",
	"ethereal", "fierce", "floral", "flowing", "flying", "forest", "gentle",
	"golden", "graceful", "grand", "hidden", "humming", "icy", "infinite",
	"jade", "jolly", "kind", "lively", "lunar", "majestic", "melodic", "mighty",
	"misty", "morning", "mountain", "noble", "ocean", "patient", "peaceful",
	"playful", "poetic", "proud", "purple", "quick", "quiet", "radiant", "rapid",
	"royal", "ruby", "rustic", "serene", "shining", "silent", "silver", "sincere",
	"singing", "skillful", "sleepy", "smiling", "snowy", "solar", "sparkling",
	"spring", "starry", "steadfast", "stormy", "summer", "sunny", "swift",
	"thoughtful", "thundering", "tranquil", "twilight", "vibrant", "wandering",
	"whispering", "wild", "winter", "wise", "zephyr",
}

// List of nouns for generating run names
var nouns = []string{
	"acorn", "archipelago", "aurora", "badger", "beacon", "bear", "birch",
	"bison", "brook", "buzzard", "canyon", "cardinal", "cascade", "cave",
	"cheetah", "cliff", "cloud", "coast", "condor", "coral", "cove", "crater",
	"creek", "dawn", "deer", "delta", "dolphin", "dove", "dragon", "driftwood",
	"dusk", "eagle", "elm", "falcon", "fern", "firefly", "fjord", "flower",
	"forest", "galaxy", "gazelle", "geyser", "glacier", "grove", "harbor",
	"hawk", "heron", "hill", "horizon", "ibex", "iceberg", "island", "jackal",
	"jaguar", "jay", "journey", "koala", "lagoon", "lantern", "leopard", "lighthouse",
	"lightning", "lynx", "maple", "marsh", "meadow", "meteor", "mist", "moon",
	"moose", "mountain", "nebula", "ocean", "osprey", "otter", "owl", "panther",
	"path", "peak", "penguin", "phoenix", "pine", "planet", "plateau", "puma",
	"quail", "rabbit", "raccoon", "rain", "raven", "reef", "ridge", "river",
	"robin", "rock", "satellite", "sea", "sequoia", "shadow", "shore", "sparrow",
	"squirrel", "star", "storm", "stream", "summit", "sun", "sunset", "swift",
	"thunder", "tiger", "tortoise", "trail", "valley", "vapor", "volcano", "wave",
	"willow", "wind", "wolf", "zenith",
}

// GenerateRunName creates a random adjective-noun combination suitable for use
// as a run name or directory name. The result follows the pattern "adjective-noun"
// with all lowercase and a hyphen as separator.
func GenerateRunName() string {
	adjective := adjectives[randomSource.Intn(len(adjectives))]
	noun := nouns[randomSource.Intn(len(nouns))]

	return adjective + "-" + noun
}

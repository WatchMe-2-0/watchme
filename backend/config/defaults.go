package config

// Avatar presets available for profiles
var AvatarPresets = []string{
	"aurora",
	"nebula",
	"comet",
	"galaxy",
	"nova",
	"pulsar",
	"quasar",
	"stellar",
	"orbit",
	"prism",
	"flux",
	"echo",
}

// Genre mapping for TMDB
var GenreMap = map[int]string{
	28:    "Action",
	12:    "Adventure",
	16:    "Animation",
	35:    "Comedy",
	80:    "Crime",
	99:    "Documentary",
	18:    "Drama",
	10751: "Family",
	14:    "Fantasy",
	36:    "History",
	27:    "Horror",
	10402: "Music",
	9648:  "Mystery",
	10749: "Romance",
	878:   "Science Fiction",
	10770: "TV Movie",
	53:    "Thriller",
	10752: "War",
	37:    "Western",
}

// Kids-safe certifications (MPAA)
var KidsSafeCertifications = []string{"G", "PG"}

// MaxPINLength is the maximum PIN length allowed
const MaxPINLength = 4

// MinPINLength is the minimum PIN length allowed
const MinPINLength = 3

// AdminPasswordLength is the required admin password length
const AdminPasswordLength = 6

// RecoveryKeyLength is the length of the generated recovery key in bytes (32 = 256-bit)
const RecoveryKeyLength = 32

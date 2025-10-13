package appdata

import (
	"gorm.io/gorm"
)

var DB *gorm.DB
var SmtpServer string
var SmtpUsername string
var SmtpPassword string
var SmtpPort uint
var JwtExpiryMinutes uint
var JwtSecret []byte
var RefreshExpiryMinutes uint
var RefreshExpiryNoRemember uint
var JwtExpiryNoRemember uint
var ResetValidMinutes uint
var LogRequests bool

const BookCount uint = 66
const OtCount uint = 39
const NtCount uint = 27

type Book struct {
	Book         string
	Abbreviation string
	Testament    uint8 // 1 = OT, 2 = NT
	Chapters     uint
	Verses       uint
}

var AvailableTranslations = []string{"TOVBSI", "KJV", "MSLVP", "ASV", "WEB", "WEBU", "GOVBSI", "OOVBSI"}

var Books = []Book{
	{
		Book:         "Genesis",
		Abbreviation: "GEN",
		Testament:    1,
		Chapters:     50,
		Verses:       1533,
	},
	{
		Book:         "Exodus",
		Abbreviation: "EXO",
		Testament:    1,
		Chapters:     40,
		Verses:       1213,
	},
	{
		Book:         "Leviticus",
		Abbreviation: "LEV",
		Testament:    1,
		Chapters:     27,
		Verses:       859,
	},
	{
		Book:         "Numbers",
		Abbreviation: "NUM",
		Testament:    1,
		Chapters:     36,
		Verses:       1288,
	},
	{
		Book:         "Deuteronomy",
		Abbreviation: "DEU",
		Testament:    1,
		Chapters:     34,
		Verses:       959,
	},
	{
		Book:         "Joshua",
		Abbreviation: "JOS",
		Testament:    1,
		Chapters:     24,
		Verses:       658,
	},
	{
		Book:         "Judges",
		Abbreviation: "JDG",
		Testament:    1,
		Chapters:     21,
		Verses:       618,
	},
	{
		Book:         "Ruth",
		Abbreviation: "RUT",
		Testament:    1,
		Chapters:     4,
		Verses:       85,
	},
	{
		Book:         "1 Samuel",
		Abbreviation: "1SA",
		Testament:    1,
		Chapters:     31,
		Verses:       810,
	},
	{
		Book:         "2 Samuel",
		Abbreviation: "2SA",
		Testament:    1,
		Chapters:     24,
		Verses:       695,
	},
	{
		Book:         "1 Kings",
		Abbreviation: "1KI",
		Testament:    1,
		Chapters:     22,
		Verses:       816,
	},
	{
		Book:         "2 Kings",
		Abbreviation: "2KI",
		Testament:    1,
		Chapters:     25,
		Verses:       719,
	},
	{
		Book:         "1 Chronicles",
		Abbreviation: "1CH",
		Testament:    1,
		Chapters:     29,
		Verses:       942,
	},
	{
		Book:         "2 Chronicles",
		Abbreviation: "2CH",
		Testament:    1,
		Chapters:     36,
		Verses:       822,
	},
	{
		Book:         "Ezra",
		Abbreviation: "EZR",
		Testament:    1,
		Chapters:     10,
		Verses:       280,
	},
	{
		Book:         "Nehemiah",
		Abbreviation: "NEH",
		Testament:    1,
		Chapters:     13,
		Verses:       406,
	},
	{
		Book:         "Esther",
		Abbreviation: "EST",
		Testament:    1,
		Chapters:     10,
		Verses:       167,
	},
	{
		Book:         "Job",
		Abbreviation: "JOB",
		Testament:    1,
		Chapters:     42,
		Verses:       1070,
	},
	{
		Book:         "Psalm",
		Abbreviation: "PSA",
		Testament:    1,
		Chapters:     150,
		Verses:       2461,
	},
	{
		Book:         "Proverbs",
		Abbreviation: "PRO",
		Testament:    1,
		Chapters:     31,
		Verses:       915,
	},
	{
		Book:         "Ecclesiastes",
		Abbreviation: "ECC",
		Testament:    1,
		Chapters:     12,
		Verses:       222,
	},
	{
		Book:         "Song of Solomon",
		Abbreviation: "SNG",
		Testament:    1,
		Chapters:     8,
		Verses:       117,
	},
	{
		Book:         "Isaiah",
		Abbreviation: "ISA",
		Testament:    1,
		Chapters:     66,
		Verses:       1292,
	},
	{
		Book:         "Jeremiah",
		Abbreviation: "JER",
		Testament:    1,
		Chapters:     52,
		Verses:       1364,
	},
	{
		Book:         "Lamentations",
		Abbreviation: "LAM",
		Testament:    1,
		Chapters:     5,
		Verses:       154,
	},
	{
		Book:         "Ezekiel",
		Abbreviation: "EZK",
		Testament:    1,
		Chapters:     48,
		Verses:       1273,
	},
	{
		Book:         "Daniel",
		Abbreviation: "DAN",
		Testament:    1,
		Chapters:     12,
		Verses:       357,
	},
	{
		Book:         "Hosea",
		Abbreviation: "HOS",
		Testament:    1,
		Chapters:     14,
		Verses:       197,
	},
	{
		Book:         "Joel",
		Abbreviation: "JOL",
		Testament:    1,
		Chapters:     3,
		Verses:       73,
	},
	{
		Book:         "Amos",
		Abbreviation: "AMO",
		Testament:    1,
		Chapters:     9,
		Verses:       146,
	},
	{
		Book:         "Obadiah",
		Abbreviation: "OBA",
		Testament:    1,
		Chapters:     1,
		Verses:       21,
	},
	{
		Book:         "Jonah",
		Abbreviation: "JON",
		Testament:    1,
		Chapters:     4,
		Verses:       48,
	},
	{
		Book:         "Micah",
		Abbreviation: "MIC",
		Testament:    1,
		Chapters:     7,
		Verses:       105,
	},
	{
		Book:         "Nahum",
		Abbreviation: "NAM",
		Testament:    1,
		Chapters:     3,
		Verses:       47,
	},
	{
		Book:         "Habakkuk",
		Abbreviation: "HAB",
		Testament:    1,
		Chapters:     3,
		Verses:       56,
	},
	{
		Book:         "Zephaniah",
		Abbreviation: "ZEP",
		Testament:    1,
		Chapters:     3,
		Verses:       53,
	},
	{
		Book:         "Haggai",
		Abbreviation: "HAG",
		Testament:    1,
		Chapters:     2,
		Verses:       38,
	},
	{
		Book:         "Zechariah",
		Abbreviation: "ZEC",
		Testament:    1,
		Chapters:     14,
		Verses:       211,
	},
	{
		Book:         "Malachi",
		Abbreviation: "MAL",
		Testament:    1,
		Chapters:     4,
		Verses:       55,
	},
	{
		Book:         "Matthew",
		Abbreviation: "MAT",
		Testament:    2,
		Chapters:     28,
		Verses:       1071,
	},
	{
		Book:         "Mark",
		Abbreviation: "MRK",
		Testament:    2,
		Chapters:     16,
		Verses:       678,
	},
	{
		Book:         "Luke",
		Abbreviation: "LUK",
		Testament:    2,
		Chapters:     24,
		Verses:       1151,
	},
	{
		Book:         "John",
		Abbreviation: "JHN",
		Testament:    2,
		Chapters:     21,
		Verses:       879,
	},
	{
		Book:         "Acts",
		Abbreviation: "ACT",
		Testament:    2,
		Chapters:     28,
		Verses:       1007,
	},
	{
		Book:         "Romans",
		Abbreviation: "ROM",
		Testament:    2,
		Chapters:     16,
		Verses:       433,
	},
	{
		Book:         "1 Corinthians",
		Abbreviation: "1CO",
		Testament:    2,
		Chapters:     16,
		Verses:       437,
	},
	{
		Book:         "2 Corinthians",
		Abbreviation: "2CO",
		Testament:    2,
		Chapters:     13,
		Verses:       257,
	},
	{
		Book:         "Galatians",
		Abbreviation: "GAL",
		Testament:    2,
		Chapters:     6,
		Verses:       149,
	},
	{
		Book:         "Ephesians",
		Abbreviation: "EPH",
		Testament:    2,
		Chapters:     6,
		Verses:       155,
	},
	{
		Book:         "Philippians",
		Abbreviation: "PHP",
		Testament:    2,
		Chapters:     4,
		Verses:       104,
	},
	{
		Book:         "Colossians",
		Abbreviation: "COL",
		Testament:    2,
		Chapters:     4,
		Verses:       95,
	},
	{
		Book:         "1 Thessalonians",
		Abbreviation: "1TH",
		Testament:    2,
		Chapters:     5,
		Verses:       89,
	},
	{
		Book:         "2 Thessalonians",
		Abbreviation: "2TH",
		Testament:    2,
		Chapters:     3,
		Verses:       47,
	},
	{
		Book:         "1 Timothy",
		Abbreviation: "1TI",
		Testament:    2,
		Chapters:     6,
		Verses:       113,
	},
	{
		Book:         "2 Timothy",
		Abbreviation: "2TI",
		Testament:    2,
		Chapters:     4,
		Verses:       83,
	},
	{
		Book:         "Titus",
		Abbreviation: "TIT",
		Testament:    2,
		Chapters:     3,
		Verses:       46,
	},
	{
		Book:         "Philemon",
		Abbreviation: "PHM",
		Testament:    2,
		Chapters:     1,
		Verses:       25,
	},
	{
		Book:         "Hebrews",
		Abbreviation: "HEB",
		Testament:    2,
		Chapters:     13,
		Verses:       303,
	},
	{
		Book:         "James",
		Abbreviation: "JAM",
		Testament:    2,
		Chapters:     5,
		Verses:       108,
	},
	{
		Book:         "1 Peter",
		Abbreviation: "1PE",
		Testament:    2,
		Chapters:     5,
		Verses:       105,
	},
	{
		Book:         "2 Peter",
		Abbreviation: "2PE",
		Testament:    2,
		Chapters:     3,
		Verses:       61,
	},
	{
		Book:         "1 John",
		Abbreviation: "1JN",
		Testament:    2,
		Chapters:     5,
		Verses:       105,
	},
	{
		Book:         "2 John",
		Abbreviation: "2JN",
		Testament:    2,
		Chapters:     1,
		Verses:       13,
	},
	{
		Book:         "3 John",
		Abbreviation: "3JN",
		Testament:    2,
		Chapters:     1,
		Verses:       14,
	},
	{
		Book:         "Jude",
		Abbreviation: "JUD",
		Testament:    2,
		Chapters:     1,
		Verses:       25,
	},
	{
		Book:         "Revelation",
		Abbreviation: "REV",
		Testament:    2,
		Chapters:     22,
		Verses:       404,
	},
}

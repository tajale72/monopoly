package types

var GameEnd = false

var YesChoice = map[string]bool{
	"yes": true,
	"Yes": true,
	"YES": true,
	"y":   true,
	"Y":   true,
}

var NoChoice = map[string]bool{
	"no": true,
	"No": true,
	"NO": true,
	"n":  true,
	"N":  true,
}

var AllProperties = []map[string]int{
	{"Mediterranean Avenue": 60},
	{"Baltic Avenue": 60},
	{"Oriental Avenue": 100},
	{"Vermont Avenue": 100},
	{"Connecticut Avenue": 120},
	{"St. Charles Place": 140},
	{"States Avenue": 140},
	{"Virginia Avenue": 160},
	{"St. James Place": 180},
	{"Tennessee Avenue": 180},
	{"Free Parking": 0},
	{"New York Avenue": 200},
	{"Kentucky Avenue": 220},
	{"Indiana Avenue": 220},
	{"Illinois Avenue": 240},
	{"Atlantic Avenue": 260},
	{"Ventnor Avenue": 260},
	{"Marvin Gardens": 280},
	{"Pacific Avenue": 300},
	{"Jail": 0},
	{"North Carolina Avenue": 300},
	{"Pennsylvania Avenue": 320},
	{"Park Place": 350},
	{"Boardwalk": 400},
}

var TestDataProperty = map[int]Property{
	0:  {PropertyName: "Mediterranean Avenue", Price: 60, Rent: 2, Owner: ""},
	1:  {PropertyName: "Baltic Avenue", Price: 60, Rent: 4, Owner: ""},
	2:  {PropertyName: "Oriental Avenue", Price: 100, Rent: 6, Owner: ""},
	3:  {PropertyName: "Vermont Avenue", Price: 100, Rent: 6, Owner: ""},
	4:  {PropertyName: "Connecticut Avenue", Price: 120, Rent: 8, Owner: ""},
	5:  {PropertyName: "St. Charles Place", Price: 140, Rent: 10, Owner: ""},
	6:  {PropertyName: "States Avenue", Price: 140, Rent: 10, Owner: ""},
	7:  {PropertyName: "Virginia Avenue", Price: 160, Rent: 12, Owner: ""},
	8:  {PropertyName: "St. James Place", Price: 180, Rent: 14, Owner: ""},
	9:  {PropertyName: "Tennessee Avenue", Price: 180, Rent: 14, Owner: ""},
	10: {PropertyName: "Free Parking", Price: 0, Rent: 0, Owner: ""},
	11: {PropertyName: "New York Avenue", Price: 200, Rent: 16, Owner: ""},
	12: {PropertyName: "Kentucky Avenue", Price: 220, Rent: 18, Owner: ""},
	13: {PropertyName: "Indiana Avenue", Price: 220, Rent: 18, Owner: ""},
	14: {PropertyName: "Illinois Avenue", Price: 240, Rent: 20, Owner: ""},
	15: {PropertyName: "Atlantic Avenue", Price: 260, Rent: 22, Owner: ""},
	16: {PropertyName: "Ventnor Avenue", Price: 260, Rent: 22, Owner: ""},
	17: {PropertyName: "Marvin Gardens", Price: 280, Rent: 24, Owner: ""},
	18: {PropertyName: "Pacific Avenue", Price: 300, Rent: 26, Owner: ""},
	19: {PropertyName: "Jail", Price: 0, Rent: 0, Owner: ""},
	20: {PropertyName: "North Carolina Avenue", Price: 300, Rent: 26, Owner: ""},
	21: {PropertyName: "Pennsylvania Avenue", Price: 320, Rent: 28, Owner: ""},
	22: {PropertyName: "Park Place", Price: 350, Rent: 35, Owner: ""},
	23: {PropertyName: "Boardwalk", Price: 400, Rent: 50, Owner: ""},
}

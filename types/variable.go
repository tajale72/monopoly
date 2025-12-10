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

type Tile struct {
	Name  string
	Type  string // corner, property, rail, util, chance, chest, tax
	Band  string // color hex if property
	Icon  string // emoji
	Price int
	Rent  int
	Owner string
}

var Board = map[int]Tile{
	0:  {Name: "GO", Type: "corner", Icon: "🚩"},
	1:  {Name: "Mediterranean Avenue", Type: "property", Band: "#955436", Icon: "🏠", Price: 60, Rent: 2},
	2:  {Name: "Community Chest", Type: "chest", Icon: "🧰"},
	3:  {Name: "Baltic Avenue", Type: "property", Band: "#955436", Icon: "🏠", Price: 60, Rent: 4},
	4:  {Name: "Income Tax", Type: "tax", Icon: "💵"},
	5:  {Name: "Reading Railroad", Type: "rail", Icon: "🚂", Price: 200, Rent: 25},
	6:  {Name: "Oriental Avenue", Type: "property", Band: "#aae0fa", Icon: "🏢", Price: 100, Rent: 6},
	7:  {Name: "Chance", Type: "chance", Icon: "❓"},
	8:  {Name: "Vermont Avenue", Type: "property", Band: "#aae0fa", Icon: "🏢", Price: 100, Rent: 6},
	9:  {Name: "Connecticut Avenue", Type: "property", Band: "#aae0fa", Icon: "🏢", Price: 120, Rent: 8},
	10: {Name: "Jail / Just Visiting", Type: "corner", Icon: "🚓"},
	11: {Name: "St. Charles Place", Type: "property", Band: "#d93a96", Icon: "🏘️", Price: 140, Rent: 10},
	12: {Name: "Electric Company", Type: "util", Icon: "💡", Price: 150, Rent: 10},
	13: {Name: "States Avenue", Type: "property", Band: "#d93a96", Icon: "🏘️", Price: 140, Rent: 10},
	14: {Name: "Virginia Avenue", Type: "property", Band: "#d93a96", Icon: "🏘️", Price: 160, Rent: 12},
	15: {Name: "Pennsylvania Railroad", Type: "rail", Icon: "🚆", Price: 200, Rent: 25},
	16: {Name: "St. James Place", Type: "property", Band: "#f7941d", Icon: "🏨", Price: 180, Rent: 14},
	17: {Name: "Community Chest", Type: "chest", Icon: "🧰"},
	18: {Name: "Tennessee Avenue", Type: "property", Band: "#f7941d", Icon: "🏨", Price: 180, Rent: 14},
	19: {Name: "New York Avenue", Type: "property", Band: "#f7941d", Icon: "🏨", Price: 200, Rent: 16},
	20: {Name: "Free Parking", Type: "corner", Icon: "🅿️"},
	21: {Name: "Kentucky Avenue", Type: "property", Band: "#ed1b24", Icon: "🏬", Price: 220, Rent: 18},
	22: {Name: "Chance", Type: "chance", Icon: "❓"},
	23: {Name: "Indiana Avenue", Type: "property", Band: "#ed1b24", Icon: "🏬", Price: 220, Rent: 18},
	24: {Name: "Illinois Avenue", Type: "property", Band: "#ed1b24", Icon: "🏬", Price: 240, Rent: 20},
	25: {Name: "B. & O. Railroad", Type: "rail", Icon: "🚄", Price: 200, Rent: 25},
	26: {Name: "Atlantic Avenue", Type: "property", Band: "#fef200", Icon: "🏢", Price: 260, Rent: 22},
	27: {Name: "Ventnor Avenue", Type: "property", Band: "#fef200", Icon: "🏢", Price: 260, Rent: 22},
	28: {Name: "Water Works", Type: "util", Icon: "🚰", Price: 150, Rent: 10},
	29: {Name: "Marvin Gardens", Type: "property", Band: "#fef200", Icon: "🏢", Price: 280, Rent: 24},
	30: {Name: "Go To Jail", Type: "corner", Icon: "🚓"},
	31: {Name: "Pacific Avenue", Type: "property", Band: "#1fb25a", Icon: "🏢", Price: 300, Rent: 26},
	32: {Name: "North Carolina Avenue", Type: "property", Band: "#1fb25a", Icon: "🏢", Price: 300, Rent: 26},
	33: {Name: "Community Chest", Type: "chest", Icon: "🧰"},
	34: {Name: "Pennsylvania Avenue", Type: "property", Band: "#1fb25a", Icon: "🏢", Price: 320, Rent: 28},
	35: {Name: "Short Line", Type: "rail", Icon: "🚃", Price: 200, Rent: 25},
	36: {Name: "Chance", Type: "chance", Icon: "❓"},
	37: {Name: "Park Place", Type: "property", Band: "#0072bb", Icon: "🏙️", Price: 350, Rent: 35},
	38: {Name: "Luxury Tax", Type: "tax", Icon: "💎"},
	39: {Name: "Boardwalk", Type: "property", Band: "#0072bb", Icon: "🏙️", Price: 400, Rent: 50},
}

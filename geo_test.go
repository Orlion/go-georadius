package geo

import (
	"testing"
)

func TestNeighborRanges(t *testing.T) {
	r := NeighborRanges(119.322741, 26.085920, 5000)
	t.Log(r)
}

func TestGetAreasByRadius(t *testing.T) {
	radius := getAreasByRadius(119.322741, 26.085920, 5000)
	if radius.neighbors.south == nil {
		t.Errorf("south is nil \n")
	}

	south := decode(radius.neighbors.south)
	if south.longitude.min != 119.1796875 {
		t.Errorf("south.longitude.min expected be %v, but %v \n", 119.1796875, south.longitude.min)
	}
	if south.longitude.max != 119.35546875 {
		t.Errorf("south.longitude.max expected be %v, but %v \n", 119.35546875, south.longitude.max)
	}
	if south.latitude.min != 25.997073543105472 {
		t.Errorf("south.latitude.min expected be %v, but %v \n", 25.997073543105472, south.latitude.min)
	}
	if south.latitude.max != 26.080131286054694 {
		t.Errorf("south.latitude.max expected be %v, but %v \n", 26.080131286054694, south.latitude.max)
	}

	east := decode(radius.neighbors.east)
	if east.longitude.min != 119.35546875 {
		t.Errorf("east.longitude.min expected be %v, but %v \n", 119.35546875, east.longitude.min)
	}
	if east.longitude.max != 119.53125 {
		t.Errorf("east.longitude.max expected be %v, but %v \n", 119.53125, east.longitude.max)
	}
	if east.latitude.min != 26.080131286054694 {
		t.Errorf("east.latitude.min expected be %v, but %v \n", 26.080131286054694, east.latitude.min)
	}
	if east.latitude.max != 26.163189029003902 {
		t.Errorf("east.latitude.max expected be %v, but %v \n", 26.163189029003902, east.latitude.max)
	}

	southEast := decode(radius.neighbors.southEast)
	if southEast.longitude.min != 119.35546875 {
		t.Errorf("southEast.longitude.min expected be %v, but %v \n", 119.35546875, southEast.longitude.min)
	}
	if southEast.longitude.max != 119.53125 {
		t.Errorf("southEast.longitude.max expected be %v, but %v \n", 119.53125, southEast.longitude.max)
	}
	if southEast.latitude.min != 25.997073543105472 {
		t.Errorf("southEast.latitude.min expected be %v, but %v \n", 25.997073543105472, southEast.latitude.min)
	}
	if southEast.latitude.max != 26.080131286054694 {
		t.Errorf("southEast.latitude.max expected be %v, but %v \n", 26.080131286054694, southEast.latitude.max)
	}
}

func TestGetNeighbors(t *testing.T) {
	fibTests := [...]struct {
		bits      uint64
		north     uint64
		east      uint64
		west      uint64
		south     uint64
		northEast uint64
		southEast uint64
		northWest uint64
		southWest uint64
	}{
		{6, 7, 12, 4, 3, 13, 9, 5, 1},
		{9, 12, 11, 3, 8, 14, 10, 6, 2},
	}

	for _, v := range fibTests {
		neighbors := getNeighbors(&Bits{
			bits: v.bits,
			step: 2,
		})
		if neighbors.north.bits != v.north {
			t.Errorf("%v's north expected %v but %v", v.bits, v.north, neighbors.north.bits)
		}
		if neighbors.east.bits != v.east {
			t.Errorf("%v's east expected %v but %v", v.bits, v.east, neighbors.east.bits)
		}
		if neighbors.west.bits != v.west {
			t.Errorf("%v's west expected %v but %v", v.bits, v.west, neighbors.west.bits)
		}
		if neighbors.south.bits != v.south {
			t.Errorf("%v's south expected %v but %v", v.bits, v.south, neighbors.south.bits)
		}
		if neighbors.northEast.bits != v.northEast {
			t.Errorf("%v's northEast expected %v but %v", v.bits, v.northEast, neighbors.northEast.bits)
		}
		if neighbors.southEast.bits != v.southEast {
			t.Errorf("%v's southEast expected %v but %v", v.bits, v.southEast, neighbors.southEast.bits)
		}
		if neighbors.northWest.bits != v.northWest {
			t.Errorf("%v's northWest expected %v but %v", v.bits, v.northWest, neighbors.northWest.bits)
		}
		if neighbors.southWest.bits != v.southWest {
			t.Errorf("%v's southWest expected %v but %v", v.bits, v.southWest, neighbors.southWest.bits)
		}
	}
}

func TestEncode(t *testing.T) {
	fibTests := [...]struct {
		lng      float64
		lat      float64
		expected uint64
	}{
		{100, 50, 4229754648981807},
		{121.268863, 30.293746, 4054538169573788},  // 上海
		{116.505021, 39.950898, 4069886986823592},  // 北京
		{120.4724433, 36.095838, 4067547047995087}, // 青岛
		{87.638893, 43.776323, 3846741834451526},   // 乌鲁木齐
		{91.147007, 29.675463, 4016577785256578},   // 拉萨
	}

	for _, v := range fibTests {
		hash := encode(v.lng, v.lat, GEO_STEP_MAX)
		if hash.bits != v.expected {
			t.Errorf("%v,%v is %v, expected to be %v", v.lng, v.lat, hash.bits, v.expected)
		}
	}
}

func TestDecode(t *testing.T) {
	fibTests := [...]struct {
		lng  float64
		lat  float64
		hash uint64
	}{
		{100, 50, 4229754648981807},
		{121.268863, 30.293746, 4054538169573788},  // 上海
		{116.505021, 39.950898, 4069886986823592},  // 北京
		{120.4724433, 36.095838, 4067547047995087}, // 青岛
		{87.638893, 43.776323, 3846741834451526},   // 乌鲁木齐
		{91.147007, 29.675463, 4016577785256578},   // 拉萨
	}

	for _, v := range fibTests {
		area := decode(&Bits{
			bits: v.hash,
			step: GEO_STEP_MAX,
		})
		if v.lat < area.latitude.min {
			t.Errorf("%v's lat %v < area.latitude.min %v", v.hash, v.lat, area.latitude.min)
		}
		if v.lat > area.latitude.max {
			t.Errorf("%v's lat %v > area.latitude.max %v", v.hash, v.lat, area.latitude.max)
		}
		if v.lng < area.longitude.min {
			t.Errorf("%v's lat %v < area.longitude.min %v", v.hash, v.lat, area.longitude.min)
		}
		if v.lng > area.longitude.max {
			t.Errorf("%v's lat %v > area.longitude.max %v", v.hash, v.lat, area.longitude.max)
		}
	}
}

func TestEstimateStepsByRadius(t *testing.T) {
	fibTests := [...]struct {
		radiusMeters float64
		lat          float64
		expected     uint8
	}{
		{610, 50, 15},
		{610, 61, 14},
		{610, 81, 13},
		{76, 50, 18},
		{19.10, 50, 20},
	}

	for _, v := range fibTests {
		step := estimateStepsByRadius(v.radiusMeters, v.lat)
		if step != v.expected {
			t.Errorf("%v, %v expected %v but %v", v.radiusMeters, v.lat, v.expected, step)
		}
	}
}

func TestGetDistance(t *testing.T) {
	fibTests := [...]struct {
		lon1     float64
		lat1     float64
		lon2     float64
		lat2     float64
		expected float64
	}{
		{100, 50, 100, 51, 111226.300000},
		{116, 39, 115, 33, 673382.764053},
	}

	for _, v := range fibTests {
		dist := getDistance(v.lon1, v.lat1, v.lon2, v.lat2)
		if int(dist*1000000) != int(v.expected*1000000) {
			t.Errorf("%v, %v => %v,%v dist=%v expected=%v", v.lon1, v.lat1, v.lon2, v.lat2, dist, v.expected)
		}
	}
}

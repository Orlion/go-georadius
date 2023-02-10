package geo

import (
	"math"
)

const (
	GEO_LONG_MAX           float64 = 180 // Limits from EPSG:900913 / EPSG:3785 / OSGEO:41001 南极和北极不能编码
	GEO_LONG_MIN           float64 = -180
	GEO_LAT_MAX            float64 = 85.05112878
	GEO_LAT_MIN            float64 = -85.05112878
	ONE_INT                int     = 1
	GEO_STEP_MAX                   = 26
	M_PI                   float64 = 3.14159265358979323846264338327950288
	D_R                    float64 = (M_PI / 180.0)
	EARTH_RADIUS_IN_METERS float64 = 6372797.560856
	MERCATOR_MAX           float64 = 20037726.37
)

const (
	magic1 uint64 = 0xaaaaaaaaaaaaaaaa
	magic2 uint64 = 0x5555555555555555
)

type Bits struct {
	bits uint64
	step uint8
}

type Range struct {
	min float64
	max float64
}

type Area struct {
	hash      *Bits
	longitude Range
	latitude  Range
}

type Radius struct {
	hash      *Bits
	area      *Area
	neighbors *Neighbors
}

type Neighbors struct {
	north     *Bits
	east      *Bits
	west      *Bits
	south     *Bits
	northEast *Bits
	southEast *Bits
	northWest *Bits
	southWest *Bits
}

func NeighborRanges(longitude, latitude, radiusMeters float64) [][2]uint64 {
	georadius := getAreasByRadius(longitude, latitude, radiusMeters)

	neighborRanges := make([][2]uint64, 0)
	min, max := scoresOfGeoHashBox(georadius.hash)
	neighborRanges = append(neighborRanges, [2]uint64{min, max})

	if georadius.neighbors.north != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.north)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}

	if georadius.neighbors.east != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.east)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.west != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.west)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.south != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.south)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.northEast != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.northEast)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.southEast != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.southEast)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.northWest != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.northWest)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	if georadius.neighbors.southWest != nil {
		min, max = scoresOfGeoHashBox(georadius.neighbors.southWest)
		neighborRanges = append(neighborRanges, [2]uint64{min, max})
	}
	return neighborRanges
}

func scoresOfGeoHashBox(hash *Bits) (min, max uint64) {
	min = align52Bits(hash)
	hash.bits++
	max = align52Bits(hash)
	return
}

func align52Bits(hash *Bits) uint64 {
	bits := hash.bits
	bits <<= (52 - hash.step*2)
	return bits
}

func getAreasByRadius(longitude, latitude, radiusMeters float64) Radius {
	minLon, maxLon, minLat, maxLat := boundingBox(longitude, latitude, radiusMeters)

	// 根据距离估算需要的精度
	steps := estimateStepsByRadius(radiusMeters, latitude)
	// 按照精度计算出当前点即中心块hash
	hash := encode(longitude, latitude, steps)
	// 获取中心块附近的八个区块
	neighbors := getNeighbors(hash)
	area := decode(hash)

	north := decode(neighbors.north)
	south := decode(neighbors.south)
	east := decode(neighbors.east)
	west := decode(neighbors.west)

	decreaseStep := false

	if getDistance(longitude, latitude, longitude, north.latitude.max) < radiusMeters {
		decreaseStep = true
	}
	if getDistance(longitude, latitude, longitude, south.latitude.min) < radiusMeters {
		decreaseStep = true
	}
	if getDistance(longitude, latitude, east.longitude.max, latitude) < radiusMeters {
		decreaseStep = true
	}
	if getDistance(longitude, latitude, west.longitude.min, latitude) < radiusMeters {
		decreaseStep = true
	}

	if steps > 1 && decreaseStep {
		steps--
		hash = encode(longitude, latitude, steps)
		neighbors = getNeighbors(hash)
		area = decode(hash)
	}

	if steps >= 2 {
		if area.latitude.min < minLat {
			neighbors.south = nil
			neighbors.southWest = nil
			neighbors.southEast = nil
		}
		if area.latitude.max > maxLat {
			neighbors.north = nil
			neighbors.northEast = nil
			neighbors.northWest = nil
		}
		if area.longitude.min < minLon {
			neighbors.west = nil
			neighbors.southWest = nil
			neighbors.northWest = nil
		}
		if area.longitude.max > maxLon {
			neighbors.east = nil
			neighbors.southEast = nil
			neighbors.northEast = nil
		}
	}

	return Radius{
		hash:      hash,
		area:      area,
		neighbors: neighbors,
	}
}

func gzero(s *Bits) {
	s.bits = 0
	s.step = 0
}

func boundingBox(longitude, latitude, radiusMeters float64) (float64, float64, float64, float64) {
	minLog := longitude - radDeg(radiusMeters/EARTH_RADIUS_IN_METERS/math.Cos(degRad(latitude)))
	maxLon := longitude + radDeg(radiusMeters/EARTH_RADIUS_IN_METERS/math.Cos(degRad(latitude)))
	minLat := latitude - radDeg(radiusMeters/EARTH_RADIUS_IN_METERS)
	maxLat := latitude + radDeg(radiusMeters/EARTH_RADIUS_IN_METERS)
	return minLog, maxLon, minLat, maxLat
}

func GetDistanceByScore(lon1d, lat1d float64, score uint64) float64 {
	area := decode(&Bits{
		bits: score,
		step: GEO_STEP_MAX,
	})

	return getDistance(lon1d, lat1d, (area.longitude.min+area.longitude.max)/2, (area.latitude.min+area.latitude.max)/2)
}

func getDistance(lon1d, lat1d, lon2d, lat2d float64) float64 {
	lat1r := degRad(lat1d)
	lon1r := degRad(lon1d)
	lat2r := degRad(lat2d)
	lon2r := degRad(lon2d)
	u := math.Sin((lat2r - lat1r) / 2)
	v := math.Sin((lon2r - lon1r) / 2)
	return 2.0 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(u*u+math.Cos(lat1r)*math.Cos(lat2r)*v*v))
}

func degRad(ang float64) float64 {
	return ang * D_R
}

func radDeg(ang float64) float64 {
	return ang / D_R
}

// 获取中心块附近的八个区块
func getNeighbors(hash *Bits) *Neighbors {
	neighbors := &Neighbors{
		north: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		east: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		west: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		south: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		northEast: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		southEast: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		northWest: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
		southWest: &Bits{
			bits: hash.bits,
			step: hash.step,
		},
	}

	moveX(neighbors.east, 1)
	moveY(neighbors.east, 0)

	moveX(neighbors.west, -1)
	moveY(neighbors.west, 0)

	moveX(neighbors.south, 0)
	moveY(neighbors.south, -1)

	moveX(neighbors.north, 0)
	moveY(neighbors.north, 1)

	moveX(neighbors.northWest, -1)
	moveY(neighbors.northWest, 1)

	moveX(neighbors.northEast, 1)
	moveY(neighbors.northEast, 1)

	moveX(neighbors.southEast, 1)
	moveY(neighbors.southEast, -1)

	moveX(neighbors.southWest, -1)
	moveY(neighbors.southWest, -1)

	return neighbors
}

func moveX(hash *Bits, d int8) {
	if d == 0 {
		return
	}

	x := hash.bits & magic1
	y := hash.bits & magic2

	zz := magic2 >> (64 - hash.step*2)

	if d > 0 {
		x = x + (zz + 1)
	} else {
		x = x | zz
		x = x - (zz + 1)
	}

	x &= magic1 >> (64 - hash.step*2)
	hash.bits = x | y
}

func moveY(hash *Bits, d int8) {
	if d == 0 {
		return
	}

	x := hash.bits & magic1
	y := hash.bits & magic2

	zz := magic1 >> (64 - hash.step*2)
	if d > 0 {
		y = y + (zz + 1)
	} else {
		y = y | zz
		y = y - (zz + 1)
	}
	y &= (magic2 >> (64 - hash.step*2))
	hash.bits = x | y
}

// 根据距离估算需要的精度
func estimateStepsByRadius(radiusMeters, latitude float64) uint8 {
	if radiusMeters == 0 {
		return 26
	}

	step := 1

	for radiusMeters < MERCATOR_MAX {
		radiusMeters *= 2
		step++
	}

	step -= 2

	if latitude > 60 || latitude < -60 {
		step--
		if latitude > 80 || latitude < -80 {
			step--
		}
	}

	if step < 1 {
		step = 1
	}

	if step > 26 {
		step = 26
	}

	return uint8(step)
}

func encode(longitude, latitude float64, step uint8) *Bits {
	latOffset := ((latitude - GEO_LAT_MIN) / (GEO_LAT_MAX - GEO_LAT_MIN)) * float64(ONE_INT<<step)
	longOffset := (longitude - GEO_LONG_MIN) / (GEO_LONG_MAX - GEO_LONG_MIN) * float64(ONE_INT<<step)

	return &Bits{
		bits: interleave64(uint32(latOffset), uint32(longOffset)),
		step: step,
	}
}

func interleave64(xlo, ylo uint32) uint64 {
	B := [5]uint64{0x5555555555555555, 0x3333333333333333, 0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF, 0x0000FFFF0000FFFF}

	S := [5]uint32{1, 2, 4, 8, 16}

	x := uint64(xlo)
	y := uint64(ylo)

	x = (x | (x << S[4])) & B[4]
	y = (y | (y << S[4])) & B[4]

	x = (x | (x << S[3])) & B[3]
	y = (y | (y << S[3])) & B[3]

	x = (x | (x << S[2])) & B[2]
	y = (y | (y << S[2])) & B[2]

	x = (x | (x << S[1])) & B[1]
	y = (y | (y << S[1])) & B[1]

	x = (x | (x << S[0])) & B[0]
	y = (y | (y << S[0])) & B[0]

	return x | (y << 1)
}

func decode(hash *Bits) *Area {
	area := &Area{
		hash: hash,
	}

	step := hash.step
	hashSep := deinterleave64(hash.bits)

	latScale := GEO_LAT_MAX - GEO_LAT_MIN
	longScale := GEO_LONG_MAX - GEO_LONG_MIN

	ilato := uint32(hashSep)
	ilono := uint32(hashSep >> 32)

	area.latitude.min = GEO_LAT_MIN + (float64(ilato)/float64(ONE_INT<<step))*latScale
	area.latitude.max = GEO_LAT_MIN + (float64(ilato+1)/float64(ONE_INT<<step))*latScale
	area.longitude.min = GEO_LONG_MIN + (float64(ilono)/float64(ONE_INT<<step))*longScale
	area.longitude.max = GEO_LONG_MIN + (float64(ilono+1)/float64(ONE_INT<<step))*longScale

	return area
}

func deinterleave64(interleaved uint64) uint64 {
	B := [6]uint64{0x5555555555555555, 0x3333333333333333, 0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF, 0x0000FFFF0000FFFF, 0x00000000FFFFFFFF}
	S := [6]uint{0, 1, 2, 4, 8, 16}

	x := interleaved
	y := interleaved >> 1

	x = (x | (x >> S[0])) & B[0]
	y = (y | (y >> S[0])) & B[0]

	x = (x | (x >> S[1])) & B[1]
	y = (y | (y >> S[1])) & B[1]

	x = (x | (x >> S[2])) & B[2]
	y = (y | (y >> S[2])) & B[2]

	x = (x | (x >> S[3])) & B[3]
	y = (y | (y >> S[3])) & B[3]

	x = (x | (x >> S[4])) & B[4]
	y = (y | (y >> S[4])) & B[4]

	x = (x | (x >> S[5])) & B[5]
	y = (y | (y >> S[5])) & B[5]

	return x | (y << 32)
}

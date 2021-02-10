// This ip2l package provides a fast lookup of country, region, city, latitude, longitude, ZIP code, time zone,
// ISP, domain name, connection type, IDD code, area code, weather station code, station name, MCC, MNC,
// mobile brand, elevation, and usage type from IP address by using IP2Location database.
package ip2l

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"os"
	"strconv"
)

type DBReader interface {
	io.ReadCloser
	io.ReaderAt
}

type ip2LocationMeta struct {
	databaseType      uint8
	databaseColumn    uint8
	databaseDay       uint8
	databaseMonth     uint8
	databaseYear      uint8
	ipv4DatabaseCount uint32
	ipv4DatabaseAddr  uint32
	ipv6DatabaseCount uint32
	ipv6DatabaseAddr  uint32
	ipv4IndexBaseAddr uint32
	ipv6IndexBaseAddr uint32
	ipv4ColumnSize    uint32
	ipv6ColumnSize    uint32
}

// The IP2LocationRecord struct stores all of the available
// geolocation info found in the IP2Location database.
type IP2LocationRecord struct {
	CountryShort       string
	CountryLong        string
	Region             string
	City               string
	Isp                string
	Latitude           float32
	Longitude          float32
	Domain             string
	ZipCode            string
	Timezone           string
	NetSpeed           string
	IddCode            string
	AreaCode           string
	WeatherStationCode string
	WeatherStationName string
	MCC                string
	MNC                string
	MobileBrand        string
	Elevation          float32
	UsageType          string
}

type DB struct {
	f    DBReader
	meta ip2LocationMeta

	countryPositionOffset            uint32
	regionPositionOffset             uint32
	cityPositionOffset               uint32
	ispPositionOffset                uint32
	domainPositionOffset             uint32
	zipcodePositionOffset            uint32
	latitudePositionOffset           uint32
	longitudePositionOffset          uint32
	timezonePositionOffset           uint32
	netSpeedPositionOffset           uint32
	iddCodePositionOffset            uint32
	areaCodePositionOffset           uint32
	weatherStationCodePositionOffset uint32
	weatherStationNamePositionOffset uint32
	mccPositionOffset                uint32
	mncPositionOffset                uint32
	mobileBrandPositionOffset        uint32
	elevationPositionOffset          uint32
	usageTypePositionOffset          uint32

	countryEnabled            bool
	regionEnabled             bool
	cityEnabled               bool
	ispEnabled                bool
	domainEnabled             bool
	zipcodeEnabled            bool
	latitudeEnabled           bool
	longitudeEnabled          bool
	timeZoneEnabled           bool
	netSpeedEnabled           bool
	iddCodeEnabled            bool
	areaCodeEnabled           bool
	weatherStationCodeEnabled bool
	weatherStationNameEnabled bool
	mccEnabled                bool
	mncEnabled                bool
	mobileBrandEnabled        bool
	elevationEnabled          bool
	usageTypeEnabled          bool

	metaOk bool
}

var countryPosition = [25]uint8{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
var regionPosition = [25]uint8{0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
var cityPosition = [25]uint8{0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}
var ispPosition = [25]uint8{0, 0, 3, 0, 5, 0, 7, 5, 7, 0, 8, 0, 9, 0, 9, 0, 9, 0, 9, 7, 9, 0, 9, 7, 9}
var latitudePosition = [25]uint8{0, 0, 0, 0, 0, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}
var longitudePosition = [25]uint8{0, 0, 0, 0, 0, 6, 6, 0, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6}
var domainPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 6, 8, 0, 9, 0, 10, 0, 10, 0, 10, 0, 10, 8, 10, 0, 10, 8, 10}
var zipCodePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 7, 7, 7, 0, 7, 7, 7, 0, 7, 0, 7, 7, 7, 0, 7}
var timeZonePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 7, 8, 8, 8, 7, 8, 0, 8, 8, 8, 0, 8}
var netSpeedPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 11, 0, 11, 8, 11, 0, 11, 0, 11, 0, 11}
var iddCodePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 12, 0, 12, 0, 12, 9, 12, 0, 12}
var areaCodePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 13, 0, 13, 0, 13, 10, 13, 0, 13}
var weatherStationCodePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 14, 0, 14, 0, 14, 0, 14}
var weatherStationNamePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 15, 0, 15, 0, 15, 0, 15}
var mccPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 16, 0, 16, 9, 16}
var mncPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 17, 0, 17, 10, 17}
var mobileBrandPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 18, 0, 18, 11, 18}
var elevationPosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 19, 0, 19}
var usageTypePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 20}

const apiVersion string = "8.4.0"

var maxIpv4Range = big.NewInt(4294967295)
var maxIpv6Range = big.NewInt(0)
var fromV4mapped = big.NewInt(281470681743360)
var toV4mapped = big.NewInt(281474976710655)
var from6to4 = big.NewInt(0)
var to6to4 = big.NewInt(0)
var fromTeredo = big.NewInt(0)
var toTeredo = big.NewInt(0)
var last32bits = big.NewInt(4294967295)

const countryShort uint32 = 0x00001
const countryLong uint32 = 0x00002
const region uint32 = 0x00004
const city uint32 = 0x00008
const isp uint32 = 0x00010
const latitude uint32 = 0x00020
const longitude uint32 = 0x00040
const domain uint32 = 0x00080
const zipCode uint32 = 0x00100
const timezone uint32 = 0x00200
const netSpeed uint32 = 0x00400
const iddCode uint32 = 0x00800
const areaCode uint32 = 0x01000
const weatherStationCode uint32 = 0x02000
const weatherStationName uint32 = 0x04000
const mcc uint32 = 0x08000
const mnc uint32 = 0x10000
const mobileBrand uint32 = 0x20000
const elevation uint32 = 0x40000
const usageType uint32 = 0x80000

const all = countryShort | countryLong | region | city | isp | latitude | longitude | domain | zipCode | timezone | netSpeed | iddCode | areaCode | weatherStationCode | weatherStationName | mcc | mnc | mobileBrand | elevation | usageType

const invalidAddress string = "Invalid IP address."
const missingFile string = "Invalid database file."
const parameterIsNotSupported string = "This parameter is unavailable for selected data file. Please upgrade the data file."

// get IP type and calculate IP number; calculates index too if exists
func (d *DB) checkIP(ip string) (ipType uint32, ipNum *big.Int, ipIndex uint32) {
	ipType = 0
	ipNum = big.NewInt(0)
	ipNumTmp := big.NewInt(0)
	ipIndex = 0
	ipaddress := net.ParseIP(ip)

	if ipaddress != nil {
		v4 := ipaddress.To4()

		if v4 != nil {
			ipType = 4
			ipNum.SetBytes(v4)
		} else {
			v6 := ipaddress.To16()

			if v6 != nil {
				ipType = 6
				ipNum.SetBytes(v6)

				if ipNum.Cmp(fromV4mapped) >= 0 && ipNum.Cmp(toV4mapped) <= 0 {
					// ipv4-mapped ipv6 should treat as ipv4 and read ipv4 data section
					ipType = 4
					ipNum.Sub(ipNum, fromV4mapped)
				} else if ipNum.Cmp(from6to4) >= 0 && ipNum.Cmp(to6to4) <= 0 {
					// 6to4 so need to remap to ipv4
					ipType = 4
					ipNum.Rsh(ipNum, 80)
					ipNum.And(ipNum, last32bits)
				} else if ipNum.Cmp(fromTeredo) >= 0 && ipNum.Cmp(toTeredo) <= 0 {
					// Teredo so need to remap to ipv4
					ipType = 4
					ipNum.Not(ipNum)
					ipNum.And(ipNum, last32bits)
				}
			}
		}
	}
	if ipType == 4 {
		if d.meta.ipv4IndexBaseAddr > 0 {
			ipNumTmp.Rsh(ipNum, 16)
			ipNumTmp.Lsh(ipNumTmp, 3)
			ipIndex = uint32(ipNumTmp.Add(ipNumTmp, big.NewInt(int64(d.meta.ipv4IndexBaseAddr))).Uint64())
		}
	} else if ipType == 6 {
		if d.meta.ipv6IndexBaseAddr > 0 {
			ipNumTmp.Rsh(ipNum, 112)
			ipNumTmp.Lsh(ipNumTmp, 3)
			ipIndex = uint32(ipNumTmp.Add(ipNumTmp, big.NewInt(int64(d.meta.ipv6IndexBaseAddr))).Uint64())
		}
	}
	return
}

// read byte
func (d *DB) readUint8(pos int64) (uint8, error) {
	data := make([]byte, 1)
	_, err := d.f.ReadAt(data, pos-1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// read unsigned 32-bit integer from slices
func (d *DB) readUint32Row(row []byte, pos uint32) uint32 {
	data := row[pos : pos+4]
	return binary.LittleEndian.Uint32(data)
}

// read unsigned 32-bit integer
func (d *DB) readUint32(pos uint32) (uint32, error) {
	pos2 := int64(pos)
	data := make([]byte, 4)
	_, err := d.f.ReadAt(data, pos2-1)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewReader(data)

	var res uint32
	err = binary.Read(buf, binary.LittleEndian, &res)
	if err != nil {
		fmt.Printf("binary read failed: %v", err)
	}
	return res, nil
}

// read unsigned 128-bit integer
func (d *DB) readUint128(pos uint32) (*big.Int, error) {
	pos2 := int64(pos)
	data := make([]byte, 16)
	_, err := d.f.ReadAt(data, pos2-1)
	if err != nil {
		return nil, err
	}

	// little endian to big endian
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return big.NewInt(0).SetBytes(data), nil
}

// read string
func (d *DB) readStr(pos uint32) (string, error) {
	pos2 := int64(pos)
	lenbyte := make([]byte, 1)
	_, err := d.f.ReadAt(lenbyte, pos2)
	if err != nil {
		return "", err
	}
	strlen := lenbyte[0]
	data := make([]byte, strlen)
	_, err = d.f.ReadAt(data, pos2+1)
	if err != nil {
		return "", err
	}
	return string(data[:strlen]), nil
}

// read float from slices
func (d *DB) readFloatRow(row []byte, pos uint32) float32 {
	data := row[pos : pos+4]
	bits := binary.LittleEndian.Uint32(data)
	return math.Float32frombits(bits)
}

// read float
func (d *DB) readFloat(pos uint32) (float32, error) {
	pos2 := int64(pos)
	data := make([]byte, 4)
	_, err := d.f.ReadAt(data, pos2-1)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewReader(data)
	var res float32
	err = binary.Read(buf, binary.LittleEndian, &res)
	if err != nil {
		fmt.Printf("binary read failed: %v", err)
	}
	return res, nil
}

func fatal(db *DB, err error) (*DB, error) {
	_ = db.f.Close()
	return nil, err
}

// Open takes the path to the IP2Location BIN database file. It will read all the metadata required to
// be able to extract the embedded geolocation data, and return the underlining DB object.
func OpenDB(dbpath string) (*DB, error) {
	f, err := os.Open(dbpath)
	if err != nil {
		return nil, err
	}

	return OpenDBWithReader(f)
}

// OpenDBWithReader takes a DBReader to the IP2Location BIN database file. It will read all the metadata required to
// be able to extract the embedded geolocation data, and return the underlining DB object.
func OpenDBWithReader(reader DBReader) (*DB, error) {
	var db = &DB{}

	maxIpv6Range.SetString("340282366920938463463374607431768211455", 10)
	from6to4.SetString("42545680458834377588178886921629466624", 10)
	to6to4.SetString("42550872755692912415807417417958686719", 10)
	fromTeredo.SetString("42540488161975842760550356425300246528", 10)
	toTeredo.SetString("42540488241204005274814694018844196863", 10)

	db.f = reader

	var err error
	db.meta.databaseType, err = db.readUint8(1)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.databaseColumn, err = db.readUint8(2)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.databaseYear, err = db.readUint8(3)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.databaseMonth, err = db.readUint8(4)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.databaseDay, err = db.readUint8(5)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv4DatabaseCount, err = db.readUint32(6)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv4DatabaseAddr, err = db.readUint32(10)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv6DatabaseCount, err = db.readUint32(14)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv6DatabaseAddr, err = db.readUint32(18)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv4IndexBaseAddr, err = db.readUint32(22)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv6IndexBaseAddr, err = db.readUint32(26)
	if err != nil {
		return fatal(db, err)
	}
	db.meta.ipv4ColumnSize = uint32(db.meta.databaseColumn << 2)              // 4 bytes each column
	db.meta.ipv6ColumnSize = uint32(16 + ((db.meta.databaseColumn - 1) << 2)) // 4 bytes each column, except IPFrom column which is 16 bytes

	dbt := db.meta.databaseType

	if countryPosition[dbt] != 0 {
		db.countryPositionOffset = uint32(countryPosition[dbt]-2) << 2
		db.countryEnabled = true
	}
	if regionPosition[dbt] != 0 {
		db.regionPositionOffset = uint32(regionPosition[dbt]-2) << 2
		db.regionEnabled = true
	}
	if cityPosition[dbt] != 0 {
		db.cityPositionOffset = uint32(cityPosition[dbt]-2) << 2
		db.cityEnabled = true
	}
	if ispPosition[dbt] != 0 {
		db.ispPositionOffset = uint32(ispPosition[dbt]-2) << 2
		db.ispEnabled = true
	}
	if domainPosition[dbt] != 0 {
		db.domainPositionOffset = uint32(domainPosition[dbt]-2) << 2
		db.domainEnabled = true
	}
	if zipCodePosition[dbt] != 0 {
		db.zipcodePositionOffset = uint32(zipCodePosition[dbt]-2) << 2
		db.zipcodeEnabled = true
	}
	if latitudePosition[dbt] != 0 {
		db.latitudePositionOffset = uint32(latitudePosition[dbt]-2) << 2
		db.latitudeEnabled = true
	}
	if longitudePosition[dbt] != 0 {
		db.longitudePositionOffset = uint32(longitudePosition[dbt]-2) << 2
		db.longitudeEnabled = true
	}
	if timeZonePosition[dbt] != 0 {
		db.timezonePositionOffset = uint32(timeZonePosition[dbt]-2) << 2
		db.timeZoneEnabled = true
	}
	if netSpeedPosition[dbt] != 0 {
		db.netSpeedPositionOffset = uint32(netSpeedPosition[dbt]-2) << 2
		db.netSpeedEnabled = true
	}
	if iddCodePosition[dbt] != 0 {
		db.iddCodePositionOffset = uint32(iddCodePosition[dbt]-2) << 2
		db.iddCodeEnabled = true
	}
	if areaCodePosition[dbt] != 0 {
		db.areaCodePositionOffset = uint32(areaCodePosition[dbt]-2) << 2
		db.areaCodeEnabled = true
	}
	if weatherStationCodePosition[dbt] != 0 {
		db.weatherStationCodePositionOffset = uint32(weatherStationCodePosition[dbt]-2) << 2
		db.weatherStationCodeEnabled = true
	}
	if weatherStationNamePosition[dbt] != 0 {
		db.weatherStationNamePositionOffset = uint32(weatherStationNamePosition[dbt]-2) << 2
		db.weatherStationNameEnabled = true
	}
	if mccPosition[dbt] != 0 {
		db.mccPositionOffset = uint32(mccPosition[dbt]-2) << 2
		db.mccEnabled = true
	}
	if mncPosition[dbt] != 0 {
		db.mncPositionOffset = uint32(mncPosition[dbt]-2) << 2
		db.mncEnabled = true
	}
	if mobileBrandPosition[dbt] != 0 {
		db.mobileBrandPositionOffset = uint32(mobileBrandPosition[dbt]-2) << 2
		db.mobileBrandEnabled = true
	}
	if elevationPosition[dbt] != 0 {
		db.elevationPositionOffset = uint32(elevationPosition[dbt]-2) << 2
		db.elevationEnabled = true
	}
	if usageTypePosition[dbt] != 0 {
		db.usageTypePositionOffset = uint32(usageTypePosition[dbt]-2) << 2
		db.usageTypeEnabled = true
	}

	db.metaOk = true

	return db, nil
}

// ApiVersion returns the version of the component.
func ApiVersion() string {
	return apiVersion
}

// populate record with message
func loadMessage(mesg string) *IP2LocationRecord {
	x := &IP2LocationRecord{}

	x.CountryShort = mesg
	x.CountryLong = mesg
	x.Region = mesg
	x.City = mesg
	x.Isp = mesg
	x.Domain = mesg
	x.ZipCode = mesg
	x.Timezone = mesg
	x.NetSpeed = mesg
	x.IddCode = mesg
	x.AreaCode = mesg
	x.WeatherStationCode = mesg
	x.WeatherStationName = mesg
	x.MCC = mesg
	x.MNC = mesg
	x.MobileBrand = mesg
	x.UsageType = mesg

	return x
}

// GetAll will return all geolocation fields based on the queried IP address.
func (d *DB) GetAll(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, all)
}

// GetCountryShort will return the ISO-3166 country code based on the queried IP address.
func (d *DB) GetCountryShort(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, countryShort)
}

// GetCountryLong will return the country name based on the queried IP address.
func (d *DB) GetCountryLong(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, countryLong)
}

// GetRegion will return the region name based on the queried IP address.
func (d *DB) GetRegion(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, region)
}

// GetCity will return the city name based on the queried IP address.
func (d *DB) GetCity(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, city)
}

// GetISP will return the Internet Service Provider name based on the queried IP address.
func (d *DB) GetISP(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, isp)
}

// GetLatitude will return the latitude based on the queried IP address.
func (d *DB) GetLatitude(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, latitude)
}

// GetLongitude will return the longitude based on the queried IP address.
func (d *DB) GetLongitude(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, longitude)
}

// GetDomain will return the domain name based on the queried IP address.
func (d *DB) GetDomain(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, domain)
}

// GetZipCode will return the postal code based on the queried IP address.
func (d *DB) GetZipCode(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, zipCode)
}

// GetTimezone will return the time zone based on the queried IP address.
func (d *DB) GetTimezone(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, timezone)
}

// GetNetSpeed will return the Internet connection speed based on the queried IP address.
func (d *DB) GetNetSpeed(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, netSpeed)
}

// GetIDDCode will return the International Direct Dialing code based on the queried IP address.
func (d *DB) GetIDDCode(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, iddCode)
}

// GetAreaCode will return the area code based on the queried IP address.
func (d *DB) GetAreaCode(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, areaCode)
}

// GetWeatherStationCode will return the weather station code based on the queried IP address.
func (d *DB) GetWeatherStationCode(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, weatherStationCode)
}

// GetWeatherStationName will return the weather station name based on the queried IP address.
func (d *DB) GetWeatherStationName(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, weatherStationName)
}

// GetMCC will return the mobile country code based on the queried IP address.
func (d *DB) GetMCC(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, mcc)
}

// GetMNC will return the mobile network code based on the queried IP address.
func (d *DB) GetMNC(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, mnc)
}

// GetMobileBrand will return the mobile carrier brand based on the queried IP address.
func (d *DB) GetMobileBrand(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, mobileBrand)
}

// GetElevation will return the elevation in meters based on the queried IP address.
func (d *DB) GetElevation(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, elevation)
}

// GetUsageType will return the usage type based on the queried IP address.
func (d *DB) GetUsageType(ip string) (*IP2LocationRecord, error) {
	return d.query(ip, usageType)
}

// main query
func (d *DB) query(ip string, mode uint32) (*IP2LocationRecord, error) {
	x := loadMessage(parameterIsNotSupported) // default message

	// read metadata
	if !d.metaOk {
		x = loadMessage(missingFile)
		return x, nil
	}

	// check IP type and return IP number & index (if exists)
	iptype, ipno, ipindex := d.checkIP(ip)

	if iptype == 0 {
		x = loadMessage(invalidAddress)
		return x, nil
	}

	var err error
	var colsize uint32
	var baseaddr uint32
	var low uint32
	var high uint32
	var mid uint32
	var rowoffset uint32
	var rowoffset2 uint32
	ipfrom := big.NewInt(0)
	ipto := big.NewInt(0)
	maxip := big.NewInt(0)

	if iptype == 4 {
		baseaddr = d.meta.ipv4DatabaseAddr
		high = d.meta.ipv4DatabaseCount
		maxip = maxIpv4Range
		colsize = d.meta.ipv4ColumnSize
	} else {
		baseaddr = d.meta.ipv6DatabaseAddr
		high = d.meta.ipv6DatabaseCount
		maxip = maxIpv6Range
		colsize = d.meta.ipv6ColumnSize
	}

	// reading index
	if ipindex > 0 {
		low, err = d.readUint32(ipindex)
		if err != nil {
			return x, err
		}
		high, err = d.readUint32(ipindex + 4)
		if err != nil {
			return x, err
		}
	}

	if ipno.Cmp(maxip) >= 0 {
		ipno.Sub(ipno, big.NewInt(1))
	}

	for low <= high {
		mid = (low + high) >> 1
		rowoffset = baseaddr + (mid * colsize)
		rowoffset2 = rowoffset + colsize

		if iptype == 4 {
			ipfrom32, err := d.readUint32(rowoffset)
			if err != nil {
				return x, err
			}
			ipfrom = big.NewInt(int64(ipfrom32))

			ipto32, err := d.readUint32(rowoffset2)
			if err != nil {
				return x, err
			}
			ipto = big.NewInt(int64(ipto32))

		} else {
			ipfrom, err = d.readUint128(rowoffset)
			if err != nil {
				return x, err
			}

			ipto, err = d.readUint128(rowoffset2)
			if err != nil {
				return x, err
			}
		}

		if ipno.Cmp(ipfrom) >= 0 && ipno.Cmp(ipto) < 0 {
			var firstcol uint32 = 4 // 4 bytes for ip from
			if iptype == 6 {
				firstcol = 16 // 16 bytes for ipv6
			}

			row := make([]byte, colsize-firstcol) // exclude the ip from field
			_, err := d.f.ReadAt(row, int64(rowoffset+firstcol-1))
			if err != nil {
				return x, err
			}

			if mode&countryShort == 1 && d.countryEnabled {
				if x.CountryShort, err = d.readStr(d.readUint32Row(row, d.countryPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&countryLong != 0 && d.countryEnabled {
				if x.CountryLong, err = d.readStr(d.readUint32Row(row, d.countryPositionOffset) + 3); err != nil {
					return x, err
				}
			}

			if mode&region != 0 && d.regionEnabled {
				if x.Region, err = d.readStr(d.readUint32Row(row, d.regionPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&city != 0 && d.cityEnabled {
				if x.City, err = d.readStr(d.readUint32Row(row, d.cityPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&isp != 0 && d.ispEnabled {
				if x.Isp, err = d.readStr(d.readUint32Row(row, d.ispPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&latitude != 0 && d.latitudeEnabled {
				x.Latitude = d.readFloatRow(row, d.latitudePositionOffset)
			}

			if mode&longitude != 0 && d.longitudeEnabled {
				x.Longitude = d.readFloatRow(row, d.longitudePositionOffset)
			}

			if mode&domain != 0 && d.domainEnabled {
				if x.Domain, err = d.readStr(d.readUint32Row(row, d.domainPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&zipCode != 0 && d.zipcodeEnabled {
				if x.ZipCode, err = d.readStr(d.readUint32Row(row, d.zipcodePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&timezone != 0 && d.timeZoneEnabled {
				if x.Timezone, err = d.readStr(d.readUint32Row(row, d.timezonePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&netSpeed != 0 && d.netSpeedEnabled {
				if x.NetSpeed, err = d.readStr(d.readUint32Row(row, d.netSpeedPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&iddCode != 0 && d.iddCodeEnabled {
				if x.IddCode, err = d.readStr(d.readUint32Row(row, d.iddCodePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&areaCode != 0 && d.areaCodeEnabled {
				if x.AreaCode, err = d.readStr(d.readUint32Row(row, d.areaCodePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&weatherStationCode != 0 && d.weatherStationCodeEnabled {
				if x.WeatherStationCode, err = d.readStr(d.readUint32Row(row, d.weatherStationCodePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&weatherStationName != 0 && d.weatherStationNameEnabled {
				if x.WeatherStationName, err = d.readStr(d.readUint32Row(row, d.weatherStationNamePositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&mcc != 0 && d.mccEnabled {
				if x.MCC, err = d.readStr(d.readUint32Row(row, d.mccPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&mnc != 0 && d.mncEnabled {
				if x.MNC, err = d.readStr(d.readUint32Row(row, d.mncPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&mobileBrand != 0 && d.mobileBrandEnabled {
				if x.MobileBrand, err = d.readStr(d.readUint32Row(row, d.mobileBrandPositionOffset)); err != nil {
					return x, err
				}
			}

			if mode&elevation != 0 && d.elevationEnabled {
				res, err := d.readStr(d.readUint32Row(row, d.elevationPositionOffset))
				if err != nil {
					return x, err
				}

				f, _ := strconv.ParseFloat(res, 32)
				x.Elevation = float32(f)
			}

			if mode&usageType != 0 && d.usageTypeEnabled {
				if x.UsageType, err = d.readStr(d.readUint32Row(row, d.usageTypePositionOffset)); err != nil {
					return x, err
				}
			}

			return x, nil
		} else {
			if ipno.Cmp(ipfrom) < 0 {
				high = mid - 1
			} else {
				low = mid + 1
			}
		}
	}
	return x, nil
}

func (d *DB) Close() {
	_ = d.f.Close()
}

// PrintRecord is used to output the geolocation data for debugging purposes.
func PrintRecord(x *IP2LocationRecord) {
	fmt.Printf("countryShort: %s\n", x.CountryShort)
	fmt.Printf("countryLong: %s\n", x.CountryLong)
	fmt.Printf("region: %s\n", x.Region)
	fmt.Printf("city: %s\n", x.City)
	fmt.Printf("isp: %s\n", x.Isp)
	fmt.Printf("latitude: %f\n", x.Latitude)
	fmt.Printf("longitude: %f\n", x.Longitude)
	fmt.Printf("domain: %s\n", x.Domain)
	fmt.Printf("zipCode: %s\n", x.ZipCode)
	fmt.Printf("timezone: %s\n", x.Timezone)
	fmt.Printf("netSpeed: %s\n", x.NetSpeed)
	fmt.Printf("iddCode: %s\n", x.IddCode)
	fmt.Printf("areaCode: %s\n", x.AreaCode)
	fmt.Printf("weatherStationCode: %s\n", x.WeatherStationCode)
	fmt.Printf("weatherStationName: %s\n", x.WeatherStationName)
	fmt.Printf("mcc: %s\n", x.MCC)
	fmt.Printf("mnc: %s\n", x.MNC)
	fmt.Printf("mobileBrand: %s\n", x.MobileBrand)
	fmt.Printf("elevation: %f\n", x.Elevation)
	fmt.Printf("usageType: %s\n", x.UsageType)
}

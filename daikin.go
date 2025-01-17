// Package daikin provides functionality to interact with Daikin split
// system air conditioners equipped with a Wifi module. It is tested to work
// with the BRP072A42 Wifi interface.
package daikin

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	uriGetBasicInfo    = "/common/basic_info"
	uriGetRemoteMethod = "/common/get_remote_method"
	uriGetModelInfo    = "/aircon/get_model_info"
	uriGetControlInfo  = "/aircon/get_control_info"
	uriGetSensorInfo   = "/aircon/get_sensor_info"
	uriGetTimer        = "/aircon/get_timer"
	uriGetPrice        = "/aircon/get_price"
	uriGetTarget       = "/aircon/get_target"
	uriGetWeekPower    = "/aircon/get_week_power"
	uriGetYearPower    = "/aircon/get_year_power"
	uriGetProgram      = "/aircon/get_program"
	uriGetScdlTimer    = "/aircon/get_scdltimer"
	uriGetNotify       = "/aircon/get_notify"
	uriSetControlInfo  = "/aircon/set_control_info"
)

/*
type Parameter interface {
	// Puts this Parameter's entry into the provided url.Values.
	setUrlValues(v url.Values)
	// Set sets this parameter's value.
	Set(string) error
	// String returns the human readable value.
	String() string
}
*/

const (
	returnOk  = "OK"
	returnBad = "PARAM NG"
)

// Power represents the power status of the unit (off/on).
type Power int

// The power status of the unit.
const (
	PowerOff Power = 0
	PowerOn  Power = 1
)

var powerMap = map[Power]string{
	PowerOff: "Off",
	PowerOn:  "On",
}

func (p *Power) setUrlValues(v url.Values) {
	v.Set("pow", strconv.Itoa(int(*p)))
}

func (p *Power) decode(s string) error {
	switch s {
	case "0":
		*p = Power(PowerOff)
	case "1":
		*p = Power(PowerOn)
	default:
		return fmt.Errorf("unknown pwr value: %s", s)
	}
	return nil
}

func (p *Power) String() string {
	v, ok := powerMap[*p]
	if !ok {
		return fmt.Sprintf("Unknown Power [%d]", int(*p))
	}
	return v
}

// Mode is the operating mode of the Daikin unit.
type Mode int

// The valid modes supported by the Daikin Wifi module (not all units
// may support all values).
const (
	ModeDehumidify Mode = 2
	ModeCool       Mode = 3
	ModeHeat       Mode = 4
	ModeFan        Mode = 6
	ModeAuto       Mode = 0
	ModeAuto1      Mode = 1
	ModeAuto7      Mode = 7
)

var modeMap = map[Mode]string{
	ModeDehumidify: "Dehumidify",
	ModeCool:       "Cool",
	ModeHeat:       "Heat",
	ModeFan:        "Fan",
	ModeAuto:       "Auto",
	ModeAuto1:      "Auto",
	ModeAuto7:      "Auto",
}

func (m *Mode) String() string {
	if v, ok := modeMap[*m]; ok {
		return v
	}
	return fmt.Sprintf("Unknown Mode [%d]", *m)
}

func (m *Mode) setUrlValues(v url.Values) {
	v.Set("mode", strconv.Itoa(int(*m)))
}

func (m *Mode) decode(s string) error {
	switch s {
	case "2":
		*m = Mode(ModeDehumidify)
	case "3":
		*m = Mode(ModeCool)
	case "4":
		*m = Mode(ModeHeat)
	case "6":
		*m = Mode(ModeFan)
	case "0":
		*m = Mode(ModeAuto)
	case "1":
		*m = Mode(ModeAuto1)
	case "7":
		*m = Mode(ModeAuto7)
	default:
		return fmt.Errorf("unknown mode value: %s", s)
	}
	return nil
}

// Fan is the fan speed of the Daikin unit.
type Fan string

// Fan values. Not all may be valid on all models.
const (
	FanAuto   Fan = "A"
	FanSilent Fan = "B"
	Fan1      Fan = "3"
	Fan2      Fan = "4"
	Fan3      Fan = "5"
	Fan4      Fan = "6"
	Fan5      Fan = "7"
)

var fanMap = map[Fan]string{
	FanAuto:   "Auto",
	FanSilent: "Silent",
	Fan1:      "1",
	Fan2:      "2",
	Fan3:      "3",
	Fan4:      "4",
	Fan5:      "5",
}

func (f *Fan) setUrlValues(v url.Values) {
	v.Set("f_rate", string(*f))
}

func (f *Fan) decode(s string) error {
	switch s {
	case "A":
		*f = Fan(FanAuto)
	case "B":
		*f = Fan(FanSilent)
	case "3":
		*f = Fan(Fan1)
	case "4":
		*f = Fan(Fan2)
	case "5":
		*f = Fan(Fan3)
	case "6":
		*f = Fan(Fan4)
	case "7":
		*f = Fan(Fan5)
	default:
		return fmt.Errorf("unknown pwr value: %s", s)
	}
	return nil
}

func (f *Fan) String() string {
	v, ok := fanMap[*f]
	if !ok {
		return fmt.Sprintf("Unknown Fan [%v]", *f)
	}
	return v
}

// FanDir is the louvre swing setting of the Daikin unit.
type FanDir int

// Supported louve settings. Not all models will support all values.
const (
	FanDirStopped    FanDir = 0
	FanDirVertical   FanDir = 1
	FanDirHorizontal FanDir = 2
	FanDirBoth       FanDir = 3
)

var fanDirMap = map[FanDir]string{
	FanDirStopped:    "Stopped",
	FanDirVertical:   "Vertical",
	FanDirHorizontal: "Horizontal",
	FanDirBoth:       "Both",
}

func (f *FanDir) setUrlValues(v url.Values) {
	v.Set("f_dir", strconv.Itoa(int(*f)))
}

func (f *FanDir) decode(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid f_dir value: %s (err=%v)", s, err)
	}
	fd := FanDir(v)
	if _, ok := fanDirMap[fd]; !ok {
		return fmt.Errorf("unknown f_dir value: %s", s)
	}
	*f = fd
	return nil
}

func (f *FanDir) String() string {
	v, ok := fanDirMap[*f]
	if !ok {
		return fmt.Sprintf("Unknown FanDir [%d]", int(*f))
	}
	return v
}

// Temperature is the set temperature of the Daikin unit, in Celcius.
type Temperature float64

func (t *Temperature) setUrlValues(v url.Values) {
	v.Set("stemp", t.String())
}

func (t *Temperature) decode(v string) error {
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("error parsing s_temp=%s: %v", v, err)
	}
	*t = Temperature(val)
	return nil
}

func (t *Temperature) String() string {
	return strconv.FormatFloat(float64(*t), 'f', 1, 64)
}

// Shum is the set humidity of the Daikin unit.
type Humidity int32

func (h *Humidity) String() string {
	return strconv.Itoa(int(*h))
}

func (h *Humidity) setUrlValues(v url.Values) {
	v.Set("shum", h.String())
}

func (h *Humidity) decode(v string) error {
	if v == "-" {
		v = "-1"
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("error parsing s_hum=%s: %v", v, err)
	}
	*h = Humidity(val)
	return nil
}

// Name is the human-readable name of the Daikin unit.
type Name string

func (n *Name) String() string {
	return string(*n)
}

func (n *Name) setUrlValues(v url.Values) {
	v.Set("name", url.PathEscape(n.String()))
}

func (n *Name) decode(s string) error {
	v, err := url.PathUnescape(s)
	if err != nil {
		return err
	}
	*n = Name(v)
	return nil
}

type WattHours int32

func (w *WattHours) String() string {
	return strconv.Itoa(int(*w))
}

func (w *WattHours) setUrlValues(v url.Values) {
	return
}

func (w *WattHours) decode(v string) error {
	if v == "-" {
		v = "-1"
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("error parsing watt hours=%s: %v", v, err)
	}
	*w = WattHours(val)
	return nil
}

type Minutes int32

func (m *Minutes) String() string {
	return strconv.Itoa(int(*m))
}

func (m *Minutes) setUrlValues(v url.Values) {
	return
}

func (m *Minutes) decode(v string) error {
	if v == "-" {
		v = "-1"
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("error parsing minutes=%s: %v", v, err)
	}
	*m = Minutes(val)
	return nil
}

// Daikin represents the settings of the Daikin unit.
type Daikin struct {
	// Address is the IP address of the unit.
	Address string
	// Some daikin units require an authentication token in the HTTP Headers
	Token string
	// Name is the human-readable name of the unit.
	Name Name
	// ControlInfo contains the environment control info.
	ControlInfo *ControlInfo
	// SensorInfo contains the environment sensor info.
	SensorInfo *SensorInfo
	// WeekPower contains daily power usage data for the past 7 days
	WeekPower *WeekPower
}

// SensorInfo represents current sensor values.
type SensorInfo struct {
	// HomeTemperature is the home (interior) temperature.
	HomeTemperature Temperature
	// OutsideTemperature is the external temperature.
	OutsideTemperature Temperature
	// Humidity is the current interior humidity.
	Humidity Humidity
}

func (s *SensorInfo) populate(values map[string]string) error {
	for k, v := range values {
		var err error
		switch k {
		case "htemp":
			err = s.HomeTemperature.decode(v)
		case "otemp":
			err = s.OutsideTemperature.decode(v)
		case "hhum":
			err = s.Humidity.decode(v)
		case "ret":
			if v != returnOk {
				err = fmt.Errorf("device returned error ret=%s", v)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SensorInfo) String() string {
	return fmt.Sprintf("in_temp: %s\nin_humidity: %s\nout_temp: %s\n", s.HomeTemperature.String(), s.Humidity.String(), s.OutsideTemperature.String())
}

// ControlInfo represents the control status of the unit.
type ControlInfo struct {
	// Power is the current power status of the unit.
	Power Power
	// Mode is the operating mode of the unit.
	Mode Mode
	// Fan is the fan speed of the unit.
	Fan Fan
	// FanDir is the fan louvre setting of the unit.
	FanDir FanDir
	// Temperature is the current set temperature of the unit.
	Temperature Temperature
	// Humidity is the set humidity of the unit.
	Humidity Humidity
}

func (c *ControlInfo) urlValues() url.Values {
	qStr := url.Values{}
	c.Power.setUrlValues(qStr)
	c.Mode.setUrlValues(qStr)
	c.Fan.setUrlValues(qStr)
	c.FanDir.setUrlValues(qStr)
	c.Temperature.setUrlValues(qStr)
	c.Humidity.setUrlValues(qStr)
	return qStr
}

func (c *ControlInfo) populate(values map[string]string) error {
	for k, v := range values {
		var err error
		switch k {
		case "pow":
			err = c.Power.decode(v)
		case "mode":
			err = c.Mode.decode(v)
		case "stemp":
			err = c.Temperature.decode(v)
		case "shum":
			err = c.Humidity.decode(v)
		case "f_rate":
			err = c.Fan.decode(v)
		case "f_dir":
			err = c.FanDir.decode(v)
		case "ret":
			if v != returnOk {
				err = fmt.Errorf("device returned error ret=%s", v)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ControlInfo) String() string {
	return fmt.Sprintf("pow: %s\nmode: %s\nstemp: %s\nshum: %s\nf_rate: %s\nf_dir: %s",
		c.Power.String(), c.Mode.String(), c.Temperature.String(), c.Humidity.String(), c.Fan.String(), c.FanDir.String())
}

// WeekPower represents power usage over the past 7 days
type WeekPower struct {
	TodayRuntime          Minutes
	TodayWattHours        WattHours
	YesterdayWattHours    WattHours
	ThreeDaysAgoWattHours WattHours
	FourDaysAgoWattHours  WattHours
	FiveDaysAgoWattHours  WattHours
	SixDaysAgoWattHours   WattHours
	SevenDaysAgoWattHours WattHours
}

// ret=OK,today_runtime=85,datas=5200/3800/5300/1800/2900/3900/1100
func (w *WeekPower) populate(values map[string]string) error {
	for k, v := range values {
		var err error
		switch k {
		case "today_runtime":
			err = w.TodayRuntime.decode(v)
		case "datas":
			elems := strings.Split(v, "/")
			if len(elems) != 7 {
				return fmt.Errorf("expected 7 elements in week power data, got %d", len(elems))
			}
			w.SevenDaysAgoWattHours.decode(elems[0])
			w.SixDaysAgoWattHours.decode(elems[1])
			w.FiveDaysAgoWattHours.decode(elems[2])
			w.FourDaysAgoWattHours.decode(elems[3])
			w.ThreeDaysAgoWattHours.decode(elems[4])
			w.YesterdayWattHours.decode(elems[5])
			w.TodayWattHours.decode(elems[6])
		case "ret":
			if v != returnOk {
				err = fmt.Errorf("device returned error ret=%s", v)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WeekPower) String() string {
	return fmt.Sprintf(
		"today_runtime: %d\nwatt_hours: %d %d %d %d %d %d %d %d\n",
		w.TodayRuntime.String(),
		w.SevenDaysAgoWattHours.String(),
		w.SixDaysAgoWattHours.String(),
		w.FiveDaysAgoWattHours.String(),
		w.FourDaysAgoWattHours.String(),
		w.ThreeDaysAgoWattHours.String(),
		w.YesterdayWattHours.String(),
		w.TodayWattHours.String(),
	)
}

func (d *Daikin) parseResponse(resp *http.Response) (map[string]string, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(body)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 1 {
		return nil, fmt.Errorf("Have %d rows of records, want just one", len(records))
	}

	values := map[string]string{}
	for _, rec := range records[0] {
		parts := strings.SplitN(rec, "=", 2)
		values[parts[0]] = parts[1]
	}
	return values, nil

}

// Set configures the current setting to the unit.
func (d *Daikin) SetControlInfo() error {
	qStr := d.ControlInfo.urlValues()
	resp, err := d.httpGet(fmt.Sprintf("%s?%s", uriSetControlInfo, qStr.Encode()))
	if err != nil {
		return err
	}
	vals, err := d.parseResponse(resp)
	if err != nil {
		return err
	}
	if v := vals["ret"]; v != "OK" {
		return fmt.Errorf("device returned error ret=%s", v)
	}
	return nil
}

// GetControlInfo gets the current control settings for the unit.
func (d *Daikin) GetControlInfo() error {
	resp, err := d.httpGet(uriGetControlInfo)
	if err != nil {
		return err
	}
	d.ControlInfo = &ControlInfo{}
	vals, err := d.parseResponse(resp)
	if err != nil {
		return fmt.Errorf("GetControlInfo: %v", err)
	}
	return d.ControlInfo.populate(vals)
}

func (d *Daikin) httpGet(path string) (*http.Response, error) {
	var scheme string
	if d.Token == "" {
		scheme = "http"
	} else {
		scheme = "https"
	}
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s%s", scheme, d.Address, path), nil)
	if d.Token != "" {
		request.Header["X-Daikin-uuid"] = []string{d.Token}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(request)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != 200 {
		return resp, fmt.Errorf("GET %s request failed: %d", path, resp.StatusCode)
	}
	return resp, err
}

// GetSensorInfo gets the current sensor values for the unit.
func (d *Daikin) GetSensorInfo() error {
	resp, err := d.httpGet(uriGetSensorInfo)
	if err != nil {
		return err
	}
	d.SensorInfo = &SensorInfo{}
	vals, err := d.parseResponse(resp)
	if err != nil {
		return fmt.Errorf("GetSensorInfo: %v", err)
	}
	return d.SensorInfo.populate(vals)
}

func (d *Daikin) GetWeekPower() error {
	resp, err := d.httpGet(uriGetWeekPower)
	if err != nil {
		return err
	}
	d.WeekPower = &WeekPower{}
	vals, err := d.parseResponse(resp)
	if err != nil {
		return fmt.Errorf("GetWeekPower: %v", err)
	}
	return d.WeekPower.populate(vals)
}

func (d *Daikin) String() string {
	return fmt.Sprintf("name: %s\n%s\n%s\n", d.Name.String(), d.ControlInfo.String(), d.SensorInfo.String())
}

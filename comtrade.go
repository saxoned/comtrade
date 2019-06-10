package comtrade

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

//CFG文件
type CFG struct {
	StationName         string
	DeviceId            string
	RevisionYear        uint16
	ChannelNumber       uint //通道总数(模拟通道+数字通道)
	AnalogChannelNumber uint //模拟通道数量
	DigitChannelNumber  uint //数字通道数量
	AnalogChannelDetail []AnalogChannel
	DigitChannelDetail  []DigitChannel
	LineFrequency       float64 //非必须
	SampleRateNumber    uint
	SampleDetails       []SampleDetail
	StartTime           time.Time
	TriggerTime         time.Time
	DataFileType        string
	TimeFactor          float64
}

//模拟通道
type AnalogChannel struct {
	Id        uint    //索引
	Name      string  // 通道标识符
	Phase     string  //相位特征
	Element   string  //被监视的回路元件
	Unit      string  //通道单位
	A         float64 //通道乘数
	B         float64 //通道偏移加数
	Skew      uint8   //通道时滞
	Min       int     //最小值
	Max       int     //最大值
	Primary   float64 //通道电压或电流变换比一次系数
	Secondary float64 //通道电压或电流变换比二次系数
	PS        string  //一次(P),二次(S)
}

//数字通道
type DigitChannel struct {
}

//采样详情
type SampleDetail struct {
	Rate   float64 //以(Hz)为单位的采样率
	Number float64 //该采样率下最后一次采样数
}

// parse .cfg file
func (cfg *CFG) UnmarshalCfg(content []byte) (err error) {
	index := 0
	line := strings.Split(string(content), "\n")

	if len(line[0]) > 0 && line[0][len(line[0])-1] == '\r' {
		line = strings.Split(string(content), "\r\n")
	}
	if len(line) < 9 {
		return errors.New("invalid cfg file")
	}
	tempList := strings.Split(line[index], ",")
	if len(tempList) < 3 {
		return checkError("cfg file", index)
	}

	//station name
	cfg.StationName = tempList[0]

	//deviceId
	cfg.DeviceId = tempList[1]

	//revision year
	year, err := strconv.ParseUint(tempList[2], 10, 32)
	if err != nil {
		return checkError("year", index)
	}
	cfg.RevisionYear = uint16(year)
	index++
	tempList = strings.Split(line[index], ",")
	if len(tempList) < 3 {
		return checkError("cfg file", index)
	}

	//channelNumber
	channelNumber, err := strconv.ParseUint(tempList[0], 10, 32)
	if err != nil {
		return checkError("channelNumber", index)
	}
	cfg.ChannelNumber = uint(channelNumber)

	//analogChannelNumber
	if !strings.HasSuffix(tempList[1], "A") {
		return checkError("analogChannelNumber", index)
	}
	analogChannelNumber, err := strconv.ParseUint(tempList[1][:len(tempList[1])-1], 10, 32)
	cfg.AnalogChannelNumber = uint(analogChannelNumber)

	//digitChannelNumber
	if !strings.HasSuffix(tempList[2], "D") {
		return checkError("digitChannelNumber", index)
	}

	digitChannelNumber, err := strconv.ParseUint(tempList[2][:len(tempList[2])-1], 10, 32)
	cfg.DigitChannelNumber = uint(digitChannelNumber)
	if channelNumber != analogChannelNumber+digitChannelNumber {
		return checkError("channelNumber", index)
	}

	//analogChannelDetail
	analogChannelDetail := make([]AnalogChannel, analogChannelNumber)
	for i := 0; i < int(analogChannelNumber); i++ {
		index++
		tempList = strings.Split(line[index], ",")
		if len(tempList) < 13 {
			return checkError("analogChannelNumber", index)
		}
		id, err := strconv.ParseUint(tempList[0], 10, 32)
		if err != nil {
			return checkError("id", index)
		}
		analogChannelDetail[i].Id = uint(id)
		analogChannelDetail[i].Name = tempList[1]
		analogChannelDetail[i].Phase = tempList[2]
		analogChannelDetail[i].Element = tempList[3]
		analogChannelDetail[i].Unit = tempList[4]
		a, err := strconv.ParseFloat(tempList[5], 32)
		if err != nil {
			return checkError("value a", index)
		}
		analogChannelDetail[i].A = a
		b, err := strconv.ParseFloat(tempList[6], 32)
		if err != nil {
			return checkError("value b", index)
		}
		analogChannelDetail[i].B = b
		skew, err := strconv.ParseUint(tempList[7], 10, 32)
		if err != nil {
			return checkError("skew", index)
		}
		analogChannelDetail[i].Skew = uint8(skew)
		min, err := strconv.ParseInt(tempList[8], 10, 32)
		if err != nil {
			return checkError("min", index)
		}
		analogChannelDetail[i].Min = int(min)
		max, err := strconv.ParseInt(tempList[9], 10, 32)
		if err != nil {
			return checkError("max", index)
		}
		analogChannelDetail[i].Max = int(max)
		primary, err := strconv.ParseFloat(tempList[10], 32)
		if err != nil {
			return checkError("primary", index)
		}
		analogChannelDetail[i].Primary = primary
		secondary, err := strconv.ParseFloat(tempList[11], 32)
		if err != nil {
			return checkError("secondary", index)
		}
		analogChannelDetail[i].Secondary = secondary
		analogChannelDetail[i].PS = tempList[12]

	}
	cfg.AnalogChannelDetail = analogChannelDetail

	//analogChannelDetail
	digitChannelDetail := make([]DigitChannel, analogChannelNumber)
	for i := 0; i < int(digitChannelNumber); i++ {
		index++
		//todo
	}
	cfg.DigitChannelDetail = digitChannelDetail
	index++

	lf, err := strconv.ParseFloat(line[index], 32)
	if err != nil {
		return checkError("line frequency", index)
	}
	cfg.LineFrequency = lf
	index++
	sampleRateNumber, err := strconv.ParseUint(line[index], 10, 32)
	if err != nil {
		return checkError("sampleRateNum", index)
	}
	cfg.SampleRateNumber = uint(sampleRateNumber)

	//sampleDetail
	sampleDetails := make([]SampleDetail, sampleRateNumber)
	for i := 0; i < int(sampleRateNumber); i++ {
		index++
		tempList = strings.Split(line[index], ",")
		if len(tempList) != 2 {
			return checkError("sampleDetail", index)
		}
		rate, err := strconv.ParseFloat(tempList[0], 32)
		if err != nil {
			return checkError("rate", index)
		}
		sampleDetails[i].Rate = rate

		number, err := strconv.ParseFloat(tempList[1], 32)
		if err != nil {
			return checkError("number", index)
		}
		sampleDetails[i].Number = number

	}
	cfg.SampleDetails = sampleDetails
	index++

	//startTime
	startTime, err := time.Parse(TimeFormat, line[index])
	if err != nil {
		return checkError("startTime", index)
	}
	cfg.StartTime = startTime
	index++

	//triggerTime
	triggerTime, err := time.Parse(TimeFormat, line[index])
	if err != nil {
		return checkError("triggerTime", index)
	}
	cfg.TriggerTime = triggerTime
	index++
	cfg.DataFileType = line[index]
	index++

	//timeFactor
	timeFactor, err := strconv.ParseFloat(line[index], 32)
	if err != nil {
		return checkError("timeFactor", index)
	}
	cfg.TimeFactor = timeFactor

	return nil

}

//parse .dat file
func (cfg *CFG) UnmarshalDat(content []byte) (result [][]int, err error) {

	nb := 8 + int(cfg.AnalogChannelNumber)<<1 + int(math.Ceil(float64(cfg.DigitChannelNumber)/float64(16)))<<1

	for n := 0; n < int(cfg.ChannelNumber); n++ {
		r := make([]int, 0)

		for i := 0; i < int(cfg.SampleDetails[0].Number); i++ {
			s := content[i*nb : i*nb+nb]

			/*var data struct {
				Sample int32
				Stamp  int32
			}
			*/
			value := make([]int16, (nb-8)/2)

			/*	err = binary.Read(bytes.NewReader(s[:8]), binary.LittleEndian, &data)
				if err != nil {
					return nil, err
				}
			*/
			err = binary.Read(bytes.NewReader(s[8:]), binary.LittleEndian, &value)
			if err != nil {
				return nil, err
			}

			r = append(r, int(float64(value[n])*cfg.AnalogChannelDetail[n].A+cfg.AnalogChannelDetail[n].B))
		}
		result = append(result, r)
	}

	return result, nil
}

func checkError(v string, index int) error {
	return errors.New(fmt.Sprintf("invalid %s in line %d", v, index+1))
}

const TimeFormat = "02/01/2006,15:04:05.000000"

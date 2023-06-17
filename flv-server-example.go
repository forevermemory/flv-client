package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// read local flv file: flv.flv, and start http server, support http-flv stream
// the video data is circular
// no audio !!!
// no audio !!!
// no audio !!!

type Tag struct {
	Type          uint8
	SoundFormat   uint8
	SoundRate     uint8
	SoundSize     uint8
	SoundType     uint8
	AACPacketType uint8
	FrameType     uint8
	VideoFormat   uint8
	AVCPacketType uint8
	Time          uint32
	CTime         int32
	StreamId      uint32
	Header, Data  []byte

	_curTagSize uint32
}

func main() {
	data, err := ioutil.ReadFile("./flv_video_autio.flv")
	if err != nil {
		fmt.Println("read flv file err", err)
		return
	}

	var length = len(data)
	var flvindex = 0
	var flvdata = data[flvindex:]

	//  length=9
	var flvheader = flvdata[:9]
	var flvbody = flvdata[9+4:]

	fmt.Println("file size:", length)
	fmt.Println("flvheader", flvheader)

	var tmp []byte
	var readSize = 0
	var tags = make([]*Tag, 0)

	var tagScript *Tag
	var tagSpspps *Tag

	for {

		var _type = flvbody[0]

		tmp = []byte{0, flvbody[1], flvbody[2], flvbody[3]}
		var _tagDataSize = binary.BigEndian.Uint32(tmp)

		tmp = []byte{0, flvbody[4], flvbody[5], flvbody[6]}
		var _timestreamp = binary.BigEndian.Uint32(tmp)

		var _timestampExtended = flvbody[7]

		tmp = []byte{0, flvbody[8], flvbody[9], flvbody[10]}
		var _streamid = binary.BigEndian.Uint32(tmp)

		// 1 + 3 + 3 + 1 + 3 + n
		var _tagData = flvbody[11 : 11+_tagDataSize]

		if 11+4+_tagDataSize >= uint32(len(flvbody)) {
			break
		}
		tmp = flvbody[11+_tagDataSize : 11+4+_tagDataSize]
		var _curTagSize = binary.BigEndian.Uint32(tmp)

		fmt.Printf("type: 0x%02X, tagDataSize: %v,\ttsp: %v,\ttspExtended: %v,\tstreamid:%v,\tpreviousTagSize: %v \n",
			_type, _tagDataSize, _timestreamp, _timestampExtended, _streamid, _curTagSize)

		// append tag
		t := Tag{
			Type:        _type,
			Header:      flvbody[:11],
			Data:        _tagData,
			_curTagSize: _curTagSize,
		}

		// move flvbody ptr
		flvbody = flvbody[_tagDataSize+11+4:]
		readSize = readSize + 11 + int(_tagDataSize) + 4

		// is read file end
		fmt.Println(readSize, len(flvbody), length)
		if readSize+11+4 >= length {
			break
		}

		if _type == 0x12 {
			tagScript = &t
			continue
		}
		if _type == 0x09 && tagSpspps == nil {
			tagSpspps = &t
			continue
		}

		tags = append(tags, &t)
	}
	var tagLength = len(tags)
	// start http flv server
	fmt.Println("tags length:", tagScript._curTagSize)
	fmt.Println("tags length:", tagSpspps._curTagSize)
	fmt.Println("tags length:", tagLength)

	http.HandleFunc("/flv.flv", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("xxxxxxxxxxxxxxxx")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "video/x-flv")
		w.Header().Set("Expires", "-1")
		w.Header().Set("Transfer-Encoding", "chunked")

		// start write flvheader script first
		w.Write(flvheader)
		w.Write([]byte{0, 0, 0, 0})

		var preSizeArr = make([]byte, 4)

		w.Write(tagScript.Header)
		w.Write(tagScript.Data)
		binary.BigEndian.PutUint32(preSizeArr, tagScript._curTagSize)
		w.Write(preSizeArr)

		w.Write(tagSpspps.Header)
		w.Write(tagSpspps.Data)
		binary.BigEndian.PutUint32(preSizeArr, tagSpspps._curTagSize)
		w.Write(preSizeArr)

		var index uint32 = 0 // fake timestamp
		for {

			for _, tag := range tags {

				// rebuild data size and timestamp
				var head = []byte{tag.Header[0], tag.Header[1], tag.Header[2], tag.Header[3]}

				var tsp = []byte{0, 0, 0, 0}
				binary.BigEndian.PutUint32(tsp, index)
				head = append(head, tsp[1], tsp[2], tsp[3])
				head = append(head, tag.Header[7:]...)
				w.Write(head)
				w.Write(tag.Data)
				fmt.Println("preTag._curTagSize ", tag._curTagSize, tag._curTagSize, len(tag.Data)+len(tag.Header))

				binary.BigEndian.PutUint32(preSizeArr, tag._curTagSize)
				if _, err := w.Write(preSizeArr); err != nil {
					fmt.Println("client exit")
					return
				}
				time.Sleep(time.Millisecond * 100)
				index += 100 // fake timestamp interval
			}

		}
	})
	fmt.Println("http server on 13370")
	http.ListenAndServe(":13370", nil)

}

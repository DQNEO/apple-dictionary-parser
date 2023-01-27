package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
)

func main() {
	rawDumpFile := os.Args[1]
	all, err := os.ReadFile(rawDumpFile)
	if err != nil {
		panic(err)
	}
	entries := bytes.Split(all, []byte{'\n'})
	for _, ent := range entries {
		if len(ent) == 0 {
			// Possibly end of file
			continue
		}
		title, rawBody, found := bytes.Cut(ent, []byte(":::"))
		if !found {
			panic("failed to Cut:" + (string(ent)))
		}

		_ = title
		var d interface{}
		rawBody = []byte(`<?xml version="1.0" encoding="UTF-8" ?>
<d:entry xmlns:d="http://www.apple.com/DTDs/DictionaryService-1.0.rng" id="m_en_gbus1179660" d:title="007" class="entry"><span class="hg x_xh0"><span role="text" class="hw">007 </span><span dialect="AmE" prxid="007_us_5d24" prlexid="optra0156151.002" class="prx"> | <span d:prn="US" dialect="AmE" class="ph t_respell">ˌdəbəl ˌō ˈsevən<d:prn></d:prn></span><span d:prn="IPA" dialect="AmE" soundFile="007#_us_1" media="online" class="ph">ˌdəbəl ˌoʊ ˈsɛvən<d:prn></d:prn></span> | </span></span><span class="sg"><span id="m_en_gbus1179660.005" class="se1 x_xd0"><span role="text" class="posg x_xdh"><span d:pos="1" class="pos"><span class="gp tg_pos">noun </span><d:pos></d:pos></span></span><span id="m_en_gbus1179660.006" class="msDict x_xd1 t_core"><span d:def="1" role="text" class="df">the fictional British secret agent James Bond, or someone based on, inspired by, or reminiscent of him<span class="gp tg_df">. </span><d:def></d:def></span></span></span></span></d:entry>
`)
		err := xml.Unmarshal(rawBody, &d)
		if err != nil {
			panic(err)
		}

		//os.Stdout.Write(title)
		//fmt.Print(":::")
		fmt.Printf("%v", d)
		//fmt.Print("\n")
		return
	}
}

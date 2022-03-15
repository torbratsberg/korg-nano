package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/micmonay/keybd_event"
	"gitlab.com/gomidi/midi"
	. "gitlab.com/gomidi/midi/midimessage/channel" // (Channel Messages)
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

// ===============================================================

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	f, err := os.Open("test.mp3")
	check(err)

	streamer, format, err := mp3.Decode(f)
	check(err)
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// ===================

	drv, err := rtmididrv.New()
	check(err)

	// make sure to close the driver at the end
	defer drv.Close()

	ins, err := drv.Ins()
	check(err)

	// takes the first input
	in := ins[0]

	fmt.Printf("opening MIDI Port %v\n", in)
	check(in.Open())

	defer in.Close()

	volume := 0.

	actions := map[int64]func(v uint8){
		52: func(uint8) { // Volume down
			volume -= 0.5
			cmd := exec.Command("osascript", "-e", fmt.Sprintf("set Volume %f", volume))
			cmd.Stdout = os.Stdout
			cmd.Run()
		},

		53: func(uint8) { // Volume up
			volume += 0.5
			cmd := exec.Command("osascript", "-e", fmt.Sprintf("set Volume %f", volume))
			cmd.Stdout = os.Stdout
			cmd.Run()
		},

		54: func(uint8) { // Pause music
			cmd := exec.Command("osascript", "-e 'tell application \"spotify\" to play'")
			fmt.Println(cmd)
			cmd.Stdout = os.Stdout
			cmd.Run()
		},

		55: func(uint8) { // Play music
			cmd := exec.Command("osascript", "-e 'tell application \"spotify\" to pause'")
			fmt.Println(cmd)
			cmd.Stdout = os.Stdout
			cmd.Run()
		},

		67: func(v uint8) { // <CR>
			kb, err := keybd_event.NewKeyBonding()
			check(err)
			if v < 120 {
				kb.SetKeys(keybd_event.VK_ENTER)
			} else {
				kb.SetKeys(keybd_event.VK_ENTER, keybd_event.VK_ENTER, keybd_event.VK_ENTER, keybd_event.VK_ENTER, keybd_event.VK_ENTER)
			}
			err = kb.Launching()
			check(err)
		},
	}

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(func(pos *reader.Position, msg midi.Message) {

			// inspect
			// fmt.Println(msg)

			switch v := msg.(type) {
			case NoteOn:
				fmt.Printf("NoteOn - key: %v velocity: %v\n", v.Key(), v.Velocity())

				if actions[int64(v.Key())] != nil {
					go actions[int64(v.Key())](v.Velocity())
				}
			}
		}),
	)

	// listen for MIDI
	err = rd.ListenTo(in)
	check(err)

	for {
	}
}

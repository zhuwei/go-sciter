package main

import (
	"log"

	"github.com/zhuwei/go-sciter"
	"github.com/zhuwei/go-sciter/window"
)

func main() {
	//w, err := window.New(sciter.SW_TITLEBAR|sciter.SW_RESIZEABLE|sciter.SW_CONTROLS|sciter.SW_MAIN|sciter.SW_ENABLE_DEBUG, &sciter.Rect{0, 0, 800, 600})
	w, err := window.NewCenter(sciter.SW_TITLEBAR|sciter.SW_RESIZEABLE|sciter.SW_CONTROLS|sciter.SW_MAIN|sciter.SW_ENABLE_DEBUG, 800, 600, true)
	if err != nil {
		log.Fatal(err)
	}

	w.LoadFile("test.html")
	w.SetTitle("Example")
	setEventHandler(w)
	w.Show()
	w.Run()
}

func setEventHandler(w *window.Window) {
	w.DefineFunction("getNetInformation", func(args ...*sciter.Value) *sciter.Value {
		log.Println("Args Length:", len(args))
		log.Println("Arg 0:", args[0], args[0].IsInt())
		log.Println("Arg 1:", args[1], args[1].IsString())
		log.Println("Arg 2: IsFunction", args[2], args[2].IsFunction())
		log.Println("Arg 2: IsObjectFunction", args[2], args[2].IsObjectFunction())
		log.Println("args[3].IsMap():", args[3].IsMap(), args[3].Length())
		log.Println("args[3].IsObject():", args[3].IsObject(), args[3].Length(), args[3].Get("str"))
		args[3].EnumerateKeyValue(func(key, val *sciter.Value) bool {
			log.Println(key, val)
			return true
		})
		fn := args[2]
		fn.Invoke(sciter.NullValue(), "[Native Script]", sciter.NewValue("OK"))
		ret := sciter.NewValue()
		ret.Set("ip", sciter.NewValue("127.0.0.1"))
		return ret
	})
}

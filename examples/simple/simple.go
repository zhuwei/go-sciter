package main

import (
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	jwlib "github.com/zhuwei/go-sciter/examples/simple/jwlib"

	"github.com/zhuwei/go-sciter"
	"github.com/zhuwei/go-sciter/window"
)

var gCh chan map[string]interface{}

func main() {

	//必须要先声明defer，否则不能捕获到panic异常
	defer func() {
		if r := recover(); r != nil {
			err := errors.New(``)
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknow panic")
			}
			jwlib.WriteLog(`[recover]` + err.Error())
		}
	}()

	jwlib.WriteLog(`程序启动`)

	jwlib.WriteLog(`系统架构:` + runtime.GOARCH)
	jwlib.WriteLog(`操作系统:` + runtime.GOOS)

	var sciterDll string
	if runtime.GOARCH == `386` {
		sciterDll = `lib\sciter32.dll`
	} else {
		sciterDll = `lib\sciter64.dll`
	}
	copyFile(sciterDll, `sciter.dll`)

	gCh = make(chan map[string]interface{}, 1)

	var err error
	//w, err := window.New(sciter.SW_TITLEBAR|sciter.SW_RESIZEABLE|sciter.SW_CONTROLS|sciter.SW_MAIN|sciter.SW_ENABLE_DEBUG, &sciter.Rect{0, 0, 800, 600})
	w, err := window.NewCenter(sciter.SW_TITLEBAR|sciter.SW_RESIZEABLE|sciter.SW_CONTROLS|sciter.SW_MAIN|sciter.SW_ENABLE_DEBUG, 800, 600, true)
	if err != nil {
		log.Fatal(err)
	}

	w.LoadFile("test.html")
	w.SetTitle("Example")
	setEventHandler(w)
	w.Show()
	w.RunWithHandler(messageHanlder)

	jwlib.WriteLog(`程序结束`)
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

	w.DefineFunction("loadData", func(args ...*sciter.Value) *sciter.Value {
		log.Println("args[0].IsObject():", args[0].IsObject(), args[0].Length(), args[0].Get("str"))

		go func(v *sciter.Value) {

			time.Sleep(10 * 1000 * time.Millisecond)

			log.Println("Thread Over")

			result := make(map[string]interface{})
			result["ip"] = "127.0.0.1"
			result["v"] = v
			gCh <- result

		}(args[1])

		return sciter.NewValue()
	})
}

func messageHanlder(w *window.Window) {
	select {
	case v, ok := <-gCh:
		// 读出来一个，v=10, ok=true
		if ok {
			log.Println(v["ip"])

			ret := sciter.NewValue()
			ret.Set("ip", sciter.NewValue("127.0.0.1"))
			root, _ := w.GetRootElement()
			root.CallFunction("loadDataCallback", ret)

			//v["v"].(*sciter.Value).Invoke(sciter.NullValue(), "[Native Script]", ret)
		} else {
			log.Println("false")
			ok = true
		}
	default:
	}
}

//拷贝文件  要拷贝的文件路径 拷贝到哪里
func copyFile(source, dest string) bool {
	if source == "" || dest == "" {
		log.Println("source or dest is null")
		return false
	}
	//打开文件资源
	source_open, err := os.Open(source)
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer source_open.Close()
	//只写模式打开文件 如果文件不存在进行创建 并赋予 644的权限。详情查看linux 权限解释
	dest_open, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 644)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	defer dest_open.Close()
	//进行数据拷贝
	_, copy_err := io.Copy(dest_open, source_open)
	if copy_err != nil {
		log.Println(copy_err.Error())
		return false
	} else {
		return true
	}
}

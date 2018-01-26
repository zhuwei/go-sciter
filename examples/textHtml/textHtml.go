package main

import (
	"log"

	"github.com/zhuwei/go-sciter"
	"github.com/zhuwei/go-sciter/window"
)

const (
	html = `
<html resizeable window-frame="solid-with-shadow">
	<head>
    	<title></title>
		<style>
		html {
			background: white;
			width: 400dip;
			height: 300dip;
			border: 1dip solid threedlightshadow
		}
		body {
        	background-image: url(bg.png);
		}
		</style>
	</head>
<body>
  <H1>test for element.Text and element.Html</H1>
  <div id="btns">
    <button id="csss">CSS!!! Button</button>
    <button id="functor">Tiscript Native Functor Button</button>
    <button id="native">Native Handler Button</button>
    <button id="sumall">Do Sum</button>
    <button id="mcall">Method Call</button>
    <form >
      <label>X (Ctrl+1):</label><input type="integer" name="x" step="20" accesskey="^1" />
      <label>Y (Ctrl+2):</label><input type="integer" name="y" step="20" accesskey="^2" />
    </form>
  </div>
  <div id="output">
    
  </div>
</body>
</html>
`
)

func main() {

	rect := sciter.NewRect(200, 400, 300, 400)
	//DefaultWindowCreateFlag = SW_TITLEBAR | SW_RESIZEABLE | SW_CONTROLS | SW_MAIN | SW_ENABLE_DEBUG
	w, err := window.New(sciter.SW_ALPHA|sciter.SW_GLASSY, rect)
	if err != nil {
		log.Fatal("sciter create window failed", err)
	}

	w.SetTitle("test for element.Text and element.Html")
	w.LoadHtml(html, "")

	root, err := w.GetRootElement()
	if err != nil {
		log.Fatal(err)
	}
	text, err := root.Text()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("text:", text)
	text, err = root.Html(false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("html:", text)

	w.SetTitle("Example")
	w.Show()
	w.Run()
}

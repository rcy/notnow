package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var ButtonClass = "border rounded bg-orange-200 hover:bg-orange-300 p-1"

func Layout(title string, body ...Node) Node {
	return HTML5(HTML5Props{
		Title: title,
		Head: []Node{
			Script(Src("https://cdn.tailwindcss.com")),
		},
		Body: body,
	})
}

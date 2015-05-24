package tpl

import (
	"bytes"
)

func Index() []byte {
	_buffer := new(bytes.Buffer)
	_buffer.WriteString("<html lang=\"zh-CN\">")
	head(_buffer, "")
	_buffer.WriteString("<body><div id=\"sidebar\"><div id=\"logo\"></div><div id=\"sidebar-content\"></div></div><div id=\"content\"><div id=\"messages\"></div><div id=\"input-box\"><div id=\"toolbar\"><button id=\"edit\" class=\"tab\" disabled>编辑</button><button id=\"preview\" class=\"tab\">预览</button><button id=\"submit\" class=\"button\" disabled>发送</button></div><textarea id=\"input\"></textarea><div id=\"preview-box\"></div></div></div><div id=\"init\"><div id=\"i\"><p>请输入用户名：</p><input id=\"inputName\"/><button id=\"submitName\">确定</button></div></div></body></html>")
	return _buffer.Bytes()
}

func head(_buffer *bytes.Buffer, title string) {
	_buffer.WriteString("<head><meta charset=\"utf-8\"><title>")
	_buffer.WriteString(title)
	_buffer.WriteString("</title><link href=\"./static/css/main.css\" media=\"all\" rel=\"stylesheet\"><link href=\"./static/css/monokai_sublime.css\" media=\"all\" rel=\"stylesheet\"><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/main.js\"></script><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/marked.js\"></script><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/highlight.pack.js\"></script></head>")
}

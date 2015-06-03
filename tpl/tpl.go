package tpl

import (
	"bytes"
)

func Index() []byte {
	_buffer := new(bytes.Buffer)
	_buffer.WriteString("<html lang=\"zh-CN\">")
	head(_buffer, "")
	_buffer.WriteString("<body><div id=\"sidebar\"><div id=\"logo\"></div><div id=\"sidebar-content\"><button id=\"new-topic\" class=\"button\">发表新主题</button></div></div><div id=\"content\"><div id=\"messages\"></div><div class=\"input-box\"><div><button id=\"edit\" class=\"tab\" disabled>编辑</button><button id=\"preview\" class=\"tab\">预览</button><button id=\"submit\" class=\"button\" disabled>发送</button></div><textarea id=\"input\" class=\"input\"></textarea><div id=\"preview-box\" class=\"preview-box\"></div></div></div><div id=\"init\"><div id=\"i\"><p>请输入用户名：</p><input id=\"inputName\" /><button id=\"submitName\">确定</button></div></div><div class=\"background\"><div id=\"new-topic-box\"><input id=\"n-title\" placeholder=\"标题\" /><div class=\"input-box\"><div class=\"toolbar\"><button id=\"n-edit\" class=\"tab\" disabled>编辑</button><button id=\"n-prev\" class=\"tab\">预览</button><button id=\"n-subm\" class=\"button\" disabled>发布</button></div><textarea id=\"n-edit-box\" class=\"input maximum\"></textarea><div id=\"n-prev-box\" class=\"preview-box maximum\"></div></div></div></div></body></html>")
	return _buffer.Bytes()
}

func head(_buffer *bytes.Buffer, title string) {
	_buffer.WriteString("<head><meta charset=\"utf-8\"><title>")
	_buffer.WriteString(title)
	_buffer.WriteString("</title><link href=\"./static/css/main.css\" media=\"all\" rel=\"stylesheet\"><link href=\"./static/css/monokai_sublime.css\" media=\"all\" rel=\"stylesheet\"><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/main.js\"></script><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/marked.js\"></script><script type=\"text/javascript\" charset=\"utf-8\" src=\"./static/js/highlight.pack.js\"></script></head>")
}

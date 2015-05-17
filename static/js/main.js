var content = document.getElementById("content")
content.style["width"] = document.body.clientWidth * 0.8 - 48 + "px"

window.onResize = function(){
    content.style["width"] = document.body.clientWidth * 0.8 - 48 + "px"
}

socket = new WebSocket(document.location.protocol == "https:" ? "wss:" : "ws:" + "//" + document.location.host + "/ws")

socket.onopen = function () {
    socket.send("Hi")
}

socket.onerror = function (e) {
    console.log("err from connect " + e)
}

socket.onclose = function (e) {
    console.log("connection closed (" + e.code + ")")
}

socket.onmessage = function (e) {
    console.log(e.data)
}
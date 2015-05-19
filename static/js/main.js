window.onload = function() {
    var content = document.getElementById("content");
    content.style["width"] = document.body.clientWidth * 0.8 - 48 + "px";

    window.onresize = function(){
        content.style["width"] = document.body.clientWidth * 0.8 - 48 + "px";
    };

    var submitBtn = document.getElementById("submit");
    var inputBox = document.getElementById("input-box");
    var wsUrl = document.location.protocol == "https:" ? "wss:" : "ws:" + "//" + document.location.host + "/ws";
    var socket = new WebSocket(wsUrl);

    inputBox.oninput = function(e) {
        if (inputBox.value === "") {
            if (!submitBtn.disabled) {
                submitBtn.disabled = true;
            }
        } else if (submitBtn.disabled && socket.readyState === socket.OPEN) {
            submitBtn.disabled = false;
        }
    };

    submitBtn.disabled = true;

    submitBtn.onclick = function(e) {
        if (inputBox.value !== "") {
            submitBtn.disabled = true;
            submitBtn.textContent = "发送中……";
            socket.send(inputBox.value);
            inputBox.value = "";
            submitBtn.textContent = "发送";
        }
    };

    socket.onopen = function(e) {
        socket.send("Hi");
    };

    socket.onerror = function(e) {
        console.log("err from connect " + e);
    };

    socket.onclose = function(e) {
        console.log("connection closed (" + e.code + ")");
    };

    var msgBox = document.getElementById('messages');

    socket.onmessage = function(e) {
        console.log(e.data);
        var data = JSON.parse(e.data);
        var msg = document.createElement("div");
        msg.innerHTML = marked(data["msg"]);

        // 高亮代码块
        var codeBlocks = msg.getElementsByTagName("pre");
        for (var i = 0; i < codeBlocks.length; i++) {
            hljs.highlightBlock(codeBlocks[i]);
        }

        msgBox.appendChild(msg);
        // 滚动到底部
        msgBox.scrollTop = msgBox.scrollHeight;
    };

    // 监听快捷键
    document.onkeydown = function(keys) {
        // Ctrl + Enter
        if(keys.ctrlKey && keys.keyCode == 13){
            submitBtn.click();
        }
    };
};
var wsUrl = document.location.protocol == "https:" ? "wss:" : "ws:" + "//" + document.location.host + "/ws";

window.onload = function() {
    window.onresize = function () {
        var i = document.getElementById("i");
        i.style.marginTop = String(document.body.clientHeight/2 - i.offsetHeight/2);
        i.style.marginLeft = String(document.body.clientWidth/2 - i.offsetWidth/2);
    };
    window.onresize();

    document.getElementById("inputName").focus();

    var submit = document.getElementById("submitName");

    document.onkeydown = function(keys) {
        // Enter
        if(keys.keyCode == 13){
            submit.click();
        }
    };

    submit.onclick = function () {
        var socket = new WebSocket(wsUrl);

        socket.onopen = function(e) {
            socket.send(document.getElementById("inputName").value);
        };

        socket.onmessage = function (e) {
            if (JSON.parse(e.data)["error"] === true) {
                alert("name already exists");
            } else {
                socket.onmessage = onMsg;
                init(socket);
                document.getElementById("init").style.display = "none";
            }
        };
    };
};

function onMsg(e) {
        console.log(e.data);
        var msgBox = document.getElementById("messages");
        var content = document.getElementById("content");
        var data = JSON.parse(e.data);
        var msg = document.createElement("div");

        msg.className = "msg";
        msg.innerHTML = '<img class="avatar" src="' + data["avatar"] + '"/>';
        msg.innerHTML += '<div class="msg-body"><p class="msg-header"><a class=msg-name>' +
            data["name"] + '</a>' + data["time"] + '</p>' + marked(data["msg"]) + '</div>';

        // 高亮代码块
        var codeBlocks = msg.getElementsByTagName("pre");
        for (var i = 0; i < codeBlocks.length; i++) {
            hljs.highlightBlock(codeBlocks[i]);
        }

        msgBox.appendChild(msg);
        // 滚动到底部
        content.scrollTop = content.scrollHeight;
}

function init(socket) {
    var submitBtn = document.getElementById("submit");
    var input = document.getElementById("input");

    input.oninput = function(e) {
        if (input.value === "") {
            if (!submitBtn.disabled) {
                submitBtn.disabled = true;
            }
        } else if (submitBtn.disabled && socket.readyState === socket.OPEN) {
            submitBtn.disabled = false;
        }
    };

    submitBtn.disabled = true;

    submitBtn.onclick = function(e) {
        if (input.value !== "") {
            submitBtn.disabled = true;
            submitBtn.textContent = "发送中……";
            socket.send(input.value);
            input.value = "";
            submitBtn.textContent = "发送";
        }
    };

    socket.onerror = function(e) {
        console.log("err from connect " + e);
    };

    socket.onclose = function(e) {
        console.log("connection closed (" + e.code + ")");
    };

    // 监听快捷键
    document.onkeydown = function(keys) {
        // Ctrl + Enter
        if(keys.ctrlKey && keys.keyCode == 13){
            submitBtn.click();
        }
    };

    var previewTab = document.getElementById("preview");
    var editTab = document.getElementById("edit");

    editTab.disabled = true;

    previewTab.onclick = function () {
        previewTab.disabled = true;
        editTab.disabled = false;
    };

    editTab.onclick = function () {
        editTab.disabled = true;
        previewTab.disabled = false;
    };
}
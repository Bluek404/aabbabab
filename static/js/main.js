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

    submit.onclick = login;
};

function login() {
    var socket = new WebSocket(wsUrl);

    socket.onopen = function(e) {
        socket.send(document.getElementById("inputName").value);
    };

    socket.onerror = function(e) {
        console.log("err from connect " + e);
        setTimeout(login, 5000);
    };

    socket.onmessage = function (e) {
        if (JSON.parse(e.data)["error"] === true) {
            alert("name already exists");
        } else {
            init(socket);
            document.getElementById("init").style.display = "none";
        }
    };
}

function highlightAll(elment) {
    var codeBlocks = elment.getElementsByTagName("pre");
    for (var i = 0; i < codeBlocks.length; i++) {
        hljs.highlightBlock(codeBlocks[i]);
    }
}

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
        highlightAll(msg);

        msgBox.appendChild(msg);
        // 滚动到底部
        content.scrollTop = content.scrollHeight;
}

function init(socket) {
    marked.setOptions({
        sanitize: true
    });
    
    var submitBtn = document.getElementById("submit");
    var input = document.getElementById("input");
    var previewBox = document.getElementById("preview-box");
    var previewTab = document.getElementById("preview");
    var editTab = document.getElementById("edit");

    input.oninput = function(e) {
        if (input.value === "") {
            if (!submitBtn.disabled) {
                submitBtn.disabled = true;
            }
        } else if (submitBtn.disabled && socket.readyState === socket.OPEN) {
            submitBtn.disabled = false;
        }
    };

    if (input.value === "") {
        submitBtn.disabled = true;
    } else {
        submitBtn.disabled = false;
    }
    submitBtn.textContent = "服务器链接已建立";
    setTimeout(function() {
        submitBtn.textContent = "发送";
    }, 1500);

    submitBtn.onclick = function(e) {
        if (input.value !== "") {
            submitBtn.disabled = true;
            submitBtn.textContent = "发送中……";
            socket.send(input.value);
            input.value = "";
            previewBox.innerHTML = "";
            editTab.click();
            submitBtn.textContent = "发送";
        }
    };


    socket.onmessage = onMsg;
    socket.onclose = function(e) {
        submitBtn.disabled = true;
        submitBtn.textContent = "与服务器链接中断";
        setTimeout(login(), 5000);
        console.log("connection closed (" + e.code + ")");
    };

    // 监听快捷键
    document.onkeydown = function(keys) {
        // Ctrl + Enter
        if(keys.ctrlKey && keys.keyCode == 13){
            submitBtn.click();
        }
    };

    previewTab.onclick = function () {
        previewTab.disabled = true;
        editTab.disabled = false;
        input.style.display = "none";
        previewBox.style.display = "inline-block";
        previewBox.innerHTML = marked(input.value);

        highlightAll(previewBox);
    };

    editTab.onclick = function () {
        editTab.disabled = true;
        previewTab.disabled = false;
        previewBox.style.display = "none";
        input.style.display = "inline-block";
    };
}
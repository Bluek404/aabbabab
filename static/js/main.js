var wsUrl = document.location.protocol === "https:" ? "wss:" : "ws:" + "//" + document.location.host + "/ws";
var socket;

var topic = document.location.hash === "" ? "hall" : document.location.hash;
var topicList;

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

    submit.onclick = function() {
        submit.disabled = true;
        login();
    };
};

var lastMsgID = "";

function login() {
    socket = new WebSocket(wsUrl);
    
    var name = document.getElementById("inputName").value;
    var submit = document.getElementById("submitName");
    if (!/^\w{3,16}$/.test(name)) {
        alert("用户名必须为字母和数字，长度大于等于3小于等于16");
        submit.disabled = false;
        return;
    }

    socket.onopen = function(e) {
        socket.send(JSON.stringify({
            "name": name,
            "topic": topic,
            "lastMsgID": lastMsgID,
        }));
    };

    socket.onerror = function(e) {
        console.log("err from connect " + e);
        setTimeout(login, 5000);
    };

    socket.onmessage = function (e) {
        if (JSON.parse(e.data)["error"]) {
            alert("name already exists");
            socket.close();
            submit.disabled = false;
            document.getElementById("init").style.display = "block";
            document.onkeydown = function(keys) {
                if(keys.keyCode == 13){
                    submit.click();
                }
            };
        } else {
            init();
            document.getElementById("init").style.display = "none";
        }
    };
}

function highlightAll(elment) {
    var codeBlocks = elment.getElementsByTagName("pre");
    for (var i = 0; i < codeBlocks.length; i++) {
        var code = codeBlocks[i].getElementsByTagName("code")[0];
        var src = code.classList[0];
        if (!/^[\w-]+$/.test(src)) {
            // 非法语言名，删除
            code.classList.remove(src);
        }
        hljs.highlightBlock(code);
    }
}

function initStar(msgElem) {
    var star = msgElem.getElementsByClassName("star")[0];

    star.onclick = function() {
        addStar(msgElem, star);
    };

    msgElem.onmouseenter = function(){
        star.style.display = "initial";
    };
    msgElem.onmouseleave = function(){
        star.style.display = "none";
    };
}

function addStar(msgElem, star) {
    socket.send(JSON.stringify({
        "type": "star",
        "id": msgElem.id,
    }));
    star.classList.add("on");
    star.onclick = function() {
        remStar(msgElem, star);
    };

    msgElem.onmouseenter = null;
    msgElem.onmouseleave = null;
}

function remStar(msgElem, star) {
    socket.send(JSON.stringify({
        "type": "unstar",
        "id": msgElem.id,
    }));
    star.classList.remove("on");
    initStar(msgElem);
}

function genMsg(jsonData) {
    var data = JSON.parse(jsonData);
    var msg = document.createElement("div");

    msg.id = data["id"];
    msg.className = "msg";
    msg.innerHTML = '<img class="avatar" src="' + data["avatar"] + '"/>' +
        '<div class="msg-body"><p class="msg-header"><a class=msg-name>' +
        data["name"] + '</a>' + data["time"] + '<span class="star">★</span>' +
        '</p>' + marked(data["msg"]) + '</div>';

    // 高亮代码块
    highlightAll(msg);

    initStar(msg);

    lastMsgID = msg.id;
    return msg;
}

function onMsg(e) {
        console.log(e.data);
        var msgBox = document.getElementById("messages");
        var content = document.getElementById("content");

        msgBox.appendChild(genMsg(e.data));
        // 滚动到底部
        content.scrollTop = content.scrollHeight;
}

function init() {
    marked.setOptions({
        sanitize: true
    });

    var submitBtn = document.getElementById("submit");
    var input = document.getElementById("input");
    var previewBox = document.getElementById("preview-box");
    var previewTab = document.getElementById("preview");
    var editTab = document.getElementById("edit");

    input.focus();

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
            socket.send(JSON.stringify({
                "type": "msg",
                "topic": topic,
                "value": input.value,
            }));
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
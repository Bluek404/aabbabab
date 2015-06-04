var wsUrl = document.location.protocol === "https:" ? "wss:" : "ws:" + "//" + document.location.host + "/ws";
var socket;

var topic = document.location.hash === "" ? "hall" : document.location.hash.substr(1);

window.onload = function() {
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

    var newTopic = document.getElementById("new-topic");
    var newTopicBox = document.getElementById("new-topic-box");

    newTopic.onclick = function() {
        newTopicBox.parentElement.style.display = "block";
    };
    newTopicBox.onclick = function(e) {
        e.stopPropagation();
    };
    newTopicBox.parentElement.onclick = function() {
        newTopicBox.parentElement.style.display = "none";
    };
};

function initTitle(title, author, time) {
    var titleBox = document.getElementById("title");
    if (topic !== "hall") {
        titleBox.innerHTML = "<h1>" + title + "</h1><p>作者: " + author +
        "   时间: " + time + "</p>";
    } else {
        titleBox.innerHTML = "<h1>" + title + "</h1>";
    }
}

var lastMsgID = new Map();

function login() {
    socket = new WebSocket(wsUrl);

    var name = document.getElementById("inputName").value;
    var submit = document.getElementById("submitName");
    if (!/^\w{3,16}$/.test(name)) {
        alert("用户名必须为字母和数字，长度不能小于3或大于16");
        submit.disabled = false;
        return;
    }

    socket.onopen = function(e) {
        socket.send(JSON.stringify({
            "name": name,
            "topic": topic,
            "lastMsgID": lastMsgID[topic],
        }));
    };

    socket.onerror = function(e) {
        console.log("err from connect " + e);
        setTimeout(login, 5000);
    };

    socket.onmessage = function (e) {
        var data = JSON.parse(e.data);
        if (data["error"]) {
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
            initTitle(data["title"], data["author"], data["time"]);
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

function genMsg(data) {
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

    lastMsgID[topic] = msg.id;
    return msg;
}

function onMsg(e) {
    var data = JSON.parse(e.data);
    switch (data["type"]) {
    case "msg":
        var msgBox = document.getElementById("messages");
        var content = document.getElementById("content");

        msgBox.appendChild(genMsg(data));
        // 滚动到底部
        content.scrollTop = content.scrollHeight;
        break;
    case "getList":
        var topicList = document.getElementById("topic-list");
        var list = "";
        for (var i=0; i < data["topics"].length; i++) {
            var topicInfo = data["topics"][i];
            list += '<div onclick="openTopic(\'' + topicInfo["id"] + '\')">' +
                "</p>" + topicInfo["title"] + "</p><p>作者: " + topicInfo["author"] +
                " 时间: " + topicInfo["time"] + "</p></div>";
        }
        topicList.innerHTML = list;
        break;
    case "new":
        var nEditTab = document.getElementById("n-edit");
        var nEditBox = document.getElementById("n-edit-box");
        var nPreviewBox = document.getElementById("n-prev-box");
        var submitBtn = document.getElementById("n-subm");
        var title = document.getElementById("n-title");
        submitBtn.textContent= "发布";
        title.value = "";
        nEditBox.value = "";
        nPreviewBox.innerHTML = "";
        nEditTab.click();
        document.getElementById("new-topic-box").parentElement.click();
        openTopic(data["id"]);
        break;
    }
}

var msgHistory = new Map();

function openTopic(topicID) {
    if (topicID === topic) {
        return;
    }
    var msgBox = document.getElementById("messages");
    msgHistory[topic] = msgBox.innerHTML;
    msgBox.innerHTML = "";
    topic = topicID;
    document.location.hash = "#" + topic;
    socket.onclose = null;
    socket.close();
    login();
}

function showPreviewBox(editTab, previewTab, editBox, previewBox) {
    previewTab.disabled = true;
    editTab.disabled = false;
    editBox.style.display = "none";
    previewBox.style.display = "inline-block";
    previewBox.innerHTML = marked(editBox.value);

    highlightAll(previewBox);
}

function showEditBox(editTab, previewTab, editBox, previewBox) {
    editTab.disabled = true;
    previewTab.disabled = false;
    previewBox.style.display = "none";
    editBox.style.display = "inline-block";
}

function initNewTopicBox() {
    var nEditTab = document.getElementById("n-edit");
    var nEditBox = document.getElementById("n-edit-box");
    var nPreviewTab = document.getElementById("n-prev");
    var nPreviewBox = document.getElementById("n-prev-box");

    nPreviewTab.onclick = function () {
        showPreviewBox(nEditTab, nPreviewTab, nEditBox, nPreviewBox);
    };

    nEditTab.onclick = function () {
        showEditBox(nEditTab, nPreviewTab, nEditBox, nPreviewBox);
    };

    var submitBtn = document.getElementById("n-subm");
    var title = document.getElementById("n-title");

    title.oninput = function(e) {
        var len = title.value.length;
        if (len < 5) {
            submitBtn.textContent= "标题过短";
            submitBtn.disabled = true;
        } else if (len > 50) {
            submitBtn.textContent= "标题过长";
            submitBtn.disabled = true;
        } else {
            submitBtn.textContent= "发布";
            submitBtn.disabled = false;
        }
    };

    submitBtn.onclick = function(e) {
        submitBtn.disabled = true;
        submitBtn.textContent= "发布中……";
        socket.send(JSON.stringify({
            "type": "new",
            "title": title.value,
            "content": nEditBox.value,
        }));
    };
}

var topicListPage;

function getTopicList(page) {
    topicListPage = page;
    socket.send(JSON.stringify({
        "type": "getList",
        "page": String(page),
    }));
}

function init() {
    marked.setOptions({
        sanitize: true
    });

    var msgBox = document.getElementById("messages");
    if (msgBox.innerHTML == "" && msgHistory[topic] != null) {
        // 恢复消息记录
        msgBox.innerHTML = msgHistory[topic];
        var messages = document.getElementsByClassName("msg");
        for (var i=0, l=messages.length; i< l; i++) {
            var msg = messages[i];
            initStar(msg);
        }
    }

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
        showPreviewBox(editTab, previewTab, input, previewBox);
    };

    editTab.onclick = function () {
        showEditBox(editTab, previewTab, input, previewBox);
    };

    initNewTopicBox();
    getTopicList(1);
}
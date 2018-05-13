(function() {
    $(function () { new ChatRoom(); });

    var ChatRoom = function() {
        this.init();
        this.listenSend();
        this.listenRecv();
    }

    var cr = ChatRoom.prototype;

    cr.ws = null;
    cr.msgId = 1;

    cr.init = function() {
        var userId = prompt("请输入id", "");
        if (!userId && userId.length == 0) {
            alert("login fail!");
            return;
        } else {
            var that = this;
            this.ws = new WebSocket("ws://127.0.0.1:80/chatroom");   
            this.ws.onopen = function(event) {
                console.log('open ws');
                switch (that.ws.readyState) {
                    case WebSocket.CONNECTING:
                        console.log('WebSocket.CONNECTING');
                        break;
                    case WebSocket.OPEN:
                        console.log('WebSocket.OPEN');
                        break;
                    case WebSocket.CLOSING:
                        console.log('WebSocket.CLOSING');
                        break;
                    case WebSocket.CLOSED:
                        console.log('WebSocket.CLOSED');
                        break;
                    default:
                        console.log('WebSocket default');
                        break;
                }
                var user = {
                    id: parseInt(userId)
                }
                that.ws.send(JSON.stringify(user));
                $("#ui").show();
                $("#userid").html('ID: ' + userId);
            };
            this.ws.onmessage = function(event) {
                console.log('ws.onmessage: ', event.data);

                // $('#msg').text($('#msg').text() + event.data + '\n');
                var paras = $('#msg p');
                $('<p>' + event.data + '</p>').insertAfter(paras[paras.length - 1]);
            };    
            this.ws.onclose = function(event) {
                console.log('ws close');
            }
            // console.log(this.ws);

        }
    }

    cr.listenSend = function() {
        var that = this;
        $('#send').click(function(event) {
            var msgToSend = $("input[id='tosend']").val();
            var dstUserId = $("input[id='touserid']").val();
            console.log("msgToSend: ", msgToSend);
            console.log("dstUserId: ", dstUserId);

            if (msgToSend.length == 0 || dstUserId.length == 0) {
                alert("msg or toUser empty!");
            } else {
                $("input[id='tosend']").val('');
                $("input[id='touserid']").val('');
                
                var dstIdNum = parseInt(dstUserId);

                if (dstIdNum) {
                    var data = {
                        type: "chat",
                        dst_id: dstIdNum,
                        data: msgToSend,
                        msg_id: that.msgId++
                    }
        
                    that.ws.send(JSON.stringify(data));
                }
            }            
        });
    }

    cr.listenRecv = function() {
        var that = this;
    }
})();
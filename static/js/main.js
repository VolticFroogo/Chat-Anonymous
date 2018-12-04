$(document).ready(function(){
    var socket = new WebSocket("wss://" + document.location.host + "/ws");
    var users = {};

    window.onbeforeunload = function() {
        socket.onclose = function () {};
        socket.close();
    };

    socket.onopen = function (event) {
        setInterval(keepAlive, 15*1000);
    };

    socket.onmessage = function (event) {
        var message = JSON.parse(event.data);

        if (message.Success) {
            // We're in a room.
            $("form").hide();
            $("#chat").show();

            for (var i = 0; i < message.Users.length; i++) {
                users[message.Users[i].UUID] = message.Users[i].Username;
            }

            return;
        } else if (typeof event.data.Success !== "undefined") {
            // We failed to get into a room.
            return;
        }

        switch (message.Type) {
            case 1: // Message
                console.log(users[message.UserUUID] + ": " + message.Message);
                break;
            case 2: // User connected
                console.log(message.User.Username + " has connected.");
                users[message.User.UUID] = message.User.Username;
                break;
            case 3: // User disconnected
                console.log(users[message.UserUUID] + " has disconnected.");
                delete users[message.UserUUID];
                break;
        }
    }

    $("#connect").click(function(){
        if ($("#room").val() !== "") {
            if ($("#username").val() !== "") {
                if (grecaptcha.getResponse() !== "") {
                    socket.send(JSON.stringify({
                        Captcha: grecaptcha.getResponse(),
                        Room: $("#room").val(),
                        Username: $("#username").val()
                    }));
                } else {
                    console.log("Please check \"I'm not a robot\".");
                }
            }
            else {
                console.log("Please enter a username.");
            }
        }
        else {
            console.log("Please enter a room.");
        }
    });

    $("#send").click(function(){
        socket.send(JSON.stringify({
            Message: $("#message").val()
        }));

        $("#message").val("");
    });

    function keepAlive() {
        socket.send(JSON.stringify({
            Type: 5
        }));
    }

    function bytesToSize(bytes) {
        var sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        if (bytes == 0) return 'n/a';
        var i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
        if (i == 0) return bytes + ' ' + sizes[i];
        return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + sizes[i];
    };
});

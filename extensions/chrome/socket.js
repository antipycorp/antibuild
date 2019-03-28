console.info("Started Antibuild socket script.")

var socket = new WebSocket("ws://" + location.host + "/__/websocket");
socket.onopen = function () {
    socket.send("hey")
    socket.onmessage = function (event) {
        console.log(event.data);
    }
}
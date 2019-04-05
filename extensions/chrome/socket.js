// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

console.info('Started Antibuild socket script.');

var socket = new WebSocket('ws://' + location.host + '/__/websocket');
socket.onopen = function() {
  socket.send('hey');
  socket.onmessage = function(event) {
    console.log(event.data);
  };
};

// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

chrome.tabs.query({ active: true, currentWindow: true }, function(tabs) {
  chrome.tabs.executeScript(tabs[0].id, { file: 'socket.js' });
});

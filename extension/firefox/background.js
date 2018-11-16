connections = []
tabs = []
function handleRequest(event){
	data = JSON.parse(event.data);
	console.log(data);
	if (data.status == "update"){
		var loc = new URL(event.originalTarget.url)
		console.log(tabs[loc.host])
		browser.tabs.reload(
			tabs[loc.host],
			{
				bypassCache:true
			}
		)
	}
}

function startWS(host){
	var socket = new WebSocket("ws://" + host + "/__/websocket");
	socket.onopen = function () {
		socket.send("JSON")
	}
	socket.onmessage = function (event) {
		console.log(event);
		socket.onmessage = handleRequest;
	}
	socket.onclose = function(event){
		connections[host] = false
	}	
}

function addWS(url, tabid){
	var loc = new URL(url)

	if (connections[loc.host] == true){
		console.log("is already used")
		return
	}
	connections[loc.host] = true
	tabs[loc.host] = tabid
	startWS(loc.host)
	console.log("started using")
}

function listener(details){
	console.log(details)
	details.responseHeaders.forEach(function(header){
		if (header.name.toLowerCase() == "server" && header.value == "antibuildmodulehost") {
			addWS(details.url, details.tabId)
		}
	});
	return
}
browser.webRequest.onHeadersReceived.addListener(
	listener, 
	{
		urls:[
			"<all_urls>"
		],
		types:["main_frame"]
	},
	["responseHeaders"]
)
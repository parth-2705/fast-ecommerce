const webSocketUrl = `ws://${window.location.host}/payments/status`;
const webSocket = new WebSocket(webSocketUrl);

webSocket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    const status = data.status;
    const paymentId = data.payment_id;
    const statusElement = document.getElementById(`payment-status-${paymentId}`);
    statusElement.innerHTML = status;
}

webSocket.onclose = function (event) {
    console.error('Websocket closed unexpectedly');
};

webSocket.onopen = function (event) {
    console.log('Websocket opened');
};

// Path: static/js/payment-status.js
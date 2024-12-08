let connectInternal = undefined
const url = import.meta.env.MODE === "development" ? "ws://localhost:8080/ws" : import.meta.env.VITE_WEBSOCKET_URL

export function connect(onMessageCallback) {
  if (connectInternal) {
    clearTimeout(connectInternal)
  }
  const socket = new WebSocket(url);
  socket.addEventListener("message", onMessageCallback);
  socket.addEventListener("error", (error) => {
    console.error(error);
    socket.close()
    if (connectInternal) {
      clearTimeout(connectInternal)
    }
    connectInternal = setTimeout(() => connect(onMessageCallback), 10_000);
  });
  socket.addEventListener("close", (error) => {
    if (connectInternal) {
      clearTimeout(connectInternal)
    }
    connectInternal = setTimeout(() => connect(onMessageCallback), 10_000);
  });
}
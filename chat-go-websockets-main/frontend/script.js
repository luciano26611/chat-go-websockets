const messagesDiv = document.getElementById("messages");
const nicknameInput = document.getElementById("nicknameInput")
const messageInput = document.getElementById("messageInput");
const sendButton = document.getElementById("sendButton");

let nickname = "Usuario"

function addChatBubble(message, isOwn) {
    const messageBubble = document.createElement("div")
    const messageElement = document.createElement("div");
    messageBubble.className = `flex ${isOwn ? 'justify-end' : 'justify-start'} mb-4`;
    messageElement.className = isOwn
        ? 'bg-blue-500 text-white p-3 rounded-lg rounded-br-none shadow'
        : 'bg-gray-200 text-gray-800 p-3 rounded-lg rounded-bl-none shadow'

    messageElement.textContent = message;
    messageBubble.appendChild(messageElement)
    messagesDiv.appendChild(messageBubble);
    messagesDiv.scrollTop = messagesDiv.scrollHeight
}

nicknameInput.addEventListener("input", () => {
    nickname = nicknameInput.value || "Usuario"
})

const ws = new WebSocket("/ws");

ws.onopen = () => {
    console.log("WebSocket connection established");
    nicknameInput.value = ""
};

ws.onmessage = (event) => {
    addChatBubble(event.data, false);
};

function sendMessage() {
    const message = messageInput.value;
    if (message.trim() !== "") {
        ws.send(`${nickname}: ${message}`);
        addChatBubble(message, true)
        messageInput.value = "";
    }
}

sendButton.addEventListener("click", sendMessage);
messageInput.addEventListener("keypress", (e) => {
    if (e.key === "Enter") {
        sendMessage();
    }
})

const messagesDiv = document.getElementById("messages");
const nicknameInput = document.getElementById("nicknameInput")
const messageInput = document.getElementById("messageInput");
const sendButton = document.getElementById("sendButton");

let nickname = "Usuario"

function addChatBubble(message, isOwn, senderUsername) {
    const messageBubble = document.createElement("div")
    const messageElement = document.createElement("div");
    messageBubble.className = `flex ${isOwn ? 'justify-end' : 'justify-start'} mb-4`;
    messageElement.className = isOwn
        ? 'bg-blue-500 text-white p-3 rounded-lg rounded-br-none shadow'
        : 'bg-gray-200 text-gray-800 p-3 rounded-lg rounded-bl-none shadow'

    if (!isOwn && senderUsername) {
        // Mostrar el nombre del usuario si no es propio
        const usernameLabel = document.createElement("div");
        usernameLabel.className = "text-xs text-gray-500 mb-1";
        usernameLabel.textContent = senderUsername;
        messageElement.appendChild(usernameLabel);
    }

    const messageText = document.createElement("div");
    messageText.textContent = message;
    messageElement.appendChild(messageText);
    
    messageBubble.appendChild(messageElement)
    messagesDiv.appendChild(messageBubble);
    messagesDiv.scrollTop = messagesDiv.scrollHeight
}

function addSystemMessage(message) {
    const systemDiv = document.createElement("div");
    systemDiv.className = "text-center my-2";
    const systemElement = document.createElement("span");
    systemElement.className = "bg-gray-100 text-gray-600 text-xs px-3 py-1 rounded-full";
    systemElement.textContent = message;
    systemDiv.appendChild(systemElement);
    messagesDiv.appendChild(systemDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
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
    try {
        const data = JSON.parse(event.data);
        
        switch(data.type) {
            case 'message':
                addChatBubble(data.message, false, data.username);
                break;
            case 'user_join':
                addSystemMessage(`${data.username || 'Usuario'} se conectó`);
                break;
            case 'user_leave':
                addSystemMessage(`${data.username || 'Usuario'} se desconectó`);
                break;
            case 'system':
                addSystemMessage(data.message);
                break;
            default:
                console.log('Evento desconocido:', data);
        }
    } catch (error) {
        // Fallback para mensajes que no son JSON
        addChatBubble(event.data, false);
    }
};

function sendMessage() {
    const message = messageInput.value;
    if (message.trim() !== "") {
        // Enviar mensaje en formato JSON
        const messageData = {
            username: nickname,
            message: message,
            timestamp: new Date().toISOString()
        };
        ws.send(JSON.stringify(messageData));
        addChatBubble(message, true, nickname)
        messageInput.value = "";
    }
}

sendButton.addEventListener("click", sendMessage);
messageInput.addEventListener("keypress", (e) => {
    if (e.key === "Enter") {
        sendMessage();
    }
})

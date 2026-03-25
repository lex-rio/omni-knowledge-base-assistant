const messagesEl = document.getElementById("messages");
const form = document.getElementById("chat-form");
const input = document.getElementById("input");
const statusEl = document.getElementById("status");

let conversationId = null;

async function checkHealth() {
  try {
    const res = await fetch("/api/health");
    if (res.ok) {
      statusEl.textContent = "online";
      statusEl.style.color = "#22c55e";
    }
  } catch {
    statusEl.textContent = "offline";
    statusEl.style.color = "#ef4444";
  }
}

function addMessage(role, content) {
  const div = document.createElement("div");
  div.className = `message ${role}`;
  div.textContent = content;
  messagesEl.appendChild(div);
  messagesEl.scrollTop = messagesEl.scrollHeight;
  return div;
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  const text = input.value.trim();
  if (!text) return;

  addMessage("user", text);
  input.value = "";
  input.disabled = true;

  const assistantEl = addMessage("assistant", "");

  try {
    const res = await fetch("/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        message: text,
        conversation_id: conversationId,
      }),
    });

    if (!res.ok) {
      assistantEl.textContent = `Error: ${res.status} ${res.statusText}`;
      return;
    }

    const newConvId = res.headers.get("X-Conversation-ID");
    if (newConvId) conversationId = newConvId;

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop();

      for (const line of lines) {
        if (!line.startsWith("data: ")) continue;
        const data = line.slice(6);
        if (data === "[DONE]") continue;
        if (data.startsWith("[ERROR]")) {
          assistantEl.textContent += `\n${data}`;
          continue;
        }
        assistantEl.textContent += data;
        messagesEl.scrollTop = messagesEl.scrollHeight;
      }
    }
  } catch (err) {
    assistantEl.textContent = `Error: ${err.message}`;
  } finally {
    input.disabled = false;
    input.focus();
  }
});

checkHealth();

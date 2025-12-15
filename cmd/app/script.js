const statusContainer = document.getElementById('status-container');
const messageDiv = document.getElementById('message');
const saveButton = document.getElementById('save-button');

async function loadStatus() {
    try {
        const response = await fetch(`/shops/${shopId}/telegram/status`);
        if (!response.ok) throw new Error('Ошибка сервера');
        const data = await response.json();

        document.getElementById('botToken').value = '';
        document.getElementById('chatId').value = data.chatId ? '****' + data.chatId.slice(-4) : '';
        document.getElementById('enabled').checked = data.enabled === true;

        let html = `<strong>Статус:</strong> ${data.enabled ? 'Включено' : 'Выключено'}<br/>`;
        html += `<strong>Chat ID:</strong> ${data.chatId ? '****' + data.chatId.slice(-4) : 'не указан'}<br/>`;
        html += `<strong>Последнее отправленное:</strong> ${data.lastSentAt ? new Date(data.lastSentAt * 1000).toLocaleString('ru-RU') : 'никогда'}<br/>`;
        html += `<strong>За последние 7 дней:</strong> отправлено ${data.sentCountLast7Days || 0}, ошибок ${data.failedCountLast7Days || 0}`;

        statusContainer.innerHTML = html;
    } catch (err) {
        statusContainer.innerHTML = 'Ошибка загрузки статуса';
        console.error(err);
    }
}

saveButton.addEventListener('click', async () => {
    const botToken = document.getElementById('botToken').value.trim();
    const chatIdInput = document.getElementById('chatId').value.trim();
    const enabled = document.getElementById('enabled').checked;

    const payload = { enabled };
    if (botToken !== '') payload.botToken = botToken;
    
    // Обработка Chat ID: если пользователь ввел новый ID (не начинается с ****) или очистил поле
    if (chatIdInput !== '') {
        if (!chatIdInput.startsWith('****')) {
            payload.chatId = chatIdInput;
        }
        // Если начинается с ****, это скрытый существующий ID, не отправляем его
    } else {
        // Пользователь очистил поле - отправляем пустое значение для сброса
        payload.chatId = '';
    }

    try {
        const response = await fetch(`/shops/${shopId}/telegram/connect`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (response.ok) {
            messageDiv.innerHTML = '<span class="success">Настройки сохранены</span>';
            setTimeout(() => { messageDiv.innerHTML = ''; }, 4000);
            await loadStatus();
        } else {
            const errorText = await response.text();
            messageDiv.innerHTML = `<span class="error">Ошибка: ${errorText || 'серверная ошибка'}</span>`;
        }
    } catch (err) {
        messageDiv.innerHTML = '<span class="error">Ошибка сети</span>';
        console.error(err);
    }
});

// Загружаем статус при открытии страницы
loadStatus();
document.addEventListener('DOMContentLoaded', function () {
    const errorContainer = document.getElementById('error-container');
    const errorMessage = document.getElementById('error-message');
    const userCard = document.getElementById('user-card');
    const userName = document.getElementById('user-name');
    const userEmail = document.getElementById('user-email');
    const avatar = document.getElementById('avatar');
    const tokenContainer = document.getElementById('token-container');
    const refreshBtn = document.getElementById('refresh-btn');
    const logoutBtn = document.getElementById('logout-btn');

    // Material Buttons
    new mdc.ripple.MDCRipple(refreshBtn);
    new mdc.ripple.MDCRipple(logoutBtn);

    function parseJwt(token) {
        try {
            const parts = token.split('.');
            if (parts.length !== 3) return null;
            const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
            const binaryString = atob(base64);
            const bytes = new Uint8Array(binaryString.length);
            for (let i = 0; i < binaryString.length; i++) {
                bytes[i] = binaryString.charCodeAt(i);
            }
            return JSON.parse(new TextDecoder().decode(bytes));
        } catch (e) {
            console.error('Ошибка парсинга JWT:', e);
            return null;
        }
    }

    function isTokenValid(tokenData) {
        if (!tokenData.exp) return true;
        const currentTime = Math.floor(Date.now() / 1000);
        return tokenData.exp > currentTime;
    }

    function displayUserData(tokenData) {
        if (tokenData) {
            if (!isTokenValid(tokenData)) {
                showError('Срок действия токена истек');
                return;
            }

            if (tokenData.name) {
                userCard.style.display = 'block';
                userName.textContent = tokenData.name;
                userEmail.textContent = tokenData.email || 'Не указан';
                avatar.src = tokenData.picture || '/img/person.png';
            }

            tokenContainer.innerHTML = '';
            for (const [key, value] of Object.entries(tokenData)) {
                const tokenInfo = document.createElement('div');
                tokenInfo.className = 'token-info';

                const label = document.createElement('div');
                label.className = 'token-label';
                label.textContent = key;

                const tokenValue = document.createElement('div');
                tokenValue.className = 'token-value';
                // Преобразование iat и exp в читаемый формат даты
                if ((key === 'iat' || key === 'exp') && typeof value === 'number') {
                    tokenValue.textContent = value;
                    const date = new Date(value * 1000); // Unix timestamp in seconds
                    const readableTime = document.createElement('span');
                    readableTime.className = 'sub-value';
                    readableTime.textContent = ` (${date.toLocaleDateString()} ${date.toLocaleTimeString()})`;
                    tokenValue.appendChild(readableTime);
                } else if (typeof value === 'object') {
                    tokenValue.textContent = JSON.stringify(value, null, 2);
                } else {
                    tokenValue.textContent = value;
                }

                tokenInfo.appendChild(label);
                tokenInfo.appendChild(tokenValue);
                tokenContainer.appendChild(tokenInfo);
            }
        } else {
            showError('Не удалось извлечь данные из токена');
        }
    }

    async function loadUserData() {
        errorContainer.style.display = 'none';
        try {
            const response = await fetch(window.location.href, {
                method: 'GET',
                cache: 'no-cache'
            });

            const authHeader = response.headers.get('Authorization');
            if (!authHeader || !authHeader.startsWith('Bearer ')) {
                showError('Токен не найден или неверного формата');
                return;
            }

            const token = authHeader.substring(7);
            const tokenData = parseJwt(token);
            displayUserData(tokenData);
        } catch (e) {
            showError(`Ошибка при обработке токена: ${e.message}`);
        }
    }

    function showError(message) {
        errorMessage.textContent = message;
        errorContainer.style.display = 'block';
        userCard.style.display = 'none';
        tokenContainer.innerHTML = `<div class="no-token">${message}</div>`;
    }

    function logout() {
        fetch('/logout', {
            method: 'GET',
            credentials: 'same-origin'
        })
            .then(response => {
                if (response.redirected) {
                    window.location.href = response.url;
                } else {
                    window.location.reload();
                }
            })
            .catch(err => {
                console.error('Logout error:', err);
                window.location.reload();
            });
    }

    refreshBtn.addEventListener('click', loadUserData);
    logoutBtn.addEventListener('click', logout);

    loadUserData();
});
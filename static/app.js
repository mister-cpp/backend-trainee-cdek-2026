const API_URL = '/api';
const STATE = {
    currentWishlistId: null,
    currentToken: null
};

const ICONS = {
    delete: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>`
};

const api = {
    async request(endpoint, method = 'GET', body = null) {
        const headers = { 'Content-Type': 'application/json' };
        const jwt = localStorage.getItem('jwt');

        if (jwt)
            headers['Authorization'] = `Bearer ${jwt}`;

        const options = { method, headers };

        if (body)
            options.body = JSON.stringify(body);

        const res = await fetch(`${API_URL}${endpoint}`, options);

        if (!res.ok) {
            const err = await res.text();

            throw new Error(err || `Error ${res.status}`);
        }

        return res.json();
    },
    get: (endpoint) => api.request(endpoint),
    post: (endpoint, body) => api.request(endpoint, 'POST', body),
    put: (endpoint, body) => api.request(endpoint, 'PUT', body),
    delete: (endpoint) => api.request(endpoint, 'DELETE')
};

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');

    toast.className = `toast ${type}`;
    toast.textContent = message;

    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        toast.style.transition = 'all 0.3s ease';

        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function formatDate(dateStr) {
    return new Intl.DateTimeFormat('ru-RU', {
        day: '2-digit', month: '2-digit', year: 'numeric'
    }).format(new Date(dateStr));
}

function showView(viewId) {
    document.querySelectorAll('.view').forEach(el => {
        el.classList.add('hidden');
        el.classList.remove('active');
    });

    const target = document.getElementById(viewId);

    if (target) {
        target.classList.remove('hidden');
        target.classList.add('active');
    }
}

function checkAuth() {
    const token = localStorage.getItem('jwt');

    const navElements = {
        login: document.getElementById('btn-show-login'),
        register: document.getElementById('btn-show-register'),
        dashboard: document.getElementById('btn-show-dashboard'),
        logout: document.getElementById('btn-logout')
    };

    if (token) {
        navElements.login.classList.add('hidden');
        navElements.register.classList.add('hidden');
        navElements.dashboard.classList.remove('hidden');
        navElements.logout.classList.remove('hidden');

        showView('view-dashboard');

        loadWishlists();
    } else {
        navElements.login.classList.remove('hidden');
        navElements.register.classList.remove('hidden');
        navElements.dashboard.classList.add('hidden');
        navElements.logout.classList.add('hidden');

        showView('view-login');
    }
}

function clearUrlAndGoHome() {
    window.history.pushState({}, document.title, window.location.pathname);

    STATE.currentToken = null;

    checkAuth();
}

document.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');

    if (token) {
        STATE.currentToken = token;

        loadPublicWishlist(token);
    } else {
        checkAuth();
    }

    flatpickr("#wl-date", {
        locale: "ru",
        dateFormat: "d/m/Y",
        disableMobile: "true",
        minDate: "today"
    });

    setupEventListeners();
});

function setupEventListeners() {
    document.getElementById('btn-show-login').addEventListener('click', () => showView('view-login'));
    document.getElementById('btn-show-register').addEventListener('click', () => showView('view-register'));
    document.getElementById('btn-show-dashboard').addEventListener('click', clearUrlAndGoHome);
    document.getElementById('cdek-logo').addEventListener('click', clearUrlAndGoHome);
    document.getElementById('btn-back-dashboard').addEventListener('click', clearUrlAndGoHome);

    document.getElementById('btn-logout').addEventListener('click', () => {
        localStorage.removeItem('jwt');

        showToast('Вы вышли из системы');

        clearUrlAndGoHome();
    });

    document.getElementById('form-login').addEventListener('submit', handleLogin);
    document.getElementById('form-register').addEventListener('submit', handleRegister);
    document.getElementById('form-create-wishlist').addEventListener('submit', handleCreateWishlist);
    document.getElementById('form-create-item').addEventListener('submit', handleCreateItem);

    document.getElementById('btn-close-modal').addEventListener('click', closeModal);
    document.getElementById('form-edit-item').addEventListener('submit', handleEditItemSubmit);

    document.getElementById('wishlists-container').addEventListener('click', handleWishlistActions);
    document.getElementById('items-container').addEventListener('click', handleItemActions);
    document.getElementById('pub-items-container').addEventListener('click', handlePublicItemActions);
}

async function handleLogin(e) {
    e.preventDefault();

    const btn = e.target.querySelector('button');

    btn.disabled = true;

    try {
        const res = await api.post('/login', {
            email: document.getElementById('login-email').value,
            password: document.getElementById('login-password').value
        });

        localStorage.setItem('jwt', res.token);

        e.target.reset();

        showToast('Успешный вход!', 'success');

        checkAuth();
    } catch (err) {
        showToast('Неверный логин или пароль', 'error');
    } finally {
        btn.disabled = false;
    }
}

async function handleRegister(e) {
    e.preventDefault();

    const btn = e.target.querySelector('button');

    btn.disabled = true;

    const email = document.getElementById('reg-email').value;

    try {
        await api.post('/register', {
            email,
            password: document.getElementById('reg-password').value
        });

        showToast("Регистрация успешна!", 'success');

        e.target.reset();

        showView('view-login');

        document.getElementById('login-email').value = email;
    } catch (err) {
        showToast("Ошибка регистрации: " + err.message, 'error');
    } finally {
        btn.disabled = false;
    }
}

async function handleCreateWishlist(e) {
    e.preventDefault();

    const btn = e.target.querySelector('button');

    btn.disabled = true;

    try {
        const [day, month, year] = document.getElementById('wl-date').value.split('/');
        const eventDate = new Date(year, month - 1, day);

        if (isNaN(eventDate.getTime()))
            throw new Error("Неверная дата");

        await api.post('/wishlists', {
            title: document.getElementById('wl-title').value,
            description: document.getElementById('wl-desc').value,
            event_date: eventDate.toISOString()
        });

        e.target.reset();

        showToast('Вишлист создан!', 'success');

        loadWishlists();
    } catch (err) {
        showToast(err.message, 'error');
    } finally {
        btn.disabled = false;
    }
}

async function handleCreateItem(e) {
    e.preventDefault();

    const btn = e.target.querySelector('button');

    btn.disabled = true;

    try {
        const priority = parseInt(document.getElementById('item-priority').value) || 1;

        await api.post(`/wishlists/${STATE.currentWishlistId}/items`, {
            title: document.getElementById('item-title').value,
            description: document.getElementById('item-desc').value,
            url: document.getElementById('item-url').value,
            priority
        });

        e.target.reset();

        showToast('Подарок добавлен!', 'success');

        loadWishlistItems(STATE.currentToken);
    } catch (err) {
        showToast(err.message, 'error');
    } finally {
        btn.disabled = false;
    }
}

async function loadWishlists() {
    try {
        const lists = await api.get('/wishlists');
        const container = document.getElementById('wishlists-container');

        container.innerHTML = '';

        if (!lists?.length) {
            container.innerHTML = '<p class="text-center" style="grid-column: 1/-1;">У вас пока нет списков</p>';

            return;
        }

        lists.forEach(wl => {
            container.insertAdjacentHTML('beforeend', `
                <div class="card">
                    <h3>${escapeHTML(wl.title)}</h3>
                    <p>${escapeHTML(wl.description || 'Нет описания')}</p>
                    <div class="card-meta">
                        <span style="color: var(--text-muted)">Событие: ${formatDate(wl.event_date)}</span>
                    </div>
                    <div class="flex gap-2 mt-4">
                        <button class="btn-primary w-full" data-action="open" data-id="${wl.id}" data-title="${escapeHTML(wl.title)}" data-token="${wl.token}">Открыть</button>
                        <button class="btn-secondary icon-only" data-action="delete" data-id="${wl.id}">${ICONS.delete}</button>
                    </div>
                </div>
            `);
        });
    } catch (err) {
        showToast(err.message, 'error');
    }
}

async function loadWishlistItems(token) {
    try {
        const wl = await api.get(`/wishlists/public/${token}`);

        renderItems(wl.items || [], 'items-container', false);
    } catch (err) {
        showToast('Ошибка загрузки', 'error');
    }
}

async function loadPublicWishlist(token) {
    const tokenObj = localStorage.getItem('jwt');

    if (!tokenObj) {
        document.getElementById('btn-show-login').classList.remove('hidden');
        document.getElementById('btn-show-register').classList.remove('hidden');
    }

    showView('view-public');

    try {
        const wl = await api.get(`/wishlists/public/${token}`);

        document.getElementById('pub-wl-title').textContent = wl.title;
        document.getElementById('pub-wl-desc').textContent = wl.description || '';

        renderItems(wl.items || [], 'pub-items-container', true);
    } catch (err) {
        document.getElementById('pub-wl-title').textContent = "Вишлист не найден";

        showToast('Неверная ссылка', 'error');
    }
}

function renderItems(items, containerId, isPublic) {
    const container = document.getElementById(containerId);

   container.innerHTML = '';

    if (!items.length) {
        container.innerHTML = '<p class="text-center" style="grid-column: 1/-1; padding: 2rem 0; color: var(--text-muted);">Список пока пуст...</p>';

        return;
    }

    items.forEach(item => {
        let actionHTML = '';

        if (isPublic) {
            actionHTML = item.is_reserved
                ? `<button class="btn-secondary w-full mt-4" disabled>Забронировано</button>`
                : `<button class="btn-primary w-full mt-4" data-action="reserve" data-id="${item.id}">Забронировать</button>`;
        } else {
            actionHTML = `
                <div class="flex gap-2 mt-4">
                    <button class="btn-secondary w-full" data-action="edit" 
                        data-id="${item.id}" 
                        data-title="${escapeHTML(item.title)}" 
                        data-desc="${escapeHTML(item.description || '')}" 
                        data-url="${escapeHTML(item.url || '')}" 
                        data-priority="${item.priority}">Ред.</button>
                    <button class="btn-secondary icon-only" data-action="delete" data-id="${item.id}">${ICONS.delete}</button>
                </div>
            `;
        }

        container.insertAdjacentHTML('beforeend', `
            <div class="card ${item.is_reserved ? 'reserved' : ''}">
                <h3>${escapeHTML(item.title)}</h3>
                <p>${escapeHTML(item.description || '')}</p>
                <div class="card-meta">
                    <span>Приоритет: ${item.priority}</span>
                    ${item.url ? `<a href="${escapeHTML(item.url)}" target="_blank" class="accent-link">Купить →</a>` : ''}
                </div>
                ${actionHTML}
            </div>
        `);
    });
}

async function handleWishlistActions(e) {
    const btn = e.target.closest('button');

    if (!btn)
        return;

    const { action, id, title, token } = btn.dataset;

    if (action === 'open') {
        STATE.currentWishlistId = id;
        STATE.currentToken = token;

        document.getElementById('current-wl-title').textContent = title;

        const publicLink = `${window.location.origin}/?token=${token}`;
        const linkEl = document.getElementById('current-wl-link');

        linkEl.href = publicLink;
        linkEl.textContent = publicLink;

        showView('view-wishlist');

        loadWishlistItems(token);
    }

    if (action === 'delete') {
        if (!confirm('Удалить этот список?'))
            return;

        try {
            await api.delete(`/wishlists/${id}`);

            showToast('Список удален');

            loadWishlists();
        } catch (err) { showToast(err.message, 'error'); }
    }
}

async function handleItemActions(e) {
    const btn = e.target.closest('button');

    if (!btn)
        return;

    const { action, id, title, desc, url, priority } = btn.dataset;

    if (action === 'delete') {
        if (!confirm('Удалить этот подарок?'))
            return;

        try {
            await api.delete(`/items/${id}`);

            showToast('Подарок удален');

            loadWishlistItems(STATE.currentToken);
        } catch (err) { showToast(err.message, 'error'); }
    }

    if (action === 'edit') {
        openEditModal({ id, title, desc, url, priority });
    }
}

async function handlePublicItemActions(e) {
    const btn = e.target.closest('button');

    if (!btn || btn.dataset.action !== 'reserve')
        return;

    try {
        await api.post(`/wishlists/public/${STATE.currentToken}/items/${btn.dataset.id}/reserve`);

        showToast("Успешно забронировано!", 'success');

        loadPublicWishlist(STATE.currentToken);
    } catch (err) {
        showToast("Ошибка: кто-то уже забронировал это", 'error');
    }
}

function openEditModal(item) {
    document.getElementById('edit-item-id').value = item.id;
    document.getElementById('edit-item-title').value = item.title;
    document.getElementById('edit-item-desc').value = item.desc;
    document.getElementById('edit-item-url').value = item.url;
    document.getElementById('edit-item-priority').value = item.priority;

    document.getElementById('modal-edit').classList.remove('hidden');
}

function closeModal() {
    document.getElementById('modal-edit').classList.add('hidden');

    document.getElementById('form-edit-item').reset();
}

async function handleEditItemSubmit(e) {
    e.preventDefault();

    const id = document.getElementById('edit-item-id').value;

    try {
        await api.put(`/items/${id}`, {
            title: document.getElementById('edit-item-title').value,
            description: document.getElementById('edit-item-desc').value,
            url: document.getElementById('edit-item-url').value,
            priority: parseInt(document.getElementById('edit-item-priority').value)
        });

        showToast('Подарок обновлен', 'success');

        closeModal();

        loadWishlistItems(STATE.currentToken);
    } catch (err) {
        showToast(err.message, 'error');
    }
}

function escapeHTML(str) {
    if (!str)
        return '';

    return str.replace(/[&<>'"]/g,
        tag => ({
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            "'": '&#39;',
            '"': '&quot;'
        }[tag] || tag)
    );
}
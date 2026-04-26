const API = '/api/v1';
let me = null; // {id, name}
let partner = null; // {id, name}
let lastMsgID = '';
let pollController = null;
let pendingFile = null;

const $ = id => document.getElementById(id);

// --- Session restore ---
function enterApp(user) {
  me = user;
  $('login-screen').style.display = 'none';
  $('app').classList.add('visible');
  $('my-name').textContent = me.name;
  $('my-id').textContent = 'ID: ' + me.id.slice(0, 8) + '...';
  loadConversations();
}

(async function restoreSession() {
  const saved = localStorage.getItem('me');
  if (!saved) return;
  try {
    const user = JSON.parse(saved);
    // Verify user still exists on server
    const resp = await fetch(`${API}/users/${user.id}`);
    if (!resp.ok) { localStorage.removeItem('me'); return; }
    enterApp(await resp.json());
  } catch {
    localStorage.removeItem('me');
  }
})();

// --- Auth ---
$('login-btn').addEventListener('click', doLogin);
$('name-input').addEventListener('keydown', e => { if (e.key === 'Enter') doLogin(); });

async function doLogin() {
  const name = $('name-input').value.trim();
  if (!name) return;
  try {
    const resp = await fetch(`${API}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    });
    if (!resp.ok) throw new Error(await resp.text());
    const user = await resp.json();
    localStorage.setItem('me', JSON.stringify(user));
    enterApp(user);
  } catch (e) {
    alert('Ошибка входа: ' + e.message);
  }
}

$('logout-btn').addEventListener('click', () => {
  localStorage.removeItem('me');
  location.reload();
});

// --- Search ---
let searchTimeout = null;
$('search-input').addEventListener('input', e => {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => searchUsers(e.target.value.trim()), 300);
});

async function searchUsers(q) {
  if (!q) { $('search-results').innerHTML = ''; return; }
  const resp = await fetch(`${API}/users?q=${encodeURIComponent(q)}`);
  const users = await resp.json();
  renderSearchResults(users || []);
}

function renderSearchResults(users) {
  const container = $('search-results');
  container.innerHTML = '';
  const filtered = (users || []).filter(u => u.id !== me.id);
  if (!filtered.length) return;
  const label = document.createElement('div');
  label.className = 'conv-section-label';
  label.textContent = 'Результаты поиска';
  container.appendChild(label);
  filtered.forEach(u => {
    const div = document.createElement('div');
    div.className = 'user-item' + (partner && partner.id === u.id ? ' active' : '');
    div.innerHTML = `<div class="name">${esc(u.name)}</div>`;
    div.addEventListener('click', () => openChat(u));
    container.appendChild(div);
  });
}

// --- Conversations ---
async function loadConversations() {
  if (!me) return;
  try {
    const resp = await fetch(`${API}/conversations?user_id=${me.id}`);
    const convs = await resp.json();
    renderConversations(convs || []);
  } catch { /* silent */ }
}

function renderConversations(convs) {
  const container = $('conversations');
  container.innerHTML = '';
  if (!convs.length) return;
  const label = document.createElement('div');
  label.className = 'conv-section-label';
  label.textContent = 'Диалоги';
  container.appendChild(label);
  convs.forEach(conv => {
    const div = document.createElement('div');
    div.className = 'user-item' + (partner && partner.id === conv.partner_id ? ' active' : '');
    div.dataset.partnerId = conv.partner_id;
    const lastText = lastMsgPreview(conv.last_message);
    div.innerHTML = `<div class="name">${esc(conv.partner_name)}</div><div class="last-msg">${esc(lastText)}</div>`;
    div.addEventListener('click', () => openChat({ id: conv.partner_id, name: conv.partner_name }));
    container.appendChild(div);
  });
}

function lastMsgPreview(msg) {
  if (!msg) return '';
  if (msg.file_name) return (isImageFile(msg.file_name) ? '🖼 ' : '📎 ') + msg.file_name;
  return msg.text || '';
}

// --- Chat ---
function openChat(user) {
  partner = user;
  $('chat-header').innerHTML = esc(user.name) + ' <span id="chat-partner-id" style="color:#888;font-size:.85rem">ID: ' + user.id.slice(0, 8) + '...</span>';
  $('messages').innerHTML = '';
  lastMsgID = '';
  // mark active in both conversations list and search results
  document.querySelectorAll('.user-item').forEach(el => el.classList.remove('active'));
  document.querySelectorAll(`.user-item[data-partner-id="${user.id}"]`).forEach(el => el.classList.add('active'));
  loadMessages();
  startPolling();
}

async function loadMessages() {
  if (!partner) return;
  const resp = await fetch(`${API}/messages?user_a=${me.id}&user_b=${partner.id}&limit=100`);
  const msgs = await resp.json();
  const list = (msgs || []).slice().reverse();
  $('messages').innerHTML = '';
  list.forEach(appendMessage);
  if (list.length) lastMsgID = list[list.length - 1].id;
  scrollToBottom();
}

function appendMessage(msg) {
  const mine = msg.sender_id === me.id;
  const div = document.createElement('div');
  div.className = 'msg ' + (mine ? 'mine' : 'theirs');
  div.dataset.id = msg.id;

  let content = '';
  if (msg.text) content += `<div>${esc(msg.text)}</div>`;
  if (msg.file_url) {
    const url = msg.file_url;
    const name = msg.file_name || '';
    if (isImageFile(name)) {
      content += `<img src="${url}" data-name="${esc(name)}" onerror="this.style.display='none'" title="Нажмите для просмотра" />`;
    } else {
      content += `<a class="file-link" href="${url}" download="${esc(name)}" target="_blank">${fileIcon(name)} ${esc(name || 'Файл')}</a>`;
    }
  }
  content += `<div class="meta">${formatTime(msg.created_at)}${msg.edited ? ' <span class="edited-badge">(изменено)</span>' : ''}</div>`;
  if (mine) {
    content += `<div class="actions">
      <button onclick="editMsg('${msg.id}')">✏️</button>
      <button onclick="deleteMsg('${msg.id}')">🗑️</button>
    </div>`;
  }
  div.innerHTML = content;
  $('messages').appendChild(div);
}

function scrollToBottom() {
  const el = $('messages');
  el.scrollTop = el.scrollHeight;
}

// --- Send ---
$('send-btn').addEventListener('click', sendMessage);
$('msg-input').addEventListener('keydown', e => {
  if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); }
});

$('file-btn').addEventListener('click', () => $('file-input').click());
$('file-input').addEventListener('change', e => {
  const f = e.target.files[0];
  if (!f) return;
  pendingFile = f;
  $('file-name').textContent = f.name;
  $('file-preview').classList.add('visible');
});
$('remove-file').addEventListener('click', () => {
  pendingFile = null;
  $('file-input').value = '';
  $('file-preview').classList.remove('visible');
});

async function sendMessage() {
  if (!partner) return;
  const text = $('msg-input').value.trim();
  let fileID = '';
  let fileName = '';

  if (pendingFile) {
    const fd = new FormData();
    fd.append('file', pendingFile);
    const r = await fetch(`${API}/files`, { method: 'POST', body: fd });
    if (r.ok) {
      const data = await r.json();
      fileID = data.id;
      fileName = pendingFile.name;
    }
    pendingFile = null;
    $('file-input').value = '';
    $('file-preview').classList.remove('visible');
  }

  if (!text && !fileID) return;

  const resp = await fetch(`${API}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sender_id: me.id, receiver_id: partner.id, text, file_id: fileID, file_name: fileName })
  });
  if (resp.ok) {
    const msg = await resp.json();
    appendMessage(msg);
    lastMsgID = msg.id;
    scrollToBottom();
    $('msg-input').value = '';
    loadConversations();
  }
}

// --- Edit / Delete ---
window.editMsg = async function(id) {
  const el = document.querySelector(`.msg[data-id="${id}"]`);
  const current = el.querySelector('div')?.textContent || '';
  const newText = prompt('Изменить сообщение:', current);
  if (newText === null || newText === current) return;
  await fetch(`${API}/messages/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id: me.id, text: newText })
  });
  loadMessages();
};

window.deleteMsg = async function(id) {
  if (!confirm('Удалить сообщение?')) return;
  await fetch(`${API}/messages/${id}?user_id=${me.id}`, { method: 'DELETE' });
  document.querySelector(`.msg[data-id="${id}"]`)?.remove();
};

// --- Long polling ---
function startPolling() {
  if (pollController) pollController.abort();
  pollController = new AbortController();
  doPoll();
}

async function doPoll() {
  if (!partner || !me) return;
  try {
    const url = `${API}/poll?user_a=${me.id}&user_b=${partner.id}&after_id=${lastMsgID}`;
    const resp = await fetch(url, { signal: pollController.signal });
    if (resp.ok) {
      const msgs = await resp.json();
      if (msgs && msgs.length > 0) {
        msgs.forEach(m => {
          if (!document.querySelector(`.msg[data-id="${m.id}"]`)) {
            appendMessage(m);
            lastMsgID = m.id;
          }
        });
        scrollToBottom();
        loadConversations();
      }
    }
  } catch (e) {
    if (e.name === 'AbortError') return;
  }
  setTimeout(doPoll, 500);
}

// --- Lightbox ---
function openLightbox(src, name) {
  $('lightbox-img').src = src;
  $('lightbox-name').textContent = name || '';
  $('lightbox-download').href = src;
  $('lightbox-download').download = name || 'image';
  $('lightbox').classList.add('open');
}

function closeLightbox() {
  $('lightbox').classList.remove('open');
  $('lightbox-img').src = '';
}

$('lightbox-close').addEventListener('click', closeLightbox);
$('lightbox').addEventListener('click', e => { if (e.target === $('lightbox')) closeLightbox(); });
document.addEventListener('keydown', e => { if (e.key === 'Escape') closeLightbox(); });

// delegate clicks on message images
document.getElementById('messages').addEventListener('click', e => {
  const img = e.target.closest('.msg img');
  if (img) openLightbox(img.src, img.dataset.name || '');
});

// --- Helpers ---
function esc(str) {
  return (str || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function isImageFile(name) {
  return /\.(png|jpg|jpeg|gif|webp|bmp|svg)$/i.test(name || '');
}

function fileIcon(name) {
  const ext = (name || '').split('.').pop().toLowerCase();
  const icons = {
    pdf: '📄', doc: '📝', docx: '📝', xls: '📊', xlsx: '📊',
    ppt: '📋', pptx: '📋', zip: '🗜️', rar: '🗜️', gz: '🗜️',
    mp3: '🎵', mp4: '🎬', mov: '🎬', avi: '🎬',
    txt: '📃', csv: '📊', json: '📃', xml: '📃',
  };
  return icons[ext] || '📎';
}

function formatTime(iso) {
  const d = new Date(iso);
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

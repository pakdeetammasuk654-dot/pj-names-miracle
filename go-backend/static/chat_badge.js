(function() {
  var k = 'chuedee_support_chat_v1';
  function updateBadge(c) {
    document.querySelectorAll('.chat-unread-badge').forEach(function(e) { e.remove(); });
    if (c > 0) {
      var t = c > 99 ? '99+' : c;
      document.querySelectorAll('.floating-chat-cta, .nav-chat-cta').forEach(function(b) {
        var s = document.createElement('span');
        s.className = 'chat-unread-badge';
        s.textContent = t;
        b.appendChild(s);
      });
    }
  }
  async function check() {
    try {
      var s = JSON.parse(localStorage.getItem(k) || 'null');
      if (!s || !s.id || !s.token) { updateBadge(0); return; }
      var r = await fetch('/api/v1/chat/conversations/' + s.id + '/unread', {
        headers: { 'X-Chat-Token': s.token }
      });
      if (r.status === 401) { updateBadge(0); return; }
      if (r.ok) {
        var d = await r.json();
        updateBadge(d.unread_count || 0);
      }
    } catch (_) {}
  }
  check();
  setInterval(check, 8000);
})();

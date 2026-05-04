/* Tiny vanilla helpers used across the panel. */
(function (global) {
    'use strict';

    const NPS = global.NPS = global.NPS || {};
    NPS.baseUrl = global.__nps_base_url || '';

    function url(path) { return NPS.baseUrl + path; }
    NPS.url = url;

    /* ---------- toast ---------- */
    let toastWrap;
    function ensureToastWrap() {
        if (toastWrap) return toastWrap;
        toastWrap = document.createElement('div');
        toastWrap.className = 'toast-wrap';
        document.body.appendChild(toastWrap);
        return toastWrap;
    }
    NPS.toast = function (msg, type) {
        const wrap = ensureToastWrap();
        const el = document.createElement('div');
        el.className = 'toast' + (type ? ' is-' + type : '');
        el.textContent = msg;
        wrap.appendChild(el);
        setTimeout(function () { el.remove(); }, 4000);
    };

    /* ---------- fetch wrapper ---------- */
    NPS.post = function (path, data) {
        const body = new URLSearchParams();
        Object.keys(data || {}).forEach(function (k) {
            const v = data[k];
            if (v === undefined || v === null) return;
            body.append(k, String(v));
        });
        return fetch(url(path), {
            method: 'POST',
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: body
        }).then(function (r) {
            const ct = r.headers.get('content-type') || '';
            if (ct.indexOf('application/json') >= 0) return r.json();
            return r.text();
        });
    };

    NPS.get = function (path) {
        return fetch(url(path), { credentials: 'same-origin' }).then(function (r) {
            return r.json();
        });
    };

    /* ---------- copy ---------- */
    NPS.copy = function (text) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function () {
                NPS.toast('Copied to clipboard', 'success');
            }, function () {
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    };
    function fallbackCopy(text) {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.left = '-1000px';
        document.body.appendChild(ta);
        ta.select();
        try { document.execCommand('copy'); NPS.toast('Copied to clipboard', 'success'); }
        catch (e) { NPS.toast('Copy failed', 'error'); }
        ta.remove();
    }

    /* ---------- formatters ---------- */
    NPS.fmt = {
        bytes: function (n) {
            if (!Number.isFinite(n)) return '–';
            const units = ['B', 'KB', 'MB', 'GB', 'TB'];
            let i = 0;
            while (n >= 1024 && i < units.length - 1) { n /= 1024; i++; }
            return n.toFixed(1) + ' ' + units[i];
        },
        uptime: function (sec) {
            sec = Math.floor(sec || 0);
            const d = Math.floor(sec / 86400); sec -= d * 86400;
            const h = Math.floor(sec / 3600); sec -= h * 3600;
            const m = Math.floor(sec / 60); sec -= m * 60;
            const out = [];
            if (d) out.push(d + 'd');
            if (h || d) out.push(h + 'h');
            out.push(m + 'm');
            out.push(sec + 's');
            return out.join(' ');
        },
        truncate: function (s, n) {
            if (!s) return '';
            return s.length > n ? s.slice(0, n) + '…' : s;
        }
    };

    /* ---------- dom helpers ---------- */
    NPS.h = function (tag, attrs, children) {
        const el = document.createElement(tag);
        Object.keys(attrs || {}).forEach(function (k) {
            if (k === 'class') el.className = attrs[k];
            else if (k === 'html') el.innerHTML = attrs[k];
            else if (k === 'text') el.textContent = attrs[k];
            else if (k.indexOf('on') === 0) el.addEventListener(k.slice(2), attrs[k]);
            else if (attrs[k] === true) el.setAttribute(k, '');
            else if (attrs[k] !== false && attrs[k] != null) el.setAttribute(k, attrs[k]);
        });
        (children || []).forEach(function (c) {
            if (c == null) return;
            if (typeof c === 'string') el.appendChild(document.createTextNode(c));
            else el.appendChild(c);
        });
        return el;
    };

    /* ---------- escape ---------- */
    NPS.escape = function (s) {
        if (s == null) return '';
        return String(s)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    };
})(window);

/* ============================================================
 * ChinaSCLM DAV · WebDAV web client ("Lumen" design)
 * ============================================================ */

const state = {
  user: null,
  route: 'login',
  dir: '/',
  search: '',
};

const $ = (sel, root) => (root || document).querySelector(sel);
const $$ = (sel, root) => Array.from((root || document).querySelectorAll(sel));

const BRAND_NAME = 'ChinaSCLM DAV';
const BRAND_TAG  = 'WebDAV · 网盘';

/* ---------------- helpers ---------------- */
function esc(s) { return String(s == null ? '' : s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#39;'); }

function fmtSize(n) {
  if (n == null) return '';
  const u = ['B','KB','MB','GB','TB']; let i = 0, v = Number(n);
  while (v >= 1024 && i < u.length-1) { v /= 1024; i++; }
  return (i === 0 ? v : v.toFixed(1)) + ' ' + u[i];
}

function fmtTime(sec) {
  if (!sec) return '';
  const d = new Date(Number(sec) * 1000);
  const pad = x => String(x).padStart(2,'0');
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function fmtDateTime(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d)) return String(iso).replace(/[-T]/g,'/').replace(/\.\d+Z$/,'').replace('Z','');
  const pad = x => String(x).padStart(2,'0');
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function initials(name) {
  const s = (name||'U').trim();
  return s.slice(0,2).toUpperCase();
}

/* ---------------- file-type icons (iconfont.cn) ---------------- */
const Fi = (d) => `<svg style="width:80%;height:80%" viewBox="0 0 1024 1024" fill="currentColor"><path d="${d}"/></svg>`;

const FileIco = {
  word: Fi('M854.6 288.7c6 6 9.4 14.1 9.4 22.6V928c0 17.7-14.3 32-32 32H192c-17.7 0-32-14.3-32-32V96c0-17.7 14.3-32 32-32h424.7c8.5 0 16.7 3.4 22.7 9.4l215.2 215.3zM790.2 326L602 137.8V326h188.2zM512 566.095l52.814 197.012a12 12 0 0 0 11.59 8.893h31.78a12 12 0 0 0 11.587-8.878l74.375-276a12 12 0 0 0 0.413-3.122c0-6.627-5.373-12-12-12h-35.576a12 12 0 0 0-11.695 9.31l-45.79 199.105-49.76-199.321A12 12 0 0 0 528.097 472h-32.192a12 12 0 0 0-11.643 9.094l-49.66 198.927-46.096-198.732a12 12 0 0 0-11.69-9.289h-35.381a12 12 0 0 0-3.115 0.411c-6.4 1.72-10.194 8.303-8.474 14.703l74.173 276A12 12 0 0 0 415.606 772h31.99a12 12 0 0 0 11.59-8.893L512 566.095z'),
  pdf: Fi('M854.6 288.7c6 6 9.4 14.1 9.4 22.6V928c0 17.7-14.3 32-32 32H192c-17.7 0-32-14.3-32-32V96c0-17.7 14.3-32 32-32h424.7c8.5 0 16.7 3.4 22.7 9.4l215.2 215.3zM790.2 326L602 137.8V326h188.2zM633.217 637.256c-15.174-0.489-31.314 0.67-49.65 2.964-24.298-14.987-40.654-35.582-52.274-65.827 0.28-1.152 0.86-3.538 1.063-4.38 0.474-1.958 0.867-3.594 1.243-5.185 4.293-18.13 6.615-31.358 7.3-44.695 0.518-10.074-0.04-19.368-1.827-27.976-3.298-18.584-16.454-29.453-33.021-30.126-15.446-0.627-29.649 7.993-33.281 21.373-5.913 21.612-2.45 50.07 10.08 98.582-15.964 38.056-37.052 82.661-51.203 107.539-18.885 9.74-33.604 18.605-45.953 28.427-16.303 12.966-26.48 26.29-29.286 40.306-1.355 6.48 0.692 14.966 5.36 21.912 5.296 7.879 13.282 12.991 22.855 13.735 24.152 1.877 53.83-23.024 86.59-79.258 3.295-1.09 6.78-2.257 11.026-3.69 2.323-0.783 10.464-3.538 11.91-4.026 7.521-2.54 12.98-4.36 18.376-6.116 23.396-7.612 41.096-12.429 57.21-15.163 27.973 14.973 60.316 24.796 82.098 24.796 17.979 0 30.126-9.319 34.515-23.985 3.857-12.886 0.794-27.824-7.473-36.084-8.56-8.41-24.3-12.434-45.658-13.123z m-247.985 128.42v-0.36l0.126-0.338c1.275-3.421 3.157-7.008 5.6-10.758 4.284-6.576 10.173-13.5 17.472-20.865 3.92-3.955 8.002-7.8 12.79-12.12 1.073-0.969 7.91-7.059 9.189-8.25l11.176-10.407-8.12 12.934c-12.326 19.638-23.46 33.78-33.013 43.004-3.507 3.387-6.6 5.9-9.091 7.505-1.027 0.662-1.916 1.144-2.613 1.424-0.409 0.163-0.771 0.268-1.13 0.302a2.202 2.202 0 0 1-1.117-0.16 2.068 2.068 0 0 1-1.269-1.911z m125.934-218.269l-2.26 4.007-1.39-4.385c-3.114-9.829-5.387-24.641-6.016-37.997-0.716-15.197 0.49-24.323 5.286-24.323 6.74 0 9.831 10.808 10.076 27.053 0.216 14.28-2.03 29.142-5.696 35.645z m-5.81 58.464l1.534-4.05 2.088 3.795c11.69 21.245 26.858 38.967 43.538 51.315l3.595 2.662-4.38 0.904c-16.328 3.372-31.544 8.457-52.34 16.842 2.174-0.876-21.623 8.863-27.638 11.169l-5.252 2.013 2.802-4.877c12.35-21.496 23.758-47.326 36.052-79.773z m157.626 76.261c-7.864 3.104-24.777 0.329-54.569-12.387l-7.561-3.227 8.199-0.607c23.295-1.724 39.807-0.44 49.422 3.08 4.09 1.498 6.824 3.388 8.037 5.553 1.31 2.336 0.71 4.81-1.362 6.31-0.448 0.427-1.155 0.88-2.166 1.278z'),
  excel: Fi('M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7zM790.2 326H602V137.8L790.2 326z m1.8 562H232V136h302v216c0 23.2 18.8 42 42 42h216v494zM514.1 580.1l-61.8-102.4c-2.2-3.6-6.1-5.8-10.3-5.8h-38.4c-2.3 0-4.5 0.6-6.4 1.9-5.6 3.5-7.3 10.9-3.7 16.6l82.3 130.4-83.4 132.8c-1.2 1.9-1.8 4.1-1.8 6.4 0 6.6 5.4 12 12 12h34.5c4.2 0 8-2.2 10.2-5.7L510 664.8l62.3 101.4c2.2 3.6 6.1 5.7 10.2 5.7H620c2.3 0 4.5-0.7 6.5-1.9 5.6-3.6 7.2-11 3.6-16.6l-84-130.4 85.3-132.5c1.2-1.9 1.9-4.2 1.9-6.5 0-6.6-5.4-12-12-12h-35.7c-4.2 0-8.1 2.2-10.3 5.8l-61.2 102.3z'),
  ppt: Fi('M854.6 288.7c6 6 9.4 14.1 9.4 22.6V928c0 17.7-14.3 32-32 32H192c-17.7 0-32-14.3-32-32V96c0-17.7 14.3-32 32-32h424.7c8.5 0 16.7 3.4 22.7 9.4l215.2 215.3zM790.2 326L602 137.8V326h188.2zM468.526 760v-91.537h59.277c60.57 0 100.197-39.655 100.197-98.125C628 512.116 588.424 472 528.016 472H424c-6.627 0-12 5.373-12 12v276c0 6.627 5.373 12 12 12h32.526c6.628 0 12-5.373 12-12z m0-139.326h34.907c47.815 0 67.186-12.937 67.186-50.336 0-32.045-18.117-50.121-49.87-50.121h-52.223v100.457z'),
  audio: Fi('M867.882 195.882l-167.764-167.764A96 96 0 0 0 632.236 0H224C170.98 0 128 42.98 128 96v832c0 53.02 42.98 96 96 96h576c53.02 0 96-42.98 96-96V263.764a96 96 0 0 0-28.118-67.882zM792.236 256H640V103.764L792.236 256zM224 928V96h320v208c0 26.51 21.49 48 48 48h208v576H224z m288-152.048c0 21.382-25.852 32.09-40.97 16.97L400 720.972h-56c-13.254 0-24-10.746-24-24v-112c0-13.254 10.746-24 24-24h56l71.03-73.894c15.12-15.12 40.97-4.412 40.97 16.97v271.904z m82.402-94.26c18.102-18.594 18.12-48.266 0.002-66.878-44.298-45.504 24.47-112.492 68.79-66.962 54.396 55.88 54.424 144.888 0.002 200.802-43.586 44.772-113.894-20.63-68.794-66.962z'),
  image: Fi('M553.1 509.1l-77.8 99.2-41.1-52.4c-3.2-4.1-9.4-4.1-12.6 0l-99.8 127.2c-4.1 5.2-0.4 12.9 6.3 12.9H696c6.7 0 10.4-7.7 6.3-12.9l-136.5-174c-3.3-4.1-9.5-4.1-12.7 0zM400 442m-40 0a40 40 0 1 0 80 0 40 40 0 1 0-80 0ZM854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7zM790.2 326H602V137.8L790.2 326z m1.8 562H232V136h302v216c0 23.2 18.8 42 42 42h216v494z'),
  video: Fi('M867.882 195.882l-167.764-167.764A96 96 0 0 0 632.236 0H224C170.98 0 128 42.98 128 96v832c0 53.02 42.98 96 96 96h576c53.02 0 96-42.98 96-96V263.764a96 96 0 0 0-28.118-67.882zM792.236 256H640V103.764L792.236 256zM224 928V96h320v208c0 26.51 21.49 48 48 48h208v576H224z m457.374-422.606L576 610.748V536c0-22.092-17.908-40-40-40H328c-22.092 0-40 17.908-40 40v208c0 22.092 17.908 40 40 40h208c22.092 0 40-17.908 40-40v-74.748l105.374 105.348C701.408 794.636 736 780.56 736 751.972V528.022c0-28.622-34.618-42.638-54.626-22.628z'),
  archive: Fi('M384.549893 320.137473v63.987503h63.987502v-63.987503z m127.975004-191.962507h-63.987502v63.987502h63.987502z m-127.975004 63.987502v63.987503h63.987502V192.162468z m127.975004 63.987503h-63.987502v63.987502h63.987502z m355.130639-60.188245L699.888303 28.194493C681.891818 10.198008 657.496583 0 632.101543 0H223.981254C170.991603 0.199961 128 43.191564 128 96.181215v831.837531c0 52.98965 42.991603 95.981254 95.981254 95.981254h575.887522c52.98965 0 95.981254-42.991603 95.981253-95.981254V263.948448c0-25.39504-10.198008-49.990236-28.194493-67.986722zM639.90002 103.979691l152.170279 152.17028H639.90002zM799.868776 928.018746H223.981254V96.181215h159.368873v31.993751h63.987502V96.181215H543.918766v207.959383c0 26.594806 21.395821 47.990627 47.990627 47.990626h207.959383zM516.324155 531.496192c-2.19957-11.197813-11.997657-19.396212-23.595391-19.396212h-44.191369v-63.987502h-63.987502v63.987502l-39.392307 194.162078C331.960164 771.249365 381.550478 832.037493 447.937512 832.037493c66.187073 0 115.777387-60.388205 102.979887-125.175552z m-67.78676 248.751416c-35.793009 0-64.787346-24.195274-64.787346-53.989455s28.994337-53.989455 64.787346-53.989455 64.787346 24.195274 64.787346 53.989455-28.994337 53.989455-64.787346 53.989455z m63.987502-396.122632h-63.987502v63.987502h63.987502z'),
  text: Fi('M854.6 288.7c6 6 9.4 14.1 9.4 22.6V928c0 17.7-14.3 32-32 32H192c-17.7 0-32-14.3-32-32V96c0-17.7 14.3-32 32-32h424.7c8.5 0 16.7 3.4 22.7 9.4l215.2 215.3zM790.2 326L602 137.8V326h188.2zM320 482a8 8 0 0 0-8 8v48a8 8 0 0 0 8 8h384a8 8 0 0 0 8-8v-48a8 8 0 0 0-8-8H320z m0 136a8 8 0 0 0-8 8v48a8 8 0 0 0 8 8h184a8 8 0 0 0 8-8v-48a8 8 0 0 0-8-8H320z'),
  generic: Fi('M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7zM790.2 326H602V137.8L790.2 326z m1.8 562H232V136h302v216c0 23.2 18.8 42 42 42h216v494z'),
};

const CAT_META = {
  documents:{ c:'#2a6df4', e: FileIco.word },
  images:   { c:'#22c55e', e: FileIco.image },
  video:    { c:'#ff4d6d', e: FileIco.video },
  audio:    { c:'#ff9500', e: FileIco.audio },
  archives: { c:'#9b5cff', e: FileIco.archive },
  other:    { c:'#94a3b8', e: FileIco.generic },
};

/* ---------------- API ---------------- */
async function api(path, opts) {
  opts = Object.assign({ credentials:'same-origin' }, opts);
  if (opts.body && typeof opts.body !== 'string' && !(opts.body instanceof FormData)) {
    opts.headers = Object.assign({ 'Content-Type':'application/json' }, opts.headers||{});
    opts.body = JSON.stringify(opts.body);
  }
  const res = await fetch('/api' + path, opts);
  let data = null;
  const ct = res.headers.get('content-type')||'';
  if (ct.includes('application/json')) { try { data = await res.json(); } catch(e){} }
  if (!res.ok) {
    const err = new Error((data && data.error) || ('请求失败 ('+res.status+')'));
    err.status = res.status; throw err;
  }
  return data;
}

/* ---------------- toast & confirm ---------------- */
function copyText(text, okMsg, errMsg) {
  const done = () => toast(okMsg || '已复制', 'ok');
  const fail = () => toast(errMsg || '复制失败', 'error');
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text).then(done).catch(() => fallbackCopy(text) ? done() : fail());
  }
  return fallbackCopy(text) ? done() : fail();
}
function fallbackCopy(text) {
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.style.position = 'fixed'; ta.style.opacity = '0';
  document.body.appendChild(ta);
  ta.select();
  let ok = false;
  try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
  ta.remove();
  return ok;
}

function toast(msg, type) {
  let box = document.getElementById('toast-wrap');
  if (!box) {
    box = document.createElement('div');
    box.id = 'toast-wrap';
    box.className = 'toast-wrap';
    document.body.appendChild(box);
  }
  const t = document.createElement('div');
  const color = type === 'error' ? 'linear-gradient(135deg,#ff4d5f,#ff7a59)' : type === 'ok' ? 'linear-gradient(135deg,#14c26b,#0e9f6e)' : 'linear-gradient(135deg,#2a2f45,#101226)';
  t.style.background = color;
  t.style.color = '#fff'; t.style.fontSize = '.82rem'; t.style.fontWeight = '600';
  t.style.padding = '.6rem 1rem'; t.style.borderRadius = '.85rem';
  t.style.boxShadow = '0 10px 30px -8px rgba(0,0,0,.35)';
  t.className = 'fade-in';
  t.textContent = msg;
  box.appendChild(t);
  setTimeout(() => { t.style.opacity = '0'; t.style.transition = 'opacity .3s'; }, 2200);
  setTimeout(() => t.remove(), 2600);
}

function iConfirm(title, text, okLabel, danger) {
  return new Promise(resolve => {
    const wrap = document.createElement('div');
    wrap.className = 'fixed inset-0 z-[150] flex items-end md:items-center justify-center p-4';
    wrap.innerHTML = `
      <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" data-close></div>
      <div class="card relative w-full max-w-sm p-6 fade-in">
        <h3 class="text-lg font-semibold mb-1.5" style="letter-spacing:-.01em">${esc(title)}</h3>
        <p class="text-sm mb-5" style="color:var(--text-2)">${esc(text)}</p>
        <div class="flex gap-3">
          <button class="btn-ghost flex-1 hairline" data-cancel>取消</button>
          <button class="flex-1 inline-flex items-center justify-center py-3 px-4 rounded-xl font-semibold text-white ${danger ? 'btn-danger !py-2.5' : ''}" style="${danger ? 'background:var(--danger)' : 'background:linear-gradient(135deg,var(--brand),var(--brand-2))'}" data-ok>${esc(okLabel)}</button>
        </div>
      </div>`;
    document.body.appendChild(wrap);
    const close = val => { wrap.remove(); resolve(val); };
    wrap.querySelector('[data-close]').onclick = () => close(false);
    wrap.querySelector('[data-cancel]').onclick = () => close(false);
    wrap.querySelector('[data-ok]').onclick = () => close(true);
  });
}

let sheetEl = null;
function closeAll() {
  if (sheetEl) { sheetEl.remove(); sheetEl = null; }
}

/* ---------------- icons ---------------- */
const I = (d,cls='') => `<svg class="${cls}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${d}</svg>`;

const Icon = {
  home: I('<path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V21h14V9.5"/>'),
  files: I('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/>'),
  share: I('<path d="M12 3v12"/><path d="m8 7 4-4 4 4"/><path d="M6 12H4a2 2 0 0 0-2 2v5a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-5a2 2 0 0 0-2-2h-2"/>'),
  trash: I('<path d="M4 7h16"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12"/><path d="M9 7V5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"/>'),
  key: I('<circle cx="8" cy="15" r="4"/><path d="M11 12 20 3m0 0h-4m4 0v4"/>'),
  gear: I('<circle cx="12" cy="12" r="3.2"/><path d="M12 2v2m0 16v2M4.9 4.9l1.4 1.4m11.4 11.4 1.4 1.4M2 12h2m16 0h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>'),
  logout: I('<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5"/><path d="M21 12H9"/>'),
  folder: I('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/>'),
  file: I('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/>'),
  upload: I('<path d="M12 16V3"/><path d="m7 8 5-5 5 5"/><path d="M4 21h16"/>'),
  folderPlus: I('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/><path d="M12 11v5m-2.5-2.5h5"/>'),
  search: I('<circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/>'),
  download: I('<path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M4 21h16"/>'),
  clock: I('<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>'),
  copy: I('<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/>'),
  close: I('<path d="M6 6l12 12M18 6 6 18"/>'),
  more: I('<circle cx="12" cy="5" r="1.5" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none"/><circle cx="12" cy="19" r="1.5" fill="currentColor" stroke="none"/>'),
  back: I('<path d="m15 18-6-6 6-6"/>'),
  restore: I('<path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 4v4h4"/>'),
  check: I('<path d="M20 6 9 17l-5-5"/>'),
  refresh: I('<path d="M20 12a8 8 0 1 1-2.3-5.6"/><path d="M20 4v4h-4"/>'),
  chevron: I('<path d="m9 18 6-6-6-6"/>'),
  user: I('<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 3.6-6 8-6s8 2 8 6"/>'),
  theme: I('<path d="M21 12.8A9 9 0 1 1 11.2 3 7 7 0 0 0 21 12.8z"/>'),
  eye: I('<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/>'),
  eyeOff: I('<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><path d="m3 3 18 18"/><path d="M10.5 10.5a2 2 0 0 0 2.8 2.8"/>'),
  link: I('<path d="M10 13a5 5 0 0 0 7.5.5l2-2a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7.5-.5l-2 2a5 5 0 0 0 7 7l1-1"/>'),
  zap: I('<path d="M13 2 3 14h7l-1 8 10-12h-7z"/>'),
  github: I('<path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.7c-2.78.6-3.37-1.34-3.37-1.34-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.64 0 0 .84-.27 2.75 1.02a9.58 9.58 0 0 1 5 0c1.91-1.3 2.75-1.02 2.75-1.02.55 1.37.2 2.39.1 2.64.64.7 1.03 1.59 1.03 2.68 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.85v2.74c0 .27.18.58.69.48A10 10 0 0 0 12 2z"/>'),
};

const LOGO = `
  <svg class="w-full h-full" viewBox="0 0 48 48" fill="none">
    <rect width="48" height="48" rx="14" fill="url(#lg1)"/>
    <path d="M30 11 18 34h6l-5 6 18-24h-6l5-5z" fill="#fff"/>
    <defs><linearGradient id="lg1" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#2a6df4"/><stop offset="1" stop-color="#7a5cff"/>
    </linearGradient></defs>
  </svg>`;

function fileGlyph(name, isDir, size) {
  if (isDir) {
    return `<div class="ficon" style="background:rgba(42,109,244,.12);color:#2a6df4">${Icon.folder.replace('class="','class="w-6 h-6 ')}</div>`;
  }
  const ext = (name.split('.').pop()||'').toLowerCase();
  const m = {
    pdf: ['#ff4d5f', FileIco.pdf],
    jpg: ['#22c55e', FileIco.image], jpeg: ['#22c55e', FileIco.image], png: ['#22c55e', FileIco.image], gif: ['#22c55e', FileIco.image], webp: ['#22c55e', FileIco.image], svg: ['#22c55e', FileIco.image], bmp: ['#22c55e', FileIco.image],
    mp4: ['#ff4d6d', FileIco.video], mov: ['#ff4d6d', FileIco.video], avi: ['#ff4d6d', FileIco.video], mkv: ['#ff4d6d', FileIco.video], webm: ['#ff4d6d', FileIco.video],
    mp3: ['#ff9500', FileIco.audio], wav: ['#ff9500', FileIco.audio], flac: ['#ff9500', FileIco.audio], m4a: ['#ff9500', FileIco.audio], ogg: ['#ff9500', FileIco.audio],
    zip: ['#9b5cff', FileIco.archive], rar: ['#9b5cff', FileIco.archive], '7z': ['#9b5cff', FileIco.archive], tar: ['#9b5cff', FileIco.archive], gz: ['#9b5cff', FileIco.archive],
    doc: ['#2a6df4', FileIco.word], docx: ['#2a6df4', FileIco.word], odt: ['#2a6df4', FileIco.word], rtf: ['#2a6df4', FileIco.word],
    xls: ['#14c26b', FileIco.excel], xlsx: ['#14c26b', FileIco.excel], csv: ['#14c26b', FileIco.excel],
    ppt: ['#ff9500', FileIco.ppt], pptx: ['#ff9500', FileIco.ppt], odp: ['#ff9500', FileIco.ppt],
    txt: ['#94a3b8', FileIco.text], md: ['#7a5cff', FileIco.text], log: ['#94a3b8', FileIco.text], json: ['#94a3b8', FileIco.text], xml: ['#94a3b8', FileIco.text], yaml: ['#94a3b8', FileIco.text], yml: ['#94a3b8', FileIco.text], ini: ['#94a3b8', FileIco.text], conf: ['#94a3b8', FileIco.text],
  };
  if (m[ext]) {
    return `<div class="ficon" style="background:${m[ext][0]}1a;color:${m[ext][0]}">${m[ext][1]}</div>`;
  }
  return `<div class="ficon" style="background:var(--bg-soft);color:var(--text-3)">${FileIco.generic}</div>`;
}

function actionBtn(icon, attr, label, variant) {
  return `<button ${attr} class="${variant||'btn-ghost'}">${icon}<span class="hidden sm:inline">${esc(label)}</span></button>`;
}

/* ---------------- preview ---------------- */
const PREVIEW_IMAGES = ['png','jpg','jpeg','gif','webp','svg','bmp','ico'];
const PREVIEW_TEXT = ['txt','md','log','json','xml','yaml','yml','ini','conf','csv','js','ts','css','html','htm','go','py','java','c','cpp','h','sql','sh','bat'];
const PREVIEW_EMBED = ['pdf','mp4','webm','mov','mp3','wav','flac','m4a','ogg'];

// Client-side extension → MIME map, used to show the file type when a file cannot be previewed.
const EXT_MIME = {
  zip:'application/zip', rar:'application/vnd.rar', '7z':'application/x-7z-compressed', tar:'application/x-tar', gz:'application/gzip',
  doc:'application/msword', docx:'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  xls:'application/vnd.ms-excel', xlsx:'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ppt:'application/vnd.ms-powerpoint', pptx:'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  exe:'application/x-msdownload', msi:'application/x-msdownload', apk:'application/vnd.android.package-archive', iso:'application/x-iso9660-image',
};

function pvImgError(img) { const p = img.parentElement; if (p) p.innerHTML = '<div class="py-12 text-[13px]" style="color:var(--text-3)">图片加载失败</div>'; }

function previewFile(path, name) {
  const ext = (name.split('.').pop()||'').toLowerCase();
  const url = '/api/preview?path='+encodeURIComponent(path);
  if (PREVIEW_IMAGES.includes(ext)) {
    openSheet(`<div class="mt-1 flex items-center justify-center min-h-[16rem]"><img src="${url}" class="w-full" style="max-height:70vh;object-fit:contain;border-radius:1rem" onerror="pvImgError(this)" /></div>`, name, true);
  } else if (PREVIEW_TEXT.includes(ext)) {
    openSheet(`<div class="mt-1"><pre id="pv-text" class="p-3 rounded-xl overflow-auto text-[12px] font-mono whitespace-pre-wrap" style="max-height:60vh;background:var(--bg-soft);color:var(--text-2)">加载中…</pre></div>`, name, true);
    fetch(url).then(r => r.ok ? r.text() : Promise.reject()).then(t => {
      const el = document.getElementById('pv-text');
      if (el) el.textContent = t.length > 200000 ? t.slice(0,200000)+'\n\n… 文件过大，仅预览部分内容' : t;
    }).catch(() => { const el = document.getElementById('pv-text'); if (el) el.textContent = '文本加载失败'; });
  } else if (PREVIEW_EMBED.includes(ext)) {
    window.open(url, '_blank');
  } else {
    const mime = EXT_MIME[ext] || 'application/octet-stream';
    openSheet(`<div class="mt-1 text-center space-y-4">
      <div class="ico" style="width:3.4rem;height:3.4rem">${Icon.file.replace('class="','class="w-6 h-6 ')}</div>
      <div>
        <div class="text-[15px] font-semibold">暂不支持预览此类文件</div>
        <div class="text-[12px] mt-1 break-all" style="color:var(--text-3)">${esc(mime)}</div>
      </div>
      <a href="/api/download?path=${encodeURIComponent(path)}" class="btn-primary w-full !py-3">下载文件</a>
    </div>`, name);
  }
}

/* ---------------- nav model ---------------- */
const NAV = [
  { key:'home',   label:'首页',   icon:'home',  hash:'#/home',   tag:'' },
  { key:'files',  label:'文件管理', icon:'files', hash:'#/files',  tag:'' },
  { key:'shared', label:'共享链接', icon:'share', hash:'#/shared', tag:'' },
  { key:'trash',  label:'回收站', icon:'trash', hash:'#/trash',  tag:'' },
  { key:'audit',  label:'审计日志', icon:'clock', hash:'#/audit',  tag:'' },
  { key:'users',  label:'应用密码',icon:'key',  hash:'#/users',  tag:'' },
  { key:'settings',label:'系统设置',icon:'gear',  hash:'#/settings',tag:'' },
];

function currentNavKey() {
  const r = state.route;
  if (r === 'login') return null;
  if (NAV.find(n => n.key === r)) return r;
  return 'home';
}

function userChip() {
  const u = state.user;
  const name = u ? (u.display_name||u.username||'') : '';
  return `
    <div class="flex items-center gap-3 px-3 pt-5 pb-3">
      <div class="avatar grad-icon" style="width:2.4rem;height:2.4rem;font-size:.78rem">${esc(initials(name||'D'))}</div>
      <div class="min-w-0 flex-1">
        <div class="text-[13px] font-semibold truncate">${esc(name||'—')}</div>
        <div class="text-[11px] text-[var(--text-3)] truncate">个人空间</div>
      </div>
    </div>`;
}

function desktopNav() {
  const k = currentNavKey();
  const topGroup = ['home','files'];
  const midGroup = ['shared','trash','audit','users'];
  const navObj = key => NAV.find(n => n.key === key);
  return `<aside class="hidden lg:flex flex-col w-64 shrink-0 h-screen sticky top-0 overflow-y-auto" style="border-right:1px solid var(--line);background:var(--bg-soft)">
    <div class="px-5 pt-6 pb-4 flex items-center gap-3">
      <div class="w-10 h-10 shrink-0 drop-shadow" style="filter:drop-shadow(0 6px 14px rgba(42,109,244,.35))">${LOGO}</div>
      <div>
        <div class="font-bold leading-none tracking-tight" style="font-size:1.02rem">ChinaSCLM<sup style="color:var(--brand);font-size:.6rem;margin-left:2px">DAV</sup></div>
        <div class="text-[11px] text-[var(--text-3)] mt-1">${esc(BRAND_TAG)}</div>
      </div>
    </div>
    <nav class="mt-1 px-3 space-y-0.5 pb-4">
      <div class="px-3 py-1 text-[10px] font-bold uppercase tracking-[.14em] text-[var(--text-3)]">空间</div>
      ${topGroup.map(key => navLink(navObj(key),k)).join('')}
      ${midGroup.map(key => navLink(navObj(key),k)).join('')}
      ${navLink(navObj('settings'),k)}
    </nav>
    <div class="mt-auto px-3 pb-5">
      <div class="card p-2" style="box-shadow:none">${userChip()}
        <button data-logout class="nav-item w-full mt-1" style="color:var(--text-2)"><span class="ic">${Icon.logout.replace('class="','class="w-full h-full ')}</span>退出登录</button>
      </div>
    </div>
  </aside>`;
}

function navLink(n, k) {
  const on = k === n.key;
  return `<a href="${n.hash}" class="nav-item ${on?'on':''}"><span class="ic">${Icon[n.icon].replace('class="','class="w-full h-full ')}</span>${esc(n.label)}${n.tag?`<span class="nav-tag">${esc(n.tag)}</span>`:''}</a>`;
}

function topBar(title, subtitle, actions) {
  return `<header class="sticky top-0 z-30" style="background:color-mix(in srgb, var(--bg) 82%, transparent);backdrop-filter:saturate(180%) blur(20px);border-bottom:1px solid var(--line)"><div class="px-5 md:px-8 h-16 flex items-center justify-between gap-3">
    <div class="min-w-0 flex-1"><div class="font-bold tracking-tight truncate" style="font-size:1.25rem">${esc(title)}</div>${subtitle ? `<div class="text-[12px] text-[var(--text-3)] truncate">${esc(subtitle)}</div>` : ''}</div>
    <div class="flex items-center gap-2 shrink-0 whitespace-nowrap">${actions||''}</div>
  </div></header>`;
}

function mobileTabBar() {
  const k = currentNavKey();
  return `<nav class="lg:hidden fixed bottom-0 inset-x-0 z-40" style="background:color-mix(in srgb, var(--bg) 86%, transparent);backdrop-filter:saturate(180%) blur(20px);border-top:1px solid var(--line)"><div class="grid grid-cols-7 h-16">
    ${NAV.map(n => {
      const on = k === n.key;
      return `<a href="${n.hash}" class="flex flex-col items-center justify-center gap-1" style="${on?'color:var(--brand)':'color:var(--text-3)'}"><span class="w-6 h-6">${Icon[n.icon].replace('class="','class="w-6 h-6 ')}</span><span class="text-[10px] font-semibold">${esc(n.label)}</span></a>`;
    }).join('')}
  </div></nav>`;
}

function appShell(body) {
  return `<div class="lg:flex min-h-screen">${desktopNav()}<div class="flex-1 min-w-0"><main class="min-h-screen pb-24 lg:pb-10">${body}</main></div>${mobileTabBar()}</div>`;
}

const SPINNER = `<div class="flex justify-center py-16"><svg class="w-7 h-7 animate-spin" style="color:var(--text-3)" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 12a8 8 0 1 1-2.3-5.6"/><path d="M20 4v4h-4"/></svg></div>`;

/* modal/sheet */
function openSheet(content, title, wide) {
  closeAll();
  const wrap = document.createElement('div');
  sheetEl = wrap;
  wrap.className = 'fixed inset-0 z-[100] flex items-end md:items-center justify-center p-3';
  wrap.innerHTML = `<div class="absolute inset-0 bg-black/30 backdrop-blur-sm" data-close></div>
    <div class="card relative w-full max-w-md max-h-[88vh] overflow-hidden flex flex-col fade-in" style="border-radius:1.6rem${wide?';max-width:44rem':''}">
      <div class="px-5 pt-4 pb-2 flex items-center justify-between">
        <h3 class="text-lg font-bold tracking-tight truncate">${esc(title)}</h3>
        <button data-close class="icon-btn"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M6 6l12 12M18 6 6 18"/></svg></button>
      </div>
      <div class="px-5 pb-5 overflow-y-auto">${content}</div>
    </div>`;
  document.body.appendChild(wrap);
  wrap.addEventListener('click', e => { if (e.target.closest('[data-close]')) { closeAll(); } });
  return wrap;
}

/* ============================================================
 * VIEWS
 * ============================================================ */

async function renderLogin() {
  const app = document.getElementById('app');
  app.innerHTML = `<div class="min-h-screen relative overflow-hidden flex items-center justify-center px-5 py-10">
    <div class="pointer-events-none absolute -top-28 -right-24 w-96 h-96 rounded-full" style="background:radial-gradient(circle, rgba(42,109,244,.16), transparent 70%); filter:blur(10px)"></div>
    <div class="pointer-events-none absolute -bottom-32 -left-24 w-96 h-96 rounded-full" style="background:radial-gradient(circle, rgba(122,92,255,.14), transparent 70%); filter:blur(10px)"></div>
    <div class="w-full max-w-[400px] raise">
      <div class="card p-8">
        <div class="flex flex-col items-center mb-7">
          <div class="w-16 h-16 mb-4 drop-shadow-xl">${LOGO}</div>
          <h1 class="text-2xl font-extrabold tracking-tight">欢迎回来</h1>
          <p class="text-[13px] mt-1" style="color:var(--text-2)">登录 ${esc(BRAND_NAME)}，访问你的 WebDAV 网盘</p>
        </div>
        <form id="login-form" class="space-y-3.5">
          <div class="field"><label>账号</label><input id="lg-user" class="input" type="text" placeholder="邮箱或用户名" autocomplete="username" /></div>
          <div class="field"><label>密码</label><input id="lg-pass" class="input" type="password" placeholder="请输入密码" autocomplete="current-password" /></div>
          <div id="totp-row" class="hidden"><label class="field" style="font-size:.8rem;font-weight:600;color:var(--text-2);display:block;margin-bottom:.4rem">两步验证码</label><input id="lg-totp" class="input" type="text" placeholder="6 位验证码" inputmode="numeric" maxlength="6" /></div>
          <p id="lg-err" class="hidden text-[13px] font-medium" style="color:var(--danger)"></p>
          <button class="btn-primary w-full !py-3 mt-1" type="submit"><span class="label">登录</span><span class="loader hidden"><svg class="w-5 h-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 12a8 8 0 1 1-2.3-5.6"/><path d="M20 4v4h-4"/></svg></span></button>
        </form>
      </div>
      <div class="flex items-center justify-center gap-2 mt-6"><span class="chip">兼容 WebDAV 客户端</span><span class="chip">支持应用密码</span></div>
    </div>
  </div>`;
  const saved = localStorage.getItem('cscd_user') || '';
  document.getElementById('lg-user').value = saved;
  document.getElementById('lg-user').focus();
  if (saved) document.getElementById('lg-pass').focus();

  document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.querySelector('#login-form .btn-primary');
    const loader = btn.querySelector('.loader');
    const label = btn.querySelector('.label');
    btn.disabled = true; loader.classList.remove('hidden'); label.textContent = '正在登录…';
    const errEl = document.getElementById('lg-err'); errEl.classList.add('hidden');
    try {
      const data = await api('/login', { method:'POST', body:{ username: document.getElementById('lg-user').value.trim(), password: document.getElementById('lg-pass').value, totp_code: document.getElementById('lg-totp').value.trim()||undefined }});
      if (data.totp_enabled === false && data.totp_forced) {
        errEl.textContent = '账号已启用两步验证，请输入验证码';
        errEl.classList.remove('hidden');
        document.getElementById('totp-row').classList.remove('hidden');
        btn.disabled = false; loader.classList.add('hidden'); label.textContent = '登录';
        return;
      }
      localStorage.setItem('cscd_user', document.getElementById('lg-user').value.trim());
      state.user = data;
      toast('登录成功','ok');
      location.hash = '#/home';
    } catch(err) { errEl.textContent = err.message; errEl.classList.remove('hidden'); }
    btn.disabled = false; loader.classList.add('hidden'); label.textContent = '登录';
  });
}

/* dashboard */
async function renderHome() {
  const app = document.getElementById('app');
  const uname = (state.user && (state.user.display_name||state.user.username))||'';
  app.innerHTML = appShell(topBar('首页', `${esc(uname)}，欢迎回来`, actionBtn(Icon.refresh,'data-refresh','刷新','btn-ghost')) + '<div class="px-5 md:px-8 py-6"><div id="home-body" class="space-y-6">' + SPINNER + '</div></div>');
  const btn = document.querySelector('[data-refresh]');
  if (btn) btn.onclick = () => { document.getElementById('home-body').innerHTML = SPINNER; loadHome(); };
  async function loadHome() {
    try {
      const d = await api('/dashboard');
      renderDashboard(d);
    } catch(e) { document.getElementById('home-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  loadHome();
}

function renderDashboard(d) {
  const body = document.getElementById('home-body');
  const total = d.total_size || 0;
  const cap = 1024*1024*1024; // 1GB reference
  const pct = Math.max(0, Math.min(100, total/cap*100));
  const R = 52, C = 2*Math.PI*R;
  const off = C*(1 - pct/100);
  const catTotal = total || 1;
  body.innerHTML = `
    <div class="card p-6 relative overflow-hidden">
      <div class="pointer-events-none absolute -top-16 -right-16 w-64 h-64 rounded-full" style="background:radial-gradient(circle, rgba(42,109,244,.18), transparent 70%);filter:blur(8px)"></div>
      <div class="flex flex-col md:flex-row md:items-center gap-6">
        <div class="flex-1 min-w-0 relative">
          <div class="text-[13px] font-semibold tracking-wide" style="color:var(--text-2)">已用存储</div>
          <div class="mt-1 grad-text font-extrabold tracking-tight" style="font-size:2.6rem;line-height:1">${fmtSize(total)}</div>
          <div class="mt-2 text-[12px]" style="color:var(--text-3)">${d.file_count} 个文件 · ${d.dir_count} 个文件夹</div>
          <div class="flex flex-wrap gap-2 mt-4">
            <span class="stat-chip"><span style="color:var(--brand);font-weight:800">${d.file_count}</span><span class="text-[12px]" style="color:var(--text-2)">文件</span></span>
            <span class="stat-chip"><span style="color:var(--brand-2);font-weight:800">${d.dir_count}</span><span class="text-[12px]" style="color:var(--text-2)">文件夹</span></span>
          </div>
        </div>
        <div class="donut mx-auto md:mx-0 relative">
          <svg width="152" height="152" viewBox="0 0 152 152">
            <defs><linearGradient id="ringgrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2a6df4"/><stop offset="1" stop-color="#7a5cff"/></linearGradient></defs>
            <circle class="track" cx="76" cy="76" r="${R}" fill="none" stroke-width="14"/>
            <circle class="val" cx="76" cy="76" r="${R}" fill="none" stroke-width="14" stroke-linecap="round" stroke-dasharray="${C.toFixed(1)}" stroke-dashoffset="${off.toFixed(1)}"/>
          </svg>
          <div class="donut-center"><div class="font-extrabold tracking-tight" style="font-size:1.6rem">${pct.toFixed(0)}<span style="font-size:.8rem">%</span></div><div class="text-[11px]" style="color:var(--text-3)">空间已用</div></div>
        </div>
      </div>
      <a href="#/files" class="btn-primary mt-6" style="width:100%">进入文件空间<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="m9 18 6-6-6-6"/></svg></a>
    </div>

    <div>
      <div class="sec-title">存储分类</div>
      <div class="grid-cats">${d.categories && d.categories.length ? d.categories.map(c => {
        const meta = CAT_META[c.key] || CAT_META.other;
        const w = Math.max(4, Math.min(100, c.size/catTotal*100));
        return `<div class="cate fade-in"><div class="flex items-center gap-2.5"><div class="ficon" style="width:2.5rem;height:2.5rem;background:${meta.c}1a">${meta.e}</div><div class="flex-1 min-w-0"><div class="text-[13.5px] font-semibold truncate">${esc(c.label)}</div><div class="text-[11.5px]" style="color:var(--text-3)">${c.count} 项</div></div></div><div class="mt-3 flex items-center justify-between text-[12.5px]"><span class="font-bold">${fmtSize(c.size)}</span><span style="color:var(--text-3)">${(c.size/catTotal*100).toFixed(0)}%</span></div><div class="mt-1.5 h-1.5 rounded-full overflow-hidden" style="background:var(--bg-soft)"><div class="h-full rounded-full" style="width:${w}%;background:${meta.c}"></div></div></div>`;
      }).join('') : '<div class="card p-8 text-center text-sm" style="color:var(--text-3);grid-column:1/-1">暂无存储分类</div>'}</div>
    </div>

    <div>
      <div class="sec-title">最近文件</div>
      <div class="card px-4 py-2">${d.recent && d.recent.length ? d.recent.map(f => `
        <div class="recent-row">
          ${fileGlyph(f.name,f.is_dir,f.size)}
          <div class="flex-1 min-w-0"><div class="text-[14px] font-semibold truncate">${esc(f.name)}</div><div class="text-[12px]" style="color:var(--text-3)">${fmtTime(f.mod_time)} · ${fmtSize(f.size)}</div></div>
          <a href="#/files" class="btn-ghost !px-3 !py-1.5 text-[12px]" style="color:var(--brand)">打开</a>
        </div>`).join('') : '<div class="py-10 text-center text-sm" style="color:var(--text-3)">还没有文件</div>'}</div>
    </div>`;
}

/* files view */
function filesCrumb(dir) {
  const parts = dir === '/' ? [] : dir.split('/').filter(Boolean);
  let acc = '';
  const segs = [{ name:'我的文件', path:'/' }];
  parts.forEach(p => { acc += '/' + p; segs.push({ name:p, path:acc||'/' }); });
  return segs;
}

async function renderFiles() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('文件', state.dir||'/', actionBtn(Icon.folderPlus,'data-newdir','新建文件夹','btn-ghost') + actionBtn(Icon.upload,'data-upload','上传','btn-primary')) + `<div class="px-5 md:px-8 py-5">
    <div class="flex items-center gap-2 mb-4">
      <div class="crumb">${filesCrumb(state.dir).map((c,i) => {
        const isLast = i===filesCrumb(state.dir).length-1;
        return `${i>0?`<span class="sep">/</span>`:''}${isLast?`<span class="font-semibold text-[15px]">${esc(c.name)}</span>`:`<a data-jump="${esc(c.path)}" style="cursor:pointer">${esc(c.name)}</a>`}`;
      }).join('')}</div>
      <div class="ml-auto flex-1 max-w-[240px]"><div class="relative">
        <span class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style="color:var(--text-3)">${Icon.search.replace('class="','class="w-4 h-4 ')}</span>
        <input id="file-search" class="input !pl-9 !py-2 text-[13px]" placeholder="搜索当前目录" value="${esc(state.search)}" />
      </div></div>
      <input type="file" id="file-upload-input" class="hidden" multiple />
    </div>
    <div id="files-body" class="space-y-2">${SPINNER}</div>
  </div>`);

  app.onclick = e => {
    const j = e.target.closest('[data-jump]');
    if (j) { e.preventDefault(); state.dir = j.dataset.jump; state.search = ''; route(); }
  };

  document.getElementById('file-search').addEventListener('keydown', e => {
    if (e.key === 'Enter') { state.search = e.target.value.trim(); loadFiles(); }
  });

  app.querySelector('[data-newdir]').onclick = () => {
    const s = openSheet(`<div class="space-y-4 mt-1"><div class="field"><label>文件夹名称</label><input id="nd-name" class="input" placeholder="新文件夹" /></div><button id="nd-ok" class="btn-primary w-full !py-3">创建</button></div>`, '新建文件夹');
    document.getElementById('nd-ok').onclick = async () => {
      const name = document.getElementById('nd-name').value.trim();
      if (!name) return toast('请输入名称','error');
      try { await api('/mkdir',{method:'POST',body:{dir:state.dir,name}}); s.remove(); toast('已创建','ok'); loadFiles(); } catch(e) { toast(e.message,'error'); }
    };
  };

  app.querySelector('[data-upload]').onclick = () => document.getElementById('file-upload-input').click();
  document.getElementById('file-upload-input').addEventListener('change', async (e) => {
    const files = e.target.files; if (!files.length) return;
    for (const f of Array.from(files)) {
      const fd = new FormData(); fd.append('file', f);
      toast(`上传 ${f.name}…`);
      try { await api('/upload?dir='+encodeURIComponent(state.dir),{method:'POST',body:fd}); } catch(err) { toast(err.message,'error'); }
    }
    e.target.value = '';
    loadFiles();
  });

  async function loadFiles() {
    try {
      let q = '/files?dir='+encodeURIComponent(state.dir||'/');
      if (state.search) q += '&q='+encodeURIComponent(state.search);
      const d = await api(q);
      renderFileList(d.entries, d.path);
    } catch(e) { document.getElementById('files-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  loadFiles();
}

function renderFileList(entries, dir) {
  const body = document.getElementById('files-body');
  if (!entries || !entries.length) {
    body.innerHTML = `<div class="empty card fade-in">
      <div class="ico">${Icon.folder.replace('class="','class="w-7 h-7 ')}</div>
      <div class="font-semibold">${state.search?'没有匹配项':'此文件夹为空'}</div>
      <div class="text-[13px] mt-1" style="color:var(--text-3)">${state.search?'尝试更换关键词':'点击「上传」添加文件，或「新建文件夹」整理'} </div>
      ${state.search?'':`<button class="btn-primary mt-5 !py-2.5" data-empty-upload>上传文件</button>`}
    </div>`;
    const up = body.querySelector('[data-empty-upload]');
    if (up) up.onclick = () => document.getElementById('file-upload-input').click();
    return;
  }
  body.innerHTML = '<div class="space-y-2">' + entries.map(f => {
    const isDir = f.is_dir;
    const path = f.path;
    return `<div class="file-row fade-in" data-path="${esc(path)}" data-is-dir="${isDir?'1':'0'}" data-name="${esc(f.name)}">
      ${fileGlyph(f.name, isDir, f.size)}
      <button data-open class="flex-1 min-w-0 text-left"><div class="text-[14px] font-semibold truncate">${esc(f.name)}</div><div class="text-[12px] mt-0.5" style="color:var(--text-3)">${isDir?'文件夹':fmtSize(f.size)} · ${fmtTime(f.mod_time)}</div></button>
      <div class="flex items-center gap-0.5 shrink-0">
        ${isDir ? '' : `<button data-act="download" class="icon-btn good" style="width:2.3rem;height:2.3rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M4 21h16"/></svg></button>`}
        ${isDir ? '' : `<button data-act="share" class="icon-btn" style="width:2.3rem;height:2.3rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3v12"/><path d="m8 7 4-4 4 4"/><path d="M6 12H4a2 2 0 0 0-2 2v5a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-5a2 2 0 0 0-2-2h-2"/></svg></button>`}
        <button data-act="more" class="icon-btn" style="width:2.3rem;height:2.3rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="5" r="1.5" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none"/><circle cx="12" cy="19" r="1.5" fill="currentColor" stroke="none"/></svg></button>
      </div>
    </div>`;
  }).join('') + '</div>';

  body.onclick = async (e) => {
    const row = e.target.closest('.file-row');
    if (!row) return;
    const path = row.dataset.path;
    const open = e.target.closest('[data-open]');
    const dl = e.target.closest('[data-act="download"]');
    const sh = e.target.closest('[data-act="share"]');
    const more = e.target.closest('[data-act="more"]');
    if (open) {
      if (row.dataset.isDir === '1') { state.dir = path; state.search = ''; route(); }
      else { previewFile(path, row.dataset.name||''); }
    }
    else if (dl) { window.location.href = '/api/download?path='+encodeURIComponent(path); }
    else if (sh) { shareFile(path); }
    else if (more) { fileActions(path); }
  });
}

function fileActions(path) {
  const s = openSheet(`<div class="space-y-0.5 mt-1">
    <button data-a="download" class="w-full flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-[rgba(128,128,128,0.08)]"><span class="w-5 h-5" style="color:var(--brand)">${Icon.download.replace('class="','class="w-5 h-5 ')}</span><span class="text-[15px] font-medium">下载</span></button>
    <button data-a="share" class="w-full flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-[rgba(128,128,128,0.08)]"><span class="w-5 h-5" style="color:var(--brand)">${Icon.share.replace('class="','class="w-5 h-5 ')}</span><span class="text-[15px] font-medium">分享链接</span></button>
    <button data-a="versions" class="w-full flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-[rgba(128,128,128,0.08)]"><span class="w-5 h-5" style="color:var(--text-2)">${Icon.clock.replace('class="','class="w-5 h-5 ')}</span><span class="text-[15px] font-medium">版本历史</span></button>
    <button data-a="rename" class="w-full flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-[rgba(128,128,128,0.08)]"><span class="w-5 h-5" style="color:var(--text-2)">${Icon.restore.replace('class="','class="w-5 h-5 ')}</span><span class="text-[15px] font-medium">重命名</span></button>
    <div class="border-t hairline my-1"></div>
    <button data-a="delete" class="w-full flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-[rgba(255,69,58,0.1)]" style="color:var(--danger)"><span class="w-5 h-5">${Icon.trash.replace('class="','class="w-5 h-5 ')}</span><span class="text-[15px] font-medium">移入回收站</span></button>
  </div>`, '操作');
  s.addEventListener('click', async (e) => {
    const btn = e.target.closest('[data-a]'); if (!btn) return;
    const act = btn.dataset.a;
    s.remove();
    if (act === 'download') { window.location.href = '/api/download?path='+encodeURIComponent(path); }
    else if (act === 'share') { shareFile(path); }
    else if (act === 'versions') { versionsSheet(path); }
    else if (act === 'rename') { renameSheet(path); }
    else if (act === 'delete') {
      if (await iConfirm('移入回收站','确定将该项目移入回收站吗？','移入',true)) {
        try { await api('/delete',{method:'POST',body:{path}}); toast('已移入回收站','ok'); route(); } catch(err) { toast(err.message,'error'); }
      }
    }
  });
}

function renameSheet(path) {
  const name = path.split('/').filter(Boolean).pop()||'文件';
  const s = openSheet(`<div class="space-y-4 mt-1"><div class="field"><label>新名称</label><input id="rn-name" class="input" value="${esc(name)}" /></div><button id="rn-ok" class="btn-primary w-full !py-3">保存</button></div>`, '重命名');
  document.getElementById('rn-ok').onclick = async () => {
    const p = path.split('/'); const base = p.slice(0,-1).join('/')||'/';
    const newName = document.getElementById('rn-name').value.trim();
    if (!newName) return toast('请输入名称','error');
    try { await api('/rename',{method:'POST',body:{path, new: (base==='/'?'/':base+'/')+newName}}); toast('已重命名','ok'); s.remove(); route(); } catch(e) { toast(e.message,'error'); }
  };
}

function shareFile(path) {
  const name = path.split('/').filter(Boolean).pop()||'文件';
  const s = openSheet(`<div class="space-y-4 mt-1"><p class="text-[13px]" style="color:var(--text-2)">创建可访问的分享链接：<span class="font-semibold" style="color:var(--text)">${esc(name)}</span></p><div class="field"><label>有效期（小时，留空为永久）</label><input id="sh-exp" class="input" type="number" placeholder="例如 24，留空永久" /></div><button id="sh-ok" class="btn-primary w-full !py-3">创建链接</button></div>`, '分享文件');
  document.getElementById('sh-ok').onclick = async () => {
    const exp = parseInt(document.getElementById('sh-exp').value, 10);
    try {
      const sh = await api('/shares',{method:'POST',body:{path, expires_in: Number.isNaN(exp)?0:exp}});
      s.remove();
      showShareResult(sh);
    } catch(e) { toast(e.message,'error'); }
  };
}

function showShareResult(sh) {
  const url = location.origin + '/s/' + sh.token;
  openSheet(`<div class="space-y-4 mt-1">
    <div class="flex items-center gap-3 p-3 rounded-xl" style="background:rgba(20,194,107,.12)"><svg class="w-5 h-5" style="color:var(--success)" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6 9 17l-5-5"/></svg><div class="text-[13px] font-medium">分享链接已创建</div></div>
    <div class="input !bg-transparent break-all !text-[13px]" style="color:var(--text)">${esc(url)}</div>
    <button id="sh-copy" class="btn-primary w-full !py-3">复制链接</button>
    <p class="text-[12px] text-center" style="color:var(--text-3)">链接开箱即用，任何人可下载此文件</p>
  </div>`, '分享链接');
  document.getElementById('sh-copy').onclick = () => navigator.clipboard.writeText(url).then(()=>toast('已复制链接','ok')).catch(()=>toast('复制失败','error'));
}

function versionsSheet(path) {
  const s = openSheet('<div class="py-6 text-center" style="color:var(--text-3)">加载中…</div>', '版本历史');
  (async () => {
    try {
      const vs = await api('/versions?path='+encodeURIComponent(path));
      const list = vs && vs.length ? vs.map(v => `<div class="flex items-center gap-3 px-1 py-3" style="border-bottom:1px solid var(--line)">${fileGlyph(v.name,false,v.size)}<div class="flex-1 min-w-0"><div class="text-[14px] font-semibold truncate">${esc(v.name)}</div><div class="text-[12px]" style="color:var(--text-3)">${fmtDateTime(v.created_at)} · ${fmtSize(v.size)}</div></div><button data-v="dl" data-id="${v.id}" class="icon-btn good" style="width:2.3rem;height:2.3rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M4 21h16"/></svg></button><button data-v="del" data-id="${v.id}" class="icon-btn danger" style="width:2.3rem;height:2.3rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 7h16"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12"/><path d="M9 7V5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"/></svg></button></div>`).join('') : '<div class="py-10 text-center text-sm" style="color:var(--text-3)">暂无版本历史</div>';
      s.innerHTML = `<div class="absolute inset-0 bg-black/30 backdrop-blur-sm" data-close></div><div class="card relative w-full max-w-md max-h-[88vh] overflow-hidden flex flex-col fade-in" style="border-radius:1.6rem"><div class="px-5 pt-4 pb-2 flex items-center justify-between"><h3 class="text-lg font-bold tracking-tight">版本历史</h3><button data-close class="icon-btn"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M6 6l12 12M18 6 6 18"/></svg></button></div><div class="px-5 pb-5 overflow-y-auto"><div>${list}</div></div></div>`;
      s.addEventListener('click', async (e) => {
        const b = e.target.closest('[data-v]'); if (!b) return;
        const id = b.dataset.id;
        if (b.dataset.v === 'dl') window.location.href = '/api/versions/download?id='+id;
        if (b.dataset.v === 'del') { try { await api('/versions/delete',{method:'POST',body:{id:Number(id)}}); toast('已删除版本','ok'); versionsSheet(path); } catch(err) { toast(err.message,'error'); } }
      });
    } catch(e) { s.innerHTML = `<div class="absolute inset-0 bg-black/30 backdrop-blur-sm" data-close></div><div class="card relative w-full max-w-md max-h-[88vh] overflow-hidden flex flex-col fade-in" style="border-radius:1.6rem"><div class="px-5 pt-4 pb-2 flex items-center justify-between"><h3 class="text-lg font-bold tracking-tight">版本历史</h3><button data-close class="icon-btn"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M6 6l12 12M18 6 6 18"/></svg></button></div><div class="px-5 pb-5 overflow-y-auto"><div class="py-10 text-center text-sm" style="color:var(--danger)">${esc(e.message)}</div></div></div>`; }
  })();
}

/* shared */
async function renderShared() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('共享链接','管理已分享的文件', actionBtn(Icon.refresh,'data-refresh','刷新','btn-ghost')) + '<div class="px-5 md:px-8 py-6"><div id="sh-body" class="space-y-2">' + SPINNER + '</div></div>');
  app.querySelector('[data-refresh]').onclick = load;
  async function load() {
    try {
      const list = await api('/shares');
      const body = document.getElementById('sh-body');
      if (list && list.length) {
        body.innerHTML = list.map(sh => {
          const exp = sh.expires_at ? fmtDateTime(sh.expires_at) : '永久有效';
          const isExpired = sh.expires_at ? new Date(sh.expires_at) < new Date() : false;
          return `<div class="file-row">
            ${fileGlyph(sh.name,false,sh.size)}
            <div class="flex-1 min-w-0"><div class="text-[14px] font-semibold truncate">${esc(sh.name)}</div><div class="text-[12px] mt-0.5 flex items-center gap-2" style="color:var(--text-3)"><span>${fmtSize(sh.size)}</span>·<span>${exp}</span>·<span>${sh.download_count} 次下载</span>${isExpired?`<span class="chip badge-red">已过期</span>`:''}</div></div>
            <button data-copy data-token="${esc(sh.token)}" class="icon-btn primary" style="width:2.3rem;height:2.3rem">${Icon.copy.replace('class="','class="w-5 h-5 ')}</button>
            <button data-del data-id="${sh.id}" class="icon-btn danger" style="width:2.3rem;height:2.3rem">${Icon.trash.replace('class="','class="w-5 h-5 ')}</button>
          </div>`;
        }).join('');
        body.addEventListener('click', async (e) => {
          const b = e.target.closest('[data-copy]'); if (b) { navigator.clipboard.writeText(location.origin+'/s/'+b.dataset.token).then(()=>toast('已复制链接','ok')); }
          const d = e.target.closest('[data-del]'); if (d) {
            if (await iConfirm('撤销分享','确定撤销此分享链接吗？','撤销',true)) {
              try { await api('/shares',{method:'DELETE',body:{id:Number(d.dataset.id)}}); toast('已撤销','ok'); load(); } catch(err) { toast(err.message,'error'); }
            }
          }
        });
      } else {
        body.innerHTML = `<div class="empty card fade-in"><div class="ico">${Icon.link.replace('class="','class="w-7 h-7 ')}</div><div class="font-semibold">暂无共享链接</div><div class="text-[13px] mt-1" style="color:var(--text-3)">前往「文件」分享文件以生成可访问的链接</div><a href="#/files" class="btn-primary mt-5">去分享</a></div>`;
      }
    } catch(e) { document.getElementById('sh-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  load();
}

/* trash */
async function renderTrash() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('回收站','删除的文件可在此恢复', '<button data-empty class="btn-danger" style="padding:.5rem .8rem">清空回收站</button>') + '<div class="px-5 md:px-8 py-6"><div id="tr-body" class="space-y-2">' + SPINNER + '</div></div>');
  app.querySelector('[data-empty]').onclick = async () => {
    if (await iConfirm('清空回收站','回收站中的文件将被永久删除，无法恢复。','清空',true)) {
      try { await api('/trash/empty',{method:'POST'}); toast('已清空','ok'); load(); } catch(e) { toast(e.message,'error'); }
    }
  };
  async function load() {
    try {
      const list = await api('/trash');
      const body = document.getElementById('tr-body');
      if (list && list.length) {
        body.innerHTML = list.map(t => `<div class="file-row">${fileGlyph(t.name,t.is_dir,t.size)}<div class="flex-1 min-w-0"><div class="text-[14px] font-semibold truncate">${esc(t.name)}</div><div class="text-[12px]" style="color:var(--text-3)">${esc(t.original_path)} · 删除于 ${fmtDateTime(t.deleted_at)}</div></div><button data-restore data-id="${t.id}" class="icon-btn good" style="width:2.3rem;height:2.3rem">${Icon.restore.replace('class="','class="w-5 h-5 ')}</button><button data-purge data-id="${t.id}" class="icon-btn danger" style="width:2.3rem;height:2.3rem">${Icon.trash.replace('class="','class="w-5 h-5 ')}</button></div>`).join('');
        body.addEventListener('click', async (e) => {
          const r = e.target.closest('[data-restore]'); if (r) { try { await api('/trash/restore',{method:'POST',body:{id:Number(r.dataset.id)}}); toast('已恢复','ok'); load(); } catch(err) { toast(err.message,'error'); } }
          const p = e.target.closest('[data-purge]'); if (p) { if (await iConfirm('彻底删除','确定永久删除此文件吗？','删除',true)) { try { await api('/trash/purge',{method:'POST',body:{id:Number(p.dataset.id)}}); toast('已删除','ok'); load(); } catch(err) { toast(err.message,'error'); } } }
        });
      } else {
        body.innerHTML = `<div class="empty card fade-in"><div class="ico">${Icon.restore.replace('class="','class="w-7 h-7 ')}</div><div class="font-semibold">回收站是空的</div></div>`;
      }
    } catch(e) { document.getElementById('tr-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  load();
}

/* app passwords */
async function renderUsers() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('应用密码','用于 WebDAV 客户端连接','') + `<div class="px-5 md:px-8 py-6 space-y-6">
    <div class="card p-5"><div class="sec-title" style="margin-bottom:.9rem">WebDAV 连接信息</div>
      <div class="space-y-3 text-[13px]">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2"><span style="color:var(--text-2)">服务器地址</span><div class="flex items-center gap-2"><code class="px-2.5 py-1 rounded-lg" style="background:var(--bg-soft);font-size:12px">${esc(location.origin)}/dav/</code><button data-copy-dav class="icon-btn" style="width:2rem;height:2rem">${Icon.copy.replace('class="','class="w-4 h-4 ')}</button></div></div>
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2"><span style="color:var(--text-2)">账户</span><div class="flex items-center gap-2"><code class="px-2.5 py-1 rounded-lg" style="background:var(--bg-soft);font-size:12px">${esc(state.user ? state.user.email||state.user.username : '')}</code><button data-copy-acc class="icon-btn" style="width:2rem;height:2rem">${Icon.copy.replace('class="','class="w-4 h-4 ')}</button></div></div>
      </div>
      <div class="flex items-start gap-2 mt-4 text-[12px] p-3 rounded-xl" style="background:var(--bg-soft);color:var(--text-2)"><span style="color:var(--warning)">${Icon.zap.replace('class="','class="w-4 h-4 shrink-0 mt-0.5 ')}</span><div>第三方客户端请使用本页面创建的应用密码，而非登录密码。</div></div>
    </div>
    <div>
      <div class="flex items-center justify-between mb-3"><div class="sec-title" style="margin-bottom:0">已授权应用</div><button data-add class="btn-primary !py-2 !px-3.5 text-[13px]">+ 添加应用</button></div>
      <div id="ap-body" class="space-y-2">${SPINNER}</div>
    </div>
  </div>`);

  app.querySelector('[data-copy-dav]').onclick = () => navigator.clipboard.writeText(location.origin+'/dav/').then(() => toast('已复制地址','ok'));
  app.querySelector('[data-copy-acc]').onclick = () => { const v = state.user ? (state.user.email||state.user.username) : ''; navigator.clipboard.writeText(v).then(() => toast('已复制账户','ok')); };
  app.querySelector('[data-add]').onclick = () => {
    const s = openSheet(`<div class="space-y-4 mt-1"><div class="field"><label>应用名称（例如 RaiDrive、坚果云）</label><input id="ap-name" class="input" placeholder="应用名称" /></div><button id="ap-ok" class="btn-primary w-full !py-3">生成应用密码</button></div>`, '添加应用');
    document.getElementById('ap-ok').onclick = async () => {
      const name = document.getElementById('ap-name').value.trim();
      if (!name) return toast('请输入应用名称','error');
      try { const r = await api('/app-passwords',{method:'POST',body:{app_name:name}}); s.remove(); showAppPassword(name, r.password); } catch(e) { toast(e.message,'error'); }
    };
  };

  async function load() {
    try {
      const list = await api('/app-passwords');
      const body = document.getElementById('ap-body');
      const pwMap = {};
      if (list && list.length) {
        list.forEach(a => pwMap[a.id] = a.password || '');
        body.innerHTML = list.map(a => {
          const hasPw = !!a.password;
          const dots = '•'.repeat(hasPw ? (a.password||'').length : 12);
          return `<div class="file-row">
          <div class="ficon" style="background:var(--bg-soft);color:var(--text-2)">${Icon.key.replace('class="','class="w-5 h-5 ')}</div>
          <div class="flex-1 min-w-0">
            <div class="text-[14px] font-semibold truncate">${esc(a.app_name)}</div>
            <div class="text-[12px] mt-0.5" style="color:var(--text-3)">${fmtDateTime(a.created_at)}</div>
          </div>
          <code id="pw-${a.id}" class="app-pw-key shrink-0 ${hasPw?'':'opacity-50'}" style="letter-spacing:.12em;color:var(--text-2)">${hasPw?dots:'未保存'}</code>
          <button data-eye data-id="${a.id}" class="icon-btn" style="width:2.3rem;height:2.3rem" title="显示/隐藏秘钥">${Icon.eye.replace('class="','class="w-5 h-5 ')}</button>
          <button data-copy data-id="${a.id}" class="icon-btn good" style="width:2.3rem;height:2.3rem" title="复制秘钥">${Icon.copy.replace('class="','class="w-5 h-5 ')}</button>
          <button data-del data-id="${a.id}" class="icon-btn danger" style="width:2.3rem;height:2.3rem" title="撤销">${Icon.trash.replace('class="','class="w-5 h-5 ')}</button>
        </div>`;
        }).join('');
        body.addEventListener('click', async (e) => {
          const eye = e.target.closest('[data-eye]');
          const copy = e.target.closest('[data-copy]');
          const d = e.target.closest('[data-del]');
          if (eye) {
            const el = document.getElementById('pw-'+eye.dataset.id);
            const pw = pwMap[eye.dataset.id];
            if (!pw) { toast('该应用密码未保存明文，无法查看，请删除后重新生成','error'); return; }
            if (el.dataset.on === '1') { el.textContent = '•'.repeat((pw||'').length||12); el.dataset.on = '0'; eye.innerHTML = Icon.eye.replace('class="','class="w-5 h-5 '); }
            else { el.textContent = pw; el.dataset.on = '1'; eye.innerHTML = Icon.eyeOff.replace('class="','class="w-5 h-5 '); }
            return;
          }
          if (copy) {
            const pw = pwMap[copy.dataset.id];
            if (!pw) { toast('该应用密码未保存明文，无法复制，请删除后重新生成','error'); return; }
            copyText(pw, '已复制秘钥'); return;
          }
          if (d) {
            if (await iConfirm('撤销授权','撤销此应用密码后，该客户端将无法连接。','撤销',true)) {
              try { await api('/app-passwords',{method:'DELETE',body:{id:Number(d.dataset.id)}}); toast('已撤销','ok'); load(); } catch(err) { toast(err.message,'error'); }
            }
          }
        });
      } else {
        body.innerHTML = '<div class="empty card fade-in"><div class="ico">'+Icon.key.replace('class="','class="w-7 h-7 ')+'</div><div class="font-semibold">还没有授权应用</div><div class="text-[13px] mt-1" style="color:var(--text-3)">点击「添加应用」为 WebDAV 客户端生成专属密码</div></div>';
      }
    } catch(e) { document.getElementById('ap-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  load();
}

function showAppPassword(name, pwd) {
  openSheet(`<div class="space-y-4 mt-1">
    <p class="text-[13px]" style="color:var(--text-2)">应用 <span class="font-semibold" style="color:var(--text)">${esc(name)}</span> 的应用密码已生成：</p>
    <div class="flex items-center gap-2"><code class="input !bg-transparent text-center font-mono !tracking-widest" id="ap-show" style="color:var(--text)">${esc(pwd)}</code><button id="ap-copy" class="icon-btn grad-icon shrink-0" style="border-radius:.85rem;width:2.7rem;height:2.7rem"><svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg></button></div>
    <button id="ap-done" class="w-full inline-flex items-center justify-center py-3 px-4 rounded-xl font-semibold text-white" style="background:linear-gradient(135deg,var(--brand),var(--brand-2))">完成</button>
    <p class="text-[12px] text-center" style="color:var(--danger)">请妥善保存，此密码仅显示一次</p>
  </div>`, '应用密码已生成');
  document.getElementById('ap-copy').onclick = () => copyText(pwd, '已复制','复制失败');
  document.getElementById('ap-done').onclick = () => { closeAll(); renderUsers(); };
}

/* settings */
async function renderSettings() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('设置','','') + `<div class="px-5 md:px-8 py-6 space-y-5">
    <section class="card p-5 space-y-3"><div class="sec-title" style="margin-bottom:.6rem">个人资料</div>
      <div class="field"><label>显示名称</label><input id="st-name" class="input" /></div>
      <div class="field"><label>邮箱</label><input id="st-email" class="input" /></div>
      <button id="st-save-profile" class="btn-primary !py-2.5">保存资料</button>
    </section>
    <section class="card p-5 space-y-3"><div class="sec-title" style="margin-bottom:.6rem">修改密码</div>
      <div class="field"><label>当前密码</label><input id="st-old" class="input" type="password" /></div>
      <div class="field"><label>新密码</label><input id="st-new" class="input" type="password" /></div>
      <button id="st-save-pass" class="btn-primary !py-2.5">更新密码</button>
    </section>
    <section class="card p-5 space-y-3"><div class="flex items-center justify-between"><span class="font-semibold" style="font-size:.95rem">两步验证 (TOTP)</span><span id="totp-state" class="chip"></span></div><div id="totp-actions"></div></section>
    <section class="card p-5 space-y-3"><div class="sec-title" style="margin-bottom:.6rem">外观</div><div class="flex items-center justify-between"><div><div class="text-[14px] font-semibold">深色模式</div><div class="text-[12px]" style="color:var(--text-3)">切换浅色 / 深色外观</div></div><button id="toggle-dark" class="btn-ghost hairline">${Icon.theme.replace('class="','class="w-5 h-5 ')} 切换</button></div></section>
    <section class="card p-5 space-y-3"><div class="sec-title" style="margin-bottom:.6rem">服务器配置</div>
      <div class="field"><label>隐藏路径（glob，逗号分隔）</label><input id="st-ignore" class="input" /></div>
      <button id="st-save-srv" class="btn-primary !py-2.5">保存配置</button>
    </section>
    <section class="card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10">${LOGO}</div><div><div class="font-bold">${esc(BRAND_NAME)}</div><div class="text-[12px] mt-0.5" style="color:var(--text-3)">v1.1.0 · Go + Tailwind · 网盘服务器</div><a href="https://github.com/vipxkw/ChinaSCLMDAV" target="_blank" rel="noopener" class="inline-flex items-center gap-1 text-[12px] mt-0.5" style="color:var(--brand)">${Icon.github.replace('class="','class="w-4 h-4 ')} github.com/vipxkw/ChinaSCLMDAV</a></div></div></section>
  </div>`);

  const u = state.user;
  document.getElementById('st-name').value = u.display_name||'';
  document.getElementById('st-email').value = u.email||'';
  document.getElementById('st-save-profile').onclick = async () => {
    try { await api('/profile',{method:'POST',body:{display_name:document.getElementById('st-name').value.trim(),email:document.getElementById('st-email').value.trim()}}); toast('已保存','ok'); state.user.display_name = document.getElementById('st-name').value.trim(); state.user.email = document.getElementById('st-email').value.trim(); } catch(e) { toast(e.message,'error'); }
  };
  document.getElementById('st-save-pass').onclick = async () => {
    try { await api('/password',{method:'POST',body:{old_password:document.getElementById('st-old').value,new_password:document.getElementById('st-new').value}}); document.getElementById('st-old').value=''; document.getElementById('st-new').value=''; toast('密码已更新','ok'); } catch(e) { toast(e.message,'error'); }
  };
  document.getElementById('toggle-dark').onclick = toggleDark;
  renderTotpSection();
  try { const s = await api('/settings'); document.getElementById('st-ignore').value = s.ignore||''; } catch(e) {}
  document.getElementById('st-save-srv').onclick = async () => {
    try { await api('/settings',{method:'PUT',body:{ignore:document.getElementById('st-ignore').value}}); toast('已保存','ok'); } catch(e) { toast(e.message,'error'); }
  };
}

function renderTotpSection() {
  const u = state.user;
  const stateEl = document.getElementById('totp-state');
  if (u.totp_enabled) { stateEl.textContent = '已启用'; stateEl.className = 'chip badge-green'; }
  else { stateEl.textContent = '未启用'; stateEl.className = 'chip'; }
  document.getElementById('totp-actions').innerHTML = u.totp_enabled
    ? `<button id="totp-disable" class="btn-danger hairline !py-2.5">
         <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="9"/></svg> 禁用两步验证</button><div class="text-[12px] mt-2" style="color:var(--text-3)">两步验证已启用，登录时需要验证码。</div>`
    : `<button id="totp-enable" class="btn-primary !py-2.5">启用两步验证</button>`;
  const enable = document.getElementById('totp-enable');
  if (enable) enable.onclick = () => {
    (async () => {
      try {
        const r = await api('/totp/secret');
        const s = openSheet(`<div class="space-y-4 mt-1">
          <div class="field"><label>1. 在认证器中添加以下密钥</label><div class="flex items-center gap-2"><code class="input !bg-transparent text-center font-mono" id="t-secret">${esc(r.secret)}</code><button id="t-copy" class="icon-btn grad-icon shrink-0" style="border-radius:.85rem;width:2.7rem;height:2.7rem">${Icon.copy.replace('class="','class="w-5 h-5 ')}</button></div></div>
          <div class="field"><label>2. 输入验证器生成的 6 位验证码</label><input id="t-code" class="input text-center tracking-[0.4em]" inputmode="numeric" maxlength="6" placeholder="000000" /></div>
          <button id="t-ok" class="btn-primary w-full !py-3">启用</button>
        </div>`, '启用两步验证');
        document.getElementById('t-copy').onclick = () => navigator.clipboard.writeText(r.secret).then(()=>toast('已复制','ok'));
        document.getElementById('t-ok').onclick = async () => {
          try { await api('/totp/enable',{method:'POST',body:{code:document.getElementById('t-code').value.trim()}}); s.remove(); state.user.totp_enabled = true; toast('已启用两步验证','ok'); renderTotpSection(); } catch(e) { toast(e.message,'error'); }
        };
      } catch(e) { toast(e.message,'error'); }
    })();
  };
  const disable = document.getElementById('totp-disable');
  if (disable) disable.onclick = () => {
    const s = openSheet(`<div class="space-y-4 mt-1"><div class="field"><label>输入当前验证码以禁用</label><input id="t-code" class="input text-center tracking-[0.4em]" inputmode="numeric" maxlength="6" placeholder="000000" /></div><button id="t-ok" class="w-full inline-flex items-center justify-center py-3 rounded-xl font-semibold text-white" style="background:var(--danger)">禁用</button></div>`, '禁用两步验证');
    document.getElementById('t-ok').onclick = async () => {
      try { await api('/totp/disable',{method:'POST',body:{code:document.getElementById('t-code').value.trim()}}); s.remove(); state.user.totp_enabled = false; toast('已禁用','ok'); renderTotpSection(); } catch(e) { toast(e.message,'error'); }
    };
  };
}

/* audit log */
const AUDIT_CATS = [
  { key:'all',      label:'全部' },
  { key:'file',     label:'文件操作' },
  { key:'share',    label:'分享' },
  { key:'trash',    label:'回收站' },
  { key:'security', label:'账号安全' },
  { key:'system',   label:'系统设置' },
];
const AUDIT_RULES = {
  file:     ['mkdir','delete','rename','upload','version_delete'],
  share:    ['share_create','share_delete'],
  trash:    ['trash_restore','trash_purge','trash_empty'],
  security: ['login','logout','password_change','profile_update','totp_enable','totp_disable','totp_policy','app_password_create','app_password_delete'],
  system:   ['settings'],
};
const AUDIT_LABELS = {
  login:'登录', logout:'退出登录', profile_update:'更新资料', password_change:'修改密码',
  totp_enable:'启用两步验证', totp_disable:'禁用两步验证', totp_policy:'TOTP 策略',
  settings:'更新设置', mkdir:'新建文件夹', delete:'移入回收站', rename:'重命名', upload:'上传文件',
  share_create:'创建分享', share_delete:'撤销分享', trash_restore:'恢复文件', trash_purge:'彻底删除', trash_empty:'清空回收站',
  version_delete:'删除版本', app_password_create:'创建应用密码', app_password_delete:'撤销应用密码',
};
const AUDIT_META = {
  file:     { c:'#2a6df4', e:'📁' },
  share:    { c:'#22c55e', e:'🔗' },
  trash:    { c:'#ff9500', e:'🗑' },
  security: { c:'#ff4d6d', e:'🔐' },
  system:   { c:'#9b5cff', e:'⚙️' },
};
function auditCat(action) {
  for (const c of AUDIT_CATS) {
    if (c.key === 'all') continue;
    if ((AUDIT_RULES[c.key]||[]).indexOf(action) >= 0) return c.key;
  }
  return 'system';
}

let _auditAll = [];
let _auditCat = 'all';
let _auditPage = 1;
const AUDIT_PAGE_SIZE = 10;

async function renderAudit() {
  const app = document.getElementById('app');
  app.innerHTML = appShell(topBar('审计日志','记录全部安全与操作事件',
    actionBtn(Icon.refresh,'data-refresh','刷新','btn-ghost') +
    `<button data-clear class="btn-ghost" style="color:var(--danger)">${Icon.trash.replace('class="','class="w-5 h-5 ')}<span class="hidden sm:inline">清空</span></button>`
  ) + `<div class="px-5 md:px-8 py-6">
    <div class="flex items-center justify-between gap-3 flex-wrap mb-4">
      <div id="audit-tabs" class="flex items-center gap-1.5 overflow-x-auto no-scrollbar">${AUDIT_CATS.map(c => `<button data-cat="${c.key}" class="audit-tab ${_auditCat===c.key?'on':''}">${esc(c.label)}</button>`).join('')}</div>
      <span id="audit-count" class="text-[12px] shrink-0" style="color:var(--text-3)"></span>
    </div>
    <div id="audit-body" class="space-y-2">${SPINNER}</div>
    <div id="audit-pager" class="flex items-center justify-center gap-3 mt-6"></div>
  </div>`);

  app.querySelector('[data-refresh]').onclick = load;
  app.querySelector('[data-clear]').onclick = async () => {
    if (!(await iConfirm('清空日志','将删除全部审计记录，且无法恢复。','清空',true))) return;
    try { await api('/audit',{method:'DELETE'}); _auditAll = []; _auditPage = 1; renderList(); toast('已清空','ok'); } catch(e) { toast(e.message,'error'); }
  };
  document.getElementById('audit-tabs').addEventListener('click', e => {
    const b = e.target.closest('[data-cat]'); if (!b) return;
    _auditCat = b.dataset.cat; _auditPage = 1;
    document.getElementById('audit-tabs').innerHTML = AUDIT_CATS.map(c => `<button data-cat="${c.key}" class="audit-tab ${_auditCat===c.key?'on':''}">${esc(c.label)}</button>`).join('');
    renderList();
  });

  function catLabel(key) { const c = AUDIT_CATS.find(x => x.key === key); return c ? c.label : key; }

  function renderList() {
    const body = document.getElementById('audit-body');
    const rows = (_auditAll||[]).filter(a => _auditCat === 'all' || auditCat(a.action) === _auditCat);
    const totalPages = Math.max(1, Math.ceil(rows.length / AUDIT_PAGE_SIZE));
    if (_auditPage > totalPages) _auditPage = totalPages;
    document.getElementById('audit-count').textContent = rows.length ? `共 ${rows.length} 条` : '';
    if (!rows.length) {
      body.innerHTML = `<div class="empty card fade-in"><div class="ico">${Icon.clock.replace('class="','class="w-7 h-7 ')}</div><div class="font-semibold">暂无审计记录</div><div class="text-[13px] mt-1" style="color:var(--text-3)">${_auditCat==='all'?'进行文件或账号操作后，这里会显示记录':'当前分类下暂无记录'}</div></div>`;
      document.getElementById('audit-pager').innerHTML = '';
      return;
    }
    const start = (_auditPage - 1) * AUDIT_PAGE_SIZE;
    const items = rows.slice(start, start + AUDIT_PAGE_SIZE);
    body.innerHTML = '<div class="space-y-2">' + items.map(a => {
      const cat = auditCat(a.action);
      const meta = AUDIT_META[cat] || AUDIT_META.system;
      return `<div class="file-row fade-in">
        <div class="ficon" style="background:${meta.c}1a">${meta.e}</div>
        <div class="flex-1 min-w-0"><div class="text-[14px] font-semibold truncate">${esc(AUDIT_LABELS[a.action] || a.action)}<span class="chip" style="margin-left:.5rem;background:${meta.c}14;color:${meta.c}">${esc(catLabel(cat))}</span></div>${a.detail ? `<div class="text-[12px] mt-0.5 truncate" style="color:var(--text-3)">${esc(a.detail)}</div>` : ''}</div>
        <div class="text-[11px] shrink-0" style="color:var(--text-3)">${fmtDateTime(a.created_at)}</div>
        <button data-del data-id="${a.id}" class="icon-btn danger" style="width:2.3rem;height:2.3rem" title="删除记录">${Icon.trash.replace('class="','class="w-4 h-4 ')}</button>
      </div>`;
    }).join('') + '</div>';

    const pager = document.getElementById('audit-pager');
    pager.innerHTML = `<button data-pg="prev" class="btn-ghost hairline" ${_auditPage<=1?'disabled':''}><svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m15 18-6-6 6-6"/></svg><span class="hidden sm:inline">上一页</span></button>
      <span class="text-[13px] font-semibold" style="color:var(--text-2)">${_auditPage} / ${totalPages}</span>
      <button data-pg="next" class="btn-ghost hairline" ${_auditPage>=totalPages?'disabled':''}><span class="hidden sm:inline">下一页</span><svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m9 18 6-6-6-6"/></svg></button>`;
    pager.onclick = e => {
      const b = e.target.closest('[data-pg]'); if (!b) return;
      if (b.dataset.pg === 'prev' && _auditPage > 1) { _auditPage--; renderList(); }
      if (b.dataset.pg === 'next' && _auditPage < totalPages) { _auditPage++; renderList(); }
    };
    body.addEventListener('click', async (e) => {
      const d = e.target.closest('[data-del]'); if (!d) return;
      if (!(await iConfirm('删除记录','确定删除这条审计记录吗？','删除',true))) return;
      try { await api('/audit?id='+d.dataset.id,{method:'DELETE'}); _auditAll = _auditAll.filter(x => String(x.id) !== d.dataset.id); renderList(); toast('已删除','ok'); } catch(err) { toast(err.message,'error'); }
    });
  }

  async function load() {
    try {
      _auditAll = await api('/audit');
      if (!_auditAll) _auditAll = [];
      renderList();
    } catch(e) { document.getElementById('audit-body').innerHTML = `<div class="p-4 rounded-xl text-sm font-medium" style="background:rgba(255,77,95,.1);color:var(--danger)">${esc(e.message)}</div>`; }
  }
  load();
}

/* dark mode */
function toggleDark() {
  const dark = !document.documentElement.classList.contains('dark');
  document.documentElement.classList.toggle('dark', dark);
  localStorage.setItem('cscd_dark', dark ? '1' : '0');
  toast(dark ? '已切换深色模式' : '已切换浅色模式');
}

function initTheme() {
  const dark = localStorage.getItem('cscd_dark') === '1';
  document.documentElement.classList.toggle('dark', dark);
}

/* router */
async function route() {
  const hash = location.hash || '#/home';
  state.route = hash.slice(2) || 'home';
  if (!state.user && state.route !== 'login') {
    try { state.user = await api('/me'); } catch(e) { state.route = 'login'; }
  }
  if (state.route === 'login') { renderLogin(); return; }
  if (!window.__logoutBound) {
    window.__logoutBound = true;
    document.addEventListener('click', (e) => {
      const b = e.target.closest('[data-logout]');
      if (b) { e.preventDefault(); doLogout(); }
    });
  }
  switch (state.route) {
    case 'files': return renderFiles();
    case 'shared': return renderShared();
    case 'trash': return renderTrash();
    case 'audit': return renderAudit();
    case 'users': return renderUsers();
    case 'settings': return renderSettings();
    default: return renderHome();
  }
}

async function doLogout() {
  try { await api('/logout',{method:'POST'}); } catch(e) {}
  state.user = null;
  location.hash = '#/login';
  route();
}

/* init */
window.addEventListener('hashchange', route);
window.addEventListener('DOMContentLoaded', () => { initTheme(); route(); });
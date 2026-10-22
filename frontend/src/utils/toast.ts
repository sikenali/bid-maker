type ToastType = 'success' | 'error' | 'warning' | 'info';

function ensureContainer() {
  let c = document.getElementById('__toast_container__');
  if (!c) {
    c = document.createElement('div');
    c.id = '__toast_container__';
    c.style.cssText = 'position:fixed;top:20px;left:50%;transform:translateX(-50%);z-index:10000;display:flex;flex-direction:column;align-items:center;gap:10px;pointer-events:none;';
    document.body.appendChild(c);
  }
  return c;
}

function showToast(message: string, type: ToastType = 'info', duration = 2500) {
  const c = ensureContainer();
  const el = document.createElement('div');
  el.style.cssText = `
    pointer-events:auto;
    display:flex;align-items:center;gap:10px;
    padding:12px 16px 12px 18px;
    border-radius:var(--radius-md, 10px);
    border:0.7px solid var(--color-border, #E5DFD3);
    box-shadow:var(--shadow-md, 0 4px 12px rgba(0,0,0,0.12));
    font-size:14px;font-weight:500;line-height:1.45;
    min-width:200px;max-width:420px;
    color:var(--color-text-body, #3A3A3A);
    background:var(--color-bg-white, #FFFFFF);
    animation:toast-in 0.25s ease;
  `;
  const accentColors: Record<ToastType, string> = {
    success: 'var(--color-success, #4E9B7A)',
    error:   'var(--color-error, #E74C3C)',
    warning: 'var(--color-warning, #F39C12)',
    info:    'var(--color-info, #4A72A8)',
  };
  el.innerHTML = `
    <span style="flex-shrink:0;width:3px;height:20px;border-radius:2px;background:${accentColors[type]};"></span>
    <span class="toast-msg">${message.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span>
    <button class="toast-close-btn" style="flex-shrink:0;width:20px;height:20px;border:none;border-radius:9999px;background:transparent;color:var(--color-text-muted);cursor:pointer;display:flex;align-items:center;justify-content:center;" aria-label="关闭">
      <span style="font-family:remixicon;font-style:normal;font-size:12px;">&#xea4c;</span>
    </button>
  `;
  const closeBtn = el.querySelector('.toast-close-btn');
  if (closeBtn) {
    closeBtn.addEventListener('click', () => { dismiss(el); });
  }
  c.appendChild(el);
  setTimeout(() => dismiss(el), duration);
}

function dismiss(el: HTMLElement) {
  el.style.transition = 'opacity 0.25s ease, transform 0.25s ease';
  el.style.opacity = '0';
  el.style.transform = 'translateY(-8px)';
  setTimeout(() => el.remove(), 250);
}

if (!document.getElementById('__toast_keyframes__')) {
  const style = document.createElement('style');
  style.id = '__toast_keyframes__';
  style.textContent = `@keyframes toast-in{from{opacity:0;transform:translateY(-12px)}to{opacity:1;transform:translateY(0)}}`;
  document.head.appendChild(style);
}

export const ElMessage = {
  success: (msg: string) => showToast(msg, 'success'),
  error:   (msg: string) => showToast(msg, 'error'),
  warning: (msg: string) => showToast(msg, 'warning'),
  info:    (msg: string) => showToast(msg, 'info', 2000),
};

export const ElMessageBox = {
  async confirm(message: string, title: string = '确认'): Promise<void> {
    return new Promise((resolve, reject) => {
      const overlay = document.createElement('div');
      overlay.style.cssText = 'position:fixed;inset:0;z-index:10001;background:rgba(0,0,0,0.4);display:flex;align-items:center;justify-content:center;';
      const box = document.createElement('div');
      box.style.cssText = 'background:var(--color-bg-white, #fff);border-radius:12px;padding:24px;max-width:420px;width:90%;box-shadow:0 20px 60px rgba(0,0,0,0.2);';
      box.innerHTML = `
        <div style="font-size:16px;font-weight:500;color:var(--color-text-primary, #1F1D1A);margin-bottom:12px;">${title}</div>
        <div style="font-size:14px;color:var(--color-text-body, #3A3A3A);line-height:1.6;margin-bottom:24px;">${message}</div>
        <div style="display:flex;gap:12px;justify-content:flex-end;">
          <button id="__cancel_btn" style="padding:8px 20px;border:1px solid var(--color-border, #E5DFD3);background:var(--color-bg-white, #fff);color:var(--color-text-body, #3A3A3A);border-radius:8px;font-size:14px;cursor:pointer;">取消</button>
          <button id="__confirm_btn" style="padding:8px 20px;border:none;background:var(--color-primary, #BE3B2E);color:#fff;border-radius:8px;font-size:14px;cursor:pointer;">确认</button>
        </div>`;
      overlay.appendChild(box);
      document.body.appendChild(overlay);
      box.querySelector('#__cancel_btn')!.addEventListener('click', () => { overlay.remove(); reject(new Error('cancel')); });
      box.querySelector('#__confirm_btn')!.addEventListener('click', () => { overlay.remove(); resolve(); });
      overlay.addEventListener('click', (e) => { if (e.target === overlay) { overlay.remove(); reject(new Error('cancel')); } });
    });
  },
};

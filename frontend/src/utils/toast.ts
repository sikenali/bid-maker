const toastContainer = () => {
  let c = document.getElementById('__toast_container__');
  if (!c) {
    c = document.createElement('div');
    c.id = '__toast_container__';
    c.style.cssText = 'position:fixed;top:20px;right:20px;z-index:10000;display:flex;flex-direction:column;gap:10px;align-items:flex-end;pointer-events:none;';
    document.body.appendChild(c);
  }
  return c;
};

const COLORS: Record<string, { bg: string; color: string; border: string }> = {
  success: { bg: '#F0FDF4', color: '#15803D', border: '#BBF7D0' },
  error:   { bg: '#FEF2F2', color: '#B91C1C', border: '#FECACA' },
  warning: { bg: '#FFFBEB', color: '#B45309', border: '#FDE68A' },
  info:    { bg: '#F0F9FF', color: '#0369A1', border: '#BAE6FD' },
};

function show(message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') {
  const c = COLORS[type] || COLORS.info;
  const el = document.createElement('div');
  el.style.cssText = `pointer-events:auto;display:flex;align-items:center;gap:8px;padding:12px 20px;border-radius:8px;background:${c.bg};color:${c.color};border:1px solid ${c.border};font-size:14px;font-family:"SourceHanSans-Regular","Noto Sans SC",sans-serif;box-shadow:0 4px 12px rgba(0,0,0,0.1);animation:toast-in 0.3s ease;max-width:400px;word-break:break-word;`;
  el.innerHTML = `<span>${message.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span>`;
  toastContainer().appendChild(el);
  setTimeout(() => {
    el.style.opacity = '0';
    el.style.transform = 'translateX(20px)';
    el.style.transition = 'opacity 0.3s, transform 0.3s';
    setTimeout(() => el.remove(), 300);
  }, 3000);
}

export const ElMessage = {
  success: (msg: string) => show(msg, 'success'),
  error:   (msg: string) => show(msg, 'error'),
  warning: (msg: string) => show(msg, 'warning'),
  info:    (msg: string) => show(msg, 'info'),
};

export const ElMessageBox = {
  async confirm(message: string, title: string = '确认'): Promise<any> {
    return new Promise((resolve, reject) => {
      const overlay = document.createElement('div');
      overlay.style.cssText = 'position:fixed;inset:0;z-index:10001;background:rgba(0,0,0,0.4);display:flex;align-items:center;justify-content:center;';
      const box = document.createElement('div');
      box.style.cssText = 'background:#fff;border-radius:12px;padding:24px;max-width:420px;width:90%;box-shadow:0 20px 60px rgba(0,0,0,0.2);font-family:"SourceHanSans-Regular","Noto Sans SC",sans-serif;';
      box.innerHTML = `
        <div style="font-size:16px;font-weight:500;color:#1F1D1A;margin-bottom:12px;">${title}</div>
        <div style="font-size:14px;color:#3A3A3A;line-height:1.6;margin-bottom:24px;">${message}</div>
        <div style="display:flex;gap:12px;justify-content:flex-end;">
          <button id="__cancel_btn" style="padding:8px 20px;border:1px solid #E5DFD3;background:#fff;color:#3A3A3A;border-radius:8px;font-size:14px;cursor:pointer;">取消</button>
          <button id="__confirm_btn" style="padding:8px 20px;border:none;background:#BE3B2E;color:#fff;border-radius:8px;font-size:14px;cursor:pointer;">确认</button>
        </div>`;
      overlay.appendChild(box);
      document.body.appendChild(overlay);
      box.querySelector('#__cancel_btn')!.addEventListener('click', () => { overlay.remove(); reject(new Error('cancel')); });
      box.querySelector('#__confirm_btn')!.addEventListener('click', () => { overlay.remove(); resolve('confirm'); });
      overlay.addEventListener('click', (e) => { if (e.target === overlay) { overlay.remove(); reject(new Error('cancel')); } });
    });
  },
};

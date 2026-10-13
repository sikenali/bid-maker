import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getTheme as getThemeFromAPI, setTheme as setThemeToAPI } from '@/api/settings';

export type ThemeType = 'parchment' | 'dark' | 'white';

const ACCENT_KEY = 'accent-color';

const themeTokens: Record<ThemeType, Record<string, string>> = {
  parchment: {
    '--color-primary': '#C0392B',
    '--color-primary-soft': '#FBE9E7',
    '--color-bg-main': '#FAF7F2',
    '--color-bg-secondary': '#F7F3EA',
    '--color-bg-card': '#FCFBF8',
    '--color-bg-white': '#FFFFFF',
    '--color-text-primary': '#1F1D1A',
    '--color-text-body': '#3A3A3A',
    '--color-text-secondary': '#6E6A60',
    '--color-text-muted': '#8A8578',
    '--color-text-faint': '#A7A296',
    '--color-border': '#E5DFD3',
    '--color-border-soft': '#EDE9E0',
    '--color-border-hover': '#D9D2C2',
    '--color-success': '#4E9B7A',
    '--color-success-strong': '#3E8A6B',
    '--color-success-soft': '#E4F0EA',
    '--color-info': '#4A72A8',
    '--color-info-soft': '#E8F0F8',
    '--color-warning': '#F39C12',
    '--color-warning-soft': '#FEF5E7',
    '--color-error': '#E74C3C',
    '--color-error-soft': '#FADCDA',
  },
  dark: {
    '--color-primary': '#E05050',
    '--color-primary-soft': '#3A2323',
    '--color-bg-main': '#17171A',
    '--color-bg-secondary': '#1C1C1F',
    '--color-bg-card': '#232324',
    '--color-bg-white': '#232324',
    '--color-text-primary': '#F2F3F5',
    '--color-text-body': '#DDE0E6',
    '--color-text-secondary': '#C9CDD4',
    '--color-text-muted': '#A7AEB9',
    '--color-text-faint': '#86909C',
    '--color-border': '#2A2B2E',
    '--color-border-soft': '#2A2B2E',
    '--color-border-hover': '#3A3B3F',
    '--color-success': '#5CB98C',
    '--color-success-strong': '#7BC8A4',
    '--color-success-soft': '#1D3329',
    '--color-info': '#82AEDD',
    '--color-info-soft': '#1B2A3A',
    '--color-warning': '#F2B23E',
    '--color-warning-soft': '#38300F',
    '--color-error': '#E86A5A',
    '--color-error-soft': '#3A2326',
  },
  white: {
    '--color-primary': '#C0392B',
    '--color-primary-soft': '#FBE9E7',
    '--color-bg-main': '#FFFFFF',
    '--color-bg-secondary': '#FAFAFA',
    '--color-bg-card': '#F5F5F5',
    '--color-bg-white': '#FFFFFF',
    '--color-text-primary': '#1A1A1A',
    '--color-text-body': '#3A3A3A',
    '--color-text-secondary': '#666666',
    '--color-text-muted': '#8A8578',
    '--color-text-faint': '#A7A296',
    '--color-border': '#E0E0E0',
    '--color-border-soft': '#E0E0E0',
    '--color-border-hover': '#D9D2C2',
    '--color-success': '#4E9B7A',
    '--color-success-strong': '#3E8A6B',
    '--color-success-soft': '#E4F0EA',
    '--color-info': '#4A72A8',
    '--color-info-soft': '#E8F0F8',
    '--color-warning': '#F39C12',
    '--color-warning-soft': '#FEF5E7',
    '--color-error': '#E74C3C',
    '--color-error-soft': '#FADCDA',
  },
};

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref<ThemeType>('parchment');
  const currentAccent = ref<string | null>(null);

  function applyAccent(color: string | null) {
    const root = document.documentElement;
    if (color) {
      root.style.setProperty('--color-primary', color);
      localStorage.setItem(ACCENT_KEY, color);
    } else {
      root.style.removeProperty('--color-primary');
      localStorage.removeItem(ACCENT_KEY);
    }
    currentAccent.value = color;
  }

  async function loadTheme() {
    try {
      const res = await getThemeFromAPI();
      const saved = res.type as ThemeType | null;
      if (saved && themeTokens[saved]) {
        currentTheme.value = saved;
        applyTheme(saved);
        localStorage.setItem('theme', saved);
        return;
      }
    } catch {
      // API unavailable, fall through to localStorage
    }
    const saved = localStorage.getItem('theme') as ThemeType | null;
    if (saved && themeTokens[saved]) {
      currentTheme.value = saved;
      applyTheme(saved);
    } else {
      applyTheme(currentTheme.value);
    }
  }

  function applyTheme(theme: ThemeType) {
    const root = document.documentElement;
    root.dataset.theme = theme;
    const tokens = themeTokens[theme];
    for (const [key, value] of Object.entries(tokens)) {
      root.style.setProperty(key, value);
    }
    // 强调色覆盖主题默认 primary（持久化，切换主题不丢失）
    const accent = localStorage.getItem(ACCENT_KEY);
    if (accent) {
      root.style.setProperty('--color-primary', accent);
      currentAccent.value = accent;
    } else {
      currentAccent.value = null;
    }
  }

  async function setTheme(theme: ThemeType) {
    currentTheme.value = theme;
    applyTheme(theme);
    localStorage.setItem('theme', theme);
    try {
      await setThemeToAPI(theme);
    } catch {
      // Backend is optional, localStorage is the primary persistence
    }
  }

  return { currentTheme, currentAccent, loadTheme, setTheme, applyAccent };
});

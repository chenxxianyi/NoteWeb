import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import * as settingsApi from '../api/settings'

export type ThemeType = 'warm' | 'blue' | 'dark'

export interface Settings {
  theme: ThemeType
  font: string
  readingMode: boolean
}

const THEME_COLORS: Record<ThemeType, string> = {
  warm: '#C67A4E',
  blue: '#2563EB',
  dark: '#1E293B',
}

const DEFAULT_SETTINGS: Settings = {
  theme: 'warm',
  font: 'Noto Serif SC',
  readingMode: false,
}

function loadSettings(): Settings {
  try {
    const stored = localStorage.getItem('noteweb_settings')
    if (stored) {
      const parsed = JSON.parse(stored)
      return {
        theme: parsed.theme || DEFAULT_SETTINGS.theme,
        font: parsed.font || DEFAULT_SETTINGS.font,
        readingMode: parsed.readingMode ?? DEFAULT_SETTINGS.readingMode,
      }
    }
  } catch {
    // ignore parsing errors
  }
  return DEFAULT_SETTINGS
}

function applyTheme(theme: ThemeType) {
  const color = THEME_COLORS[theme]
  document.documentElement.style.setProperty('--accent', color)
  // Adjust accent-light based on theme
  if (theme === 'dark') {
    document.documentElement.style.setProperty('--accent-light', 'rgba(30, 41, 59, 0.12)')
  } else if (theme === 'blue') {
    document.documentElement.style.setProperty('--accent-light', 'rgba(37, 99, 235, 0.12)')
  } else {
    document.documentElement.style.setProperty('--accent-light', 'rgba(198, 122, 78, 0.12)')
  }
}

function applyFont(font: string) {
  document.documentElement.style.setProperty('--font-body', font)
}

function applyReadingMode(enabled: boolean) {
  if (enabled) {
    document.documentElement.classList.add('reading-mode')
  } else {
    document.documentElement.classList.remove('reading-mode')
  }
}

export const useSettingsStore = defineStore('settings', () => {
  const theme = ref<ThemeType>(DEFAULT_SETTINGS.theme)
  const font = ref<string>(DEFAULT_SETTINGS.font)
  const readingMode = ref<boolean>(DEFAULT_SETTINGS.readingMode)
  const loading = ref(false)
  const syncEnabled = ref(true) // Flag to control backend sync

  // Load settings from localStorage on init
  const initial = loadSettings()
  theme.value = initial.theme
  font.value = initial.font
  readingMode.value = initial.readingMode

  // Apply initial settings
  applyTheme(theme.value)
  applyFont(font.value)
  applyReadingMode(readingMode.value)

  // Watch for changes and persist + apply
  watch([theme, font, readingMode], () => {
    const settings: Settings = {
      theme: theme.value,
      font: font.value,
      readingMode: readingMode.value,
    }
    localStorage.setItem('noteweb_settings', JSON.stringify(settings))
    applyTheme(theme.value)
    applyFont(font.value)
    applyReadingMode(readingMode.value)
  })

  // Fetch settings from backend
  async function fetchSettings() {
    loading.value = true
    try {
      const res = await settingsApi.getSettings()
      if (res.data) {
        theme.value = res.data.theme as ThemeType
        font.value = res.data.font
        readingMode.value = res.data.reading_mode
        // Update localStorage
        localStorage.setItem('noteweb_settings', JSON.stringify({
          theme: theme.value,
          font: font.value,
          readingMode: readingMode.value,
        }))
        applyTheme(theme.value)
        applyFont(font.value)
        applyReadingMode(readingMode.value)
      }
    } catch (e) {
      // Failed to fetch from backend, use localStorage defaults
      console.log('Using local settings (backend not available)')
    } finally {
      loading.value = false
    }
  }

  // Sync settings to backend
  async function syncToBackend() {
    if (!syncEnabled.value) return
    try {
      await settingsApi.updateSettings({
        theme: theme.value,
        font: font.value,
        reading_mode: readingMode.value,
      })
    } catch (e) {
      // Failed to sync, settings are still saved locally
      console.log('Settings saved locally (backend sync failed)')
    }
  }

  function setTheme(newTheme: ThemeType) {
    theme.value = newTheme
    syncToBackend()
  }

  function setFont(newFont: string) {
    font.value = newFont
    syncToBackend()
  }

  function setReadingMode(enabled: boolean) {
    readingMode.value = enabled
    syncToBackend()
  }

  const themeColor = computed(() => THEME_COLORS[theme.value])

  return {
    theme,
    font,
    readingMode,
    loading,
    themeColor,
    setTheme,
    setFont,
    setReadingMode,
    fetchSettings,
    syncToBackend,
  }
})
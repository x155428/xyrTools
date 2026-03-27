const THEME_KEY = 'app-theme'

export function setTheme(theme) {
     document.documentElement.setAttribute('data-theme', theme)
     localStorage.setItem(THEME_KEY, theme)
}

export function toggleTheme() {
     const current =
          document.documentElement.getAttribute('data-theme') || 'light'
     setTheme(current === 'light' ? 'dark' : 'light')
}

export function initTheme() {
     const saved = localStorage.getItem(THEME_KEY)
     setTheme(saved || 'light')
}

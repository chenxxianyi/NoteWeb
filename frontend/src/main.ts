import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/design-tokens.css'

console.log('[main.ts] NoteWeb 前端正在启动...')
const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
console.log('[main.ts] Vue 应用挂载完成')

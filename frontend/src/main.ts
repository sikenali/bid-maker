import { createApp } from 'vue';
import { createPinia } from 'pinia';
import * as RemixIcons from '@remixicon/vue';
import App from './App.vue';
import router from './router';
import './assets/global.css';
import { BRAND_TITLE } from './constants/brand';

document.title = BRAND_TITLE;

const app = createApp(App);
app.use(createPinia());
app.use(router);

for (const [name, component] of Object.entries(RemixIcons)) {
  app.component(name, component);
}

app.mount('#app');

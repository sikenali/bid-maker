import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import App from './App.vue';
import router from './router';
import 'remixicon/fonts/remixicon.css';
import './assets/global.css';
import { BRAND_TITLE } from './constants/brand';

document.title = BRAND_TITLE;

const app = createApp(App);
app.use(createPinia());
app.use(ElementPlus);
app.use(router);
app.mount('#app');

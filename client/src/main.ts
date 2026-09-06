import { createApp } from "vue";
import App from "./App.vue";

import "./assets/index.css";
import { router } from "./lib/router.ts";

createApp(App).use(router).mount("#app");

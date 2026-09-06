import LandingPage from "@/components/view/LandingPage.vue";
import MissionDetailView from "@/components/view/MissionDetailView.vue";
import MissionsView from "@/components/view/MissionsView.vue";
import NewMissionWizard from "@/components/view/NewMissionWizard.vue";
import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  { path: "/", component: LandingPage },
  { path: "/missions", component: MissionsView },
  { path: "/missions/new", component: NewMissionWizard },
  {path: "/missions/:id", component: MissionDetailView}
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

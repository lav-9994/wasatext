import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import {getUserId} from '../services/session.js'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: HomeView},
		{path: '/login', component: LoginView},
		{path: '/conversations/:id', component: HomeView},
		{path: '/:pathMatch(.*)*', redirect: '/'},
	]
})

router.beforeEach((to) => {
	const loggedIn = getUserId() !== null;
	if (!loggedIn && to.path !== '/login') {
		return '/login';
	}
	if (loggedIn && to.path === '/login') {
		return '/';
	}
	return true;
})

export default router

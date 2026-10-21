import axios from "axios";
import router from "../router";
import {getUserId, clearUserId} from "./session.js";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

instance.interceptors.request.use((config) => {
	const userId = getUserId();
	if (userId !== null) {
		config.headers.Authorization = `Bearer ${userId}`;
	}
	return config;
});

instance.interceptors.response.use(
	(response) => response,
	(error) => {
		if (error.response && error.response.status === 401) {
			clearUserId();
			router.replace("/login");
		}
		return Promise.reject(error);
	}
);

export default instance;

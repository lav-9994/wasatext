const userIdKey = "wasatext-user-id";

export function saveUserId(id) {
	sessionStorage.setItem(userIdKey, String(id));
}

export function getUserId() {
	const value = sessionStorage.getItem(userIdKey);
	return value === null ? null : Number(value);
}

export function clearUserId() {
	sessionStorage.removeItem(userIdKey);
}

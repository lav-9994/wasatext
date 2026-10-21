function twoDigits(value) {
	return String(value).padStart(2, "0");
}

export function formatDateTime(timestamp) {
	const date = new Date(timestamp);
	if (isNaN(date.getTime())) {
		return "";
	}
	const day = `${twoDigits(date.getDate())}/${twoDigits(date.getMonth() + 1)}/${twoDigits(date.getFullYear() % 100)}`;
	const time = `${twoDigits(date.getHours())}:${twoDigits(date.getMinutes())}`;
	return `${day} ${time}`;
}

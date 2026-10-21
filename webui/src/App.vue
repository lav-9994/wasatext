<script>
export default {
	data() {
		return {
			theme: localStorage.getItem('theme') ||
				(window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'),
		}
	},
	watch: {
		theme: {
			immediate: true,
			handler(value) {
				document.documentElement.setAttribute('data-bs-theme', value)
				localStorage.setItem('theme', value)
			},
		},
	},
	methods: {
		toggleTheme() {
			this.theme = this.theme === 'dark' ? 'light' : 'dark'
		},
	},
}
</script>

<template>
	<RouterView />
</template>

<style>
</style>

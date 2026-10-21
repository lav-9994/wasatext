<script>
import UserSearchList from './UserSearchList.vue'

export default {
	components: {UserSearchList},
	emits: ['back', 'opened'],
	data() {
		return {
			errormsg: null,
			opening: false,
		}
	},
	methods: {
		async open(user) {
			if (this.opening) {
				return;
			}
			this.opening = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.post("/conversations", {userId: user.id});
				this.$emit('opened');
				this.$router.push(`/conversations/${response.data.id}`);
			} catch (e) {
				this.errormsg = e.response ? `Impossibile aprire la chat (errore ${e.response.status}).` : "Impossibile contattare il server.";
			}
			this.opening = false;
		},
	},
}
</script>

<template>
	<div class="d-flex flex-column h-100">
		<header class="d-flex align-items-center gap-2 p-2 border-bottom">
			<button type="button" class="btn btn-sm btn-outline-secondary" title="Indietro" @click="$emit('back')">
				<i class="bi bi-arrow-left"></i>
			</button>
			<span class="fw-semibold">Nuova chat</span>
		</header>
		<ErrorMsg v-if="errormsg" :msg="errormsg" class="m-2 py-2"></ErrorMsg>
		<UserSearchList class="panel-fill flex-grow-1" @pick="open" />
	</div>
</template>

<style>
.panel-fill {
	min-height: 0;
}
</style>

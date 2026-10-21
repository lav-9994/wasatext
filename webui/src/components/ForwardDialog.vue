<script>
import UserAvatar from './UserAvatar.vue'

export default {
	components: {UserAvatar},
	props: {
		conversationId: {type: Number, required: true},
		message: {type: Object, required: true},
	},
	emits: ['close', 'forwarded'],
	data() {
		return {
			search: '',
			conversations: [],
			users: [],
			loading: true,
			sending: false,
			errormsg: null,
		}
	},
	computed: {
		filteredConversations() {
			const search = this.search.trim().toLowerCase();
			return this.conversations.filter((c) => c.displayName.toLowerCase().includes(search));
		},
		otherUsers() {
			const directNames = this.conversations.filter((c) => c.type === 'direct').map((c) => c.displayName);
			return this.users.filter((u) => !directNames.includes(u.username));
		},
	},
	watch: {
		search() {
			this.loadUsers();
		},
	},
	methods: {
		async load() {
			try {
				const response = await this.$axios.get("/conversations");
				this.conversations = response.data;
				await this.loadUsers();
			} catch (e) {
				this.errormsg = this.describe(e);
			}
			this.loading = false;
		},
		async loadUsers() {
			try {
				const response = await this.$axios.get("/users", {params: {username: this.search.trim()}});
				this.users = response.data;
			} catch (e) {
				this.errormsg = this.describe(e);
			}
		},
		async forwardTo(targetId) {
			this.sending = true;
			this.errormsg = null;
			try {
				await this.$axios.post(`/conversations/${this.conversationId}/messages/${this.message.id}/forwards`,
					{conversationId: targetId}, {timeout: 60000});
				this.$emit('forwarded', targetId);
			} catch (e) {
				this.errormsg = this.describe(e);
			}
			this.sending = false;
		},
		async forwardToUser(user) {
			this.sending = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.post("/conversations", {userId: user.id});
				await this.forwardTo(response.data.id);
			} catch (e) {
				this.errormsg = this.describe(e);
				this.sending = false;
			}
		},
		describe(e) {
			if (!e.response) {
				return "Impossibile contattare il server.";
			}
			return `Inoltro non riuscito (errore ${e.response.status}).`;
		},
		onKeydown(event) {
			if (event.key === 'Escape') {
				this.$emit('close');
			}
		},
	},
	mounted() {
		window.addEventListener('keydown', this.onKeydown);
		this.load();
	},
	beforeUnmount() {
		window.removeEventListener('keydown', this.onKeydown);
	},
}
</script>

<template>
	<div class="dialog-backdrop d-flex align-items-center justify-content-center p-3" @click.self="$emit('close')">
		<div class="dialog-card card shadow">
			<div class="card-header d-flex align-items-center">
				<span class="fw-semibold">Inoltra a…</span>
				<button type="button" class="btn-close ms-auto" title="Chiudi" @click="$emit('close')"></button>
			</div>
			<div class="p-2 border-bottom">
				<input v-model="search" type="search" class="form-control form-control-sm" maxlength="16"
					placeholder="Cerca chat o utenti">
			</div>
			<div class="dialog-body overflow-auto">
				<ErrorMsg v-if="errormsg" :msg="errormsg" class="m-2"></ErrorMsg>
				<LoadingSpinner :loading="loading">
					<div v-if="filteredConversations.length > 0" class="small text-body-secondary px-3 pt-2">Chat</div>
					<button v-for="c in filteredConversations" :key="'c' + c.id" type="button"
						class="dialog-item d-flex align-items-center gap-2 w-100 border-0 bg-transparent px-3 py-2 text-start"
						:disabled="sending" @click="forwardTo(c.id)">
						<UserAvatar :picture="c.displayPicture" :name="c.displayName" :size="36" />
						<span class="text-truncate">{{ c.displayName }}</span>
						<i v-if="c.type === 'group'" class="bi bi-people text-body-secondary ms-auto"></i>
					</button>

					<div v-if="otherUsers.length > 0" class="small text-body-secondary px-3 pt-2">Altri utenti</div>
					<button v-for="u in otherUsers" :key="'u' + u.id" type="button"
						class="dialog-item d-flex align-items-center gap-2 w-100 border-0 bg-transparent px-3 py-2 text-start"
						:disabled="sending" @click="forwardToUser(u)">
						<UserAvatar :picture="u.profilePicture" :name="u.username" :size="36" />
						<span class="text-truncate">{{ u.username }}</span>
					</button>

					<p v-if="filteredConversations.length === 0 && otherUsers.length === 0"
						class="text-center text-body-secondary p-3 mb-0">
						Nessun risultato.
					</p>
				</LoadingSpinner>
			</div>
		</div>
	</div>
</template>

<style>
.dialog-backdrop {
	position: fixed;
	inset: 0;
	z-index: 1050;
	background-color: rgba(0, 0, 0, 0.5);
}

.dialog-card {
	width: 100%;
	max-width: 420px;
	max-height: 80vh;
}

.dialog-body {
	min-height: 0;
}

.dialog-item:hover:not(:disabled) {
	background-color: var(--bs-tertiary-bg) !important;
}
</style>

<script>
import UserAvatar from './UserAvatar.vue'

const refreshInterval = 2500;

export default {
	components: {UserAvatar},
	props: {
		userId: {type: Number, required: true},
	},
	emits: ['close'],
	data() {
		return {
			user: null,
			groups: [],
			notFound: false,
			errormsg: null,
			opening: false,
			active: false,
			timer: null,
		}
	},
	methods: {
		async load() {
			try {
				const [profile, common] = await Promise.all([
					this.$axios.get(`/users/${this.userId}`),
					this.$axios.get(`/users/${this.userId}/common-groups`),
				]);
				this.user = profile.data;
				this.groups = common.data;
				this.errormsg = null;
			} catch (e) {
				if (e.response && e.response.status === 404) {
					this.notFound = true;
					this.active = false;
				} else if (!e.response) {
					this.errormsg = "Impossibile contattare il server.";
				} else if (e.response.status !== 401) {
					this.errormsg = `Errore durante l'aggiornamento (errore ${e.response.status}).`;
				}
			}
		},
		async poll() {
			await this.load();
			if (this.active) {
				this.timer = setTimeout(this.poll, refreshInterval);
			}
		},
		async write() {
			this.opening = true;
			try {
				const response = await this.$axios.post("/conversations", {userId: this.userId});
				this.$router.push(`/conversations/${response.data.id}`);
				this.$emit('close');
			} catch (e) {
				this.errormsg = e.response ? `Impossibile aprire la chat (errore ${e.response.status}).` : "Impossibile contattare il server.";
			}
			this.opening = false;
		},
	},
	mounted() {
		this.active = true;
		this.poll();
	},
	beforeUnmount() {
		this.active = false;
		clearTimeout(this.timer);
	},
}
</script>

<template>
	<div class="d-flex flex-column h-100">
		<header class="d-flex align-items-center gap-2 p-2 border-bottom">
			<span class="fw-semibold">Profilo</span>
			<button type="button" class="btn-close ms-auto" title="Chiudi" @click="$emit('close')"></button>
		</header>

		<div class="flex-grow-1 overflow-auto p-3">
			<ErrorMsg v-if="errormsg" :msg="errormsg" class="py-2"></ErrorMsg>
			<p v-if="notFound" class="text-body-secondary text-center">Utente non trovato.</p>

			<LoadingSpinner :loading="user === null && !notFound && !errormsg">
				<template v-if="user">
					<div class="d-flex flex-column align-items-center mb-3">
						<UserAvatar :picture="user.profilePicture" :name="user.username" :size="112" />
						<div class="fs-5 fw-semibold mt-2 text-break text-center">{{ user.username }}</div>
					</div>

					<div class="mb-3">
						<div class="small text-body-secondary mb-1">Bio</div>
						<div v-if="user.bio" class="text-break bio-text">{{ user.bio }}</div>
						<div v-else class="text-body-secondary fst-italic">Nessuna bio</div>
					</div>

					<button type="button" class="btn btn-primary w-100 mb-3" :disabled="opening" @click="write">
						<i class="bi bi-chat-text me-1"></i> Scrivi
					</button>

					<div class="small text-body-secondary mb-1">Gruppi in comune</div>
					<p v-if="groups.length === 0" class="text-body-secondary fst-italic">Nessun gruppo in comune</p>
					<RouterLink v-for="g in groups" :key="g.id" :to="`/conversations/${g.id}`"
						class="d-flex align-items-center gap-2 py-1 text-decoration-none text-body">
						<UserAvatar :picture="g.displayPicture" :name="g.displayName" :size="32" />
						<span class="text-truncate">{{ g.displayName }}</span>
					</RouterLink>
				</template>
			</LoadingSpinner>
		</div>
	</div>
</template>

<style>
</style>

<script>
import UserAvatar from './UserAvatar.vue'

const refreshInterval = 2500;

export default {
	components: {UserAvatar},
	props: {
		selectable: {type: Boolean, default: false},
		selectedIds: {type: Array, default: () => []},
		excludeIds: {type: Array, default: () => []},
	},
	emits: ['pick'],
	data() {
		return {
			search: '',
			users: [],
			loaded: false,
			errormsg: null,
			active: false,
			timer: null,
		}
	},
	computed: {
		visibleUsers() {
			return this.users.filter((u) => !this.excludeIds.includes(u.id));
		},
	},
	watch: {
		search() {
			this.load();
		},
	},
	methods: {
		async load() {
			try {
				const response = await this.$axios.get("/users", {params: {username: this.search.trim()}});
				this.users = response.data;
				this.errormsg = null;
			} catch (e) {
				if (!e.response) {
					this.errormsg = "Impossibile contattare il server.";
				} else if (e.response.status !== 401) {
					this.errormsg = `Ricerca non riuscita (errore ${e.response.status}).`;
				}
			}
			this.loaded = true;
		},
		async poll() {
			await this.load();
			if (this.active) {
				this.timer = setTimeout(this.poll, refreshInterval);
			}
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
	<div class="d-flex flex-column">
		<div class="p-2 border-bottom">
			<input v-model="search" type="search" class="form-control form-control-sm" maxlength="16"
				placeholder="Cerca utenti">
		</div>
		<ErrorMsg v-if="errormsg" :msg="errormsg" class="m-2 py-2"></ErrorMsg>
		<div class="user-search-items flex-grow-1 overflow-auto">
			<p v-if="loaded && visibleUsers.length === 0" class="text-center text-body-secondary p-3 mb-0">
				Nessun utente trovato.
			</p>
			<button v-for="u in visibleUsers" :key="u.id" type="button"
				class="user-search-item d-flex align-items-center gap-2 w-100 border-0 bg-transparent px-3 py-2 text-start text-body"
				@click="$emit('pick', u)">
				<UserAvatar :picture="u.profilePicture" :name="u.username" :size="40" />
				<div class="user-search-body flex-grow-1">
					<div class="fw-semibold text-truncate">{{ u.username }}</div>
					<div v-if="u.bio" class="small text-body-secondary text-truncate">{{ u.bio }}</div>
				</div>
				<i v-if="selectable" class="fs-5"
					:class="selectedIds.includes(u.id) ? 'bi bi-check-circle-fill text-primary' : 'bi bi-circle text-body-secondary'"></i>
			</button>
		</div>
	</div>
</template>

<style>
.user-search-items {
	min-height: 0;
}

.user-search-item:hover {
	background-color: var(--bs-tertiary-bg) !important;
}

.user-search-body {
	min-width: 0;
}
</style>
